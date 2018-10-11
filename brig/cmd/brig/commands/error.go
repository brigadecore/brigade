package commands

// BrigError represents a Brigade pipeline error, consisting of an exit code
// and cause
type BrigError struct {
	Code  int
	cause error
}

func (e BrigError) Error() string {
	return e.cause.Error()
}
