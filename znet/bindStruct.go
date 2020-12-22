package znet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

// BindStruct Bind Struct
func (e *Engine) BindStruct(prefix string, s interface{}, handle ...HandlerFunc) error {
	g := e.Group(prefix)
	if len(handle) > 0 {
		for _, v := range handle {
			g.Use(v)
		}
	}
	_ = reflect.TypeOf(s)
	valueOf := reflect.ValueOf(s)
	if valueOf.IsValid() {
		initFn := valueOf.MethodByName("Init")
		if initFn.IsValid() {
			before, ok := initFn.Interface().(func(e *Engine))
			if !ok {
				return fmt.Errorf("func: [%s] is not an effective routing method\n", "Init")
			}
			before(g)
		}
	}
	return zutil.GetAllMethod(s, func(numMethod int, m reflect.Method) error {
		info, err := zstring.RegexExtract(
			`^(ID|Full){0,}(?i)(ANY|GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)(.*)`, m.Name)
		infoLen := len(info)
		if err != nil || infoLen != 4 {
			if e.IsDebug() && m.Name != "Init" {
				e.Log.Warnf("matching rule error: [%s] %v\n", m.Name, err)
			}
			return nil
		}
		path := info[3]
		method := strings.ToUpper(info[2])
		key := strings.ToLower(info[1])
		fn, ok := valueOf.Method(numMethod).Interface().(func(*Context))
		if !ok {
			return fmt.Errorf("func: [%s] is not an effective routing method\n", m.Name)
		}
		// valueOf.Method(numMethod).Call([]reflect.Value{reflect.ValueOf(c)})
		if e.BindStructDelimiter != "" {
			path = zstring.CamelCaseToSnakeCase(path, e.BindStructDelimiter)
		}
		if path == "" {
			path = "/"
		}

		if key != "" {
			if strings.HasSuffix(path, "/") {
				path += ":" + key
			} else {
				path += "/:" + key
			}
		} else if path != "/" && e.BindStructSuffix != "" {
			path = path + e.BindStructSuffix
		}
		if method == "ANY" {
			g.Any(path, fn)
			return nil
		}
		g.Handle(method, path, fn)
		return nil
	})
}
