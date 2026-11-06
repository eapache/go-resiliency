package retrier

import "time"

type errWithBackoff struct {
	err     error
	backoff time.Duration
}

// ErrWithBackoff wraps err with a delay that overrides the retrier's backoff.
// The delay is used without jitter; non-positive values retry immediately.
// It preserves err for errors.Is and errors.As, and returns nil if err is nil.
// The classifier and retry limit still determine whether a retry occurs.
func ErrWithBackoff(err error, backoff time.Duration) error {
	if err == nil {
		return nil
	}
	return &errWithBackoff{err: err, backoff: backoff}
}

func (e *errWithBackoff) Error() string {
	return e.err.Error()
}

// Backoff returns the delay before the next retry.
func (e *errWithBackoff) Backoff() time.Duration {
	return e.backoff
}

// Unwrap returns the underlying error.
func (e *errWithBackoff) Unwrap() error {
	return e.err
}
