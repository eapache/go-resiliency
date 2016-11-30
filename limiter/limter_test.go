package limiter

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {

	leaky := New(5, 2.0)
	defer leaky.Close()

	var expectedCount = 7

	var count int
	quit := time.After(1250 * time.Millisecond)

	// goroutines to be rate limited
	go func() {
		for {
			<-leaky.Limiter()
			count++
		}
	}()

	go func() {
		for {
			<-leaky.Limiter()
			count++
		}
	}()

	<-quit

	if count != expectedCount {
		t.Error("Rate limt was not at expected count", expectedCount, ", actual count was", count)
	}
}
