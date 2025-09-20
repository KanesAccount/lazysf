package cmdlog

import (
    "fmt"
    "strings"
    "sync"
    "time"
)

type Entry struct {
    Time     time.Time
    Command  string
    Code     int
    Duration time.Duration
    Stderr   string
}

var (
    mu       sync.Mutex
    entries  []Entry
    maxSize  = 1000
    subsMu   sync.Mutex
    subs     []chan Entry
)

func Add(e Entry) {
    mu.Lock()
    if len(entries) >= maxSize {
        // drop oldest
        copy(entries, entries[1:])
        entries = entries[:maxSize-1]
    }
    entries = append(entries, e)
    mu.Unlock()

    subsMu.Lock()
    for _, ch := range subs {
        select { case ch <- e: default: }
    }
    subsMu.Unlock()
}

func Recent(n int) []Entry {
    mu.Lock(); defer mu.Unlock()
    if n <= 0 || n > len(entries) { n = len(entries) }
    out := make([]Entry, n)
    copy(out, entries[len(entries)-n:])
    return out
}

// Subscribe returns a channel that receives new entries. Caller should not block indefinitely.
func Subscribe() chan Entry {
    ch := make(chan Entry, 50)
    subsMu.Lock(); subs = append(subs, ch); subsMu.Unlock()
    return ch
}

// Helper to format a one-line summary for UI
func (e Entry) Summary() string {
    code := e.Code
    dur := e.Duration.Milliseconds()
    stderr := e.Stderr
    if len(stderr) > 80 {
        stderr = stderr[:77] + "..."
    }
    if stderr != "" {
        stderr = strings.ReplaceAll(stderr, "\n", "  ")
    }
    return e.Time.Format("15:04:05") +
        "  [" + itoa(code) + "]  " + e.Command +
        "  (" + itoa64(dur) + "ms)  " + stderr
}

func itoa(i int) string { return fmt.Sprintf("%d", i) }
func itoa64(i int64) string { return fmt.Sprintf("%d", i) }
