// Package cron emulate linux crontab
package cron

import (
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

type (
	// Job cron job
	Job struct {
		expr     *Expression
		NextTime time.Time
		run      func()
	}
	JobTable struct {
		table sync.Map
		sync.RWMutex
		stop bool
	}
)

func New() *JobTable {
	return &JobTable{}
}

func (c *JobTable) Add(cronLine string, fn func()) (remove func(), err error) {
	var expr *Expression
	expr, err = Parse(cronLine)
	if err != nil {
		return
	}
	task := &Job{
		expr:     expr,
		run:      fn,
		NextTime: expr.Next(time.Now()),
	}
	key := zstring.UUID()
	c.table.Store(key, task)
	remove = func() {
		c.table.Delete(key)
	}
	return
}

func (c *JobTable) ForceRun() (nextTime time.Duration) {
	now := time.Now()
	nextTime = 1 * time.Second
	// todo there is a sequence problem
	// todo later optimization directly obtains the next execution time
	c.table.Range(func(key, value interface{}) bool {
		cronjob, ok := value.(*Job)
		if ok {
			if cronjob.NextTime.Before(now) || cronjob.NextTime.Equal(now) {
				go cronjob.run()
				cronjob.NextTime = cronjob.expr.Next(now)
			}
		}
		next := time.Duration(cronjob.NextTime.UnixNano() - now.UnixNano())
		if nextTime > next {
			nextTime = next
		}
		return true
	})
	return nextTime
}

func (c *JobTable) Run() {
	go func() {
		t := time.NewTimer(time.Second)
		for {
			c.RLock()
			stop := c.stop
			c.RUnlock()
			if stop {
				break
			}
			NextTime := c.ForceRun()
			t.Reset(NextTime)
			<-t.C
		}
	}()
}

func (c *JobTable) Stop() {
	c.Lock()
	c.stop = true
	c.Unlock()
}
