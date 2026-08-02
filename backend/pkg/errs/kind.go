package errs

type kind int

const (
	KindRequest kind = iota
	KindTimeout
	KindConflict
	KindNotFound
	KindUnauth
	KindInternal
)

// KindError wraps an error with a specific kind/category.
//
// It implements the error interface and provides access to the underlying error
// and its kind, allowing for both error inspection and classification.
type KindError struct {
	kind kind
	err  error
}

// Error returns the underlying error message.
//
// Implements the error interface.
func (ke *KindError) Error() string {
	return ke.err.Error()
}

// Kind returns kind
//
// Returned value can be used to:
//   - setting HTTP code
//   - implementing retry-logic
//   - logging & monitoring
func (ke *KindError) Kind() kind {
	return ke.kind
}

// Error returns the underlying error message.
//
// Implements the error interface.
func (ke *KindError) Unwrap() error {
	return ke.err
}

// NewKindError creates a new KindError with the specified kind and error.
//
// Parameters:
//   - k: the error kind/category
//   - err: the underlying error
//
// Returns a new KindError instance.
//
// Example:
//
//	return NewKindError(KindNotFound, fmt.Errorf("user with id %s not found", id))
func NewKindError(k kind, err error) *KindError {
	return &KindError{
		kind: k,
		err:  err,
	}
}
