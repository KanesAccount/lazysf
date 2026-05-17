package gui

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
    "unicode/utf8"

    "github.com/jesseduffield/gocui"
    "lazysf/internal/cmdlog"
    "lazysf/internal/sf"
    "lazysf/internal/utils"
)

type Gui struct {
    g            *gocui.Gui
    focusedIndex int
    hasFocused   bool

    // orgs panel state
    orgs              []OrgItem
    orgSel            int
    orgTop            int
    orgsLoading       bool
    orgsSpinnerIndex  int
    orgsLoadingMsg    string

    // logs panel state
    logs    []LogItem
    logsSel int
    logsTop int
    currentLogID string

    // main panel scroll state
    mainTop        int
    mainTotalLines int
    mainLoading    bool
    spinnerIndex   int

    // vim-style chord state (for gg)
    lastGTime  time.Time
    lastGPanel int

    // panel 2 tabs: 0=Logs,1=Filters
    panel2Mode int
    // filters state
    filters    []string
    filtersSel int
    filtersTop int
    // current main text for search
    mainText    string
    searchTerm  string
    searchHits  []int
    searchIndex int
    searchPrevHits []int

    // input modal state
    inputReturnFocus int

    // command log subscription
    // (we don't need to store the channel; we'll re-render on each event)

    // cached current alias (to avoid repeated sf calls)
    currentAlias string
    currentUser  string

    // traces panel state
    traces    []TraceItem
    tracesSel int
    tracesTop int
    users     []sf.User

    // toast state
    toastTimer       *time.Timer
    // logs panel status spinner
    logsStatusLoading bool
    logsStatusIndex   int
    logsStatusMsg     string

    // command log expanded flag
    cmdlogExpanded bool
    cmdlogTop      int // -1 means follow bottom

    // main panel expanded flag
    mainExpanded bool

    // traces and logs expanded flags
    tracesExpanded bool
    logsExpanded   bool
}

type TraceItem struct {
    Id        string
    UserId    string
    UserName  string
    Username  string
    Start     string
    Expire    string
    LogType   string
}

type OrgItem struct {
    Alias           string
    Username        string
    ConnectedStatus string
    DefaultMarker   string
}

// FetchOrgsPublic exposes org retrieval to other packages.
func (gui *Gui) FetchOrgsPublic() { gui.fetchOrgs() }
func (gui *Gui) FetchLogsPublic() { gui.fetchLogs() }
func (gui *Gui) SetCurrentAlias(alias string) { gui.currentAlias = alias }
func (gui *Gui) FetchTracesPublic() { gui.fetchTraces() }
func (gui *Gui) SetCurrentUser(user string) { gui.currentUser = user }
func (gui *Gui) RefreshStatusPanel() { gui.refreshStatusPanel() }

func (gui *Gui) refreshStatusPanel() {
    alias := gui.currentAlias
    username := gui.currentUser
    if alias == "" || username == "" {
        if a, u, err := sf.GetCurrentOrgDisplay(); err == nil {
            alias, username = a, u
            gui.currentAlias, gui.currentUser = a, u
        }
    }
    // detect NVIM server from env
    server := os.Getenv("LAZYSF_NVIM_SERVER")
    if server == "" { server = os.Getenv("NVIM_LISTEN_ADDRESS") }
    lines := []string{}
    if alias != "" { lines = append(lines, alias+" → "+username) }
    if server != "" { lines = append(lines, "NVIM: "+shortServer(server)) }
    if len(lines) == 0 { lines = append(lines, "No target org set") }
    gui.SetStatusText(lines...)
}

func New() (*Gui, error) {
    g, err := gocui.NewGui(gocui.NewGuiOpts{
        OutputMode: gocui.OutputNormal,
    })
    if err != nil {
        return nil, err
    }
    // enable visible frame highlighting for the focused view
    g.Highlight = true
    g.SelFrameColor = gocui.ColorGreen
    g.SelFgColor = gocui.ColorGreen
    gui := &Gui{g: g, focusedIndex: 0, cmdlogTop: -1}
    utils.Debugf("init gui: set manager func")
    g.SetManagerFunc(gui.layout)
    if err := gui.setKeybindings(); err != nil {
        g.Close()
        return nil, err
    }
    return gui, nil
}

func (gui *Gui) Run() error {
    defer gui.g.Close()
    utils.Debugf("starting main loop")
    go gui.watchCmdLog()
    if err := gui.g.MainLoop(); err != nil {
        utils.Debugf("main loop returned error: %v", err)
        return err
    }
    utils.Debugf("main loop exited cleanly")
    return nil
}

// orgs panel helpers
func (gui *Gui) renderOrgs() {
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View(panelName(3))
        if err != nil { return nil }
        v.Clear()
        v.Highlight = true
        v.SelFgColor = gocui.ColorGreen
        for i, o := range gui.orgs {
            selected := i == gui.orgSel
            isCurrent := o.Alias == gui.currentAlias
            prefix := "  "
            if isCurrent && selected {
                prefix = "*> "
            } else if isCurrent {
                prefix = "*  "
            } else if selected {
                prefix = "> "
            }
            line := fmt.Sprintf("%s%-24s  %-28s  %s %s", prefix, o.Alias, o.Username, o.ConnectedStatus, o.DefaultMarker)
            fmt.Fprintln(v, line)
        }
        // compute viewport height and adjust origin
        _, h := v.Size()
        if h <= 0 { h = 1 }
        // clamp orgTop
        if gui.orgTop < 0 { gui.orgTop = 0 }
        if gui.orgSel < gui.orgTop { gui.orgTop = gui.orgSel }
        if gui.orgSel >= gui.orgTop+h { gui.orgTop = gui.orgSel - h + 1 }
        v.SetOrigin(0, gui.orgTop)
        // set cursor within viewport
        cy := gui.orgSel - gui.orgTop
        if cy < 0 { cy = 0 }
        v.SetCursor(0, cy)
        // show loading message at bottom right if loading (after scrolling setup)
        if gui.orgsLoading && gui.orgsLoadingMsg != "" {
            w, h := v.Size()
            if h > 0 && w > 0 {
                msgLen := len(gui.orgsLoadingMsg)
                if msgLen < w {
                    // Position message in bottom right of visible area
                    x := w - msgLen - 1
                    y := h - 1
                    // Save current cursor position
                    oldX, oldY := v.Cursor()
                    // Write message
                    v.SetCursor(x, y)
                    fmt.Fprint(v, gui.orgsLoadingMsg)
                    // Restore cursor position
                    v.SetCursor(oldX, oldY)
                }
            }
        }
        return nil
    })
}

