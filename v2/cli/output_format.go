package main

import (
	"strings"

	"github.com/pkg/errors"
)

// validateOutputFormat validates that the requested output format (for
// commands) that support this option, is valid. It returns when an unrecognized
// format is requested.
func validateOutputFormat(outputFormat string) error {
	switch strings.ToLower(outputFormat) {
	case flagOutputTable:
	case flagOutputYAML:
	case flagOutputJSON:
	default:
		return errors.Errorf("unknown output format %q", outputFormat)
	}
	return nil
}
