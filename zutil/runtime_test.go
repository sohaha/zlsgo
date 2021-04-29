package zutil_test

import (
	"testing"

	"github.com/sohaha/zlsgo/zutil"
)

func TestGetGid(t *testing.T) {
	t.Log(zutil.GetGid())
	c := make(chan struct{}, 0)
	go func() {
		t.Log(zutil.GetGid())
		c <- struct{}{}
	}()
	<-c
	t.Log(zutil.GetGid())
}
