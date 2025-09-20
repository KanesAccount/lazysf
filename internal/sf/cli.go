package sf

import (
    "bytes"
    "encoding/json"
    "errors"
    "os/exec"
    "strings"
    "time"
    "lazysf/internal/cmdlog"
)

type CmdResult struct {
    Stdout string
    Stderr string
    Code   int
}

// Exec runs a command and returns outputs. It does not interpret JSON.
func Exec(name string, args ...string) (*CmdResult, error) {
    start := time.Now()
    cmd := exec.Command(name, args...)
    var outBuf, errBuf bytes.Buffer
    cmd.Stdout = &outBuf
    cmd.Stderr = &errBuf
    err := cmd.Run()
    res := &CmdResult{Stdout: outBuf.String(), Stderr: errBuf.String(), Code: 0}
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            res.Code = exitErr.ExitCode()
        } else {
            res.Code = -1
        }
    }
    // push to command log (suppress per-ID ApexLog deletes to avoid flooding)
    suppress := false
    if name == "sf" {
        joined := " " + strings.Join(args, " ") + " "
        if strings.Contains(joined, " data ") && strings.Contains(joined, " delete ") && strings.Contains(joined, " record ") && strings.Contains(strings.ToLower(joined), " apexlog ") {
            suppress = true
        }
    }
    if !suppress {
        e := cmdlog.Entry{Time: time.Now(), Command: name + " " + strings.Join(args, " "), Code: res.Code, Duration: time.Since(start)}
        if res.Stderr != "" {
            parts := strings.SplitN(res.Stderr, "\n", 2)
            line := parts[0]
            low := strings.ToLower(line)
            if !strings.Contains(low, "@salesforce/cli update") {
                e.Stderr = line
            }
        }
        cmdlog.Add(e)
    }
    return res, err
}

// ParseSfJson attempts to strip warnings and decode the trailing JSON.
func ParseSfJson(stdout string, v any) error {
    // Some sf commands print warnings before JSON. Find first '{' or '['
    idxObj := strings.Index(stdout, "{")
    idxArr := strings.Index(stdout, "[")
    start := -1
    if idxObj >= 0 && idxArr >= 0 {
        if idxObj < idxArr {
            start = idxObj
        } else {
            start = idxArr
        }
    } else if idxObj >= 0 {
        start = idxObj
    } else if idxArr >= 0 {
        start = idxArr
    }
    if start < 0 {
        return errors.New("no JSON object/array found in sf output")
    }
    payload := stdout[start:]
    dec := json.NewDecoder(strings.NewReader(payload))
    return dec.Decode(v)
}
