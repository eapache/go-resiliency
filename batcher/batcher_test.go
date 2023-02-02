package batcher

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var errSomeError = errors.New("errSomeError")

func returnsError(params []interface{}) error {
	return errSomeError
}

func returnsSuccess(params []interface{}) error {
	return nil
}

func TestBatcherSuccess(t *testing.T) {
	b := New(10*time.Millisecond, returnsSuccess)

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			if err := b.Run(nil); err != nil {
				t.Error(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	b = New(0, returnsSuccess)
	for i := 0; i < 10; i++ {
		if err := b.Run(nil); err != nil {
			t.Error(err)
		}
	}
}

func TestShutdownSuccess(t *testing.T) {
	sleepDuration := 5 * time.Millisecond
	durationLimit := 2 * sleepDuration
	timeout := 2 * durationLimit
	total := 0
	doSum := func(params []interface{}) error {
		for _, param := range params {
			intValue, ok := param.(int)
			if !ok {
				t.Error("expected type int")
			}
			total += intValue
		}
		return nil
	}

	b := New(timeout, doSum)
	go func() {
		time.Sleep(sleepDuration)
		b.Shutdown(true)
	}()

	wg := &sync.WaitGroup{}
	expectedTotal := 0
	start := time.Now()
	for i := 0; i < 10; i++ {
		expectedTotal += i
		wg.Add(1)
		go func(i int) {
			if err := b.Run(i); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	duration := time.Since(start)
	if duration >= durationLimit {
		t.Errorf("expected duration[%v] < durationLimit[%v]", duration, durationLimit)
	}

	if total != expectedTotal {
		t.Errorf("expected processed count[%v] < actual[%v]", expectedTotal, total)
	}
}

func TestShutdownEmpty(t *testing.T) {
	b := New(10*time.Millisecond, returnsSuccess)
	b.Shutdown(true)
}

func TestBatcherError(t *testing.T) {
	b := New(10*time.Millisecond, returnsError)

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			if err := b.Run(nil); err != errSomeError {
				t.Error(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestBatcherPrefilter(t *testing.T) {
	b := New(1*time.Millisecond, returnsSuccess)

	b.Prefilter(func(param interface{}) error {
		if param == nil {
			return errSomeError
		}
		return nil
	})

	if err := b.Run(nil); err != errSomeError {
		t.Error(err)
	}

	if err := b.Run(1); err != nil {
		t.Error(err)
	}
}

func TestBatcherMultipleBatches(t *testing.T) {
	var iters uint32

	b := New(10*time.Millisecond, func(params []interface{}) error {
		atomic.AddUint32(&iters, 1)
		return nil
	})

	wg := &sync.WaitGroup{}

	for group := 0; group < 5; group++ {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				if err := b.Run(nil); err != nil {
					t.Error(err)
				}
				wg.Done()
			}()
		}
		time.Sleep(15 * time.Millisecond)
	}

	wg.Wait()

	if iters != 5 {
		t.Error("Wrong number of iters:", iters)
	}
}

func ExampleBatcher() {
	b := New(10*time.Millisecond, func(params []interface{}) error {
		// do something with the batch of parameters
		return nil
	})

	b.Prefilter(func(param interface{}) error {
		// do some sort of sanity check on the parameter, and return an error if it fails
		return nil
	})

	for i := 0; i < 10; i++ {
		go b.Run(i)
	}
}