func (gui *Gui) fetchOrgs() {
    gui.fetchOrgsWithSpinner(true)
}

func (gui *Gui) fetchOrgsWithSpinner(showSpinner bool) {
    if showSpinner {
        gui.startOrgsSpinner("Loading organizations…")
    }
    go func() {
        list, err := sf.ListAuthorizedOrgs()
        if err != nil {
            utils.Debugf("fetchOrgs error: %v", err)
            if showSpinner {
                gui.stopOrgsSpinner()
            }
            return
        }
        items := make([]OrgItem, 0, len(list))
        for _, o := range list {
            items = append(items, OrgItem{
                Alias: o.Alias, Username: o.Username, ConnectedStatus: o.ConnectedStatus, DefaultMarker: o.DefaultMarker,
            })
        }
        gui.orgs = items
        if gui.orgSel >= len(gui.orgs) { gui.orgSel = 0 }
        if showSpinner {
            gui.stopOrgsSpinner()
        }
        gui.renderOrgs()
    }()
}

// traces panel
func (gui *Gui) fetchTraces() {
    go func() {
        users, err := sf.ListActiveUsers(); if err != nil { utils.Debugf("users error: %v", err); return }
        flags, err := sf.ListTraceFlags(); if err != nil { utils.Debugf("trace flags error: %v", err); return }
        // build user lookup
        uById := map[string]sf.User{}
        for _, u := range users { uById[u.Id] = u }
        items := make([]TraceItem, 0, len(flags))
        for _, f := range flags {
            u := uById[f.TracedEntityId]
            items = append(items, TraceItem{Id: f.Id, UserId: f.TracedEntityId, UserName: u.Name, Username: u.Username, Start: f.StartDate, Expire: f.ExpirationDate, LogType: f.LogType})
        }
        gui.users = users
        gui.traces = items
        if gui.tracesSel >= len(gui.traces) { gui.tracesSel = 0 }
        gui.renderTraces()
    }()
}

func (gui *Gui) renderTraces() {
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View(panelName(1)); if err != nil { return nil }
        v.Clear(); v.Highlight = true; v.SelFgColor = gocui.ColorGreen; v.Wrap = false
        // dynamic column widths to ensure Start/End and Left are always visible
        w, _ := v.Size()
        if w < 40 { w = 40 }
        padLeft := 2 // align with content prefix ("  ", "> ")
        ww := w - padLeft
        if ww < 20 { ww = 20 }
        leftW := 6   // e.g. 1h30m
        timesW := 20 // e.g. 17 14:01 -> 17 15:01 (8+4+8)
        spacing := 2 // spaces between columns
        fixed := leftW + timesW + spacing*3
        usable := ww - fixed
        if usable < 4 { usable = 4 }
        // Prioritize compact User & Username so times/left are visible
        // Make Username shorter by default, with a lower cap.
        minNameW := 10
        minUserW := 10
        maxUserW := 20
        // Default split: 45% name, 55% username (will cap username later)
        nameW := usable * 45 / 100
        userW := usable - nameW
        if nameW < minNameW {
            nameW = minNameW
            userW = usable - nameW
        }
        if userW < minUserW {
            userW = minUserW
            // re-adjust name to use remaining
            nameW = usable - userW
            if nameW < 4 { nameW = 4 }
        }
        // cap overly large username column first
        if userW > maxUserW {
            userW = maxUserW
            nameW = usable - userW
            if nameW < minNameW { nameW = minNameW }
        }
        // cap overly large name column (keeps things tidy)
        if nameW > 26 {
            nameW = 26
            userW = usable - nameW
            if userW < minUserW { userW = minUserW }
        }

        // header
        header := fmt.Sprintf("%-*s%*s%-*s%*s%-*s%*s%-*s",
            nameW, "User", spacing, "", userW, "Username", spacing, "",
            timesW, "Start -> End", spacing, "", leftW, "Left")
        fmt.Fprintln(v, "  "+header)
        for i, t := range gui.traces {
            prefix := "  "; if i == gui.tracesSel { prefix = "> " }
            left := ""
            if exp, err := parseSFTime(t.Expire); err == nil { left = fmtDurationShort(exp.UTC().Sub(time.Now().UTC())) }
            // truncate values to fit
            uname := truncate(t.UserName, nameW)
            ulogin := truncate(t.Username, userW)
            start := formatTraceTime(t.Start)
            end := formatTraceTime(t.Expire)
            times := start + " -> " + end
            line := fmt.Sprintf("%s%-*s%*s%-*s%*s%-*s%*s%-*s",
                prefix,
                nameW, uname, spacing, "",
                userW, ulogin, spacing, "",
                timesW, times, spacing, "",
                leftW, left,
            )
            fmt.Fprintln(v, line)
        }
        // scroll management
        _, h := v.Size(); if h<=0{h=1}
        if gui.tracesTop < 0 { gui.tracesTop = 0 }
        if gui.tracesSel < gui.tracesTop { gui.tracesTop = gui.tracesSel }
        // account for header line
        contentH := h - 1; if contentH <= 0 { contentH = 1 }
        if gui.tracesSel >= gui.tracesTop+contentH { gui.tracesTop = gui.tracesSel - contentH + 1 }
        v.SetOrigin(0, gui.tracesTop)
        cy := gui.tracesSel - gui.tracesTop + 1; if cy < 1 { cy = 1 }
        v.SetCursor(0, cy)
        return nil
    })
}

func shortTime(s string) string {
    if len(s) >= 16 { return s[:16] } // rough trim
    return s
}

