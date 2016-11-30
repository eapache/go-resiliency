package limiter

import "time"

type LeakyBucket chan struct{}

func (rl LeakyBucket) fill(burst int) {
	for i := 0; i < burst; i++ {
		rl <- struct{}{}
	}
}

func (rl LeakyBucket) drip(drip float64) {
	driptime := float64(1000000000) / drip
	for {
		time.Sleep(time.Duration(driptime) * time.Nanosecond)
		rl <- struct{}{}
	}
}

func New(burst int, drip float64) LeakyBucket {
	rl := LeakyBucket(make(chan struct{}, burst))
	rl.fill(burst)
	go rl.drip(drip)
	return rl
}
