package znet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

// BindStruct Bind Struct
func (e *Engine) BindStruct(prefix string, s interface{}, handle ...Handler) error {
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
			return fmt.Errorf("func: [%s] is not an effective routing method", "Init")
		}
		before(g)
	}
	handleName := "reflect.methodValueCall"
	typeOf := typ.TypeOf()

	return zutil.TryCatch(func() error {
		return typ.ForEachMethod(func(i int, m reflect.Method, value reflect.Value) error {
			if m.Name == "Init" {
				return nil
			}
			info, err := zstring.RegexExtract(
				`^(ID|Full|Name){0,}(?i)(ANY|GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)(.*)`, m.Name)
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
			fn := value.Interface()
			handleName = strings.Join([]string{typeOf.PkgPath(), typeOf.Name(), m.Name}, ".")
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

			var (
				p  string
				l  int
				ok bool
			)

			if method == "ANY" {
				p, l, ok = g.handleAny(path, handlerFunc(fn), nil, nil)
			} else {
				p, l, ok = g.addHandle(method, path, handlerFunc(fn), nil, nil)
			}

			if ok && e.IsDebug() {
				e.Log.Debug(routeLog(e.Log, fmt.Sprintf("%%s %%-40s -> %s (%d handlers)", handleName, l), method, p))
			}
			return nil
		})
	})
}
