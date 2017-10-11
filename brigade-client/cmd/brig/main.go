package main

import (
	"fmt"
	"os"

	"github.com/deis/brigade/brigade-client/cmd/brig/commands"
)

func main() {
	if err := commands.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
