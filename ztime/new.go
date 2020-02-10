package ztime

import (
	"fmt"
	"strings"
	"time"
)

const timeTpl string = "2006-01-02 15:04:05"

type TimeEngine struct {
	zone *time.Location
}

// Zone eastEightTimeZone
func Zone(zone ...int) *time.Location {
	if len(zone) > 0 {
		return time.FixedZone("zTimeZone", zone[0]*3600)
	}
	return time.Local
}

// FormatTlp FormatTlp
func FormatTlp(tpl string) string {
	tpl = strings.ToLower(tpl)
	tpl = strings.Replace(tpl, "y", "2006", -1)
	tpl = strings.Replace(tpl, "m", "01", -1)
	tpl = strings.Replace(tpl, "d", "02", -1)
	tpl = strings.Replace(tpl, "h", "15", -1)
	tpl = strings.Replace(tpl, "i", "04", -1)
	tpl = strings.Replace(tpl, "s", "05", -1)
	return tpl
}

// New new timeEngine
func New(zone ...int) *TimeEngine {
	e := &TimeEngine{}
	timezone := Zone(zone...)
	if timezone != time.Local {
		e.zone = timezone
	}
	return e
}

// SetTimeZone SetTimeZone
func (e *TimeEngine) SetTimeZone(zone int) *TimeEngine {
	e.zone = Zone(zone)
	return e
}

// GetTimeZone GetTimeZone
func (e *TimeEngine) GetTimeZone() *time.Location {
	if e.zone == nil {
		return time.Local
	}
	return e.zone
}

func (e *TimeEngine) in(t time.Time) time.Time {
	if e.zone == nil {
		return t
	}
	return t.In(e.zone)
}

// FormatTime string format of return time
func (e *TimeEngine) FormatTime(t time.Time, format ...string) string {
	t = e.in(t)
	tpl := timeTpl
	if len(format) > 0 {
		tpl = format[0]
	}
	return t.Format(FormatTlp(tpl))
}

// FormatTimestamp convert UNIX time to time string
func (e *TimeEngine) FormatTimestamp(timestamp int64, format ...string) string {
	return e.FormatTime(time.Unix(timestamp, 0), format...)
}

// Parse Parse
func (e *TimeEngine) Parse(str string) (time.Time, error) {
	return time.ParseInLocation(timeTpl, str, e.GetTimeZone())
}

func (e *TimeEngine) Week(t time.Time) int {
	week := e.in(t).Weekday().String()
	switch week {
	case "Monday":
		return 1
	case "Tuesday":
		return 2
	case "Wednesday":
		return 3
	case "Thursday":
		return 4
	case "Friday":
		return 5
	case "Saturday":
		return 6
	// case "Sunday":
	default:
		return 7
	}
}

// MonthRange gets the start and end UNIX times for the specified year and month
func (e *TimeEngine) MonthRange(year int, month int) (beginTime, endTime int64, err error) {
	var monthStr string
	t := e.in(time.Now())

	if year == 0 {
		year = t.Year()
	}

	if month == 0 {
		month = int(t.Month())
	}
	if month <= 9 {
		monthStr = fmt.Sprintf("0%d", month)
	} else {
		monthStr = fmt.Sprint(month)
	}
	yearStr := fmt.Sprint(year)
	str := yearStr + "-" + monthStr + "-01 00:00:00"
	begin, err := e.Parse(str)
	if err != nil {
		return
	}
	beginTime = begin.Unix()
	month = int(begin.Month())
	day := 30
	if month == 2 {
		day = 28
	} else if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		day = 31
	}

	str = yearStr + "-" + monthStr + "-" + fmt.Sprint(day) + " 23:59:59"
	end, _ := e.Parse(str)
	endTime = end.Unix()
	return
}
