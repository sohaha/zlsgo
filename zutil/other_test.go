package zutil

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestUnescapeHTML(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	s := UnescapeHTML("")
	t.Log(s)
}
