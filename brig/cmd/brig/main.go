package main

import (
	"fmt"
	"os"

	"github.com/Azure/brigade/brig/cmd/brig/commands"
)

func main() {
	if err := commands.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		switch e := err.(type) {
		case commands.BrigError:
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
}
