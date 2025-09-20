package sf

type OrgList struct {
    Status   int           `json:"status"`
    Result   OrgListResult `json:"result"`
    Warnings []string      `json:"warnings"`
}

type OrgListResult struct {
    Other          []OrgInfo `json:"other"`
    Sandboxes      []OrgInfo `json:"sandboxes"`
    NonScratchOrgs []OrgInfo `json:"nonScratchOrgs"`
    DevHubs        []OrgInfo `json:"devHubs"`
    ScratchOrgs    []OrgInfo `json:"scratchOrgs"`
}

type OrgInfo struct {
    Alias               string `json:"alias"`
    Username            string `json:"username"`
    ConnectedStatus     string `json:"connectedStatus"`
    IsDefaultUsername   bool   `json:"isDefaultUsername"`
    IsDefaultDevHubUsername bool `json:"isDefaultDevHubUsername"`
    DefaultMarker       string `json:"defaultMarker"`
    Name                string `json:"name"`
    OrgId               string `json:"orgId"`
}

type OrgSummary struct {
    Alias           string
    Username        string
    ConnectedStatus string
    DefaultMarker   string
}

// ListAuthorizedOrgs returns a flattened list of orgs from all categories.
func ListAuthorizedOrgs() ([]OrgSummary, error) {
    res, err := Exec("sf", "org", "list", "--json")
    if err != nil && res.Stdout == "" {
        return nil, err
    }
    var payload OrgList
    if err := ParseSfJson(res.Stdout, &payload); err != nil {
        return nil, err
    }
    var out []OrgSummary
    appendOrgs := func(list []OrgInfo) {
        for _, o := range list {
            alias := o.Alias
            if alias == "" {
                alias = o.Username
            }
            out = append(out, OrgSummary{
                Alias:           alias,
                Username:        o.Username,
                ConnectedStatus: o.ConnectedStatus,
                DefaultMarker:   o.DefaultMarker,
            })
        }
    }
    appendOrgs(payload.Result.DevHubs)
    appendOrgs(payload.Result.NonScratchOrgs)
    appendOrgs(payload.Result.ScratchOrgs)
    appendOrgs(payload.Result.Other)
    appendOrgs(payload.Result.Sandboxes)
    return out, nil
}

