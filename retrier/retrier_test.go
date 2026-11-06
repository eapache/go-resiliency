package retrier

import (
	"context"
	"errors"
	"fmt"
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

func genWorkWithCtx() func(ctx context.Context) error {
	i = 0
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return errFoo
		default:
			i++
		}
		return nil
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

func TestRetrierCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	r := New([]time.Duration{0, 10 * time.Millisecond}, WhitelistClassifier{})

	err := r.RunCtx(ctx, genWorkWithCtx())
	if err != nil {
		t.Error(err)
	}
	if i != 1 {
		t.Error("run wrong number of times")
	}

	cancel()

	err = r.RunCtx(ctx, genWorkWithCtx())
	if err != errFoo {
		t.Error("context must be cancelled")
	}
	if i != 0 {
		t.Error("run wrong number of times")
	}
}

func TestRetrierCtxError(t *testing.T) {
	ctx := context.Background()
	r := New([]time.Duration{0, 10 * time.Millisecond}, nil)
	errExpected := []error{errFoo, errFoo, errBar, errBaz}
	retries := 0
	err := r.RunCtx(ctx, func(ctx context.Context) error {
		if retries >= len(errExpected) {
			return nil
		}
		err := errExpected[retries]
		retries++
		return err
	})
	if err != errBar {
		t.Error(err)
	}
}

func TestRetrierRunFnError(t *testing.T) {
	ctx := context.Background()
	r := New([]time.Duration{0, 10 * time.Millisecond}, nil)
	errExpected := []error{errFoo, errFoo, errBar, errBaz}

	err := r.RunFn(ctx, func(ctx context.Context, retries int) error {
		if retries >= len(errExpected) {
			return nil
		}
		return errExpected[retries]
	})
	if err != errBar {
		t.Error(err)
	}
}

func TestRetrierCtxWithInfinite(t *testing.T) {
	ctx := context.Background()
	r := New([]time.Duration{0, 10 * time.Millisecond}, nil).WithInfiniteRetry()
	errExpected := []error{errFoo, errFoo, errFoo, errBar, errBaz}
	retries := 0
	err := r.RunCtx(ctx, func(ctx context.Context) error {
		if retries >= len(errExpected) {
			return nil
		}
		err := errExpected[retries]
		retries++
		return err
	})
	if err != nil {
		t.Error(err)
	}
}

func TestRetrierRunFnWithInfinite(t *testing.T) {
	ctx := context.Background()
	r := New([]time.Duration{0, 10 * time.Millisecond}, nil).WithInfiniteRetry()
	errExpected := []error{errFoo, errFoo, errFoo, errBar, errBaz}

	err := r.RunFn(ctx, func(ctx context.Context, retries int) error {
		if retries >= len(errExpected) {
			return nil
		}
		return errExpected[retries]
	})
	if err != nil {
		t.Error(err)
	}
}

type backoffError struct {
	error
	delay time.Duration
}

func (e backoffError) Backoff() time.Duration {
	return e.delay
}

func TestRetrierCalcSleep(t *testing.T) {
	custom := backoffError{error: errFoo, delay: 7 * time.Millisecond}
	tests := []struct {
		name  string
		retry int
		err   error
		want  time.Duration
	}{
		{name: "nil error", want: time.Second},
		{name: "ordinary error", retry: 1, err: errFoo, want: 2 * time.Second},
		{name: "last static delay", retry: 3, err: errFoo, want: 2 * time.Second},
		{name: "helper", err: ErrWithBackoff(errFoo, 10*time.Millisecond), want: 10 * time.Millisecond},
		{name: "zero override", err: ErrWithBackoff(errFoo, 0)},
		{name: "negative override", err: ErrWithBackoff(errFoo, -time.Millisecond), want: -time.Millisecond},
		{name: "custom error", err: custom, want: 7 * time.Millisecond},
		{name: "wrapped helper", err: fmt.Errorf("work: %w", ErrWithBackoff(errFoo, 10*time.Millisecond)), want: 10 * time.Millisecond},
		{name: "wrapped custom error", err: fmt.Errorf("work: %w", custom), want: 7 * time.Millisecond},
		{name: "outer override", err: ErrWithBackoff(custom, 10*time.Millisecond), want: 10 * time.Millisecond},
		{name: "infinite retry index", retry: 3, err: custom, want: 7 * time.Millisecond},
		{name: "ordinary error after override", retry: 1, err: errFoo, want: 2 * time.Second},
	}
	r := New([]time.Duration{time.Second, 2 * time.Second}, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.calcSleep(tt.retry, tt.err); got != tt.want {
				t.Errorf("calcSleep(%d, %v) = %v, want %v", tt.retry, tt.err, got, tt.want)
			}
		})
	}
}

