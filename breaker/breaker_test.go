package breaker

import (
	"errors"
	"testing"
	"time"
)

var someError = errors.New("someError")

func returnsError() error {
	return someError
}

func returnsSuccess() error {
	return nil
}

func TestBreakerErrorExpiry(t *testing.T) {
	breaker := New(2, 1, 1*time.Second)

	for i := 0; i < 5; i++ {
		if err := breaker.Run(returnsError); err != someError {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)
	}
}

func TestBreakerStateTransitions(t *testing.T) {
	breaker := New(3, 2, 1*time.Second)

	// three errors opens the breaker
	for i := 0; i < 3; i++ {
		if err := breaker.Run(returnsError); err != someError {
			t.Error(err)
		}
	}

	// breaker is open
	for i := 0; i < 5; i++ {
		if err := breaker.Run(returnsError); err != BreakerOpen {
			t.Error(err)
		}
	}

	// wait for it to half-close
	time.Sleep(2 * time.Second)
	// one success works, but is not enough to fully close
	if err := breaker.Run(returnsSuccess); err != nil {
		t.Error(err)
	}
	// error works, but re-opens immediately
	if err := breaker.Run(returnsError); err != someError {
		t.Error(err)
	}
	// breaker is open
	if err := breaker.Run(returnsError); err != BreakerOpen {
		t.Error(err)
	}

	// wait for it to half-close
	time.Sleep(2 * time.Second)
	// two successes is enough to close it for good
	for i := 0; i < 2; i++ {
		if err := breaker.Run(returnsSuccess); err != nil {
			t.Error(err)
		}
	}
	// error works
	if err := breaker.Run(returnsError); err != someError {
		t.Error(err)
	}
	// breaker is still closed
	if err := breaker.Run(returnsSuccess); err != nil {
		t.Error(err)
	}
}
