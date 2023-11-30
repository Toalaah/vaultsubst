package main

import (
	"fmt"
	"os"

	_ "embed"
	cmd "github.com/toalaah/vaultsubst/cmd/vaultsubst"
)

//go:embed version.txt
var version string

func main() {
	if err := cmd.New(version).Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
