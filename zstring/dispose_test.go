package zstring_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

func TestFilter(t *testing.T) {
	tt := zlsgo.NewTest(t)
	f := zstring.NewFilter([]string{"我", "你", "他", "初音", ""}, '#')
	res, keywords, found := f.Filter("你是谁,我是谁,他又是谁")
	tt.EqualTrue(found)
	tt.EqualExit(3, len(keywords))
	tt.EqualExit("#是谁,#是谁,#又是谁", res)

	_, _, found = f.Filter("")
	tt.EqualExit(false, found)

	t.Log(2, len(f.Find("你是谁,初音又是谁")))
	t.Log(0, len(f.Find("")))
}

func TestReplacer(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := zstring.NewReplacer(map[string]string{"你": "初音", "它": "犬夜叉"})
	res := r.Replace("你是谁,我是谁,它又是谁")
	tt.EqualExit("初音是谁,我是谁,犬夜叉又是谁", res)
	t.Log("", r.Replace(""))
}
