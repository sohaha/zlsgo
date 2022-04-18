package ztime

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

const (
	timePattern = `(\d{4}[-/\.]\d{1,2}[-/\.]\d{1,2})[:\sT-]*(\d{0,2}:{0,1}\d{0,2}:{0,1}\d{0,2}){0,1}\.{0,1}(\d{0,9})([\sZ]{0,1})([\+-]{0,1})([:\d]*)`
)

var (
	TimeTpl      = "2006-01-02 15:04:05"
	formatKeyTpl = map[byte]string{
		'd': "02",
		'D': "Mon",
		'w': "Monday",
		'N': "Monday",
		'S': "02",
		'l': "Monday",
		'F': "January",
		'm': "01",
		'M': "Jan",
		'n': "1",
		'Y': "2006",
		'y': "06",
		'a': "pm",
		'A': "PM",
		'g': "3",
		'h': "03",
		'H': "15",
		'i': "04",
		's': "05",
		'O': "-0700",
		'P': "-07:00",
		'T': "MST",
		'c': "2006-01-02T15:04:05-07:00",
		'r': "Mon, 02 Jan 06 15:04 MST",
	}
	GetLocationName = func(zone int) string {
		switch zone {
		case 8:
			return "Asia/Shanghai"
		}
		return "UTC"
	}
)

type TimeEngine struct {
	zone *time.Location
}

// Zone eastEightTimeZone
func Zone(zone ...int) *time.Location {
	if len(zone) > 0 {
		return time.FixedZone(GetLocationName(zone[0]), zone[0]*3600)
	}
	return time.Local
}

// FormatTlp format template
func FormatTlp(format string) string {
	runes := []rune(format)
	buffer := bytes.NewBuffer(nil)
	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case '\\':
			if i < len(runes)-1 {
				buffer.WriteRune(runes[i+1])
				i += 1
				continue
			} else {
				return buffer.String()
			}
		default:
			if runes[i] > 255 {
				buffer.WriteRune(runes[i])
				break
			}
			if f, ok := formatKeyTpl[byte(runes[i])]; ok {
				buffer.WriteString(f)
			} else {
				buffer.WriteRune(runes[i])
			}
		}
	}
	return buffer.String()
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
	tpl := TimeTpl
	if len(format) > 0 {
		tpl = FormatTlp(format[0])
	}
	return t.Format(tpl)
}

// FormatTimestamp convert UNIX time to time string
func (e *TimeEngine) FormatTimestamp(timestamp int64, format ...string) string {
	return e.FormatTime(time.Unix(timestamp, 0), format...)
}

// Unix int to time
func (e *TimeEngine) Unix(tt int64) time.Time {
	return e.in(time.Unix(tt, 0))
}

// Parse Parse
func (e *TimeEngine) Parse(str string, format ...string) (t time.Time, err error) {
	if len(format) > 0 {
		return time.ParseInLocation(FormatTlp(format[0]), str, e.GetTimeZone())
	}
	var year, month, day, hour, min, sec string
	match, err := zstring.RegexExtract(timePattern, str)
	if err != nil {
		return
	}
	matchLen := len(match)
	if matchLen == 0 {
		err = errors.New("cannot parse")
		return
	}
	if matchLen > 1 && match[1] != "" {
		for k, v := range match {
			match[k] = strings.TrimSpace(v)
		}
		arr := make([]string, 3)
		for _, v := range []string{"-", "/", "."} {
			arr = strings.Split(match[1], v)
			if len(arr) >= 3 {
				break
			}
		}
		if len(arr) < 3 {
			err = errors.New("cannot parse date")
			return
		}
		year = arr[0]
		month = zstring.Pad(arr[1], 2, "0", zstring.PadLeft)
		day = zstring.Pad(arr[2], 2, "0", zstring.PadLeft)
	}
	if len(match[2]) > 0 {
		s := strings.Replace(match[2], ":", "", -1)
		if len(s) < 6 {
			s += strings.Repeat("0", 6-len(s))
		}
		hour = zstring.Pad(s[0:2], 2, "0", zstring.PadLeft)
		min = zstring.Pad(s[2:4], 2, "0", zstring.PadLeft)
		sec = zstring.Pad(s[4:6], 2, "0", zstring.PadLeft)
	}
	return time.ParseInLocation(TimeTpl, fmt.Sprintf("%s-%s-%s %s:%s:%s", year, month, day, hour, min, sec), e.GetTimeZone())

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
