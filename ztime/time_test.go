package ztime

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestNewTime(t *testing.T) {
	tt := zlsgo.NewTest(t)
	now := int64(1580809099)
	nowTime := time.Unix(now, 0)

	tt.Equal(time.Local, GetTimeZone())

	SetTimeZone(0)

	tt.EqualExit("2020-02-04 09:38:19", FormatTimestamp(now, "Y-m-d H:i:s"))
	tt.EqualExit("2020-02-04 09:38:19", FormatTimestamp(now))
	tt.EqualExit("20/02/ 04", FormatTimestamp(now, "y/m/ d"))

	tt.EqualExit("2020-02-04 10:38:19", New(1).FormatTimestamp(now, "Y-m-d H:i:s"))
	tt.EqualExit(New(24).Week(nowTime), Week(nowTime)+1)
	tt.EqualExit("2020-02-04T09:38:19+00:00", FormatTimestamp(now, "c"))

	SetTimeZone(8)

	t.Log(GetTimeZone().String())

	currentDate := "2020-02-04 17:38:19"
	tt.Equal(New(8).FormatTimestamp(now, "Y-m-d H:i:s"), FormatTimestamp(now, "Y-m-d H:i:s"))

	for i, v := range []int{2, 3, 4, 5, 6, 7, 1} {
		tt.Equal(v, Week(nowTime.Add((time.Hour*24)*time.Duration(i))))
	}
	tt.Equal(2, Week(nowTime))
	tt.Equal(currentDate, FormatTimestamp(now, "Y-m-d H:i:s"))
	tt.Equal(currentDate, FormatTime(nowTime))
	tt.Equal(currentDate, FormatTimestamp(now))

	beginTime, endTime, err := MonthRange(2020, 10)
	tt.EqualNil(err)
	tt.Equal(int64(1601481600), beginTime)
	tt.Equal(int64(1604159999), endTime)

	_, _, err = MonthRange(0, 0)
	tt.EqualNil(err)

	_, _, err = MonthRange(0, 30)
	tt.Equal(true, err != nil)

	SetTimeZone(0)
	t.Log(GetTimeZone().String())

	t.Log(Now())
}

func TestFormatTlp(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Equal("06-01-d", FormatTlp("y-m-\\d"))
	t.Equal("地球时间：2006y01-d", FormatTlp("地球时间：Y\\ym-\\d"))
	t.Equal("06-01-02 00:00:00", FormatTlp("y-m-d \\0\\0:\\0\\0:\\0\\0"))
}

func TestUnix(t *testing.T) {
	t.Log(Unix(1648879934))

	ti := New(2)
	t.Log(ti.Unix(1648879934))
}

func TestParse(t *testing.T) {
	tt := zlsgo.NewTest(t)
	date, err := Parse("2020-02-04 09:38:19")
	tt.EqualNil(err)
	t.Log(date)

	date, err = Parse("2020/2/4 09:38:19")
	tt.EqualNil(err)
	t.Log(date, err)

	date, err = Parse("2020.02.04", "Y.m.d")
	tt.EqualNil(err)
	t.Log(date, err)

	date, err = Parse("地球时间:2020y02m04 11:11:11", "地球时间:Y\\ym\\md h:i:s")
	tt.EqualNil(err)
	t.Log(date, err)

	date, err = Parse("2020.2.4 33")
	tt.Equal(true, err != nil)
	t.Log(date, err)

	s := Now("Y-m-d H")
	date, _ = New(2).Parse(s, "Y-m-d H")
	t.Log(
		s,
		FormatTime(date, "Y-m-d H"),
		New(24).FormatTime(date, "Y-m-d H"),
		New(0).FormatTime(date, "Y-m-d H"),
	)
}
