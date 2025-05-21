package ztime_test

import (
	"testing"

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
