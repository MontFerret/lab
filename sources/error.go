package sources

type Error struct {
	error
	Filename string
	Message  string
}

func NewError(filename string, message string) Error {
	return Error{
		Filename: filename,
		Message:  message,
	}
}

func NewErrorFrom(filename string, err error) Error {
	return Error{
		Filename: filename,
		Message:  err.Error(),
	}
}

func (e Error) Error() string {
	return e.Filename + ":" + e.Message
}
