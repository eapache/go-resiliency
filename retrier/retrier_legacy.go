// +build !go1.7

package retrier

import "time"

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
			time.Sleep(r.calcSleep(retries))
			retries++
		}
	}
}