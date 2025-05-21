//go:build go1.18
// +build go1.18

package ztime_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztime"
)

func TestDiff(t *testing.T) {
	tt := zlsgo.NewTest(t)

	t1, err := ztime.Diff("2021-02-01 00:00:00", "2021-03-01 00:00:00")
	tt.Log(t1, err)
	tt.EqualNil(err, true)
	day := int(t1.Hours() / 24)
	tt.Equal(28, day, true)

	t1, err = ztime.Diff("2021-03-01 00:00:00", "2021-02-01 00:00:00", "Y-m-d H:i:s")
	tt.Log(t1, err)
	tt.EqualNil(err, true)
	day = int(t1.Hours() / 24)
	tt.Equal(-28, day, true)

	t1, err = ztime.Diff(time.Now(), time.Now().AddDate(0, 0, 1))
	tt.Log(t1, err)
	tt.EqualNil(err, true)
	day = int(t1.Hours() / 24)
	tt.Equal(1, day, true)

	t1, err = ztime.Diff("", "")
	tt.Log(t1, err)
	tt.NotNil(err)
}

func TestFindRange(t *testing.T) {
	tt := zlsgo.NewTest(t)

	t1 := "2021-02-01 00:00:00"
	t2 := "2021-03-01 00:00:00"
	t3 := "2021-02-02 00:00:00"
	t4 := "2022-02-02 10:00:00"

	tt.Run("empty", func(tt *zlsgo.TestUtil) {
		st, et, err := ztime.FindRange([]string{}, "Y-m-d H:i:s")
		tt.Log(st, et, err)
		tt.NotNil(err, true)
		tt.Equal(st.IsZero(), true)
		tt.Equal(et.IsZero(), true)
	})

	tt.Run("one", func(tt *zlsgo.TestUtil) {
		st, et, err := ztime.FindRange([]string{t1}, "Y-m-d H:i:s")
		tt.Log(st, et, err)
		tt.NoError(err, true)
		tt.Equal(ztime.FormatTime(st, "Y-m-d H:i:s"), t1)
		tt.Equal(ztime.FormatTime(et, "Y-m-d H:i:s"), t1)
	})

	tt.Run("two", func(tt *zlsgo.TestUtil) {
		st, et, err := ztime.FindRange([]string{t1, t2}, "Y-m-d H:i:s")
		tt.Log(st, et, err)
		tt.NoError(err, true)
		tt.Equal(ztime.FormatTime(st, "Y-m-d H:i:s"), t1)
		tt.Equal(ztime.FormatTime(et, "Y-m-d H:i:s"), t2)
	})

	tt.Run("three", func(tt *zlsgo.TestUtil) {
		st, et, err := ztime.FindRange([]string{t1, t2, t3}, "Y-m-d H:i:s")
		tt.Log(st, et, err)
		tt.NoError(err, true)
		tt.Equal(ztime.FormatTime(st, "Y-m-d H:i:s"), t1)
		tt.Equal(ztime.FormatTime(et, "Y-m-d H:i:s"), t2)
	})

	tt.Run("four", func(tt *zlsgo.TestUtil) {
		st, et, err := ztime.FindRange([]string{t1, t2, t3, t4}, "Y-m-d H:i:s")
		tt.Log(st, et, err)
		tt.NoError(err, true)
		tt.Equal(ztime.FormatTime(st, "Y-m-d H:i:s"), t1)
		tt.Equal(ztime.FormatTime(et, "Y-m-d H:i:s"), t4)
	})

	tt.Run("others", func(tt *zlsgo.TestUtil) {
		st, et, err := ztime.FindRange([]string{"2021-12-12", "2032-11-01", "1990-03-28"}, "Y-m-d")
		tt.Log(st, et, err)
		tt.NoError(err, true)
		tt.Equal(ztime.FormatTime(st, "Y-m-d"), "1990-03-28")
		tt.Equal(ztime.FormatTime(et, "Y-m-d"), "2032-11-01")
	})
}

func TestSequence(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("days", func(tt *zlsgo.TestUtil) {
		days, err := ztime.Sequence("2023-01-01", "2023-01-05", time.Hour*24, "Y-m-d")
		tt.Log(days)
		tt.NoError(err)
		tt.Equal(5, len(days))
		tt.Equal("2023-01-01", days[0])
		tt.Equal("2023-01-05", days[4])
	})

	tt.Run("hours", func(tt *zlsgo.TestUtil) {
		hours, err := ztime.Sequence("2023-01-01 00:00", "2023-01-01 05:00", time.Hour, "Y-m-d H:i")
		tt.NoError(err)
		tt.Equal(6, len(hours))
		tt.Equal("2023-01-01 00:00", hours[0])
		tt.Equal("2023-01-01 05:00", hours[5])
	})

	tt.Run("only minutes", func(tt *zlsgo.TestUtil) {
		minutes, err := ztime.Sequence("00:00", "15:00", time.Minute*3, "H:i")
		tt.NoError(err)
		tt.Equal(301, len(minutes))
		tt.Equal("00:00", minutes[0])
		tt.Equal("00:03", minutes[1])
		tt.Equal("15:00", minutes[len(minutes)-1])
	})

	tt.Run("invalid range", func(tt *zlsgo.TestUtil) {
		_, err := ztime.Sequence("2023-01-05", "2023-01-01", time.Hour*24, "Y-m-d")
		tt.NotNil(err)
	})

	tt.Run("step length", func(tt *zlsgo.TestUtil) {
		equalRange, err := ztime.Sequence("2023-01-01 00:00", "2023-01-02 00:00", time.Hour*24, "Y-m-d H:i")
		tt.NoError(err)
		tt.Log("equalRange:", equalRange)
		tt.Equal(2, len(equalRange))
		tt.Equal("2023-01-01 00:00", equalRange[0])
		tt.Equal("2023-01-02 00:00", equalRange[1])

		largerStep, err := ztime.Sequence("2023-01-01 00:00", "2023-01-01 20:00", time.Hour*24, "Y-m-d H:i")
		tt.NoError(err)
		tt.Log("largerStep:", largerStep)
		tt.Equal(2, len(largerStep))
		if len(largerStep) >= 1 {
			tt.Equal("2023-01-01 00:00", largerStep[0])
		}
		if len(largerStep) >= 2 {
			tt.Equal("2023-01-01 20:00", largerStep[1])
		}

		muchLargerStep, err := ztime.Sequence("2023-01-01 00:00", "2023-01-01 01:00", time.Hour*24, "Y-m-d H:i")
		tt.NoError(err)
		tt.Log("muchLargerStep:", muchLargerStep)
		tt.Equal(2, len(muchLargerStep))
		if len(muchLargerStep) >= 1 {
			tt.Equal("2023-01-01 00:00", muchLargerStep[0])
		}
		if len(muchLargerStep) >= 2 {
			tt.Equal("2023-01-01 01:00", muchLargerStep[1])
		}

		start, _ := time.Parse("2006-01-02 15:04", "2023-01-01 00:00")
		end, _ := time.Parse("2006-01-02 15:04", "2023-01-01 01:00")
		timeTypeInput, err := ztime.Sequence(start, end, time.Hour*24, "Y-m-d H:i")
		tt.NoError(err)
		tt.Log("timeTypeInput:", timeTypeInput)
		tt.Equal(2, len(timeTypeInput))
		if len(timeTypeInput) >= 1 {
			tt.Equal("2023-01-01 00:00", timeTypeInput[0])
		}
		if len(timeTypeInput) >= 2 {
			tt.Equal("2023-01-01 01:00", timeTypeInput[1])
		}
	})
}
