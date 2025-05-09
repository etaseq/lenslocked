package errs

// publicError satisfies both the standard 'error' interface and
// the custom 'public' interface defined in views/template.go
// by implementing the std Error() and the custom Public() methods.
// It is a custom error type that does more than a regular error:
// - It behaves like a regular error (by implementing Error())
// - Also gives a user-friendly message (by implementing Public())

type publicError struct {
	err error  // the original/internal error
	msg string // the user-friendly/public message
}

func (pe publicError) Error() string {
	return pe.err.Error()
}

func (pe publicError) Public() string {
	return pe.msg
}

func (pe publicError) Unwrap() error {
	return pe.err
}

// Public is just a constructor that wraps the original error with
// a new publicError. When you use this function, you're creating
// an instance of publicError that holds both the internal error
// (err) and a public message (msg).
// This allows you to manage errors in a structured way.
func Public(err error, msg string) error {
	return publicError{err, msg}
}
