package main

import "github.com/urfave/cli/v2"

const (
	flagAborted        = "aborted"
	flagAnyPhase       = "any-phase"
	flagBrowse         = "browse"
	flagCanceled       = "canceled"
	flagContainer      = "container"
	flagContinue       = "continue"
	flagDescription    = "description"
	flagEvent          = "event"
	flagFailed         = "failed"
	flagFile           = "file"
	flagFollow         = "follow"
	flagID             = "id"
	flagInsecure       = "insecure"
	flagJob            = "job"
	flagLabel          = "label"
	flagNonInteractive = "non-interactive"
	flagNonTerminal    = "non-terminal"
	flagOutput         = "output"
	flagPassword       = "password"
	flagPayload        = "payload"
	flagPayloadFile    = "payload-file"
	flagPending        = "pending"
	flagProject        = "project"
	flagQualifier      = "qualifier"
	flagRole           = "role"
	flagRoot           = "root"
	flagRunning        = "running"
	flagServer         = "server"
	flagServiceAccount = "service-account"
	flagSet            = "set"
	flagSource         = "source"
	flagStarting       = "starting"
	flagSucceeded      = "succeeded"
	flagTerminal       = "terminal"
	flagTimedOut       = "timedout"
	flagType           = "type"
	flagUnknown        = "unknown"
	flagUnset          = "unset"
	flagUser           = "user"
	flagYes            = "yes"
)

const (
	flagOutputJSON      = "json"
	flagOutputTable     = "table"
	flagOutputYAML      = "yaml"
	flagOutputPlaintext = "plaintext"
)

var (
	cliFlagOutput = &cli.StringFlag{
		Name:    flagOutput,
		Aliases: []string{"o"},
		Usage: "Return output in the specified format; supported formats: table, " +
			"yaml, json",
		Value: flagOutputTable,
	}
)