func shortServer(s string) string {
    if s == "" { return "" }
    // prefer last path element
    base := filepath.Base(s)
    if base != "." && base != "/" { return base }
    return s
}

func truncate(s string, width int) string {
    if width <= 0 { return "" }
    if len(s) <= width { return s }
    if width <= 1 { return s[:width] }
    // add ellipsis if room
    if width >= 3 { return s[:width-3] + "..." }
    return s[:width]
}

// formatTraceTime converts an SF time string to "DD HH:MM" for compact display
func formatTraceTime(s string) string {
    if t, err := parseSFTime(s); err == nil {
        return t.Format("02 15:04")
    }
    // fallback: crude slice
    if len(s) >= 16 { return s[8:10] + " " + s[11:16] }
    return s
}

func (gui *Gui) tracesUp(g *gocui.Gui, v *gocui.View) error { if gui.tracesSel>0 { gui.tracesSel-- }; gui.renderTraces(); return nil }
func (gui *Gui) tracesDown(g *gocui.Gui, v *gocui.View) error { if gui.tracesSel < len(gui.traces)-1 { gui.tracesSel++ }; gui.renderTraces(); return nil }
func (gui *Gui) tracesRefresh(g *gocui.Gui, v *gocui.View) error { gui.fetchTraces(); return nil }

func (gui *Gui) tracesAdd(g *gocui.Gui, v *gocui.View) error {
    // prompt for user filter
    gui.openInput("Add trace for user (name/username)", "", func(text string){
        if text == "" { return }
        // find best match
        var chosen *sf.User
        tl := strings.ToLower(text)
        for i := range gui.users {
            u := &gui.users[i]
            if strings.Contains(strings.ToLower(u.Name), tl) || strings.Contains(strings.ToLower(u.Username), tl) {
                chosen = u; break
            }
        }
        if chosen == nil { gui.SetStatusText("No matching user for:", text); return }
        // run update asynchronously with a spinner in panel 2 (bottom-right)
        gui.startLogsStatusSpinner("Updating trace…")
        go func(u sf.User){
            dl, err := sf.EnsureDebugLevelId(); if err != nil { utils.Debugf("debug level err: %v", err); gui.stopLogsStatusSpinner(); gui.SetStatusText("DebugLevel error"); return }
            if err := sf.UpsertTraceForUser(u.Id, dl, time.Hour); err != nil {
                utils.Debugf("upsert trace err: %v", err)
                gui.stopLogsStatusSpinner()
                gui.openConfirm("Clear Apex Logs?", "Trace update failed. Clear all Apex logs and retry?", func(){
                    go func(){
                        gui.showToast("Deleting logs…")
                        n, err := sf.DeleteAllApexLogsFast()
                        if err != nil { gui.SetStatusText("Failed to delete logs") ; return }
                        gui.SetStatusText("Deleted ", strconv.Itoa(n), " logs. Retrying…")
                        if err2 := sf.UpsertTraceForUser(u.Id, dl, time.Hour); err2 == nil {
                            gui.showToast("Trace enabled for " + u.Username)
                        }
                        gui.fetchTraces(); gui.fetchLogs()
                    }()
                }, nil)
                return
            }
            gui.stopLogsStatusSpinner()
            gui.showToast("Trace enabled for " + u.Username)
            gui.fetchTraces()
        }(*chosen)
    })
    return nil
}

func (gui *Gui) tracesDelete(g *gocui.Gui, v *gocui.View) error {
    if len(gui.traces) == 0 { return nil }
    id := gui.traces[gui.tracesSel].Id
    if id == "" { return nil }
    if err := sf.DeleteTrace(id); err != nil { utils.Debugf("delete trace err: %v", err); return nil }
    gui.showToast("Trace deleted")
    gui.fetchTraces(); return nil
}

// Parse user input like "30m", "2h", "90m" into duration
func parseDurationInput(s string) (time.Duration, error) {
    s = strings.TrimSpace(strings.ToLower(s))
    if s == "" { return 0, fmt.Errorf("empty duration") }
    // if suffix provided, let time.ParseDuration handle (supports h, m, s)
    if strings.HasSuffix(s, "h") || strings.HasSuffix(s, "m") || strings.HasSuffix(s, "s") {
        return time.ParseDuration(s)
    }
    // no suffix: treat as minutes
    if n, err := strconv.Atoi(s); err == nil {
        return time.Duration(n) * time.Minute, nil
    }
    return 0, fmt.Errorf("invalid duration: %s", s)
}

// Parse SF time strings with or without milliseconds
func parseSFTime(s string) (time.Time, error) {
    layouts := []string{
        "2006-01-02T15:04:05.000-0700",
        "2006-01-02T15:04:05-0700",
        time.RFC3339,
    }
    var lastErr error
    for _, layout := range layouts {
        if t, err := time.Parse(layout, s); err == nil {
            return t, nil
        } else { lastErr = err }
    }
    return time.Time{}, lastErr
}

func fmtDurationShort(d time.Duration) string {
    if d < 0 { d = 0 }
    h := d / time.Hour
    m := (d % time.Hour) / time.Minute
    if h > 0 {
        if m > 0 { return fmt.Sprintf("%dh%dm", h, m) }
        return fmt.Sprintf("%dh", h)
    }
    return fmt.Sprintf("%dm", m)
}

