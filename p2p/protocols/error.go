package protocols

//// error codes used by this  protocol scheme
//const (
//	ErrMsgTooLong = iota
//	ErrDecode
//	ErrWrite
//	ErrInvalidMsgCode
//	ErrInvalidMsgType
//	ErrHandshake
//	ErrNoHandler
//	ErrHandler
//)
//
//// error description strings associated with the codes
//var errorToString = map[int]string{
//	ErrMsgTooLong:     "Message too long",
//	ErrDecode:         "Invalid message (RLP error)",
//	ErrWrite:          "Error sending message",
//	ErrInvalidMsgCode: "Invalid message code",
//	ErrInvalidMsgType: "Invalid message type",
//	ErrHandshake:      "Handshake error",
//	ErrNoHandler:      "No handler registered error",
//	ErrHandler:        "Message handler error",
//}
//
///*
//Error implements the standard go error interface.
//Use:
//
//  errorf(code, format, params ...interface{})
//
//Prints as:
//
// <description>: <details>
//
//where description is given by code in errorToString
//and details is fmt.Sprintf(format, params...)
//
//exported field Code can be checked
//*/
//type Error struct {
//	Code    int
//	message string
//	format  string
//	params  []interface{}
//}
//
//func (e Error) Error() (message string) {
//	if len(e.message) == 0 {
//		name, ok := errorToString[e.Code]
//		if !ok {
//			panic("invalid message code")
//		}
//		e.message = name
//		if e.format != "" {
//			e.message += ": " + fmt.Sprintf(e.format, e.params...)
//		}
//	}
//	return e.message
//}
//
//func errorf(code int, format string, params ...interface{}) *Error {
//	return &Error{
//		Code:   code,
//		format: format,
//		params: params,
//	}
//}

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
