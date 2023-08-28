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
	of := zreflect.ValueOf(s)
	if !of.IsValid() {
		return nil
	}
	initFn := of.MethodByName("Init")
	if initFn.IsValid() {
		before, ok := initFn.Interface().(func(e *Engine))
		if ok {
			before(g)
		} else {
			if before, ok := initFn.Interface().(func(e *Engine) error); !ok {
				return fmt.Errorf("func: [%s] is not an effective routing method", "Init")
			} else {
				if err := before(g); err != nil {
					return err
				}
			}
		}

	}
	typeOf := reflect.Indirect(of).Type()
	return zutil.TryCatch(func() error {
		return zreflect.ForEachMethod(of, func(i int, m reflect.Method, value reflect.Value) error {
			if m.Name == "Init" {
				return nil
			}
			path, method, key := "", "", ""
			methods := `ANY|GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS`
			regex := `^(?i)(` + methods + `)(.*)$`
			info, err := zstring.RegexExtract(regex, m.Name)
			infoLen := len(info)
			if err != nil || infoLen != 3 {
				indexs := zstring.RegexFind(`(?i)(`+methods+`)`, m.Name, 1)
				if len(indexs) == 0 {
					if e.IsDebug() && m.Name != "Init" {
						e.Log.Warnf("matching rule error: %s%s\n", m.Name, m.Func.String())
					}
					return nil
				}

				index := indexs[0]
				method = strings.ToUpper(m.Name[index[0]:index[1]])
				path = m.Name[index[1]:]
				key = strings.ToLower(m.Name[:index[0]])
			} else {
				path = info[2]
				method = strings.ToUpper(info[1])
			}

			fn := value.Interface()
			handleName := strings.Join([]string{typeOf.PkgPath(), typeOf.Name(), m.Name}, ".")
			if e.BindStructCase != nil {
				path = e.BindStructCase(path)
			} else if e.BindStructDelimiter != "" {
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
				p, l, ok = g.handleAny(path, Utils.ParseHandlerFunc(fn), nil, nil)
			} else {
				p, l, ok = g.addHandle(method, path, Utils.ParseHandlerFunc(fn), nil, nil)
			}

			if ok && e.IsDebug() {
				e.Log.Debug(routeLog(e.Log, fmt.Sprintf("%%s %%-40s -> %s (%d handlers)", handleName, l), method, p))
			}
			return nil
		})
	})
}
