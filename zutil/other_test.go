package zutil_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestUnescapeHTML(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	s := zutil.UnescapeHTML("")
	t.Log(s)
}
