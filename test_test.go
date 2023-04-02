package zlsgo_test

import (
	"fmt"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestNewTest(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(1, 1)
	tt.EqualExit(1, 1)
	tt.EqualTrue(true)
	tt.EqualNil(nil)
	tt.NoError(nil)
	tt.Log("ok")
}

func TestPath(t *testing.T) {
	p("yy.yy",
		map[string]interface{}{
			"yy": map[string]interface{}{
				"yy": 123,
			},
		})
	p("yy.xx.o456o",
		map[string]interface{}{
			"yy": map[string]interface{}{
				"xx": map[string]string{"o456o": "999"},
			},
		})
	p("yy\\.yy",
		map[string]interface{}{
			"yy.yy": map[string]interface{}{
				"yy": 123,
			},
		})
	p("yy\\.yy.yy",
		map[string]interface{}{
			"yy.yy": map[string]interface{}{
				"yy": 123,
			},
		})

	p("yy\\\\.yy.yy", nil)
}

func p(path string, v interface{}) {
	keys := []string{}
	t := 0
	i := 0
	val := v
	pp := func(p string, v interface{}) interface{} {
		if v == nil {
			return nil
		}
		switch val := v.(type) {
		case map[string]interface{}:
			return val[p]
		case map[string]string:
			return val[p]
		}
		return nil
	}
	for ; i < len(path); i++ {
		switch path[i] {
		case '\\':
			ss := path[t:i]
			i++
			path = ss + path[i:]
		case '.':
			keys = append(keys, path[t:i])
			val = pp(path[t:i], val)
			t = i + 1
		}
	}
	if i != t {
		keys = append(keys, path[t:])
		val = pp(path[t:], val)
	}
	fmt.Println(path, keys, val)
}
