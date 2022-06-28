go-resiliency
=============

[![Golang CI](https://github.com/eapache/go-resiliency/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/eapache/go-resiliency/actions/workflows/golang-ci.yml)
[![GoDoc](https://godoc.org/github.com/eapache/go-resiliency?status.svg)](https://godoc.org/github.com/eapache/go-resiliency)
[![Code of Conduct](https://img.shields.io/badge/code%20of%20conduct-active-blue.svg)](https://eapache.github.io/conduct.html)

Resiliency patterns for golang.
Based in part on [Hystrix](https://github.com/Netflix/Hystrix),
[Semian](https://github.com/Shopify/semian), and others.

Currently implemented patterns include:
- circuit-breaker (in the `breaker` directory)
- semaphore (in the `semaphore` directory)
- deadline/timeout (in the `deadline` directory)
- batching (in the `batcher` directory)
- retriable (in the `retrier` directory)

*Note: I will occasionally bump the minimum required Golang version without
bumping the major version of this package, which violates the official Golang
packaging convention around breaking changes. Typically the versions being
dropped are multiple years old and long unsupported.*
