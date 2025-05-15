package znet

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"runtime"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
)

// routeLog formats route information for logging purposes.
// It colorizes HTTP methods and pads strings to ensure consistent log formatting.
func routeLog(log *zlog.Logger, tf, method, path string) string {
	mLen := zstring.Len(method)
	var mtd string
	min := 6
	if mLen < min {
		mtd = zstring.Pad(method, min, " ", zstring.PadLeft)
	} else {
		mtd = zstring.Substr(method, 0, min)
	}

	switch method {
	case http.MethodGet:
		method = log.ColorTextWrap(zlog.ColorLightCyan, mtd)
	case http.MethodPost:
		method = log.ColorTextWrap(zlog.ColorLightBlue, mtd)
	case http.MethodPut:
		method = log.ColorTextWrap(zlog.ColorLightGreen, mtd)
	case http.MethodDelete:
		method = log.ColorTextWrap(zlog.ColorRed, mtd)
	case anyMethod:
		method = log.ColorTextWrap(zlog.ColorLightMagenta, mtd)
	case http.MethodOptions:
		method = log.ColorTextWrap(zlog.ColorLightMagenta, mtd)
	case "FILE":
		method = log.ColorTextWrap(zlog.ColorLightMagenta, mtd)
	default:
		method = log.ColorTextWrap(zlog.ColorDefault, mtd)
	}
	path = zstring.Pad(path, 20, " ", zstring.PadRight)
	return fmt.Sprintf(tf, method, path)
}

// templatesDebug logs information about loaded HTML templates when in debug mode.
// It lists all named templates that have been loaded into the engine.
func templatesDebug(e *Engine, t *template.Template) {
	l := 0
	buf := zstring.Buffer()
	for _, t := range t.Templates() {
		n := t.Name()
		if n == "" {
			continue
		}
		buf.WriteString("\t  - " + n + "\n")
		l++
	}
	e.Log.Debugf("Loaded HTML Templates (%d): \n%s", l, buf.String())
}

// routeAddLog logs information about a newly added route when in debug mode.
// It includes the HTTP method, path, handler function name, and middleware count.
func routeAddLog(e *Engine, method string, path string, action Handler, middlewareCount int) {
	if e.IsDebug() {
		v := zreflect.ValueOf(action)
		if e.webMode == testCode {
			e.Log.Debug(routeLog(e.Log, "%s %-40s", method, path))
			return
		}

		if v.Kind() == reflect.Func {
			e.Log.Debug(routeLog(e.Log, fmt.Sprintf("%%s %%-40s -> %s (%d handlers)", runtime.FuncForPC(v.Pointer()).Name(), middlewareCount), method, path))
		} else {
			e.Log.Warn(routeLog(e.Log, fmt.Sprintf("%%s %%-40s -> %s (%d handlers)", v.Type().String(), middlewareCount), method, path))
		}
	}
}
