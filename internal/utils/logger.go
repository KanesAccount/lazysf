package utils

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
)

var (
    debugEnabled  = true
    stderrEnabled = os.Getenv("LAZYSF_DEBUG_STDERR") == "1"
    logFile       *os.File
    logOnce       sync.Once
)

func logInit() {
    if !debugEnabled {
        return
    }
    // write logs to ./logs relative to current working directory
    cwd, err := os.Getwd()
    if err != nil {
        cwd = "."
    }
    logsDir := filepath.Join(cwd, "logs")
    if mkErr := os.MkdirAll(logsDir, 0o755); mkErr != nil {
        // fallback to temp dir if project logs dir is not writable
        logsDir = os.TempDir()
    }
    name := fmt.Sprintf("lazysf-%d.log", time.Now().UnixNano())
    path := filepath.Join(logsDir, name)
    f, err := os.Create(path)
    if err == nil {
        logFile = f
        if stderrEnabled {
            fmt.Fprintf(os.Stderr, "[lazysf] debug log file: %s\n", path)
        }
    } else {
        if stderrEnabled {
            fmt.Fprintf(os.Stderr, "[lazysf] failed to create log file: %v\n", err)
        }
    }
}

func Debugf(format string, a ...any) {
    if !debugEnabled {
        return
    }
    logOnce.Do(logInit)
    msg := fmt.Sprintf("[lazysf] "+format+"\n", a...)
    if stderrEnabled {
        fmt.Fprint(os.Stderr, msg)
    }
    if logFile != nil {
        _, _ = logFile.WriteString(msg)
    }
}
