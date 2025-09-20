package sf

import (
    "fmt"
    "time"
)

type TraceFlag struct {
    Id             string `json:"Id"`
    TracedEntityId string `json:"TracedEntityId"`
    DebugLevelId   string `json:"DebugLevelId"`
    StartDate      string `json:"StartDate"`
    ExpirationDate string `json:"ExpirationDate"`
    LogType        string `json:"LogType"`
}

type TraceQueryResult struct {
    Status int `json:"status"`
    Result struct {
        Records []TraceFlag `json:"records"`
        TotalSize int `json:"totalSize"`
        Done bool `json:"done"`
    } `json:"result"`
}

func ListTraceFlags() ([]TraceFlag, error) {
    soql := "SELECT Id, TracedEntityId, DebugLevelId, StartDate, ExpirationDate, LogType FROM TraceFlag"
    res, err := Exec("sf", "data", "query", "--json", "--use-tooling-api", "-q", soql)
    if err != nil && res.Stdout == "" { return nil, err }
    var payload TraceQueryResult
    if err := ParseSfJson(res.Stdout, &payload); err != nil { return nil, err }
    return payload.Result.Records, nil
}

type DebugLevel struct {
    Id string `json:"Id"`
    DeveloperName string `json:"DeveloperName"`
}
type DebugLevelResult struct {
    Status int `json:"status"`
    Result struct { Records []DebugLevel `json:"records"` } `json:"result"`
}

func EnsureDebugLevelId() (string, error) {
    soql := "SELECT Id, DeveloperName FROM DebugLevel ORDER BY CreatedDate LIMIT 1"
    res, err := Exec("sf", "data", "query", "--json", "--use-tooling-api", "-q", soql)
    if err != nil && res.Stdout == "" { return "", err }
    var payload DebugLevelResult
    if err := ParseSfJson(res.Stdout, &payload); err != nil { return "", err }
    if len(payload.Result.Records) == 0 { return "", fmt.Errorf("no DebugLevel found") }
    return payload.Result.Records[0].Id, nil
}

func isoTimeUTC(t time.Time) string {
    // format: 2006-01-02T15:04:05.000+0000
    return t.UTC().Format("2006-01-02T15:04:05.000-0700")
}

// UpsertTraceForUser creates or updates a trace flag for the user for the given duration.
func UpsertTraceForUser(userId, debugLevelId string, dur time.Duration) error {
    traces, err := ListTraceFlags()
    if err != nil { return err }
    var existingId string
    for _, tf := range traces {
        if tf.TracedEntityId == userId { existingId = tf.Id; break }
    }
    start := isoTimeUTC(time.Now())
    exp := isoTimeUTC(time.Now().Add(dur))
    if existingId != "" {
        // update
        _, err := Exec("sf", "data", "record", "update", "--sobject", "TraceFlag", "--record-id", existingId,
            "-v", fmt.Sprintf("StartDate=%s ExpirationDate=%s DebugLevelId=%s", start, exp, debugLevelId), "--use-tooling-api", "--json")
        return err
    }
    // create
    _, err = Exec("sf", "data", "record", "create", "--sobject", "TraceFlag",
        "-v", fmt.Sprintf("TracedEntityId=%s LogType=USER_DEBUG DebugLevelId=%s StartDate=%s ExpirationDate=%s", userId, debugLevelId, start, exp),
        "--use-tooling-api", "--json")
    return err
}

func DeleteTrace(traceId string) error {
    _, err := Exec("sf", "data", "delete", "record", "--sobject", "TraceFlag", "--record-id", traceId, "--use-tooling-api", "--json")
    return err
}