func (gui *Gui) tracesEditDuration(g *gocui.Gui, v *gocui.View) error {
    if len(gui.traces) == 0 { return nil }
    gui.openInput("New duration (e.g. 30m, 2h)", "1h", func(text string){
        dur, err := parseDurationInput(text)
        if err != nil { gui.SetStatusText("Invalid duration:", text); return }
        t := gui.traces[gui.tracesSel]
        gui.startLogsStatusSpinner("Updating trace…")
        go func(userId string){
            dl, err := sf.EnsureDebugLevelId(); if err != nil { utils.Debugf("debug level err: %v", err); gui.stopLogsStatusSpinner(); gui.SetStatusText("DebugLevel error"); return }
            if err := sf.UpsertTraceForUser(userId, dl, dur); err != nil {
                utils.Debugf("upsert trace err: %v", err)
                gui.stopLogsStatusSpinner()
                gui.openConfirm("Clear Apex Logs?", "Trace update failed. Clear all Apex logs and retry?", func(){
                    go func(){
                        gui.showToast("Deleting logs…")
                        n, err := sf.DeleteAllApexLogsFast()
                        if err != nil { gui.SetStatusText("Failed to delete logs") ; return }
                        gui.SetStatusText("Deleted ", strconv.Itoa(n), " logs. Retrying…")
                        if err2 := sf.UpsertTraceForUser(userId, dl, dur); err2 == nil {
                            gui.showToast("Trace duration updated")
                        }
                        gui.fetchTraces(); gui.fetchLogs()
                    }()
                }, nil)
                return
            }
            gui.stopLogsStatusSpinner()
            gui.showToast("Trace duration updated")
            gui.fetchTraces()
        }(t.UserId)
    })
    return nil
}

func (gui *Gui) orgsUp(g *gocui.Gui, v *gocui.View) error {
    if gui.orgSel > 0 { gui.orgSel-- }
    gui.renderOrgs()
    return nil
}
func (gui *Gui) orgsDown(g *gocui.Gui, v *gocui.View) error {
    if gui.orgSel < len(gui.orgs)-1 { gui.orgSel++ }
    gui.renderOrgs()
    return nil
}
func (gui *Gui) orgsHome(g *gocui.Gui, v *gocui.View) error { gui.orgSel = 0; gui.renderOrgs(); return nil }
func (gui *Gui) orgsEnd(g *gocui.Gui, v *gocui.View) error {
    if len(gui.orgs) > 0 { gui.orgSel = len(gui.orgs)-1 }
    gui.renderOrgs(); return nil
}
func (gui *Gui) orgsSwitch(g *gocui.Gui, v *gocui.View) error {
    if len(gui.orgs) == 0 { return nil }
    alias := gui.orgs[gui.orgSel].Alias
    gui.startOrgsSpinner("Switching org…")
    go func() {
        // switch target org
        if _, err := sf.Exec("sf", "config", "set", "target-org", alias); err != nil {
            utils.Debugf("switch org error: %v", err)
            gui.stopOrgsSpinner()
            return
        }
        // refresh status
        _, username, err := sf.GetCurrentOrgDisplay()
        if err == nil {
            gui.SetStatusText(alias+" → "+username)
        }
        gui.currentAlias = alias
        gui.stopOrgsSpinner()
        // refresh orgs to reflect markers
        gui.fetchOrgsWithSpinner(false)
    }()
    return nil
}
func (gui *Gui) orgsAuth(g *gocui.Gui, v *gocui.View) error {
    // open web auth
    go func(){ _, _ = sf.Exec("sf", "org", "login", "web") }()
    return nil
}

// logs panel helpers (Panel 2)
type LogItem struct {
    Id        string
    Status    string
    StartTime string
    Length    int
    Operation string
}

func (gui *Gui) renderLogs() {
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View(panelName(2))
        if err != nil { return nil }
        v.Clear()
        v.Highlight = true
        v.SelFgColor = gocui.ColorGreen
        // Title reflect tab mode
        if gui.panel2Mode == 0 {
            v.Title = "2. Logs / Filters [Logs]"
            for i, l := range gui.logs {
                selected := i == gui.logsSel
                isOpen := l.Id == gui.currentLogID
                prefix := "  "
                if isOpen && selected {
                    prefix = "*> "
                } else if isOpen {
                    prefix = "*  "
                } else if selected {
                    prefix = "> "
                }
                line := fmt.Sprintf("%s%s  %-7s  %-5dkB  %s  %s", prefix, l.Id, l.Status, l.Length/1024, l.StartTime, l.Operation)
                fmt.Fprintln(v, line)
            }
            // scrolling
            _, h := v.Size(); if h <= 0 { h = 1 }
            if gui.logsTop < 0 { gui.logsTop = 0 }
            if gui.logsSel < gui.logsTop { gui.logsTop = gui.logsSel }
            if gui.logsSel >= gui.logsTop+h { gui.logsTop = gui.logsSel - h + 1 }
            v.SetOrigin(0, gui.logsTop)
            cy := gui.logsSel - gui.logsTop; if cy < 0 { cy = 0 }
            v.SetCursor(0, cy)
        } else {
            v.Title = "2. Logs / Filters [Filters]"
            for i, f := range gui.filters {
                prefix := "  "; if i == gui.filtersSel { prefix = "> " }
                fmt.Fprintf(v, "%s%s\n", prefix, f)
            }
            // scrolling for filters list
            _, h := v.Size(); if h <= 0 { h = 1 }
            if gui.filtersTop < 0 { gui.filtersTop = 0 }
            if gui.filtersSel < gui.filtersTop { gui.filtersTop = gui.filtersSel }
            if gui.filtersSel >= gui.filtersTop+h { gui.filtersTop = gui.filtersSel - h + 1 }
            v.SetOrigin(0, gui.filtersTop)
            cy := gui.filtersSel - gui.filtersTop; if cy < 0 { cy = 0 }
            v.SetCursor(0, cy)
        }
        return nil
    })
}

func (gui *Gui) fetchLogs() {
    go func() {
        logs, err := sf.ListApexLogs()
        if err != nil {
            utils.Debugf("fetchLogs error: %v", err)
            return
        }
        items := make([]LogItem, 0, len(logs))
        for _, l := range logs {
            items = append(items, LogItem{Id: l.Id, Status: l.Status, StartTime: l.StartTime, Length: l.LogLength, Operation: l.Operation})
        }
        gui.logs = items
        if gui.logsSel >= len(gui.logs) { gui.logsSel = 0 }
        // reset scroll to top on refresh
        gui.logsTop = 0
        gui.renderLogs()
    }()
}

