// Package ztime provides time related operations
package ztime

import (
	"time"
)

var newZtime = New()

// SetTimeZone SetTimeZone
func SetTimeZone(zone int) *TimeEngine {
	return newZtime.SetTimeZone(zone)
}

// GetTimeZone getTimeZone
func GetTimeZone() *time.Location {
	return newZtime.GetTimeZone()
}

func FormatTime(t time.Time, format ...string) string {
	return newZtime.FormatTime(t, format...)
}

func FormatTimestamp(timestamp int64, format ...string) string {
	return newZtime.FormatTimestamp(timestamp, format...)
}

func Week(t time.Time) int {
	return newZtime.Week(t)
}

func MonthRange(year int, month int) (beginTime, endTime int64, err error) {
	return newZtime.MonthRange(year, month)
}

func Parse(str string) (time.Time, error) {
	return newZtime.Parse(str)
}
