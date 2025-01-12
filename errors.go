package betalinkauth

// ServerError is an error type that represents
// an internal server error
type ServerError struct {
	Message string
}

// Error returns the error message
func (e *ServerError) Error() string {
	return e.Message
}

// ValidationError is an error type that represents
// a validation error
type ValidationError struct {
	Message string
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return e.Message
}
