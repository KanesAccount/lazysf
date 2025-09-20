package gui

import (
    "github.com/jesseduffield/gocui"
)

// openConfirm shows a yes/no modal. onYes/onNo run on the UI thread via Update.
func (gui *Gui) openConfirm(title, message string, onYes func(), onNo func()) {
    gui.g.Update(func(g *gocui.Gui) error {
        maxX, maxY := g.Size()
        w := maxX * 2 / 3
        if w < 40 { w = maxX - 4 }
        h := 6
        x0 := (maxX - w) / 2
        y0 := (maxY - h) / 2
        v, err := g.SetView("confirm", x0, y0, x0+w, y0+h, 0)
        if err != nil && !isUnknownView(err) { return nil }
        gui.inputReturnFocus = gui.focusedIndex
        v.Title = title
        v.Wrap = true
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.Clear()
        v.Write([]byte(message + "\n\n[y] Yes    [n] No"))
        g.SetCurrentView("confirm")
        // keybindings
        _ = g.SetKeybinding("confirm", 'y', gocui.ModNone, func(g *gocui.Gui, vv *gocui.View) error { gui.closeConfirm(); if onYes!=nil { onYes() }; return nil })
        _ = g.SetKeybinding("confirm", 'Y', gocui.ModNone, func(g *gocui.Gui, vv *gocui.View) error { gui.closeConfirm(); if onYes!=nil { onYes() }; return nil })
        _ = g.SetKeybinding("confirm", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, vv *gocui.View) error { gui.closeConfirm(); if onYes!=nil { onYes() }; return nil })
        _ = g.SetKeybinding("confirm", 'n', gocui.ModNone, func(g *gocui.Gui, vv *gocui.View) error { gui.closeConfirm(); if onNo!=nil { onNo() }; return nil })
        _ = g.SetKeybinding("confirm", 'N', gocui.ModNone, func(g *gocui.Gui, vv *gocui.View) error { gui.closeConfirm(); if onNo!=nil { onNo() }; return nil })
        _ = g.SetKeybinding("confirm", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, vv *gocui.View) error { gui.closeConfirm(); if onNo!=nil { onNo() }; return nil })
        return nil
    })
}

func (gui *Gui) closeConfirm() {
    gui.g.Update(func(g *gocui.Gui) error {
        g.DeleteViewKeybindings("confirm")
        _ = g.DeleteView("confirm")
        _ = gui.focusPanel(gui.inputReturnFocus)
        return nil
    })
}

