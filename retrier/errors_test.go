package retrier

import (
	"errors"
	"testing"
	"time"
)

func TestErrWithBackoff(t *testing.T) {
	err := ErrWithBackoff(errFoo, 10*time.Millisecond)
	if !errors.Is(err, errFoo) {
		t.Errorf("errors.Is(%v, errFoo) = false, want true", err)
	}
	if got := err.Error(); got != errFoo.Error() {
		t.Errorf("Error() = %q, want %q", got, errFoo.Error())
	}

	var backoff interface{ Backoff() time.Duration }
	if !errors.As(err, &backoff) {
		t.Fatal("ErrWithBackoff does not implement Backoff() time.Duration")
	}
	if got := backoff.Backoff(); got != 10*time.Millisecond {
		t.Errorf("Backoff() = %v, want 10ms", got)
	}
}

func TestErrWithBackoffNil(t *testing.T) {
	if got := ErrWithBackoff(nil, time.Second); got != nil {
		t.Error("ErrWithBackoff(nil, 1s) returned a non-nil error")
	}
}

func TestErrWithBackoffAs(t *testing.T) {
	cause := &time.ParseError{Value: "invalid"}
	err := ErrWithBackoff(cause, 0)
	var got *time.ParseError
	if !errors.As(err, &got) || got != cause {
		t.Errorf("errors.As did not recover the original *time.ParseError: got %p, want %p", got, cause)
	}
}
