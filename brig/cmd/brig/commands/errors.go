package commands

import "fmt"

type BuildFailure struct {
	msg string
}

func (f BuildFailure) Error() string {
	return f.msg
}

func NewBuildFailure(message string, a ...interface{}) BuildFailure {
	f := BuildFailure{
		msg: fmt.Sprintf(message, a...),
	}
	return f
}