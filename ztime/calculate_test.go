package ztime_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztime"
)

func TestBetween(t *testing.T) {
	tt := zlsgo.NewTest(t)

	t1, err := ztime.Between("2021-02-01 00:00:00", "2021-03-01 00:00:00", "Y-m-d H:i:s")
	tt.Log(t1, err)
	tt.EqualNil(err, true)
	day := int(t1.Hours() / 24)
	tt.Equal(28, day, true)

	t1, err = ztime.Between("2021-03-01 00:00:00", "2021-02-01 00:00:00", "Y-m-d H:i:s")
	tt.Log(t1, err)
	tt.EqualNil(err, true)
	day = int(t1.Hours() / 24)
	tt.Equal(-28, day, true)
}
