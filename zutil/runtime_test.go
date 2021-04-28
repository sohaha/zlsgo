package zutil_test

import (
	"github.com/sohaha/zlsgo/zutil"
	"testing"
)

func TestGetGid(t *testing.T) {
	t.Log(zutil.GetGid())
	var c chan struct{}
	go func() {
		t.Log(zutil.GetGid())
		c <- struct{}{}
	}()
	<-c
}
