// +build go1.7

package retrier

import (
	"context"
	"log"
	"testing"
	"time"
)

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
		log.Println(i)
		t.Error("run wrong number of times")
	}
}
