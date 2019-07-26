package main

import (
	"os"

	"github.com/brigadecore/brigade/brig/cmd/brig/commands"
)

func main() {
	if err := commands.Root.Execute(); err != nil {
		switch e := err.(type) {
		case commands.BrigError:
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
}