func (gui *Gui) logsUp(g *gocui.Gui, v *gocui.View) error {
    if gui.panel2Mode == 0 {
        // Logs mode
        if gui.logsSel > 0 { gui.logsSel-- }
    } else {
        // Filters mode
        if gui.filtersSel > 0 { gui.filtersSel-- }
    }
    gui.renderLogs(); return nil
}
func (gui *Gui) logsDown(g *gocui.Gui, v *gocui.View) error {
    if gui.panel2Mode == 0 {
        // Logs mode
        if gui.logsSel < len(gui.logs)-1 { gui.logsSel++ }
    } else {
        // Filters mode
        if gui.filtersSel < len(gui.filters)-1 { gui.filtersSel++ }
    }
    gui.renderLogs(); return nil
}
func (gui *Gui) logsPgUp(g *gocui.Gui, v *gocui.View) error { _, h := v.Size(); if h<=0{h=1}; gui.logsSel -= h; if gui.logsSel<0{gui.logsSel=0}; gui.renderLogs(); return nil }
func (gui *Gui) logsPgDn(g *gocui.Gui, v *gocui.View) error { _, h := v.Size(); if h<=0{h=1}; gui.logsSel += h; if gui.logsSel>len(gui.logs)-1{gui.logsSel=len(gui.logs)-1}; gui.renderLogs(); return nil }
func (gui *Gui) logsHome(g *gocui.Gui, v *gocui.View) error { gui.logsSel = 0; gui.renderLogs(); return nil }
func (gui *Gui) logsEnd(g *gocui.Gui, v *gocui.View) error { if len(gui.logs)>0 { gui.logsSel = len(gui.logs)-1 }; gui.renderLogs(); return nil }

func (gui *Gui) logsRefresh(g *gocui.Gui, v *gocui.View) error {
    // ensure we're on Logs tab when refreshing
    gui.panel2Mode = 0
    gui.fetchLogs()
    return nil
}
func (gui *Gui) logsOpenEditor(g *gocui.Gui, v *gocui.View) error { path, ok := gui.ensureLogOnDisk(); if !ok { return nil }; cmd := utils.OpenInEditor(path); _ = cmd.Start(); return nil }
func (gui *Gui) mainOpenEditorAt(g *gocui.Gui, v *gocui.View) error {
    path, ok := gui.ensureLogOnDisk(); if !ok { return nil }
    // prefer current search hit line; otherwise top
    line := 1
    if len(gui.searchHits) > 0 {
        idx := gui.searchIndex
        if idx < 0 || idx >= len(gui.searchHits) { idx = 0 }
        line = gui.searchHits[idx] + 1 // 1-based for editor
    } else if gui.mainTop > 0 {
        line = gui.mainTop + 1
    }
    cmd := utils.OpenInEditorAt(path, line)
    _ = cmd.Start()
    return nil
}
func (gui *Gui) mainSetNvimServer(g *gocui.Gui, v *gocui.View) error {
    init := os.Getenv("LAZYSF_NVIM_SERVER")
    gui.openInput("Set nvim server (v:servername)", init, func(text string){
        if text != "" { _ = os.Setenv("LAZYSF_NVIM_SERVER", text); gui.refreshStatusPanel() }
    })
    return nil
}
func (gui *Gui) logsEnter(g *gocui.Gui, v *gocui.View) error {
    if gui.panel2Mode != 0 { return nil }
    // load selected log on demand with a spinner
    if len(gui.logs) == 0 { return nil }
    sel := gui.logs[gui.logsSel]
    gui.startSpinner("Loading " + sel.Id + " …")
    go func(id string){
        // ensure on disk
        path, ok := gui.ensureLogOnDisk(); if !ok { gui.stopSpinner(); return }
        data, err := ioutil.ReadFile(path); if err != nil { gui.stopSpinner(); return }
        gui.stopSpinner()
        gui.currentLogID = id
        gui.renderMainText(string(data))
        gui.renderLogs()
    }(sel.Id)
    // do not auto-focus main panel; user can press Tab to switch
    return nil
}

func (gui *Gui) logsDeleteAll(g *gocui.Gui, v *gocui.View) error {
    gui.openConfirm("Delete all Apex logs?", "This will remove all Apex log records in the current org.", func(){
        go func(){
            gui.showToast("Deleting logs…")
            n, err := sf.DeleteAllApexLogsFast()
            if err != nil { gui.SetStatusText("Failed to delete logs") ; return }
            gui.SetStatusText("Deleted ", strconv.Itoa(n), " logs")
            gui.showToast("Deleted " + strconv.Itoa(n) + " logs")
            gui.fetchLogs()
        }()
    }, nil)
    return nil
}

func (gui *Gui) showSelectedLog() {
    path, ok := gui.ensureLogOnDisk(); if !ok { return }
    data, err := ioutil.ReadFile(path); if err != nil { utils.Debugf("read log error: %v", err); return }
    // update main panel content and reset scroll state
    text := string(data)
    gui.renderMainText(text)
}

