package cron

import (
	"fmt"
	"github.com/sohaha/zlsgo/zstring"
	"sync"
	"time"
)

type (
	// Job cron job
	Job struct {
		expr     *Expression
		NextTime time.Time
		run      func()
	}
	CronJobTable struct {
		table sync.Map
	}
)

func New() *CronJobTable {
	return &CronJobTable{}
}

func (c *CronJobTable) Add(cronLine string, fn func()) (remove func(), err error) {
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

func (c *CronJobTable) Run() {
	go func() {
		for {
			now := time.Now()
			c.table.Range(func(key, value interface{}) bool {
				cronjob, ok := value.(*Job)
				if ok {
					if cronjob.NextTime.Before(now) || cronjob.NextTime.Equal(now) {
						cronjob.run()
					}
					cronjob.NextTime = cronjob.expr.Next(now)
				}

				return true
			})
			select {
			case <-time.NewTimer(200 * time.Millisecond).C:
			}
		}
	}()
}
