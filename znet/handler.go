package znet

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zutil"
)

type (
	invokerCodeText func() (int, string)
)

var (
	_ zdi.PreInvoker = (*invokerCodeText)(nil)
)

func (h invokerCodeText) Invoke(_ []interface{}) ([]reflect.Value, error) {
	code, text := h()
	return []reflect.Value{zreflect.ValueOf(code), reflect.ValueOf(text)}, nil
}

func defErrorHandler() ErrHandlerFunc {
	return func(c *Context, err error) {
		c.String(500, err.Error())
	}
}

// RewriteErrorHandler rewrite error handler
func RewriteErrorHandler(handler ErrHandlerFunc) Handler {
	return func(c *Context) {
		c.renderError = handler
		c.Next()
	}
}

// Recovery is a middleware that recovers from panics anywhere in the chain
func Recovery(handler ErrHandlerFunc) Handler {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				errMsg, ok := err.(error)
				if !ok {
					errMsg = errors.New(fmt.Sprint(err))
				}
				handler(c, errMsg)
			}
		}()
		c.Next()
	}
}

func requestLog(c *Context) {
	if c.Engine.IsDebug() {
		var status string
		end := time.Now()
		statusCode := zutil.GetBuff()
		latency := end.Sub(c.startTime)
		code := c.prevData.Code.Load()
		statusCode.WriteString(" ")
		statusCode.WriteString(strconv.FormatInt(int64(code), 10))
		statusCode.WriteString(" ")
		s := statusCode.String()
		zutil.PutBuff(statusCode)
		switch {
		case code >= 200 && code <= 299:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorGreen, s)
		case code >= 300 && code <= 399:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorYellow, s)
		default:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorRed, s)
		}
		clientIP := c.GetClientIP()
		if clientIP == "" {
			clientIP = "unknown"
		}
		ft := fmt.Sprintf("%s %15s %15v %%s %%s", status, clientIP, latency)
		c.Log.Success(routeLog(c.Log, ft, c.Request.Method, c.Request.RequestURI))
	}
}

const errURLQuerySemicolon = "http: URL query contains semicolon, which is no longer a supported separator; parts of the query may be stripped when parsed; see golang.org/issue/25192\n"

func allowQuerySemicolons(r *http.Request) {
	// clopy of net/http.AllowQuerySemicolons.
	if s := r.URL.RawQuery; strings.Contains(s, ";") {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.RawQuery = strings.Replace(s, ";", "&", -1)
		*r = *r2
	}
}
