package gui

import (
    "errors"
    "strings"

    "github.com/jesseduffield/gocui"
)

// isUnknownView returns true if the error corresponds to gocui's unknown view.
// Some gocui builds wrap the error without supporting errors.Is, so we also
// fall back to substring matching.
func isUnknownView(err error) bool {
    if err == nil {
        return false
    }
    if errors.Is(err, gocui.ErrUnknownView) {
        return true
    }
    // fallback: text match
    return strings.Contains(strings.ToLower(err.Error()), "unknown view")
}

