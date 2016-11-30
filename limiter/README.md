limiter
========

[![Build Status](https://travis-ci.org/eapache/go-resiliency.svg?branch=master)](https://travis-ci.org/eapache/go-resiliency)
[![GoDoc](https://godoc.org/github.com/eapache/go-resiliency/limiter?status.svg)](https://godoc.org/github.com/limiter/go-resiliency/deadline)
[![Code of Conduct](https://img.shields.io/badge/code%20of%20conduct-active-blue.svg)](https://eapache.github.io/conduct.html)

The leaky-bucket rate-limiter resiliency pattern for golang.

Creating a leaky bucket takes two parameters:
- initial bust amount allowed by limiter
- drips/sec rate after burst is used up

```go
// returned limiter is safe to use in multiple go routines
leaky := limiter.New(100, 1) // allows for 100 initial calls, then subsequent calls of 1 per second


// goroutines of some worker process
go func() {
  for {
    <-leaky.Limiter()
    // do something that needs to be rate limited
  }
}()
}
```
