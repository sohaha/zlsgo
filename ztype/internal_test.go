package ztype

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestProcessFieldTagIgnore(t *testing.T) {
	tt := zlsgo.NewTest(t)
	type Emb struct{ Y int }
	type T struct{ Emb }
	c := &Conver{TagName: tagName, Squash: true, IgnoreTagName: true}
	val := reflect.Indirect(reflect.ValueOf(T{}))
	sf := val.Type().Field(0)
	squash, remain := c.processFieldTag(sf, val.Field(0))
	tt.EqualTrue(squash && !remain)
}

func TestCacheGetStructInfoCoverage(t *testing.T) {
	tt := zlsgo.NewTest(t)
	type A struct {
		X int `z:"x"`
	}
	info1 := getStructInfo(reflect.TypeOf((*A)(nil)))
	tt.EqualTrue(len(info1) == 1 && info1[0].Name == "x")
	info2 := getStructInfo(reflect.TypeOf((*A)(nil)))
	tt.Equal(1, len(info2))
	v := getStructInfo(reflect.TypeOf(1))
	tt.EqualTrue(v == nil)
}

func TestUnescapeAndInvalidToken(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal("", unescapePathKey(""))
	tok := pathToken{kind: 99}
	v, ok := executePathToken(tok, nil)
	tt.EqualTrue(!ok && v == nil)
}

func TestProcessFieldTagVariants(t *testing.T) {
	tt := zlsgo.NewTest(t)
	type NoTag struct{ A int }
	type WithName struct {
		A int `z:"a"`
	}
	type OnlySquash struct {
		Em struct{} `z:"Em,squash"`
	}
	type OnlyRemain struct {
		R map[interface{}]interface{} `z:"R,remain"`
	}
	type Multi struct {
		Em struct{} `z:"Em,squash,remain"`
	}

	c := &Conver{TagName: tagName, Squash: true}

	{
		v := reflect.Indirect(reflect.ValueOf(WithName{}))
		sf := v.Type().Field(0)
		sq, rm := c.processFieldTag(sf, v.Field(0))
		tt.EqualTrue(!sq && !rm)
	}

	{
		v := reflect.Indirect(reflect.ValueOf(NoTag{}))
		sf := v.Type().Field(0)
		sq, rm := c.processFieldTag(sf, v.Field(0))
		tt.EqualTrue(!sq && !rm)
	}

	{
		v := reflect.Indirect(reflect.ValueOf(OnlySquash{}))
		sf := v.Type().Field(0)
		sq, rm := c.processFieldTag(sf, v.Field(0))
		tt.EqualTrue(sq && !rm)
	}

	{
		v := reflect.Indirect(reflect.ValueOf(OnlyRemain{}))
		sf := v.Type().Field(0)
		sq, rm := c.processFieldTag(sf, v.Field(0))
		tt.EqualTrue(!sq && rm)
	}

	{
		v := reflect.Indirect(reflect.ValueOf(Multi{}))
		sf := v.Type().Field(0)
		sq, rm := c.processFieldTag(sf, v.Field(0))
		tt.EqualTrue(sq && rm)
	}
}

func TestParseTagOptions(t *testing.T) {
	tt := zlsgo.NewTest(t)
	opts := parseTagOptions("")
	tt.EqualTrue(opts == nil)
	opts = parseTagOptions("omitempty, squash, custom ")
	tt.Equal(3, len(opts))
}
