package main

import (
    "fmt"
    "os"

    "lazysf/internal/app"
    "lazysf/internal/utils"
)

func main() {
    utils.Debugf("starting lazysf main")
    a, err := app.New()
    if err != nil {
        fmt.Fprintf(os.Stderr, "init error: %v\n", err)
        os.Exit(1)
    }
    if err := a.Run(); err != nil {
        utils.Debugf("run error: %v", err)
        fmt.Fprintf(os.Stderr, "run error: %v\n", err)
        os.Exit(1)
    }
}
