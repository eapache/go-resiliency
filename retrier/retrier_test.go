package retrier

import (
	"testing"
	"time"
)

var i int

func genWork(returns []error) func() error {
	i = 0
	return func() error {
		i++
		if i > len(returns) {
			return nil
		}
		return returns[i-1]
	}
}

func TestRetrier(t *testing.T) {
	r := New([]time.Duration{0, 10 * time.Millisecond}, WhitelistClassifier{errFoo})

	err := r.Run(genWork([]error{errFoo, errFoo}))
	if err != nil {
		t.Error(err)
	}
	if i != 3 {
		t.Error("run wrong number of times")
	}

	err = r.Run(genWork([]error{errFoo, errBar}))
	if err != errBar {
		t.Error(err)
	}
	if i != 2 {
		t.Error("run wrong number of times")
	}

	err = r.Run(genWork([]error{errBar, errBaz}))
	if err != errBar {
		t.Error(err)
	}
	if i != 1 {
		t.Error("run wrong number of times")
	}
}

func TestRetrierNone(t *testing.T) {
	r := New(nil, nil)

	i = 0
	err := r.Run(func() error {
		i++
		return errFoo
	})
	if err != errFoo {
		t.Error(err)
	}
	if i != 1 {
		t.Error("run wrong number of times")
	}

	i = 0
	err = r.Run(func() error {
		i++
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if i != 1 {
		t.Error("run wrong number of times")
	}
}

func ExampleRetrier() {
	r := New(ConstantBackoff(3, 100*time.Millisecond), nil)

	err := r.Run(func() error {
		// do some work
		return nil
	})

	if err != nil {
		// handle the case where the work failed three times
	}
}
