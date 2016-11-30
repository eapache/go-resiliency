package limiter

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {

	rl := New(5, 2.0)
	var expectedCount = 7

	var count int
	quit := time.After(1250 * time.Millisecond)

	// goroutines to be rate limited
	go func() {
		for {
			<-rl
			count++
		}
	}()

	go func() {
		for {
			<-rl
			count++
		}
	}()

	<-quit

	if count != expectedCount {
		t.Error("Rate limt was not at expected count", expectedCount, ", actual count was", count)
	}
}
