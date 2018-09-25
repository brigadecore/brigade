package commands

type BrigError struct {
	Code  int
	cause error
}

func (e BrigError) Error() string {
	return e.cause.Error()
}
