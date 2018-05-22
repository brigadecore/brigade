// Package decolorizer provides utilities for removing color codes from data
package decolorizer

import (
	"io"
	"regexp"
)

// filter matches ASCII color code sequences.
// See https://stackoverflow.com/questions/4842424/list-of-ansi-color-escape-sequences
var filter = regexp.MustCompile(`\x1b\[[0-9;]+m`)

// New creates a new Writer that removes colors before writing to the underlying writer.
func New(destination io.Writer) *Writer {
	return &Writer{dest: destination}
}

// Writer wraps a decolorizing filter around an embedded writer.
type Writer struct {
	dest io.Writer
}

// Write takes an unfiltered input []byte and filters out any ASCII terminal color sequences.
//
// The cleaned data is then written to the internal writer.
func (w *Writer) Write(p []byte) (n int, err error) {
	// For now, we're assuming that the calls to Write are not breaking in between
	// the color escape sequence. This is probably not a safe assumption for all
	// streams of data, but is safe for log files sent over the Kubernetes API.
	//
	// We could either buffer lines (since color sequences can't cross lines) or
	// implement a buffering scanner.
	buf := filter.ReplaceAll(p, []byte{})
	return w.dest.Write(buf)
}
