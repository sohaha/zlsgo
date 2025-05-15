package zutil

import (
	"errors"
	"math/rand"
	"time"
)

// RetryConf holds configuration options for the retry mechanism.
// It controls the number of retries, intervals between attempts, timeout,
// and backoff strategy.
type RetryConf struct {
	// maxRetry is the maximum number of retries
	maxRetry int
	// Interval is the base interval between retry attempts
	Interval time.Duration
	// MaxRetryInterval is the maximum interval between retry attempts
	// when using exponential backoff
	MaxRetryInterval time.Duration
	// Timeout is the maximum total duration for all retry attempts
	Timeout time.Duration
	// BackOffDelay determines whether to use exponential backoff
	// for increasing the interval between retries
	BackOffDelay bool
}

// DoRetry executes a function with retry logic based on the provided configuration.
// It will retry the function up to 'sum' times or until it succeeds.
// Additional options can be provided to customize retry behavior.
func DoRetry(sum int, fn func() error, opt ...func(*RetryConf)) (err error) {
	o := RetryConf{
		maxRetry:         sum,
		Interval:         time.Second,
		MaxRetryInterval: time.Minute,
	}
	for i := range opt {
		opt[i](&o)
	}

	err = fn()
	if err == nil {
		return
	}

	if o.maxRetry == 0 {
		return errors.New("maxRetry must be greater than 0")
	}

	i, now := 1, time.Now()
	for ; ; i++ {

		var interval time.Duration
		if o.BackOffDelay {
			interval = BackOffDelay(i, o.Interval, o.MaxRetryInterval)
		} else {
			interval = o.Interval
		}

		time.Sleep(interval)

		if o.maxRetry > 0 && i > o.maxRetry {
			break
		}

		if o.Timeout > 0 && time.Since(now) > o.Timeout {
			break
		}

		err = fn()
		if err == nil {
			break
		}
	}

	return
}

// BackOffDelay calculates the delay duration for exponential backoff retry strategy.
// It increases the delay exponentially based on the attempt number and adds jitter
// to prevent synchronized retries in distributed systems.
func BackOffDelay(attempt int, retryInterval, maxRetryInterval time.Duration) time.Duration {
	attempt = attempt - 1
	if attempt < 0 {
		return 0
	}

	retryFactor := 1 << uint(attempt)
	jitter := rand.Float64()
	waitDuration := time.Duration(retryFactor) * retryInterval
	waitDuration = waitDuration + time.Duration(jitter*float64(waitDuration))

	if waitDuration > maxRetryInterval {
		return maxRetryInterval
	}

	return waitDuration
}
