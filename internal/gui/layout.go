package gui

import (
    "github.com/jesseduffield/gocui"
    "lazysf/internal/utils"
)

// Layout:
// Panel 0 (very small, top-left)
// Left column: Panel 1 (top), Panel 2 (middle), Panel 3 (bottom)
// Right column: Panel 4 fills most space
// Bottom-right: Panel 5 (command log)
func (gui *Gui) layout(g *gocui.Gui) error {
    maxX, maxY := g.Size()

    // guard for tiny terminals
    if maxX < 20 || maxY < 10 {
        return nil
    }

    // if command log is expanded, render it full screen (excluding helpbar)
    if gui.cmdlogExpanded {
        utils.Debugf("layout: cmdlog expanded")
        if v, err := g.SetView(panelName(5), 0, 0, maxX-1, maxY-2, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Title = baseTitle(5) + " [Expanded — Esc to collapse]"
            v.Autoscroll = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        } else {
            v.Title = baseTitle(5) + " [Expanded — Esc to collapse]"
            v.Autoscroll = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        }
        // helpbar at bottom row
        if v, err := g.SetView("helpbar", 0, maxY-1, maxX-1, maxY-1, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Frame = false
            v.BgColor = gocui.ColorBlack
            v.FgColor = gocui.ColorCyan
        }
        return nil
    }

    // if main is expanded, render it full screen (excluding helpbar)
    if gui.mainExpanded {
        utils.Debugf("layout: main expanded")
        cmdLogHeight := 6
        if v, err := g.SetView(panelName(4), 0, 0, maxX-1, maxY-cmdLogHeight-2, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Title = baseTitle(4) + " [Expanded — Esc to collapse]"
            v.Wrap = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        } else {
            v.Title = baseTitle(4) + " [Expanded — Esc to collapse]"
            v.Wrap = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        }
        _, _ = g.SetViewOnTop(panelName(4))
        if v, err := g.SetView(panelName(5), 0, maxY-cmdLogHeight-1, maxX-1, maxY-2, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Title = baseTitle(5)
            v.Autoscroll = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        } else {
            v.Title = baseTitle(5)
            v.Autoscroll = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        }
        _, _ = g.SetViewOnTop(panelName(5))
        if v, err := g.SetView("helpbar", 0, maxY-1, maxX-1, maxY-1, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Frame = false
            v.BgColor = gocui.ColorBlack
            v.FgColor = gocui.ColorCyan
        }
        return nil
    }

    // if traces is expanded
    if gui.tracesExpanded {
        utils.Debugf("layout: traces expanded")
        cmdLogHeight := 6
        if v, err := g.SetView(panelName(1), 0, 0, maxX-1, maxY-cmdLogHeight-2, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Title = baseTitle(1) + " [Expanded — Esc to collapse]"
            v.Highlight = true
            v.SelFgColor = gocui.ColorGreen
            v.Wrap = false
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        } else {
            v.Title = baseTitle(1) + " [Expanded — Esc to collapse]"
            v.Wrap = false
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        }
        // bring expanded view to top so it obscures other panels
        _, _ = g.SetViewOnTop(panelName(1))
        if v, err := g.SetView(panelName(5), 0, maxY-cmdLogHeight-1, maxX-1, maxY-2, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Title = baseTitle(5)
            v.Autoscroll = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        } else {
            v.Title = baseTitle(5)
            v.Autoscroll = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        }
        // ensure cmdlog is on top of traces if overlapping
        _, _ = g.SetViewOnTop(panelName(5))
        if v, err := g.SetView("helpbar", 0, maxY-1, maxX-1, maxY-1, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Frame = false
            v.BgColor = gocui.ColorBlack
            v.FgColor = gocui.ColorCyan
        }
        return nil
    }

    // if logs is expanded
    if gui.logsExpanded {
        utils.Debugf("layout: logs expanded")
        cmdLogHeight := 6
        if v, err := g.SetView(panelName(2), 0, 0, maxX-1, maxY-cmdLogHeight-2, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Title = baseTitle(2) + " [Expanded — Esc to collapse]"
            v.Highlight = true
            v.SelFgColor = gocui.ColorGreen
            v.Wrap = false
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        } else {
            v.Title = baseTitle(2) + " [Expanded — Esc to collapse]"
            v.Wrap = false
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        }
        _, _ = g.SetViewOnTop(panelName(2))
        if v, err := g.SetView(panelName(5), 0, maxY-cmdLogHeight-1, maxX-1, maxY-2, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Title = baseTitle(5)
            v.Autoscroll = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        } else {
            v.Title = baseTitle(5)
            v.Autoscroll = true
            v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        }
        _, _ = g.SetViewOnTop(panelName(5))
        if v, err := g.SetView("helpbar", 0, maxY-1, maxX-1, maxY-1, 0); err != nil {
            if !isUnknownView(err) { return err }
            v.Frame = false
            v.BgColor = gocui.ColorBlack
            v.FgColor = gocui.ColorCyan
        }
        return nil
    }

    // dimensions (make left column a little thinner than half)
    leftColWidth := maxX * 2 / 5 // ~40%
    if leftColWidth < 20 { leftColWidth = 20 }
    rightColX0 := leftColWidth + 1

    utils.Debugf("layout start: maxX=%d maxY=%d", maxX, maxY)
    // Panel 0: small status bar at top spanning leftColWidth
    if v, err := g.SetView(panelName(0), 0, 0, leftColWidth, 2, 0); err != nil {
        utils.Debugf("SetView(%s) -> err=%v", panelName(0), err)
        if !isUnknownView(err) {
            return err
        }
        v.Title = baseTitle(0)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.Wrap = true
        setViewContent(v, "loading org status…")
        // ensure we have a current view set during first layout
        if _, err := g.SetCurrentView(panelName(0)); err != nil {
            utils.Debugf("SetCurrentView(%s) during create err: %v", panelName(0), err)
        } else {
            gui.hasFocused = true
            utils.Debugf("focused initial view during create: %s", panelName(0))
        }
    } else {
        utils.Debugf("SetView(%s) ok", panelName(0))
        v.Title = baseTitle(0)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
    }

    // Remaining left column vertical split into 3 roughly-equal panels
    leftTopY0 := 3
    leftHeight := maxY - leftTopY0 - 1
    if leftHeight < 6 {
        leftHeight = 6
    }
    third := leftHeight / 3
    y1 := leftTopY0 + third
    y2 := leftTopY0 + 2*third

    if v, err := g.SetView(panelName(1), 0, leftTopY0, leftColWidth, y1-1, 0); err != nil {
        utils.Debugf("SetView(%s) -> err=%v", panelName(1), err)
        if !isUnknownView(err) {
            return err
        }
        v.Title = baseTitle(1)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.Highlight = true
        v.SelFgColor = gocui.ColorGreen
        setViewContent(v, "Traces will appear here", "(a) add  (e) edit  (d) delete  (r) refresh")
    } else {
        utils.Debugf("SetView(%s) ok", panelName(1))
        v.Title = baseTitle(1)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
    }
    if v, err := g.SetView(panelName(2), 0, y1, leftColWidth, y2-1, 0); err != nil {
        utils.Debugf("SetView(%s) -> err=%v", panelName(2), err)
        if !isUnknownView(err) {
            return err
        }
        v.Title = baseTitle(2)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.Highlight = true
        v.SelFgColor = gocui.ColorGreen
        setViewContent(v, "Logs tab (t to toggle tabs)\n— use arrows to navigate, r to refresh\n\nFilters tab\n— a/e/d to manage terms")
    } else {
        utils.Debugf("SetView(%s) ok", panelName(2))
        v.Title = baseTitle(2)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
    }
    if v, err := g.SetView(panelName(3), 0, y2, leftColWidth, maxY-2, 0); err != nil {
        utils.Debugf("SetView(%s) -> err=%v", panelName(3), err)
        if !isUnknownView(err) {
            return err
        }
        v.Title = baseTitle(3)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.Highlight = true
        v.SelFgColor = gocui.ColorGreen
        setViewContent(v, "Authorized orgs will be listed here", "(space) switch  (A) auth web")
    } else {
        utils.Debugf("SetView(%s) ok", panelName(3))
        v.Title = baseTitle(3)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
    }

    // Right column main panel and command log at bottom
    cmdLogHeight := 6
    if v, err := g.SetView(panelName(4), rightColX0, 0, maxX-1, maxY-cmdLogHeight-2, 0); err != nil {
        utils.Debugf("SetView(%s) -> err=%v", panelName(4), err)
        if !isUnknownView(err) {
            return err
        }
        v.Title = baseTitle(4)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.Wrap = true
        setViewContent(v, "Main view", "Select a log with Space to load")
    } else {
        utils.Debugf("SetView(%s) ok", panelName(4))
        v.Title = baseTitle(4)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
    }
    if v, err := g.SetView(panelName(5), rightColX0, maxY-cmdLogHeight-1, maxX-1, maxY-2, 0); err != nil {
        utils.Debugf("SetView(%s) -> err=%v", panelName(5), err)
        if !isUnknownView(err) {
            return err
        }
        v.Title = baseTitle(5)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.Autoscroll = true
        setViewContent(v, "Command log…")
    } else {
        utils.Debugf("SetView(%s) ok", panelName(5))
        v.Title = baseTitle(5)
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
    }

    // ensure a focused view only after views exist
    if g.CurrentView() == nil {
        name := panelName(gui.focusedIndex)
        if v, err := g.View(name); err == nil && v != nil {
            _ = gui.focusPanel(gui.focusedIndex)
        }
    }

    // bottom status line
    // single-line help bar at the very bottom row
    if v, err := g.SetView("helpbar", 0, maxY-1, maxX-1, maxY-1, 0); err != nil {
        utils.Debugf("SetView(%s) -> err=%v", "helpbar", err)
        if !isUnknownView(err) {
            return err
        }
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.Frame = false
        v.BgColor = gocui.ColorBlack
        v.FgColor = gocui.ColorCyan
    } else {
        utils.Debugf("SetView(%s) ok", "helpbar")
        v.Clear()
        v.Frame = false
        v.BgColor = gocui.ColorBlack
        v.FgColor = gocui.ColorCyan
    }
    // set initial focus once, after views have been created
    if !gui.hasFocused {
        name := panelName(gui.focusedIndex)
        if _, err := g.SetCurrentView(name); err != nil {
            utils.Debugf("post-layout SetCurrentView(%s) err: %v", name, err)
        } else {
            gui.hasFocused = true
            utils.Debugf("focused initial view: %s", name)
        }
    }
    return nil
}
