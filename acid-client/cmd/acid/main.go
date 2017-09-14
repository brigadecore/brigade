package main

import (
	"fmt"
	"os"

	"github.com/deis/acid/acid-client/cmd/acid/commands"
)

func main() {
	if err := commands.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