// main panel scrolling helpers
func (gui *Gui) mainUp(g *gocui.Gui, v *gocui.View) error {
    if gui.mainTop > 0 { gui.mainTop-- }
    v.SetOrigin(0, gui.mainTop)
    return nil
}
func (gui *Gui) mainDown(g *gocui.Gui, v *gocui.View) error {
    _, h := v.Size(); if h <= 0 { h = 1 }
    maxTop := gui.mainTotalLines - h
    if maxTop < 0 { maxTop = 0 }
    if gui.mainTop < maxTop { gui.mainTop++ }
    v.SetOrigin(0, gui.mainTop)
    return nil
}
func (gui *Gui) mainPgUp(g *gocui.Gui, v *gocui.View) error {
    _, h := v.Size(); if h <= 0 { h = 1 }
    gui.mainTop -= h
    if gui.mainTop < 0 { gui.mainTop = 0 }
    v.SetOrigin(0, gui.mainTop)
    return nil
}
func (gui *Gui) mainPgDn(g *gocui.Gui, v *gocui.View) error {
    _, h := v.Size(); if h <= 0 { h = 1 }
    maxTop := gui.mainTotalLines - h
    if maxTop < 0 { maxTop = 0 }
    gui.mainTop += h
    if gui.mainTop > maxTop { gui.mainTop = maxTop }
    v.SetOrigin(0, gui.mainTop)
    return nil
}
// Vim-style quick jumps with Shift+U (up) / Shift+D (down)
func (gui *Gui) mainJumpUp(g *gocui.Gui, v *gocui.View) error {
    _, h := v.Size(); if h <= 0 { h = 1 }
    delta := h / 2
    if delta < 3 { delta = 3 }
    gui.mainTop -= delta
    if gui.mainTop < 0 { gui.mainTop = 0 }
    v.SetOrigin(0, gui.mainTop)
    return nil
}
func (gui *Gui) mainJumpDown(g *gocui.Gui, v *gocui.View) error {
    _, h := v.Size(); if h <= 0 { h = 1 }
    delta := h / 2
    if delta < 3 { delta = 3 }
    maxTop := gui.mainTotalLines - h
    if maxTop < 0 { maxTop = 0 }
    gui.mainTop += delta
    if gui.mainTop > maxTop { gui.mainTop = maxTop }
    v.SetOrigin(0, gui.mainTop)
    return nil
}
func (gui *Gui) mainHome(g *gocui.Gui, v *gocui.View) error { gui.mainTop = 0; v.SetOrigin(0, 0); return nil }
func (gui *Gui) mainEnd(g *gocui.Gui, v *gocui.View) error {
    _, h := v.Size(); if h <= 0 { h = 1 }
    maxTop := gui.mainTotalLines - h
    if maxTop < 0 { maxTop = 0 }
    gui.mainTop = maxTop
    v.SetOrigin(0, gui.mainTop)
    return nil
}

// vim-style 'gg' chord handlers
func (gui *Gui) chordG(panelIdx int, home func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
    return func(g *gocui.Gui, v *gocui.View) error {
        now := time.Now()
        if gui.lastGPanel == panelIdx && now.Sub(gui.lastGTime) < 700*time.Millisecond {
            gui.lastGTime = time.Time{}
            return home(g, v)
        }
        gui.lastGPanel = panelIdx
        gui.lastGTime = now
        return nil
    }
}
func (gui *Gui) orgsG(g *gocui.Gui, v *gocui.View) error { return gui.chordG(3, gui.orgsHome)(g, v) }
func (gui *Gui) logsG(g *gocui.Gui, v *gocui.View) error { return gui.chordG(2, gui.logsHome)(g, v) }
func (gui *Gui) mainG(g *gocui.Gui, v *gocui.View) error { return gui.chordG(4, gui.mainHome)(g, v) }

func (gui *Gui) ensureLogOnDisk() (string, bool) {
    if len(gui.logs) == 0 { return "", false }
    sel := gui.logs[gui.logsSel]
    alias := gui.currentAlias
    if alias == "" {
        if a, err := sf.GetTargetOrgAlias(); err == nil { alias = a; gui.currentAlias = a } else { return "", false }
    }
    root, logsDir, _, err := utils.OrgPaths(alias); if err != nil { return "", false }
    _ = utils.EnsureDirs(root, logsDir)
    path := filepath.Join(logsDir, sel.Id+".log")
    if _, err := ioutil.ReadFile(path); err == nil { return path, true }
    if err := sf.GetApexLogToDir(sel.Id, logsDir); err != nil { utils.Debugf("download log error: %v", err); return "", false }
    return path, true
}

// Panel 2 tab toggle and filters operations
func (gui *Gui) panel2Toggle(g *gocui.Gui, v *gocui.View) error {
    if gui.panel2Mode == 0 { gui.panel2Mode = 1 } else { gui.panel2Mode = 0 }
    gui.renderLogs()
    return nil
}

func (gui *Gui) filtersAdd(g *gocui.Gui, v *gocui.View) error {
    if gui.panel2Mode != 1 { return nil }
    gui.openInput("Add filter term", "", func(text string){ if text == "" { return }; gui.filters = append(gui.filters, text); gui.renderLogs(); gui.applyFiltersToMain() })
    return nil
}
func (gui *Gui) filtersEdit(g *gocui.Gui, v *gocui.View) error {
    if gui.panel2Mode != 1 || len(gui.filters) == 0 { return nil }
    idx := gui.filtersSel; initial := gui.filters[idx]
    gui.openInput("Edit filter term", initial, func(text string){ if text == "" { return }; gui.filters[idx] = text; gui.renderLogs(); gui.applyFiltersToMain() })
    return nil
}
func (gui *Gui) filtersDelete(g *gocui.Gui, v *gocui.View) error {
    if gui.panel2Mode != 1 || len(gui.filters) == 0 { return nil }
    i := gui.filtersSel
    gui.filters = append(gui.filters[:i], gui.filters[i+1:]...)
    if gui.filtersSel >= len(gui.filters) && gui.filtersSel > 0 { gui.filtersSel-- }
    gui.renderLogs(); gui.applyFiltersToMain(); return nil
}

func (gui *Gui) applyFiltersToMain() {
    // if no log selected or no filters, show full content
    if len(gui.filters) == 0 { gui.showSelectedLog(); return }
    // ensure log on disk
    path, ok := gui.ensureLogOnDisk(); if !ok { return }
    // build combined pattern
    pattern := strings.Join(gui.filters, "|")
    // run ripgrep
    res, err := sf.Exec("rg", "-n", "-C25", "-e", pattern, path)
    if err != nil && res.Stdout == "" { return }
    text := res.Stdout
    gui.renderMainText(text)
}

