package znet

import (
	"fmt"
	"html/template"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
)

func showRouteDebug(log *zlog.Logger, tf, method, path string) string {
	mLen := zstring.Len(method)
	var mtd string
	min := 6
	if mLen < min {
		mtd = zstring.Pad(method, min, " ", zstring.PadLeft)
	} else {
		mtd = zstring.Substr(method, 0, min)
	}

	switch method {
	case "GET":
		method = log.ColorTextWrap(zlog.ColorLightCyan, mtd)
	case "POST":
		method = log.ColorTextWrap(zlog.ColorLightBlue, mtd)
	case "PUT":
		method = log.ColorTextWrap(zlog.ColorLightGreen, mtd)
	case "DELETE":
		method = log.ColorTextWrap(zlog.ColorRed, mtd)
	case "Any":
		method = log.ColorTextWrap(zlog.ColorLightMagenta, mtd)
	case "OPTIONS":
		method = log.ColorTextWrap(zlog.ColorLightMagenta, mtd)
	case "FILE":
		method = log.ColorTextWrap(zlog.ColorLightMagenta, mtd)
	default:
		method = log.ColorTextWrap(zlog.ColorDefault, mtd)
	}
	path = zstring.Pad(path, 20, " ", zstring.PadRight)
	return fmt.Sprintf(tf, method, path)
}

func templatesDebug(t *template.Template) {
	l := len(t.Templates())
	var buf string
	for _, t := range t.Templates() {
		n := t.Name()
		if n == "" {
			continue
		}
		buf = "\t  - " + n + "\n"
	}
	Log.Debugf("Loaded HTML Templates (%d): \n%s", l, buf)
}
