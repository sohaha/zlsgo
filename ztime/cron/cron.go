package cron

import (
	"fmt"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

type (
	// Job represents a scheduled task
	// Contains the cron expression, next execution time, and the function to execute
	Job struct {
		expr     *Expression // Parsed cron expression
		NextTime time.Time   // Next time the job will run
		run      func()      // Function to execute
		mu       sync.Mutex  // Mutex to protect NextTime field during concurrent access
	}

	// JobTable manages multiple scheduled tasks
	// Provides functionality to add, run, and stop tasks
	JobTable struct {
		table        sync.Map // Thread-safe map for storing tasks
		sync.RWMutex          // Read-write mutex to protect the stop field
		stop         bool     // Flag indicating whether the job table has been stopped
	}
)

// New creates and returns a new JobTable instance
// Used for managing scheduled tasks
func New() *JobTable {
	return &JobTable{}
}

// Add adds a new scheduled task to the job table
// cronLine parameter is a standard cron expression, e.g., "0 * * * * *" means run every minute
// fn parameter is the function to execute
// Returns a function to remove the task and a possible error
func (c *JobTable) Add(cronLine string, fn func()) (remove func(), err error) {
	if fn == nil {
		return nil, fmt.Errorf("task function cannot be nil")
	}

	var expr *Expression
	expr, err = Parse(cronLine)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression '%s': %w", cronLine, err)
	}

	key := zstring.UUID()
	c.table.Store(key, &Job{
		expr:     expr,
		run:      fn,
		NextTime: expr.Next(time.Now()),
	})
	remove = func() {
		c.table.Delete(key)
	}
	return remove, nil
}

// ForceRun immediately checks and executes all due tasks.
func (c *JobTable) ForceRun() (nextTime time.Duration) {
	now := time.Now()
	nextTime = 1 * time.Second

	c.table.Range(func(key, value interface{}) bool {
		cronjob, ok := value.(*Job)
		if !ok {
			return true
		}

		cronjob.mu.Lock()
		shouldRun := cronjob.NextTime.Before(now) || cronjob.NextTime.Equal(now)
		cronjob.mu.Unlock()

		if shouldRun {
			go cronjob.run()

			cronjob.mu.Lock()
			cronjob.NextTime = cronjob.expr.Next(now)
			cronjob.mu.Unlock()
		}

		cronjob.mu.Lock()
		next := time.Duration(cronjob.NextTime.UnixNano() - now.UnixNano())
		cronjob.mu.Unlock()

		if next > 0 && nextTime > next {
			nextTime = next
		}
		return true
	})
	return nextTime
}

// Run starts the task scheduler, beginning to execute tasks according to their schedule.
func (c *JobTable) Run(block ...bool) {
	run := func() {
		for {
			c.RLock()
			stop := c.stop
			c.RUnlock()
			if stop {
				break
			}
			<-time.After(c.ForceRun())
		}
	}
	if len(block) > 0 && block[0] {
		run()
		return
	}

	go run()
}

// Stop stops the task scheduler.
func (c *JobTable) Stop() {
	c.Lock()
	c.stop = true
	c.Unlock()
}