// main panel search
func (gui *Gui) mainSearchPrompt(g *gocui.Gui, v *gocui.View) error {
    gui.openInput("Search", gui.searchTerm, func(text string){ gui.searchTerm = text; gui.computeSearchHits(); gui.mainSearchJump(0) })
    return nil
}
func (gui *Gui) computeSearchHits() {
    gui.searchPrevHits = gui.searchHits
    gui.searchHits = gui.searchHits[:0]
    if gui.searchTerm == "" || gui.mainText == "" { return }
    // scan lines
    line := 0
    start := 0
    for i := 0; i <= len(gui.mainText); i++ {
        if i == len(gui.mainText) || gui.mainText[i] == '\n' {
            seg := gui.mainText[start:i]
            if strings.Contains(strings.ToLower(seg), strings.ToLower(gui.searchTerm)) {
                gui.searchHits = append(gui.searchHits, line)
            }
            line++
            start = i+1
        }
    }
    gui.searchIndex = 0
    gui.applySearchHighlights()
}
func (gui *Gui) mainSearchJump(delta int) {
    if len(gui.searchHits) == 0 { return }
    gui.searchIndex += delta
    if gui.searchIndex < 0 { gui.searchIndex = len(gui.searchHits)-1 }
    if gui.searchIndex >= len(gui.searchHits) { gui.searchIndex = 0 }
    target := gui.searchHits[gui.searchIndex]
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View(panelName(4)); if err != nil { return nil }
        _, h := v.Size(); if h<=0{h=1}
        top := target - h/2; if top < 0 { top = 0 }
        gui.mainTop = top
        v.SetOrigin(0, top)
        return nil
    })
}
func (gui *Gui) mainSearchNext(g *gocui.Gui, v *gocui.View) error { gui.mainSearchJump(+1); return nil }
func (gui *Gui) mainSearchPrev(g *gocui.Gui, v *gocui.View) error { gui.mainSearchJump(-1); return nil }

// helper to render main panel text and reset scroll, preserving search highlighting
func (gui *Gui) renderMainText(text string) {
    // compute total lines
    total := 0
    for i := 0; i < len(text); i++ { if text[i] == '\n' { total++ } }
    if len(text) > 0 && text[len(text)-1] != '\n' { total++ }
    gui.mainText = text
    gui.mainTotalLines = total
    gui.mainTop = 0
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View(panelName(4)); if err != nil { return nil }
        v.Wrap = true
        v.Clear(); fmt.Fprint(v, text)
        v.SetOrigin(0, 0)
        // apply search highlights if any
        gui.applySearchHighlightsToView(v)
        return nil
    })
}

func (gui *Gui) applySearchHighlights() {
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View(panelName(4)); if err != nil { return nil }
        gui.applySearchHighlightsToView(v)
        return nil
    })
}

func (gui *Gui) applySearchHighlightsToView(v *gocui.View) {
    // turn off previous highlights
    for _, ln := range gui.searchPrevHits {
        v.SetHighlight(ln, false)
    }
    // turn on current highlights
    for _, ln := range gui.searchHits {
        v.SetHighlight(ln, true)
    }
}

// spinner helpers for main panel
func (gui *Gui) startSpinner(msg string) {
    gui.mainLoading = true
    gui.spinnerIndex = 0
    frames := []rune{'⠋','⠙','⠹','⠸','⠼','⠴','⠦','⠧','⠇','⠏'}
    go func(){
        for gui.mainLoading {
            ch := frames[gui.spinnerIndex%len(frames)]
            gui.g.Update(func(g *gocui.Gui) error {
                v, err := g.View(panelName(4)); if err != nil { return nil }
                v.Clear(); v.Wrap = true
                fmt.Fprintf(v, "%c %s\n", ch, msg)
                return nil
            })
            gui.spinnerIndex++
            time.Sleep(100 * time.Millisecond)
        }
    }()
}

func (gui *Gui) stopSpinner() { gui.mainLoading = false }

// spinner helpers for orgs panel
func (gui *Gui) startOrgsSpinner(msg string) {
    gui.orgsLoading = true
    gui.orgsSpinnerIndex = 0
    gui.orgsLoadingMsg = msg
    frames := []rune{'⠋','⠙','⠹','⠸','⠼','⠴','⠦','⠧','⠇','⠏'}
    go func(){
        for gui.orgsLoading {
            ch := frames[gui.orgsSpinnerIndex%len(frames)]
            gui.orgsLoadingMsg = fmt.Sprintf("%c %s", ch, msg)
            gui.orgsSpinnerIndex++
            gui.renderOrgs()
            time.Sleep(100 * time.Millisecond)
        }
        gui.renderOrgs()
    }()
}

func (gui *Gui) stopOrgsSpinner() { 
    gui.orgsLoading = false 
    gui.orgsLoadingMsg = ""
}

// command log rendering
func (gui *Gui) watchCmdLog() {
    sub := cmdlog.Subscribe()
    for range sub {
        gui.renderCmdLog()
    }
}

func (gui *Gui) renderCmdLog() {
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View(panelName(5)); if err != nil { return nil }
        v.Clear()
        v.Wrap = false
        lines := cmdlog.Recent(200)
        for _, e := range lines { fmt.Fprintln(v, e.Summary()) }
        // add a right-aligned hint only when not expanded
        if !gui.cmdlogExpanded {
            if w, _ := v.Size(); w > 0 {
                hint := "?: help"
                pad := w - utf8.RuneCountInString(hint) - 2
                if pad < 0 { pad = 0 }
                fmt.Fprintln(v, strings.Repeat(" ", pad)+hint)
            }
        }
        // handle scrolling: set origin based on cmdlogTop; -1 means follow bottom
        _, h := v.Size(); if h <= 0 { h = 1 }
        total := len(lines)
        if !gui.cmdlogExpanded { total++ } // account for hint line
        if gui.cmdlogTop < 0 {
            // follow bottom
            top := total - h
            if top < 0 { top = 0 }
            v.SetOrigin(0, top)
        } else {
            // clamp and set
            maxTop := total - h
            if maxTop < 0 { maxTop = 0 }
            if gui.cmdlogTop > maxTop { gui.cmdlogTop = maxTop }
            v.SetOrigin(0, gui.cmdlogTop)
        }
        return nil
    })
}

// logs panel (panel 2) bottom-right status spinner
func (gui *Gui) startLogsStatusSpinner(msg string) {
    gui.logsStatusLoading = true
    gui.logsStatusIndex = 0
    gui.logsStatusMsg = msg
    go func(){
        for gui.logsStatusLoading {
            gui.logsStatusIndex++
            gui.renderLogs()
            time.Sleep(120 * time.Millisecond)
        }
        gui.renderLogs()
    }()
}

