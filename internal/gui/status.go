package gui

import (
    "fmt"
    "github.com/jesseduffield/gocui"
)

// SetStatusText updates the text in the status view (panel 0).
func (gui *Gui) SetStatusText(lines ...string) {
    gui.g.Update(func(g *gocui.Gui) error {
        v, err := g.View(panelName(0))
        if err != nil {
            return nil
        }
        v.Clear()
        for _, line := range lines {
            fmt.Fprintln(v, line)
        }
        return nil
    })
}
