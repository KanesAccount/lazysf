package gui

import (
    "strings"

    "github.com/jesseduffield/gocui"
)

type inputEditor struct {
    gui      *Gui
    onSubmit func(string)
}

func (e inputEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
    switch key {
    case gocui.KeyEnter:
        text := strings.TrimSpace(v.Buffer())
        e.gui.closeInput()
        if e.onSubmit != nil { e.onSubmit(text) }
        return true
    case gocui.KeyEsc:
        e.gui.closeInput()
        return true
    default:
        return gocui.SimpleEditor(v, key, ch, mod)
    }
}

func (gui *Gui) openInput(title, initial string, onSubmit func(string)) {
    gui.g.Update(func(g *gocui.Gui) error {
        maxX, maxY := g.Size()
        w := maxX * 2 / 3
        if w < 30 { w = maxX - 4 }
        h := 4
        x0 := (maxX - w) / 2
        y0 := (maxY - h) / 2
        v, err := g.SetView("input", x0, y0, x0+w, y0+h, 0)
        if err != nil && !isUnknownView(err) {
            return nil
        }
        // remember current focused panel to restore after closing
        gui.inputReturnFocus = gui.focusedIndex
        v.Clear()
        v.Title = title
        v.Editable = true
        v.Wrap = false
        v.Editor = inputEditor{gui: gui, onSubmit: onSubmit}
        v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
        v.SetCursor(0, 0)
        if initial != "" {
            v.Write([]byte(initial))
            v.SetCursor(len(initial), 0)
        }
        g.SetCurrentView("input")
        return nil
    })
}

func (gui *Gui) closeInput() {
    gui.g.Update(func(g *gocui.Gui) error {
        _ = g.DeleteView("input")
        // restore focus to prior panel
        _ = gui.focusPanel(gui.inputReturnFocus)
        return nil
    })
}
