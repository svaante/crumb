package main

import (
    "os"

    "github.com/svaante/crumb/internal"
)

func main() {
    args := os.Args[1:]
    crumb.Parse(args)
}
