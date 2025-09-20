package sf

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "lazysf/internal/cmdlog"
)

// DeleteAllApexLogsREST attempts to delete ApexLog records via the Tooling API sObject Collections endpoint
// using a batched DELETE, which is significantly faster than one-by-one deletion.
// Returns number of logs deleted on success.
func DeleteAllApexLogsREST() (int, error) {
    start := time.Now()
    // get IDs via CLI (fast and consistent with auth context)
    soql := "SELECT Id FROM ApexLog"
    res, err := Exec("sf", "data", "query", "--json", "--use-tooling-api", "-q", soql)
    if err != nil && res.Stdout == "" { return 0, err }
    var payload struct {
        Status int `json:"status"`
        Result struct { Records []struct{ Id string `json:"Id"` } `json:"records"` } `json:"result"`
    }
    if err := ParseSfJson(res.Stdout, &payload); err != nil { return 0, err }
    ids := make([]string, 0, len(payload.Result.Records))
    for _, r := range payload.Result.Records { if r.Id != "" { ids = append(ids, r.Id) } }
    if len(ids) == 0 { return 0, nil }

    instanceURL, accessToken, apiVersion, err := GetAuthInfo()
    if err != nil { return 0, fmt.Errorf("get auth info: %w", err) }
    if apiVersion == "" { apiVersion = "64.0" }

    client := &http.Client{ Timeout: 30 * time.Second }
    deleted := 0
    // Use Tooling Composite REST API: up to 25 subrequests per composite call
    for i := 0; i < len(ids); i += 25 {
        j := i + 25
        if j > len(ids) { j = len(ids) }
        batch := ids[i:j]
        // build composite request payload
        type subReq struct {
            Method      string `json:"method"`
            URL         string `json:"url"`
            ReferenceID string `json:"referenceId"`
        }
        body := struct {
            AllOrNone        bool     `json:"allOrNone"`
            CompositeRequest []subReq `json:"compositeRequest"`
        }{ AllOrNone: false }
        for k, id := range batch {
            body.CompositeRequest = append(body.CompositeRequest, subReq{
                Method: "DELETE",
                // When posting to tooling composite endpoint, URLs are relative to that base
                URL:    fmt.Sprintf("/sobjects/ApexLog/%s", id),
                ReferenceID: fmt.Sprintf("ref_%d", k),
            })
        }
        payload, _ := json.Marshal(body)
        url := fmt.Sprintf("%s/services/data/v%s/tooling/composite", instanceURL, apiVersion)
        req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
        req.Header.Set("Authorization", "Bearer "+accessToken)
        req.Header.Set("Content-Type", "application/json")
        resp, err := client.Do(req)
        if err != nil { return deleted, err }
        if resp.StatusCode < 200 || resp.StatusCode >= 300 {
            resp.Body.Close()
            return deleted, fmt.Errorf("tooling composite delete failed: HTTP %d", resp.StatusCode)
        }
        // parse response to count successful deletes
        var compResp struct {
            CompositeResponse []struct {
                HttpStatusCode int `json:"httpStatusCode"`
            } `json:"compositeResponse"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&compResp); err != nil {
            resp.Body.Close()
            return deleted, err
        }
        resp.Body.Close()
        batchDeleted := 0
        for _, r := range compResp.CompositeResponse {
            if r.HttpStatusCode >= 200 && r.HttpStatusCode < 300 { deleted++ }
            if r.HttpStatusCode >= 200 && r.HttpStatusCode < 300 { batchDeleted++ }
        }
        // if the composite call returned 200 but none of the subrequests succeeded, treat as failure for fallback
        if batchDeleted == 0 && len(batch) > 0 {
            return deleted, fmt.Errorf("tooling composite delete returned 200 but subrequests failed")
        }
    }
    // surface success in command log
    cmdlog.Add(cmdlog.Entry{
        Time: time.Now(),
        Command: fmt.Sprintf("REST tooling delete ApexLog (%d)", deleted),
        Code: 0,
        Duration: time.Since(start),
    })
    return deleted, nil
}

// DeleteAllApexLogsFast tries REST batch delete, then falls back to CLI-based concurrent deletion.
func DeleteAllApexLogsFast() (int, error) {
    if n, err := DeleteAllApexLogsREST(); err == nil {
        return n, nil
    } else {
        // surface REST error in command log and fallback
        cmdlog.Add(cmdlog.Entry{
            Time: time.Now(),
            Command: "REST tooling delete ApexLog",
            Code: -1,
            Duration: 0,
            Stderr: err.Error(),
        })
    }
    return DeleteAllApexLogs()
}
