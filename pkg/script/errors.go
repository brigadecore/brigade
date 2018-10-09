package script

import "fmt"

// BuildFailure represents a failure from a Brigade build
type BuildFailure struct {
	msg string
}

func (f BuildFailure) Error() string {
	return f.msg
}

// NewBuildFailure returns a BuildFailure object containing a given message
func NewBuildFailure(message string, a ...interface{}) BuildFailure {
	f := BuildFailure{
		msg: fmt.Sprintf(message, a...),
	}
	return f
}
