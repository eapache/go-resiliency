go-resiliency
=============

[![Build Status](https://travis-ci.org/eapache/go-resiliency.svg?branch=master)](https://travis-ci.org/eapache/go-resiliency)
[![GoDoc](https://godoc.org/github.com/eapache/go-resiliency?status.svg)](https://godoc.org/github.com/eapache/go-resiliency)

Resiliency patterns for golang. Currently implemented are:
- circuit-breaker pattern (in the `breaker` directory)
- semaphore pattern (in the `semaphore` directory)
- deadline/timeout pattern (in the `deadline` directory)
- batching pattern (in the `batcher` directory)
- retriable pattern (in the `retrier` directory)
