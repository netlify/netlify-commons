package metrics

type errString struct {
	msg string
}

func (e errString) Error() string {
	return e.msg
}

// InitError indicates that the env hasn't been setup right
type InitError struct {
	errString
}

// UnknownMetricTypeError inidcates that we're sending a type we didn't expect
type UnknownMetricTypeError struct {
	errString
}

// NotStartedError indicates that stop is called before start on a timer
type NotStartedError struct {
	errString
}
