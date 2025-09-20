package sf

type User struct {
    Id       string `json:"Id"`
    Name     string `json:"Name"`
    Username string `json:"Username"`
}

type UserQueryResult struct {
    Status int `json:"status"`
    Result struct {
        Records []User `json:"records"`
        TotalSize int `json:"totalSize"`
        Done bool `json:"done"`
    } `json:"result"`
}

func ListActiveUsers() ([]User, error) {
    // Using Tooling/REST data query
    soql := "SELECT Id, Name, Username FROM User WHERE IsActive = true ORDER BY Name"
    res, err := Exec("sf", "data", "query", "--json", "-q", soql)
    if err != nil && res.Stdout == "" {
        return nil, err
    }
    var payload UserQueryResult
    if err := ParseSfJson(res.Stdout, &payload); err != nil {
        return nil, err
    }
    return payload.Result.Records, nil
}

