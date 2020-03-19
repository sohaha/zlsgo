package ztime

import (
	"github.com/sohaha/zlsgo"
	"testing"
	"time"
)

func TestNewTime(T *testing.T) {
	t := zlsgo.NewTest(T)
	now := int64(1580809099)
	nowTime := time.Unix(now, 0)

	t.Equal(time.Local, GetTimeZone())
	SetTimeZone(0)
	t.EqualExit("2020-02-04 09:38:19", FormatTimestamp(now, "Y-m-d H:i:s"))
	t.Log(New(24).FormatTimestamp(now, "Y-m-d H:i:s"), FormatTimestamp(now, "Y-m-d H:i:s"))
	t.Log(New(24).Week(nowTime), Week(nowTime))

	SetTimeZone(8)
	currentDate := "2020-02-04 17:38:19"
	t.Equal(New(8).FormatTimestamp(now, "Y-m-d H:i:s"), FormatTimestamp(now, "Y-m-d H:i:s"))
	for i := 1; i <= 7; i++ {
		t.Log(Week(nowTime.Add((time.Hour * 24) * time.Duration(i))))
	}
	t.Equal(2, Week(nowTime))
	t.Equal(currentDate, FormatTimestamp(now, "Y-m-d H:i:s"))
	t.Equal(currentDate, FormatTime(nowTime))
	t.Equal(currentDate, FormatTimestamp(now))

	t.Log(MonthRange(0, 0))
	t.Log(MonthRange(0, 10))
	_, _, err := MonthRange(0, 30)
	t.Equal(true, err != nil)
	t.Log(GetTimeZone())
	t.Log(Now())
}
