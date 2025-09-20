package app

import (
    "lazysf/internal/gui"
    "lazysf/internal/sf"
    "lazysf/internal/utils"
)

type App struct {
    gui *gui.Gui
}

func New() (*App, error) {
    g, err := gui.New()
    if err != nil {
        return nil, err
    }
    return &App{gui: g}, nil
}

func (a *App) Run() error {
    // kick off bootstrap tasks that will update the GUI
    go a.bootstrap()
    return a.gui.Run()
}

func (a *App) bootstrap() {
    // Detect target org and show in status panel
    alias, err := sf.GetTargetOrgAlias()
    if err != nil {
        utils.Debugf("error getting target org: %v", err)
        a.gui.SetStatusText("sf not detected or no target org set")
        return
    }
    if alias == "" {
        a.gui.SetStatusText("No target org set", "Press '3' then Space to switch or 'A' to auth")
        return
    }
    a.gui.SetStatusText("Target org:", alias)
    // Get display info for alias->username
    _, username, err := sf.GetCurrentOrgDisplay()
    if err != nil {
        utils.Debugf("error getting org display: %v", err)
        a.gui.SetStatusText(alias+" → (unknown username)")
        return
    }
    a.gui.SetCurrentAlias(alias)
    a.gui.SetCurrentUser(username)
    a.gui.RefreshStatusPanel()
    // Fetch orgs panel data and logs
    a.gui.FetchOrgsPublic()
    a.gui.FetchLogsPublic()
    a.gui.FetchTracesPublic()
}
