package zutil

import (
	"errors"
	"math/rand"
	"time"
)

type RetryConf struct {
	// maxRetry is the maximum number of retries
	maxRetry int
	// Interval is the interval between retries
	Interval time.Duration
	// MaxRetryInterval is the maximum interval between retries
	MaxRetryInterval time.Duration
	// Timeout is the timeout of the entire retry
	Timeout time.Duration
	// BackOffDelay is whether to increase the interval between retries
	BackOffDelay bool
}

// DoRetry is a general retry function
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

		var interval time.Duration
		if o.BackOffDelay {
			interval = BackOffDelay(i, o.MaxRetryInterval)
		} else {
			interval = o.Interval
		}

		time.Sleep(interval)
	}

	return
}

func BackOffDelay(attempt int, maxRetryInterval time.Duration) time.Duration {
	attempt = attempt - 1
	if attempt < 0 {
		return 0
	}

	retryFactor := 1 << uint(attempt)
	jitter := rand.Float64()
	waitDuration := time.Duration(retryFactor) * time.Second
	waitDuration = waitDuration + time.Duration(jitter*float64(waitDuration))

	if waitDuration > maxRetryInterval {
		return maxRetryInterval
	}

	return waitDuration
}
