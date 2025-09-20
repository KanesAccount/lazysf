package gui

import (
    "github.com/jesseduffield/gocui"
)

func (gui *Gui) setKeybindings() error {
    // quit
    if err := gui.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, gui.wrap(gui.quit)); err != nil {
        return err
    }
    if err := gui.g.SetKeybinding("", 'q', gocui.ModNone, gui.wrap(gui.quit)); err != nil {
        return err
    }
    // global escape: close modal if open
    if err := gui.g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, gui.wrap(gui.handleEsc)); err != nil {
        return err
    }
    // global help
    if err := gui.g.SetKeybinding("", '?', gocui.ModNone, gui.wrap(gui.openHelp)); err != nil { return err }

    // cycle panels
    if err := gui.g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, gui.wrap(func(g *gocui.Gui, v *gocui.View) error {
        return gui.cycleFocus(1)
    })); err != nil {
        return err
    }

    // number keys to jump to panel 0..5
    keys := []rune{'0', '1', '2', '3', '4', '5'}
    for i, ch := range keys {
        idx := i
        if err := gui.g.SetKeybinding("", ch, gocui.ModNone, gui.wrap(func(g *gocui.Gui, v *gocui.View) error {
            return gui.focusPanel(idx)
        })); err != nil {
            return err
        }
    }

    // Orgs panel navigation/actions
    if err := gui.g.SetKeybinding(panelName(3), gocui.KeyArrowUp, gocui.ModNone, gui.wrap(gui.orgsUp)); err != nil {
        return err
    }
    if err := gui.g.SetKeybinding(panelName(3), gocui.KeyArrowDown, gocui.ModNone, gui.wrap(gui.orgsDown)); err != nil {
        return err
    }
    if err := gui.g.SetKeybinding(panelName(3), 'k', gocui.ModNone, gui.wrap(gui.orgsUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(3), 'j', gocui.ModNone, gui.wrap(gui.orgsDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(3), 'g', gocui.ModNone, gui.wrap(gui.orgsG)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(3), 'G', gocui.ModNone, gui.wrap(gui.orgsEnd)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(3), gocui.KeyPgup, gocui.ModNone, gui.wrap(func(g *gocui.Gui, v *gocui.View) error {
        _, h := v.Size(); if h <= 0 { h = 1 }
        gui.orgSel -= h
        if gui.orgSel < 0 { gui.orgSel = 0 }
        gui.renderOrgs(); return nil
    })); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(3), gocui.KeyPgdn, gocui.ModNone, gui.wrap(func(g *gocui.Gui, v *gocui.View) error {
        _, h := v.Size(); if h <= 0 { h = 1 }
        gui.orgSel += h
        if gui.orgSel > len(gui.orgs)-1 { gui.orgSel = len(gui.orgs)-1 }
        gui.renderOrgs(); return nil
    })); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(3), gocui.KeyHome, gocui.ModNone, gui.wrap(func(g *gocui.Gui, v *gocui.View) error {
        gui.orgSel = 0; gui.renderOrgs(); return nil
    })); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(3), gocui.KeyEnd, gocui.ModNone, gui.wrap(func(g *gocui.Gui, v *gocui.View) error {
        if len(gui.orgs) > 0 { gui.orgSel = len(gui.orgs)-1 }
        gui.renderOrgs(); return nil
    })); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(3), ' ', gocui.ModNone, gui.wrap(gui.orgsSwitch)); err != nil {
        return err
    }
    if err := gui.g.SetKeybinding(panelName(3), 'A', gocui.ModNone, gui.wrap(gui.orgsAuth)); err != nil {
        return err
    }

    // Traces panel (1)
    if err := gui.g.SetKeybinding(panelName(1), gocui.KeyArrowUp, gocui.ModNone, gui.wrap(gui.tracesUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(1), gocui.KeyArrowDown, gocui.ModNone, gui.wrap(gui.tracesDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(1), 'k', gocui.ModNone, gui.wrap(gui.tracesUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(1), 'j', gocui.ModNone, gui.wrap(gui.tracesDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(1), 'r', gocui.ModNone, gui.wrap(gui.tracesRefresh)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(1), 'a', gocui.ModNone, gui.wrap(gui.tracesAdd)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(1), 'e', gocui.ModNone, gui.wrap(gui.tracesEditDuration)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(1), 'd', gocui.ModNone, gui.wrap(gui.tracesDelete)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(1), gocui.KeyEnter, gocui.ModNone, gui.wrap(gui.toggleTracesExpand)); err != nil { return err }

    // Logs panel navigation/actions
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeyArrowUp, gocui.ModNone, gui.wrap(gui.logsUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeyArrowDown, gocui.ModNone, gui.wrap(gui.logsDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'k', gocui.ModNone, gui.wrap(gui.logsUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'j', gocui.ModNone, gui.wrap(gui.logsDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), ' ', gocui.ModNone, gui.wrap(gui.logsEnter)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeySpace, gocui.ModNone, gui.wrap(gui.logsEnter)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeyEnter, gocui.ModNone, gui.wrap(gui.toggleLogsExpand)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'g', gocui.ModNone, gui.wrap(gui.logsG)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'G', gocui.ModNone, gui.wrap(gui.logsEnd)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeyPgup, gocui.ModNone, gui.wrap(gui.logsPgUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeyPgdn, gocui.ModNone, gui.wrap(gui.logsPgDn)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeyHome, gocui.ModNone, gui.wrap(gui.logsHome)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeyEnd, gocui.ModNone, gui.wrap(gui.logsEnd)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'r', gocui.ModNone, gui.wrap(gui.logsRefresh)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'o', gocui.ModNone, gui.wrap(gui.logsOpenEditor)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'D', gocui.ModNone, gui.wrap(gui.logsDeleteAll)); err != nil { return err }
    // Spacebar previews log in main (Panel 4)
    if err := gui.g.SetKeybinding(panelName(2), ' ', gocui.ModNone, gui.wrap(gui.logsEnter)); err != nil { return err }
    // Enter expands/collapses logs view
    if err := gui.g.SetKeybinding(panelName(2), gocui.KeyEnter, gocui.ModNone, gui.wrap(gui.toggleLogsExpand)); err != nil { return err }
    // Panel 2 tabs toggle and filters actions
    if err := gui.g.SetKeybinding(panelName(2), 't', gocui.ModNone, gui.wrap(gui.panel2Toggle)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'a', gocui.ModNone, gui.wrap(gui.filtersAdd)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'e', gocui.ModNone, gui.wrap(gui.filtersEdit)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(2), 'd', gocui.ModNone, gui.wrap(gui.filtersDelete)); err != nil { return err }
    // (Alt+Enter no longer needed; regular Enter toggles expansion)

    // Main panel scrolling (Panel 4)
    if err := gui.g.SetKeybinding(panelName(4), gocui.KeyArrowUp, gocui.ModNone, gui.wrap(gui.mainUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), gocui.KeyArrowDown, gocui.ModNone, gui.wrap(gui.mainDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'k', gocui.ModNone, gui.wrap(gui.mainUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'j', gocui.ModNone, gui.wrap(gui.mainDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'g', gocui.ModNone, gui.wrap(gui.mainG)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'G', gocui.ModNone, gui.wrap(gui.mainEnd)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'U', gocui.ModNone, gui.wrap(gui.mainJumpUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'D', gocui.ModNone, gui.wrap(gui.mainJumpDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), '/', gocui.ModNone, gui.wrap(gui.mainSearchPrompt)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'n', gocui.ModNone, gui.wrap(gui.mainSearchNext)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'N', gocui.ModNone, gui.wrap(gui.mainSearchPrev)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'o', gocui.ModNone, gui.wrap(gui.mainOpenEditorAt)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), 'S', gocui.ModNone, gui.wrap(gui.mainSetNvimServer)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), gocui.KeyPgup, gocui.ModNone, gui.wrap(gui.mainPgUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), gocui.KeyPgdn, gocui.ModNone, gui.wrap(gui.mainPgDn)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), gocui.KeyHome, gocui.ModNone, gui.wrap(gui.mainHome)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), gocui.KeyEnd, gocui.ModNone, gui.wrap(gui.mainEnd)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(4), gocui.KeyEnter, gocui.ModNone, gui.wrap(gui.toggleMainExpand)); err != nil { return err }

    // Command log (Panel 5)
    if err := gui.g.SetKeybinding(panelName(5), gocui.KeyEnter, gocui.ModNone, gui.wrap(gui.toggleCmdlogExpand)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(5), gocui.KeyArrowUp, gocui.ModNone, gui.wrap(gui.cmdlogUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(5), gocui.KeyArrowDown, gocui.ModNone, gui.wrap(gui.cmdlogDown)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(5), gocui.KeyPgup, gocui.ModNone, gui.wrap(gui.cmdlogPgUp)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(5), gocui.KeyPgdn, gocui.ModNone, gui.wrap(gui.cmdlogPgDn)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(5), 'g', gocui.ModNone, gui.wrap(gui.cmdlogTopKey)); err != nil { return err }
    if err := gui.g.SetKeybinding(panelName(5), 'G', gocui.ModNone, gui.wrap(gui.cmdlogBottomKey)); err != nil { return err }

    return nil
}

func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
    return gocui.ErrQuit
}

func (gui *Gui) wrap(f func(*gocui.Gui, *gocui.View) error) func(*gocui.Gui, *gocui.View) error {
    return func(g *gocui.Gui, v *gocui.View) error {
        return f(g, v)
    }
}

// handleEsc closes any open modal input view
func (gui *Gui) handleEsc(g *gocui.Gui, v *gocui.View) error {
    if gui.tracesExpanded { gui.tracesExpanded = false; gui.g.Update(func(*gocui.Gui) error { return nil }); _ = gui.focusPanel(1); return nil }
    if gui.logsExpanded { gui.logsExpanded = false; gui.g.Update(func(*gocui.Gui) error { return nil }); _ = gui.focusPanel(2); return nil }
    if gui.mainExpanded { gui.mainExpanded = false; gui.g.Update(func(*gocui.Gui) error { return nil }); _ = gui.focusPanel(4); return nil }
    if gui.cmdlogExpanded { gui.cmdlogExpanded = false; gui.g.Update(func(*gocui.Gui) error { return nil }); _ = gui.focusPanel(5); return nil }
    if _, err := g.View("input"); err == nil {
        gui.closeInput()
        return nil
    }
    if _, err := g.View("help"); err == nil {
        gui.closeHelp(g, v)
        return nil
    }
    return nil
}
