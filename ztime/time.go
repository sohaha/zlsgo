// Package ztime provides time related operations
package ztime

import (
	"sync/atomic"
	"time"
)

var inlay = New()

// Now format current time
func Now(format ...string) string {
	return inlay.FormatTime(UnixMicro(Clock()), format...)
}

// Time With the time zone of the time
func Time() time.Time {
	return inlay.In(UnixMicro(Clock()))
}

// SetTimeZone SetTimeZone
func SetTimeZone(zone int) *TimeEngine {
	return inlay.SetTimeZone(zone)
}

// GetTimeZone getTimeZone
func GetTimeZone() *time.Location {
	return inlay.GetTimeZone()
}

// FormatTime format time
func FormatTime(t time.Time, format ...string) string {
	return inlay.FormatTime(t, format...)
}

// FormatTimestamp format timestamp
func FormatTimestamp(timestamp int64, format ...string) string {
	return inlay.FormatTimestamp(timestamp, format...)
}

func Week(t time.Time) int {
	return inlay.Week(t)
}

func MonthRange(year int, month int) (beginTime, endTime int64, err error) {
	return inlay.MonthRange(year, month)
}

// Parse string to time
func Parse(str string, format ...string) (time.Time, error) {
	return inlay.Parse(str, format...)
}

// Unix int to time
func Unix(tt int64) time.Time {
	return inlay.Unix(tt)
}

// UnixMicro int to time
func UnixMicro(tt int64) time.Time {
	return inlay.UnixMicro(tt)
}

// In time to time
func In(tt time.Time) time.Time {
	return inlay.In(tt)
}

var clock = time.Now().UnixNano() / 1000

func init() {
	go func() {
		m := 90 * time.Millisecond
		t := int64(m)
		for {
			atomic.StoreInt64(&clock, time.Now().UnixNano()/1000)
			for i := 0; i < 10; i++ {
				time.Sleep(m)
				atomic.AddInt64(&clock, t)
			}
			time.Sleep(m)
		}
	}()
}

// Clock The current microsecond timestamp has an accuracy of 100ms.
func Clock() int64 {
	return atomic.LoadInt64(&clock)
}
