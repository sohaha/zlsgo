package znet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

var preInvokers = make([]reflect.Type, 0)

func registerPreInvoker(preInvokers []reflect.Type, invoker ...zdi.PreInvoker) ([]reflect.Type, error) {
	for i := range invoker {
		typ := zreflect.TypeOf(invoker[i])
		for i := range preInvokers {
			if preInvokers[i] == typ {
				return preInvokers, fmt.Errorf("pre invoker %s already registered", typ)
			}
		}

		preInvokers = append(preInvokers, typ)
	}
	return preInvokers, nil
}

// RegisterRender Register Render
func RegisterRender(invoker ...zdi.PreInvoker) (err error) {
	preInvokers, err = registerPreInvoker(preInvokers, invoker...)
	return err
}

// RegisterRender Register Render
func (e *Engine) RegisterRender(invoker ...zdi.PreInvoker) (err error) {
	e.customRenderings, err = registerPreInvoker(e.customRenderings, invoker...)
	return err
}

// BindStruct Bind Struct
func (e *Engine) BindStruct(prefix string, s interface{}, handle ...Handler) error {
	g := e.Group(prefix)
	if len(handle) > 0 {
		for _, v := range handle {
			g.Use(v)
		}
	}
	of := reflect.ValueOf(s)
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
			methods := strings.Join(methodsKeys, "|")
			regex := `^(?i)(` + methods + `)(.*)$`
			info, err := zstring.RegexExtract(regex, m.Name)
			infoLen := len(info)
			if err != nil || infoLen != 3 {
				indexs := zstring.RegexFind(`(?i)(`+methods+`)`, m.Name, 1)
				if len(indexs) == 0 {
					if g.IsDebug() && m.Name != "Init" {
						g.Log.Warnf("matching rule error: %s%s\n", m.Name, m.Func.String())
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
			if g.BindStructCase != nil {
				path = g.BindStructCase(path)
			} else if g.BindStructDelimiter != "" {
				path = zstring.CamelCaseToSnakeCase(path, g.BindStructDelimiter)
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
			} else if path != "/" && g.BindStructSuffix != "" {
				path = path + g.BindStructSuffix
			}
			if path == "/" {
				path = ""
			} else if path == "s" {
				path = "/"
			}

			var (
				p  string
				l  int
				ok bool
			)

			p, l, ok = g.addHandle(method, path, Utils.ParseHandlerFunc(fn, e.customRenderings...), nil, nil)

			if ok && g.IsDebug() {
				f := fmt.Sprintf("%%s %%-40s -> %s (%d handlers)", handleName, l)
				if g.webMode == testCode {
					f = "%s %-40s"
				}
				g.Log.Debug(routeLog(g.Log, f, method, p))
			}
			return nil
		})
	})
}
