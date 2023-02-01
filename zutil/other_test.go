package zutil_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestUnescapeHTML(t *testing.T) {
	tt := zlsgo.NewTest(t)
	s := zutil.UnescapeHTML("")
	tt.Log(s)
}
