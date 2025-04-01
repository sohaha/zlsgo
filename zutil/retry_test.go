package zutil

import (
	"errors"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestRetry(tt *testing.T) {
	t := zlsgo.NewTest(tt)

	t.Run("Success", func(t *zlsgo.TestUtil) {
		i := 0
		now := time.Now()
		err := DoRetry(5, func() error {
			if i < 3 {
				i++
				return errors.New("error")
			}
			return nil
		}, func(rc *RetryConf) {
			rc.Interval = time.Second / 5
		})
		t.NoError(err)
		t.EqualTrue(time.Since(now).Seconds() < 1)
		t.Equal(3, i)
	})

	t.Run("Success BackOffDelay", func(t *zlsgo.TestUtil) {
		i := 0
		now := time.Now()
		err := DoRetry(5, func() error {
			t.Log(i, time.Since(now).Seconds())
			if i < 3 {
				i++
				return errors.New("error")
			}
			return nil
		}, func(rc *RetryConf) {
			rc.BackOffDelay = true
			rc.Interval = time.Second / 5
		})
		t.NoError(err)
		t.EqualTrue(time.Since(now).Seconds() < 3)
		t.EqualTrue(time.Since(now).Seconds() > 1.5)
		t.Equal(3, i)
	})

	t.Run("Failed", func(t *zlsgo.TestUtil) {
		i := 0
		now := time.Now()
		err := DoRetry(5, func() error {
			i++
			return errors.New("error")
		}, func(rc *RetryConf) {
			rc.Interval = time.Second / 5
		})
		t.EqualTrue(err != nil)
		t.EqualTrue(time.Since(now).Seconds() > 1)
		t.Equal(6, i)
	})
}

func Test_backOffDelay(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(BackOffDelay(i, time.Second, time.Minute))
	}
}
