package sf

import (
    "os"
    "strconv"
    "sync"
    "sync/atomic"
    "time"
    "lazysf/internal/cmdlog"
)

type ApexLogList struct {
    Status   int         `json:"status"`
    Result   []ApexLog   `json:"result"`
    Warnings []string    `json:"warnings"`
}

type ApexLog struct {
    Id                 string `json:"Id"`
    LogLength          int    `json:"LogLength"`
    StartTime          string `json:"StartTime"`
    Status             string `json:"Status"`
    Operation          string `json:"Operation"`
    Location           string `json:"Location"`
    Application        string `json:"Application"`
}

func ListApexLogs() ([]ApexLog, error) {
    res, err := Exec("sf", "apex", "log", "list", "--json")
    if err != nil && res.Stdout == "" {
        return nil, err
    }
    var payload ApexLogList
    if err := ParseSfJson(res.Stdout, &payload); err != nil {
        return nil, err
    }
    return payload.Result, nil
}

// GetApexLogToDir downloads a log to the provided directory.
func GetApexLogToDir(logId, dir string) error {
    _, err := Exec("sf", "apex", "log", "get", "--log-id", logId, "--output-dir", dir)
    return err
}

// Delete all ApexLog records via Tooling API. Returns number of deleted logs.
func DeleteAllApexLogs() (int, error) {
    start := time.Now()
    // query all ApexLog Ids
    soql := "SELECT Id FROM ApexLog"
    res, err := Exec("sf", "data", "query", "--json", "--use-tooling-api", "-q", soql)
    if err != nil && res.Stdout == "" {
        return 0, err
    }
    // reuse UserQueryResult-like shape
    var payload struct {
        Status int `json:"status"`
        Result struct {
            Records []struct{ Id string `json:"Id"` } `json:"records"`
        } `json:"result"`
    }
    if err := ParseSfJson(res.Stdout, &payload); err != nil {
        return 0, err
    }
    // Build list of IDs
    ids := make([]string, 0, len(payload.Result.Records))
    for _, r := range payload.Result.Records { if r.Id != "" { ids = append(ids, r.Id) } }
    if len(ids) == 0 { return 0, nil }
    // Concurrency (default 10; override via LAZYSF_DELETE_CONCURRENCY)
    conc := 10
    if v := os.Getenv("LAZYSF_DELETE_CONCURRENCY"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 { conc = n }
    }
    sem := make(chan struct{}, conc)
    var okCount int32
    var wg sync.WaitGroup
    for _, id := range ids {
        wg.Add(1)
        sem <- struct{}{}
        go func(logID string){
            defer wg.Done()
            defer func(){ <-sem }()
            if _, err := Exec("sf", "data", "delete", "record", "--sobject", "ApexLog", "--record-id", logID, "--use-tooling-api", "--json"); err == nil {
                atomic.AddInt32(&okCount, 1)
            }
        }(id)
    }
    wg.Wait()
    // log a summary entry to command log
    cmdlog.Add(cmdlog.Entry{ Time: time.Now(), Command: "CLI tooling delete ApexLog ("+strconv.Itoa(int(okCount))+")", Code: 0, Duration: time.Since(start) })
    return int(okCount), nil
}
