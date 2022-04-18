// Package ztime provides time related operations
package ztime

import (
	"time"
)

var newZtime = New()

func Now(format ...string) string {
	return newZtime.FormatTime(time.Now(), format...)
}

// SetTimeZone SetTimeZone
func SetTimeZone(zone int) *TimeEngine {
	return newZtime.SetTimeZone(zone)
}

// GetTimeZone getTimeZone
func GetTimeZone() *time.Location {
	return newZtime.GetTimeZone()
}

// FormatTime format time
func FormatTime(t time.Time, format ...string) string {
	return newZtime.FormatTime(t, format...)
}

// FormatTimestamp format timestamp
func FormatTimestamp(timestamp int64, format ...string) string {
	return newZtime.FormatTimestamp(timestamp, format...)
}

func Week(t time.Time) int {
	return newZtime.Week(t)
}

func MonthRange(year int, month int) (beginTime, endTime int64, err error) {
	return newZtime.MonthRange(year, month)
}

// Parse string to time
func Parse(str string, format ...string) (time.Time, error) {
	return newZtime.Parse(str, format...)
}

// Unix int to time
func Unix(tt int64) time.Time {
	return newZtime.Unix(tt)
}
