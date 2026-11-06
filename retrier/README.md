retrier
=======

[![Golang CI](https://github.com/eapache/go-resiliency/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/eapache/go-resiliency/actions/workflows/golang-ci.yml)
[![GoDoc](https://godoc.org/github.com/eapache/go-resiliency/retrier?status.svg)](https://godoc.org/github.com/eapache/go-resiliency/retrier)
[![Code of Conduct](https://img.shields.io/badge/code%20of%20conduct-active-blue.svg)](https://eapache.github.io/conduct.html)

The retriable resiliency pattern for golang.

Creating a retrier takes two parameters:
- the times to back-off between retries (and implicitly the number of times to
  retry)
- the classifier that determines which errors to retry

```go
r := retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), nil)

err := r.Run(func() error {
	// do some work
	return nil
})

if err != nil {
	// handle the case where the work failed three times
}
```

A retried error can override the delay by implementing `Backoff() time.Duration`.
The retrier also looks through wrapped errors. For an ordinary error, use
`retrier.ErrWithBackoff(err, delay)`; it preserves the original error for
`errors.Is`, `errors.As`, and the built-in classifiers, and returns nil for nil.

Error-provided delays are used without jitter. Zero or negative delays retry
immediately. Errors without a backoff use the configured schedule and jitter.
The classifier and the number of allowed retries still apply.