func TestRetrierDynamicBackoffJitter(t *testing.T) {
	r := New([]time.Duration{time.Hour}, nil)
	r.SetJitter(0.5)
	r.rand.Seed(1)
	err := ErrWithBackoff(errFoo, 10*time.Millisecond)
	if got := r.calcSleep(0, err); got != 10*time.Millisecond {
		t.Errorf("dynamic delay with jitter = %v, want 10ms", got)
	}
}

func TestRetrierWithDynamicBackoff(t *testing.T) {
	tests := []struct {
		name      string
		class     Classifier
		backoff   []time.Duration
		wantCalls int
	}{
		{name: "default", backoff: []time.Duration{0, 0}, wantCalls: 3},
		{name: "whitelist", class: WhitelistClassifier{errFoo}, backoff: []time.Duration{0, 0}, wantCalls: 3},
		{name: "blacklist", class: BlacklistClassifier{errFoo}, backoff: []time.Duration{0, 0}, wantCalls: 1},
		{name: "no retries", wantCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.backoff, tt.class)
			workErr := ErrWithBackoff(errFoo, 0)
			calls := 0
			err := r.Run(func() error {
				calls++
				return workErr
			})
			if err != workErr {
				t.Errorf("Run() = %v, want original work error %v", err, workErr)
			}
			if calls != tt.wantCalls {
				t.Errorf("work calls = %d, want %d", calls, tt.wantCalls)
			}
		})
	}
}

func TestRetrierDynamicBackoffCancellation(t *testing.T) {
	for _, surfaceWorkErrors := range []bool{false, true} {
		t.Run(fmt.Sprintf("surfaceWorkErrors=%t", surfaceWorkErrors), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			r := New([]time.Duration{0}, nil)
			workErr := ErrWithBackoff(errFoo, time.Hour)
			want := error(context.Canceled)
			if surfaceWorkErrors {
				r.WithSurfaceWorkErrors()
				want = workErr
			}

			calls := 0
			err := r.RunCtx(ctx, func(context.Context) error {
				calls++
				cancel()
				return workErr
			})
			if err != want {
				t.Errorf("RunCtx() = %v, want %v", err, want)
			}
			if calls != 1 {
				t.Errorf("work calls = %d, want 1", calls)
			}
		})
	}
}

func TestRetrierRunFnWithSurfaceWorkErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := New([]time.Duration{0, 10 * time.Millisecond}, nil).WithSurfaceWorkErrors()
	errExpected := []error{errFoo, errBar, errBaz}

	err := r.RunFn(ctx, func(ctx context.Context, retries int) error {
		if retries >= len(errExpected) {
			return nil
		}
		if retries == 1 {
			// Context canceled inside second call to work function.
			cancel()
		}
		err := errExpected[retries]
		retries++
		return err
	})
	if err != errBar {
		t.Error(err)
	}
}

func TestRetrierRunFnWithoutSurfaceWorkErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := New([]time.Duration{0, 10 * time.Millisecond}, nil)
	errExpected := []error{errFoo, errBar, errBaz}

	err := r.RunFn(ctx, func(ctx context.Context, retries int) error {
		if retries >= len(errExpected) {
			return nil
		}
		if retries == 1 {
			// Context canceled inside second call to work function.
			cancel()
		}
		err := errExpected[retries]
		retries++
		return err
	})
	if err != context.Canceled {
		t.Error(err)
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

func TestRetrierJitter(t *testing.T) {
	r := New([]time.Duration{0, 10 * time.Millisecond, 4 * time.Hour}, nil)

	if r.calcSleep(0, nil) != 0 {
		t.Error("Incorrect sleep calculated")
	}
	if r.calcSleep(1, nil) != 10*time.Millisecond {
		t.Error("Incorrect sleep calculated")
	}
	if r.calcSleep(2, nil) != 4*time.Hour {
		t.Error("Incorrect sleep calculated")
	}

	r.SetJitter(0.25)
	for i := 0; i < 20; i++ {
		if got := r.calcSleep(0, nil); got != 0 {
			t.Errorf("sample %d: calcSleep(0, nil) = %v, want 0", i, got)
		}

		slp := r.calcSleep(1, nil)
		if slp < 7500*time.Microsecond || slp > 12500*time.Microsecond {
			t.Errorf("sample %d: calcSleep(1, nil) = %v, want [7.5ms, 12.5ms]", i, slp)
		}

		slp = r.calcSleep(2, nil)
		if slp < 3*time.Hour || slp > 5*time.Hour {
			t.Errorf("sample %d: calcSleep(2, nil) = %v, want [3h, 5h]", i, slp)
		}
	}

	r.SetJitter(-1)
	if r.jitter != 0.25 {
		t.Error("Invalid jitter value accepted")
	}

	r.SetJitter(2)
	if r.jitter != 0.25 {
		t.Error("Invalid jitter value accepted")
	}
}

func TestRetrierThreadSafety(t *testing.T) {
	r := New([]time.Duration{0}, nil)
	for i := 0; i < 2; i++ {
		go func() {
			r.Run(func() error {
				return errors.New("error")
			})
		}()
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
