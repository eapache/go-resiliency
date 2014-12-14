// Package retrier implements the "retriable" resiliency pattern for Go.
package retrier

import "time"

// Retrier implements the "retriable" resiliency pattern, abstracting out the process of retrying a failed action
// a certain number of times with an optional back-off between each retry.
type Retrier struct {
	backoff []time.Duration
	class   Classifier
}

// New constructs a Retrier with the given backoff pattern and classifier. The length of the backoff pattern
// indicates how many times an action will be retried, and the value at each index indicates the amount of time
// waited before each subsequent retry. The classifier is used to determine which errors should be retried and
// which should cause the retrier to fail fast. The DefaultClassifier is used if nil is passed.
func New(backoff []time.Duration, class Classifier) *Retrier {
	if class == nil {
		class = DefaultClassifier{}
	}

	return &Retrier{
		backoff: backoff,
		class:   class,
	}
}

// Run executes the given work function, then classifies its return value based on the classifier used
// to construct the Retrier. If the result is Succeed or Fail, the return value of the work function is
// returned to the caller. If the result is Retry, then Run sleeps according to the its backoff policy
// before retrying. If the total number of retries is exceeded then the return value of the work function
// is returned to the caller regardless.
func (r *Retrier) Run(work func() error) error {
	retries := 0
	for {
		ret := work()

		switch r.class.Classify(ret) {
		case Succeed, Fail:
			return ret
		case Retry:
			if retries >= len(r.backoff) {
				return ret
			}
			time.Sleep(r.backoff[retries])
			retries++
		}
	}
}
