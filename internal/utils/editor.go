package utils

import (
    "os"
    "os/exec"
    "strings"
)

// OpenInEditor opens a file in the user's editor (see OpenInEditorAt for details).
func OpenInEditor(path string) *exec.Cmd { return OpenInEditorAt(path, 0) }

// OpenInEditorAt tries to open the given file in the user's editor at a given 1-based line.
// If `nvr` is available it will target an existing Neovim instance, preferring
// LAZYSF_NVIM_SERVER, then NVIM_LISTEN_ADDRESS, else a best-effort nvr --remote-tab.
// Otherwise it falls back to spawning $EDITOR (or nvim) in the current terminal.
func OpenInEditorAt(path string, line int) *exec.Cmd {
    if _, err := exec.LookPath("nvr"); err == nil {
        // build args
        args := []string{}
        if addr := os.Getenv("LAZYSF_NVIM_SERVER"); addr != "" {
            args = append(args, "--servername", addr)
        } else if addr := os.Getenv("NVIM_LISTEN_ADDRESS"); addr != "" {
            args = append(args, "--servername", addr)
        }
        if line > 0 { args = append(args, "+call cursor("+itoa(line)+",1)") }
        // ensure log filetype in the opened buffer
        args = append(args, "+set ft=log")
        args = append(args, "--remote-tab", path)
        return exec.Command("nvr", args...)
    }

    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = "nvim"
    }
    // try to pass line to vim/nvim
    var cmd *exec.Cmd
    lower := strings.ToLower(editor)
    if strings.Contains(lower, "nvim") || strings.Contains(lower, "vim") {
        args := []string{}
        if line > 0 { args = append(args, "+"+itoa(line)) }
        args = append(args, "+set ft=log", path)
        cmd = exec.Command(editor, args...)
    } else { cmd = exec.Command(editor, path) }
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    return cmd
}

func itoa(i int) string {
    return strconvItoa(i)
}

// tiny helper to avoid importing fmt; exec args are small
func strconvItoa(i int) string {
    // simple conversion without fmt to keep deps light
    if i == 0 { return "0" }
    neg := false
    if i < 0 { neg = true; i = -i }
    buf := make([]byte, 0, 12)
    for i > 0 {
        buf = append(buf, byte('0'+(i%10)))
        i /= 10
    }
    // reverse
    for l, r := 0, len(buf)-1; l < r; l, r = l+1, r-1 { buf[l], buf[r] = buf[r], buf[l] }
    if neg { buf = append([]byte{'-'}, buf...) }
    return string(buf)
}
