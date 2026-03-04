package handler

// assertionError is a simple error type for testing
type assertionError string

func (e assertionError) Error() string { return string(e) }
