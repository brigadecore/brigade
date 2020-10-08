package main

import "github.com/urfave/cli/v2"

const (
	flagBrowse   = "browse"
	flagFile     = "file"
	flagID       = "id"
	flagInsecure = "insecure"
	flagOutput   = "output"
	flagPassword = "password"
	flagRoot     = "root"
	flagServer   = "server"
	flagYes      = "yes"
)

var (
	cliFlagOutput = &cli.StringFlag{
		Name:    flagOutput,
		Aliases: []string{"o"},
		Usage: "Return output in the specified format; supported formats: table, " +
			"yaml, json",
		Value: "table",
	}
)
