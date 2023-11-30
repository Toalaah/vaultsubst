package main

import (
	"fmt"
	"os"

	cmd "github.com/toalaah/vaultsubst/cmd/vaultsubst"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
