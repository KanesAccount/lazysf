package sf

type ConfigGetTargetOrg struct {
    Status  int               `json:"status"`
    Result  []ConfigKeyResult `json:"result"`
    Warnings []string         `json:"warnings"`
}

type ConfigKeyResult struct {
    Name     string `json:"name"`
    Key      string `json:"key"`
    Value    string `json:"value"`
    Path     string `json:"path"`
    Success  bool   `json:"success"`
    Location string `json:"location"`
}

type OrgDisplay struct {
    Status  int          `json:"status"`
    Result  OrgDisplayResult `json:"result"`
    Warnings []string    `json:"warnings"`
}

type OrgDisplayResult struct {
    Alias    string `json:"alias"`
    Username string `json:"username"`
}

// GetTargetOrgAlias returns the configured target org alias if set.
func GetTargetOrgAlias() (string, error) {
    res, err := Exec("sf", "config", "get", "target-org", "--json")
    if err != nil && res.Stdout == "" {
        return "", err
    }
    var payload ConfigGetTargetOrg
    if err := ParseSfJson(res.Stdout, &payload); err != nil {
        return "", err
    }
    if len(payload.Result) == 0 {
        return "", nil
    }
    return payload.Result[0].Value, nil
}

// GetCurrentOrgDisplay returns alias and username for the current target org.
func GetCurrentOrgDisplay() (alias, username string, err error) {
    res, e := Exec("sf", "org", "display", "--json")
    if e != nil && res.Stdout == "" {
        err = e
        return
    }
    var payload OrgDisplay
    if e := ParseSfJson(res.Stdout, &payload); e != nil {
        err = e
        return
    }
    alias = payload.Result.Alias
    username = payload.Result.Username
    return
}

// Internal struct for full auth info from `sf org display --json`
type orgDisplayFull struct {
    Status  int `json:"status"`
    Result struct {
        InstanceUrl string `json:"instanceUrl"`
        AccessToken string `json:"accessToken"`
        ApiVersion  string `json:"apiVersion"`
    } `json:"result"`
}

// GetAuthInfo returns instance URL, access token, and API version for the current org.
func GetAuthInfo() (instanceURL, accessToken, apiVersion string, err error) {
    res, e := Exec("sf", "org", "display", "--json")
    if e != nil && res.Stdout == "" { err = e; return }
    var payload orgDisplayFull
    if e := ParseSfJson(res.Stdout, &payload); e != nil { err = e; return }
    instanceURL = payload.Result.InstanceUrl
    accessToken = payload.Result.AccessToken
    apiVersion = payload.Result.ApiVersion
    return
}
