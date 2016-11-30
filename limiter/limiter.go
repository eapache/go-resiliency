package limiter

import "time"

type leakyBucket chan struct{}

func (rl leakyBucket) fill(burst int) {
	for i := 0; i < burst; i++ {
		rl <- struct{}{}
	}
}

func (rl leakyBucket) drip(drip float64) {
	driptime := float64(1000000000) / drip
	for {
		time.Sleep(time.Duration(driptime) * time.Nanosecond)
		// close
		if rl == nil {
			return
		}
		rl <- struct{}{}
	}
}

// New returns a read-only signal channel that will send signals accoding to a
// leaky bucket implementation. The limiter . The constructor takes two argements;
// burst is the initial burst of calls allowed by the limiter, drip is the drip
// rate per sec. IMPORTANT! must set the returned channel to nil when finished with it,
// as this stops dripping background process.
func New(burst int, drip float64) <-chan struct{} {

	rl := leakyBucket(make(chan struct{}, burst))

	rl.fill(burst)
	go rl.drip(drip)
	return rl
}
