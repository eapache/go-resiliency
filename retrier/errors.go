package retrier

import "time"

type errWithBackoff struct {
	err     error
	backoff time.Duration
}

func ErrWithBackoff(err error, backoff time.Duration) error {
	return &errWithBackoff{
		err:     err,
		backoff: backoff,
	}
}

func (e *errWithBackoff) Error() string {
	return e.err.Error()
}
