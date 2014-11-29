package breaker

import (
	"errors"
	"sync"
	"time"
)

// BreakerOpen is the error returned from Run() when the function is not executed
// because the breaker is currently open.
var BreakerOpen = errors.New("circuit breaker is open")

type state int

const (
	closed state = iota
	open
	halfOpen
)

// Breaker implements the circuit-breaker resiliency pattern
type Breaker struct {
	errorThreshold, successThreshold int
	timeout                          time.Duration

	lock              sync.RWMutex
	state             state
	errors, successes int
}

// New constructs a new circuit-breaker.
func New(errorThreshold, successThreshold int, timeout time.Duration) *Breaker {
	return &Breaker{
		errorThreshold:   errorThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Run will either return BreakerOpen immediately if the circuit-breaker is
// already open, or it will run the given function and pass along its return
// value.
func (b *Breaker) Run(x func() error) error {
	b.lock.RLock()
	state := b.state
	b.lock.RUnlock()

	if state == open {
		return BreakerOpen
	}

	var panicValue interface{}

	result := func() error {
		defer func() {
			panicValue = recover()
		}()
		return x()
	}()

	b.lock.Lock()
	defer b.lock.Unlock()

	if result == nil && panicValue == nil {
		if b.state == halfOpen {
			b.successes++
			if b.successes == b.successThreshold {
				b.closeBreaker()
			}
		}
	} else {
		switch b.state {
		case closed:
			b.errors++
			if b.errors == b.errorThreshold {
				b.openBreaker()
			}
		case halfOpen:
			b.openBreaker()
		}
	}

	if panicValue != nil {
		// as close as Go lets us come to a "rethrow" although unfortunately
		// we lose the original panicing location
		panic(panicValue)
	}

	return result
}

func (b *Breaker) openBreaker() {
	b.changeState(open)
	go b.timer()
}

func (b *Breaker) closeBreaker() {
	b.changeState(closed)
}

func (b *Breaker) timer() {
	time.Sleep(b.timeout)

	b.lock.Lock()
	defer b.lock.Unlock()

	b.changeState(halfOpen)
}

func (b *Breaker) changeState(newState state) {
	b.errors = 0
	b.successes = 0
	b.state = newState
}
