package zutil

import (
	"fmt"
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
func DoRetry(sum int, fn func() bool, opt ...func(*RetryConf)) (ok bool) {
	o := RetryConf{
		maxRetry:         sum,
		Interval:         time.Second,
		MaxRetryInterval: time.Minute,
	}
	for i := range opt {
		opt[i](&o)
	}

	ok = fn()
	if ok {
		return
	}

	if o.maxRetry == 0 {
		return false
	}

	i, now := 1, time.Now()
	for ; ; i++ {
		if o.maxRetry > 0 && i > o.maxRetry {
			break
		}

		if o.Timeout > 0 && time.Since(now) > o.Timeout {
			break
		}

		ok = fn()
		if ok {
			break
		}

		var interval time.Duration
		if o.BackOffDelay {
			interval = backOffDelay(i, o.MaxRetryInterval)
			fmt.Println(i, interval)
		} else {
			interval = o.Interval
		}

		time.Sleep(interval)
	}

	return
}

func backOffDelay(attempt int, maxRetryInterval time.Duration) time.Duration {
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
