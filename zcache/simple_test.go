package zcache_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/ztime"
)

func TestSimple(t *testing.T) {
	tt := zlsgo.NewTest(t)

	now := ztime.Now()
	zcache.Set("TestSimple1", now, time.Second/5)

	v, ok := zcache.GetAny("TestSimple1")
	tt.EqualTrue(ok)
	tt.Equal(now, v.String())

	time.Sleep(time.Second / 4)

	vv, ok := zcache.Get("TestSimple1")
	t.Log(vv, ok)
	tt.EqualTrue(!ok)

	v2, ok := zcache.ProvideGet("TestSimple2", func() (interface{}, bool) {
		return now, true
	})
	tt.EqualTrue(ok)
	tt.Equal(now, v2.(string))

	zcache.Delete("TestSimple2")
	_, ok = zcache.Get("TestSimple2")
	tt.EqualTrue(!ok)
}
