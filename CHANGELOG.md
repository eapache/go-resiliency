# Changelog

*Note: I will occasionally bump the minimum required Golang version without
bumping the major version of this package, which violates the official Golang
packaging convention around breaking changes. Typically the versions being
dropped are multiple years old and long unsupported.*

#### Version 1.4.0 (2023-08-14)

 - Adds `Batcher.Shutdown()` to flush any pending work without waiting for the
   timer, e.g. on application shutdown (thanks to Ivan Stankov).
 - Fix possible memory leaks of Timer objects in Deadline, Retrier, and
   Semaphore (thanks to Dmytro Nozdrin).

#### Version 1.3.0 (2022-06-27)

 - Increased minimum Golang version to 1.13.
 - Fix a goroutine leak in `Deadline.Run()` on `ErrTimeOut`.
 - Add a `go.mod` file to conform to more recent Golang version standards.
 - Use `errors.Is` when classifying errors for the `Retrier` (thanks to Taufik
   Rama).
 - Add implementation of `LimitedExponentialBackoff` for the `Retrier` (thanks
   to tukeJonny).

#### Version 1.2.0 (2019-06-14)

 - Increased minimum Golang version to 1.7.
 - Add `RunCtx` method on `Retrier` to support running with a context.
 - Ensure the `Retrier`'s use of random numbers is concurrency-safe.
 - Bump CI to ensure we support newer Golang versions.

#### Version 1.1.0 (2018-03-26)

 - Improve documentation and fix some typos.
 - Bump CI to ensure we support newer Golang versions.
 - Add `IsEmpty()` method on `Semaphore`.

#### Version 1.0.0 (2015-02-13)

Initial release.
