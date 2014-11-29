package breaker

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var BreakerOpen = errors.New("circuit breaker is open")

type Panic struct {
	Value interface{}
}

func (p Panic) Error() string {
	return fmt.Sprint("panic:", p.Value)
}

type state int

const (
	closed state = iota
	open
	halfOpen
)

type Breaker struct {
	errorThreshold, successThreshold int
	timeout                          time.Duration

	lock              sync.RWMutex
	state             state
	errors, successes int
}

func New(errorThreshold, successThreshold int, timeout time.Duration) *Breaker {
	return &Breaker{
		errorThreshold:   errorThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

func (b *Breaker) Run(x func() error) error {
	b.lock.RLock()
	state := b.state
	b.lock.RUnlock()

	if state == open {
		return BreakerOpen
	}

	result := func() (err error) {
		defer func() {
			if val := recover(); val != nil {
				err = Panic{Value: val}
			}
		}()
		return x()
	}()

	b.lock.Lock()
	defer b.lock.Unlock()

	if result == nil {
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
