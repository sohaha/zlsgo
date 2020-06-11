package znet

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

func (e *Engine) BindStruct(prefix string, s interface{},
	handle ...HandlerFunc) error {
	g := e.Group(prefix)
	if len(handle) > 0 {
		for _, v := range handle {
			g.Use(v)
		}
	}
	_ = reflect.TypeOf(s)
	valueOf := reflect.ValueOf(s)
	return zutil.GetAllMethod(s, func(numMethod int, m reflect.Method) error {
		info, err := zstring.RegexExtract(
			`(?i)(ANY|GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)(.*)`, m.Name)
		if err != nil || len(info) != 3 {
			if e.IsDebug() {
				e.Log.Warnf("matching rule error: [%s] %v", m.Name, err)
			}
			return nil
		}
		fn := func(c *Context) {
			valueOf.Method(numMethod).Call([]reflect.Value{reflect.ValueOf(c)})
		}
		name := info[2]
		if BindStructDelimiter != "" {
			name = zstring.CamelCaseToSnakeCase(info[2], BindStructDelimiter) + BindStructSuffix
		}
		switch strings.ToUpper(info[1]) {
		case "GET":
			g.GET(name, fn)
		case "POST":
			g.POST(name, fn)
		case "PUT":
			g.PUT(name, fn)
		case "DELETE":
			g.DELETE(name, fn)
		case "PATCH":
			g.PATCH(name, fn)
		case "HEAD":
			g.HEAD(name, fn)
		case "OPTIONS":
			g.OPTIONS(name, fn)
		default:
			g.Any(name, fn)
		}
		return nil
	})
}
