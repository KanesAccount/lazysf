package gui

import (
    "github.com/jesseduffield/gocui"
)

func (gui *Gui) openHelp(g *gocui.Gui, v *gocui.View) error {
    maxX, maxY := g.Size()
    // make the help modal about half the screen width
    w := maxX / 2
    h := maxY * 3 / 4
    if w < 60 { w = 60 }
    if h < 20 { h = maxY - 4 }
    x0 := (maxX - w) / 2
    y0 := (maxY - h) / 2
    view, err := g.SetView("help", x0, y0, x0+w, y0+h, 0)
    if err != nil {
        if !isUnknownView(err) { return nil }
        gui.inputReturnFocus = gui.focusedIndex
        view.Title = "Keybindings"
        view.Wrap = true
        view.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

        // content
        lines := []string{
            "Global:",
            "  Tab: cycle panels     0..5: focus panel     q/Ctrl-C: quit",
            "  ?: open this help     Esc: close modals",
            "",
            "Status (0):",
            "  Shows current org alias → username, and active NVIM server",
            "",
            "Traces (1):",
            "  j/k, ↑/↓: move    r: refresh    a: add trace    e: edit duration    d: delete",
            "",
            "Logs/Filters (2):",
            "  Logs: j/k, ↑/↓ move   PgUp/PgDn/Home/End navigate",
            "        Space: load (marks with *), Enter: expand/collapse",
            "        Enter: expand/collapse logs",
            "  r: refresh logs       o: open in editor",
            "  t: toggle tabs        Filters: a: add  e: edit  d: delete (applies rg -C25)",
            "",
            "Orgs (3):",
            "  j/k, ↑/↓ move   PgUp/PgDn/Home/End navigate   Space: switch target org",
            "  A: auth via web",
            "",
            "Main (4):",
            "  j/k, ↑/↓ scroll   PgUp/PgDn/Home/End   gg/G top/bottom   U/D half-page",
            "  /: search   n/N next/prev match   o: open in editor at current match/top",
            "  Tip: default message shows until you load with Space",
            "  S: set NVIM server (prefers LAZYSF_NVIM_SERVER → NVIM_LISTEN_ADDRESS)",
            "",
            "Command Log (5): shows recent commands, exit codes and durations",
        }
        view.Clear()
        for _, line := range lines { 
            view.Write([]byte(line + "\n"))
        }

        // focus and bind close keys for help view
        g.SetCurrentView("help")
        _ = g.SetKeybinding("help", 'q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return gui.closeHelp(g, v) })
        _ = g.SetKeybinding("help", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return gui.closeHelp(g, v) })
    } else {
        g.SetCurrentView("help")
    }
    return nil
}

func (gui *Gui) closeHelp(g *gocui.Gui, v *gocui.View) error {
    g.DeleteViewKeybindings("help")
    _ = g.DeleteView("help")
    _ = gui.focusPanel(gui.inputReturnFocus)
    return nil
}
