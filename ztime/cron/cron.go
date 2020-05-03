// Package cron emulate linux crontab
package cron

import (
	"fmt"
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
		expr: expr,
		run:  fn,
	}
	now := time.Now().UnixNano()
	key := fmt.Sprintf("__z__cronJob__%d__%d", now, zstring.RandInt(100, 999))
	c.table.Store(key, task)
	remove = func() {
		c.table.Delete(key)
	}
	return
}

func (c *JobTable) Run() {
	go func() {
		for {
			now := time.Now()
			// todo there is a sequence problem
			// todo later optimization directly obtains the next execution time
			c.table.Range(func(key, value interface{}) bool {
				cronjob, ok := value.(*Job)
				if ok {
					if cronjob.NextTime.Before(now) || cronjob.NextTime.Equal(now) {
						go cronjob.run()
					}
					cronjob.NextTime = cronjob.expr.Next(now)
				}

				return true
			})
			<-time.NewTimer(200 * time.Millisecond).C
		}
	}()
}
