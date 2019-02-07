// +build go1.7

// Package retrier implements the "retriable" resiliency pattern for Go.
package retrier

import (
	"context"
	"time"
)

// Run executes the given work function by executing RunCtx without context.Context.
func (r *Retrier) Run(work func() error) error {
	return r.RunCtx(context.Background(), func(ctx context.Context) error {
		// never use ctx
		return work()
	})
}

// RunCtx executes the given work function, then classifies its return value based on the classifier used
// to construct the Retrier. If the result is Succeed or Fail, the return value of the work function is
// returned to the caller. If the result is Retry, then Run sleeps according to the its backoff policy
// before retrying. If the total number of retries is exceeded then the return value of the work function
// is returned to the caller regardless.
func (r *Retrier) RunCtx(ctx context.Context, work func(ctx context.Context) error) error {
	retries := 0
	for {
		ret := work(ctx)

		switch r.class.Classify(ret) {
		case Succeed, Fail:
			return ret
		case Retry:
			if retries >= len(r.backoff) {
				return ret
			}

			timeout := time.After(r.calcSleep(retries))
			if err := r.sleep(ctx, timeout); err != nil {
				return err
			}

			retries++
		}
	}
}

func (r *Retrier) sleep(ctx context.Context, t <-chan time.Time) error {
	select {
	case <-t:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
