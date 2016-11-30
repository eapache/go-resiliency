package limiter

import "time"

type LeakyBucket struct {
	bucket chan struct{}
	quit   chan struct{}
}

func (rl *LeakyBucket) fill(burst int) {
	for i := 0; i < burst; i++ {
		rl.bucket <- struct{}{}
	}
}

func (rl *LeakyBucket) drip(drip float64) {
	driptime := float64(1000000000) / drip
	for {
		time.Sleep(time.Duration(driptime) * time.Nanosecond)
		// close
		select {
		case rl.bucket <- struct{}{}:
		case <-rl.quit:
			return
		default:
			continue
		}
	}
}

// Close MUST be called by client when leaky bucket is not requited. Otherise will leak a goroutine.
func (rl *LeakyBucket) Close() {
	rl.quit <- struct{}{}
}

// Limiter returns a read-only channel signal that will send signals according to a leaky buckety alogrithm.
func (rl *LeakyBucket) Limiter() <-chan struct{} {
	return rl.bucket
}

// New returns an instance of LeakyBucket. The limiter . The constructor takes two argements;
// burst is the initial burst of calls allowed by the limiter, drip is the drip
// rate per sec.
func New(burst int, drip float64) *LeakyBucket {

	rl := &LeakyBucket{
		bucket: make(chan struct{}, burst),
		quit:   make(chan struct{}),
	}

	rl.fill(burst)
	go rl.drip(drip)
	return rl
}
