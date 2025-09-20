package utils

import (
    "errors"
    "os"
    "path/filepath"
)

// ConfigRoot returns ~/.config/.lazysf or ~/.lazysf if ~/.config is missing
func ConfigRoot() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    configDir := filepath.Join(home, ".config")
    if st, err := os.Stat(configDir); err == nil && st.IsDir() {
        return filepath.Join(configDir, ".lazysf"), nil
    }
    return filepath.Join(home, ".lazysf"), nil
}

func EnsureDirs(paths ...string) error {
    for _, p := range paths {
        if err := os.MkdirAll(p, 0o755); err != nil {
            return err
        }
    }
    return nil
}

func OrgPaths(alias string) (root, logs, cache string, err error) {
    if alias == "" {
        return "", "", "", errors.New("empty alias")
    }
    base, err := ConfigRoot()
    if err != nil {
        return
    }
    root = filepath.Join(base, "orgs", alias)
    logs = filepath.Join(root, "logs")
    cache = filepath.Join(root, "cache")
    return
}

