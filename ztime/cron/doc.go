/*
Package cron provides functionality similar to Linux crontab for scheduling tasks.

This package allows users to schedule and manage tasks using standard cron expression syntax,
with support for second-level precision. Key features include:

  - Standard cron expression syntax support
  - Second-level precision for task scheduling
  - Support for adding, running, and stopping tasks
  - Thread-safe task management

Basic usage example:

	// Create a new job table
	crontab := cron.New()
	
	// Add a task that runs every minute
	removeFunc, err := crontab.Add("0 * * * * *", func() {
		// Task code goes here
		fmt.Println("Running every minute")
	})
	if err != nil {
		// Handle error
	}
	
	// Start the cron scheduler (non-blocking mode)
	crontab.Run()
	
	// For blocking mode
	// crontab.Run(true)
	
	// Remove a specific task
	removeFunc()
	
	// Stop all tasks
	crontab.Stop()

Cron expression format:

	second minute hour day month weekday

Field descriptions:
  - second: 0-59
  - minute: 0-59
  - hour: 0-23
  - day: 1-31
  - month: 1-12
  - weekday: 0-6 (0 represents Sunday)

Supported special characters:
  - *: represents all possible values
  - ,: used to separate multiple values
  - -: represents a range
  - /: represents an increment
  - L: used in the day field to represent the last day of the month, or in the weekday field to represent the last day of the week
  - W: used in the day field to represent the nearest weekday

Examples:
  - "0 0 12 * * *": Run at 12:00 PM every day
  - "0 15 10 * * *": Run at 10:15 AM every day
  - "0 0/5 * * * *": Run every 5 minutes
  - "0 0 12 1 * *": Run at 12:00 PM on the first day of every month
*/
package cron
