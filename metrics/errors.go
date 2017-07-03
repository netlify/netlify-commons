package metrics

// InitError indicates that the env hasn't been setup right
type InitError struct {
	error
}

// UnknownMetricTypeError inidcates that we're sending a type we didn't expect
type UnknownMetricTypeError struct {
	error
}

// NotStartedError indicates that stop is called before start on a timer
type NotStartedError struct {
	error
}
