package znet

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
)

// BindStruct Bind Struct
func (e *Engine) BindStruct(prefix string, s interface{}, handle ...HandlerFunc) error {
	g := e.Group(prefix)
	if len(handle) > 0 {
		for _, v := range handle {
			g.Use(v)
		}
	}
	of := reflect.ValueOf(s)
	typ, err := zreflect.NewVal(of)
	if err != nil {
		return err
	}
	initFn := of.MethodByName("Init")
	if initFn.IsValid() {
		before, ok := initFn.Interface().(func(e *Engine))
		if !ok {
			return fmt.Errorf("func: [%s] is not an effective routing method\n", "Init")
		}
		before(g)
	}
	handleName := "reflect.methodValueCall"
	typeOf := typ.TypeOf()
	return typ.ForEachMethod(func(i int, m reflect.Method, value reflect.Value) error {
		info, err := zstring.RegexExtract(
			`^(ID|Full){0,}(?i)(ANY|GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)(.*)`, m.Name)
		infoLen := len(info)
		if err != nil || infoLen != 4 {
			if e.IsDebug() && m.Name != "Init" {
				e.Log.Warnf("matching rule error: %s%s\n", m.Name, m.Func.String())
			}
			return nil
		}
		path := info[3]
		method := strings.ToUpper(info[2])
		key := strings.ToLower(info[1])
		fn, ok := value.Interface().(func(*Context))
		handleName = runtime.FuncForPC(value.Pointer()).Name()
		handleName = strings.Join([]string{typeOf.PkgPath(), typeOf.Name(), m.Name}, ".")
		if !ok {
			if errFn, ok := value.Interface().(func(*Context) error); !ok {
				return fmt.Errorf("func: [%s] is not an effective routing method\n", m.Name)
			} else {
				fn = func(c *Context) {
					if err := errFn(c); err != nil {
						if e.router.panic != nil {
							e.router.panic(c, err)
						} else {
							panic(err)
						}
					}
				}
			}
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
		if path == "/" {
			path = ""
		}

		if method == "ANY" {
			g.Any(path, fn)
			return nil
		}

		_ = g.handle(method, path, handleName, fn)
		return nil
	})
}
