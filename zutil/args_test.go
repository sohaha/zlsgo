package zutil

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

func TestArgs(t *testing.T) {
	tt := zlsgo.NewTest(t)

	sum := NewInt32(0)

	moreNum := []interface{}{"abc " + zstring.Pad("", 88, "? ", zstring.PadRight) + " |[]", "abc " + zstring.Pad("", 88, "$? ", zstring.PadRight)}
	for _, num := range strings.Split(zstring.Rand(88), "") {
		moreNum = append(moreNum, num)
	}
	moreNum[0] = fmt.Sprintf("abc "+zstring.Pad("", 88, "? ", zstring.PadRight)+"|%v", moreNum[2:])
	tests := [][]interface{}{
		{"abc ? ?|[123 321]", "abc $? $?", 123, 321},
		{"abc ? ?|[123 123]", "abc $? $0", 123, 321},
		{"abc ? |[456]", "abc $0 ", 456},
		{"abc ? |[]", "abc ? ", 456},
		{"abc  |[]", "abc $1 ", 123},
		{"abc ?-?|[6 123]", "abc ${s}-${a}", Named("a", 123), Named("s", func() interface{} { return sum.Load() })},
		{"abc   |[]", "abc ${unknown}  ", 123},
		{"abc $ |[]", "abc $$ ", 123},
		{"abc$|[]", "abc$", 123},
		{"abc ? ? ? ? |[123 456 123 456]", "abc $? $? $0 $? ", 123, 456, 789},
		moreNum,
	}
	for _, c := range tests {
		sum.Add(1)
		args := NewArgs()
		for i := 2; i < len(c); i++ {
			args.Map(c[i])
		}
		query, values := args.Compile(c[1].(string))
		t.Log(query, values)
		actual := fmt.Sprintf("%v|%v", query, values)
		tt.Equal(c[0].(string), actual)
	}
}

func TestArgsOnlyNamed(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := [][]interface{}{
		{"abc $? $?|[]", "abc $? $?", 123, 321},
		{"abc $0 |[]", "abc $0 ", 456},
		{"abc $a |[]", "abc $$a ", Named("a", 123)},
		{"abc ? |[123]", "abc ${a} ", Named("a", 123), Named("a", 321)},
		{"abc |[]", "abc ${xxx}"},
		{"abc @sql_a |[{{} sql_a 123}]", "abc ${a} ", Named("a", sql.Named("sql_a", "123")), Named("ab", sql.Named("sql_a", "123"))},
		{"abc ?-?|[99 123]", "abc ${s}-${a}", Named("a", 123), Named("s", 99)},
		{"abc ?-?|[s.k__ 123]", "abc ${s.k}-${a}", Named("a", 123), Named("s.*", func(k string) interface{} {
			return k + "__"
		})},
	}
	for _, c := range tests {
		args := NewArgs(WithOnlyNamed())
		for i := 2; i < len(c); i++ {
			args.Map(c[i])
		}
		query, values := args.Compile(c[1].(string))
		t.Log(query, values)
		actual := fmt.Sprintf("%v|%v", query, values)
		tt.Equal(c[0].(string), actual)
	}
}

func TestArgsString(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tests := [][]interface{}{
		{"123", "$0", 123},
		{"abc 123 321", "abc $? $?", 123, 321},
		{"abc 123 321", "abc $? $?", 123, 321},
		{"abc 123 ", "abc ${a} ", Named("a", sql.Named("sql_a", "123"))},
		{"abc 99-123", "abc ${s}-${a}", Named("a", 123), Named("s", 99)},
		{"abc 99${a}123", "abc ${s}$${a}${a}", Named("a", 123), Named("s", 99)},
		{"abc s.k__-123", "abc ${s.k}-${a}", Named("a", 123), Named("s.*", func(k string) interface{} {
			return k + "__"
		})},
	}
	for _, c := range tests {
		args := NewArgs()
		for i := 2; i < len(c); i++ {
			args.Map(c[i])
		}
		result := args.CompileString(c[1].(string))
		t.Log(result)
		tt.Equal(c[0].(string), result)
	}
}
