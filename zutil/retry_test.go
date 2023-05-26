package zutil

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestRetry(tt *testing.T) {
	t := zlsgo.NewTest(tt)

	t.Run("Success", func(t *zlsgo.TestUtil) {
		i := 0
		now := time.Now()
		ok := DoRetry(5, func() bool {
			if i < 3 {
				i++
				return false
			}
			return true
		}, func(rc *RetryConf) {
			rc.Interval = time.Second / 5
		})
		t.EqualTrue(ok)
		t.EqualTrue(time.Since(now).Seconds() < 1)
		t.Equal(3, i)
	})

	t.Run("Success BackOffDelay", func(t *zlsgo.TestUtil) {
		i := 0
		now := time.Now()
		ok := DoRetry(5, func() bool {
			if i < 3 {
				i++
				return false
			}
			return true
		}, func(rc *RetryConf) {
			rc.BackOffDelay = true
			rc.Interval = time.Second / 5
		})
		t.EqualTrue(ok)
		t.EqualTrue(time.Since(now).Seconds() < 6)
		t.EqualTrue(time.Since(now).Seconds() > 2)
		t.Equal(3, i)
	})

	t.Run("Failed", func(t *zlsgo.TestUtil) {
		i := 0
		now := time.Now()
		ok := DoRetry(5, func() bool {
			i++
			return false
		}, func(rc *RetryConf) {
			rc.Interval = time.Second / 5
		})
		t.EqualTrue(!ok)
		t.EqualTrue(time.Since(now).Seconds() > 1)
		t.Equal(6, i)
	})

}

func Test_backOffDelay(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(BackOffDelay((i), time.Minute))
	}
}
