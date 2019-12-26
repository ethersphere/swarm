package protocols

// HandlerError wraps standard error
// This error is handled specially by protocol.Run
// It causes the protocol to return with ErrHandler(err)
type breakError struct {
	err error
}

func BreakError(err error) *breakError {
	return &breakError{
		err: err,
	}
}

// Unwrap returns an underlying error
func (e *breakError) Unwrap() error { return e.err }

// Error implements function of the standard go error interface
func (w *breakError) Error() string {
	return w.err.Error()
}
