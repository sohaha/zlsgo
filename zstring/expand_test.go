package zstring

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestExpand(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal("hello zlsgo", Expand("hello $world", func(key string) string {
		return "zlsgo"
	}))

	tt.Equal("hello {zlsgo}", Expand("hello {$world}", func(key string) string {
		return "zlsgo"
	}))

	tt.Equal("hello zlsgo", Expand("hello ${world}", func(key string) string {
		t.Log(key)
		return "zlsgo"
	}))

	var keys []string
	Expand("${a} $b $c.d ${e.f} $1 - ${*}", func(key string) string {
		t.Log(key)
		keys = append(keys, key)
		return ""
	})
	tt.Equal([]string{"a", "b", "c", "e.f", "1", "*"}, keys)
}