func (gui *Gui) stopLogsStatusSpinner() { gui.logsStatusLoading = false }

// toggle command log expansion to full width/height (excluding helpbar)
func (gui *Gui) toggleCmdlogExpand(g *gocui.Gui, v *gocui.View) error {
    // collapse others when expanding
    newVal := !gui.cmdlogExpanded
    gui.cmdlogExpanded = newVal
    if newVal {
        gui.mainExpanded = false
        gui.tracesExpanded = false
        gui.logsExpanded = false
    }
    // trigger a re-layout
    gui.g.Update(func(*gocui.Gui) error { return nil })
    // ensure focus stays on cmdlog view
    _ = gui.focusPanel(5)
    return nil
}

func (gui *Gui) toggleMainExpand(g *gocui.Gui, v *gocui.View) error {
    newVal := !gui.mainExpanded
    gui.mainExpanded = newVal
    if newVal {
        gui.tracesExpanded = false
        gui.logsExpanded = false
        gui.cmdlogExpanded = false
    }
    gui.g.Update(func(*gocui.Gui) error { return nil })
    _ = gui.focusPanel(4)
    return nil
}

// command log scroll handlers
func (gui *Gui) cmdlogUp(g *gocui.Gui, v *gocui.View) error {
    if gui.cmdlogTop < 0 { _, h := v.Size(); gui.cmdlogTop = h }
    if gui.cmdlogTop > 0 { gui.cmdlogTop-- }
    gui.renderCmdLog(); return nil
}
func (gui *Gui) cmdlogDown(g *gocui.Gui, v *gocui.View) error {
    if gui.cmdlogTop < 0 { return nil }
    gui.cmdlogTop++
    gui.renderCmdLog(); return nil
}
func (gui *Gui) cmdlogPgUp(g *gocui.Gui, v *gocui.View) error {
    _, h := v.Size(); if h <= 0 { h = 1 }
    if gui.cmdlogTop < 0 { gui.cmdlogTop = h }
    gui.cmdlogTop -= h
    if gui.cmdlogTop < 0 { gui.cmdlogTop = 0 }
    gui.renderCmdLog(); return nil
}
func (gui *Gui) cmdlogPgDn(g *gocui.Gui, v *gocui.View) error {
    _, h := v.Size(); if h <= 0 { h = 1 }
    if gui.cmdlogTop < 0 { return nil }
    gui.cmdlogTop += h
    gui.renderCmdLog(); return nil
}
func (gui *Gui) cmdlogTopKey(g *gocui.Gui, v *gocui.View) error { gui.cmdlogTop = 0; gui.renderCmdLog(); return nil }
func (gui *Gui) cmdlogBottomKey(g *gocui.Gui, v *gocui.View) error { gui.cmdlogTop = -1; gui.renderCmdLog(); return nil }

func (gui *Gui) toggleTracesExpand(g *gocui.Gui, v *gocui.View) error {
    newVal := !gui.tracesExpanded
    gui.tracesExpanded = newVal
    if newVal {
        gui.mainExpanded = false
        gui.logsExpanded = false
        gui.cmdlogExpanded = false
    }
    gui.g.Update(func(*gocui.Gui) error { return nil })
    _ = gui.focusPanel(1)
    return nil
}

func (gui *Gui) toggleLogsExpand(g *gocui.Gui, v *gocui.View) error {
    newVal := !gui.logsExpanded
    gui.logsExpanded = newVal
    if newVal {
        gui.mainExpanded = false
        gui.tracesExpanded = false
        gui.cmdlogExpanded = false
    }
    gui.g.Update(func(*gocui.Gui) error { return nil })
    _ = gui.focusPanel(2)
    return nil
}

// toast helpers (use the bottom helpbar to show brief messages)
func (gui *Gui) showToast(msg string) {
    if gui.toastTimer != nil { gui.toastTimer.Stop() }
    gui.renderHelpbar(msg)
    gui.toastTimer = time.AfterFunc(2500*time.Millisecond, func() { gui.renderHelpbar("") })
}

func (gui *Gui) renderHelpbar(msg string) {
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View("helpbar"); if err != nil { return nil }
        v.Clear()
        if msg != "" { fmt.Fprint(v, msg) }
        return nil
    })
}
// helpbar spinner for long-running tasks (e.g., trace updates)
// (old helpbar spinner removed; we now show spinner in Panel 2 bottom-right)

func (gui *Gui) cycleFocus(delta int) error {
    gui.focusedIndex = (gui.focusedIndex + delta + 6) % 6
    return gui.focusPanel(gui.focusedIndex)
}

func (gui *Gui) focusPanel(idx int) error {
    gui.focusedIndex = idx
    name := panelName(idx)
    if _, err := gui.g.SetCurrentView(name); err != nil {
        utils.Debugf("SetCurrentView(%s) err: %v", name, err)
        if isUnknownView(err) {
            return nil
        }
        return err
    }
    // visually emphasize focused view by updating title
    for i := 0; i < 6; i++ {
        if v, err := gui.g.View(panelName(i)); err == nil && v != nil {
            base := baseTitle(i)
            if i == idx {
                v.Title = fmt.Sprintf("[%s]", base)
            } else {
                v.Title = base
            }
        }
    }
    return nil
}

func panelName(i int) string {
    switch i {
    case 0:
        return "status"
    case 1:
        return "traces"
    case 2:
        return "logs"
    case 3:
        return "orgs"
    case 4:
        return "main"
    case 5:
        return "cmdlog"
    default:
        return "main"
    }
}

func baseTitle(i int) string {
    // prefixed with number for quick reference
    switch i {
    case 0:
        return "0. Status"
    case 1:
        return "1. Debug Traces"
    case 2:
        return "2. Logs / Filters"
    case 3:
        return "3. Organizations"
    case 4:
        return "4. Main"
    case 5:
        return "5. Command Log"
    default:
        return "Panel"
    }
}

// helper to write placeholder text
func setViewContent(v *gocui.View, lines ...string) {
    v.Clear()
    fmt.Fprintln(v, strings.Join(lines, "\n"))
}
