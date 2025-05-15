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
	// invokerCodeText is a function type that returns an HTTP status code and text.
	// It implements the zdi.PreInvoker interface for dependency injection.
	invokerCodeText func() (int, string)
)

// Ensure invokerCodeText implements zdi.PreInvoker interface
var _ zdi.PreInvoker = (*invokerCodeText)(nil)

// Invoke implements the zdi.PreInvoker interface.
// It calls the wrapped function and returns its results as reflect.Value objects.
func (h invokerCodeText) Invoke(_ []interface{}) ([]reflect.Value, error) {
	code, text := h()
	return []reflect.Value{zreflect.ValueOf(code), reflect.ValueOf(text)}, nil
}

// defErrorHandler returns the default error handler function.
// The default handler responds with a 500 status code and the error message as plain text.
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

// requestLog is a middleware function that logs HTTP request details.
// It records the request method, path, status code, and response time.
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

// errURLQuerySemicolon is the error message produced by Go's standard library
// when a URL query contains semicolons, which are no longer supported as separators.
const errURLQuerySemicolon = "http: URL query contains semicolon, which is no longer a supported separator; parts of the query may be stripped when parsed; see golang.org/issue/25192\n"

// allowQuerySemicolons modifies a request to allow semicolons in URL query parameters.
// This is a workaround for Go's standard library behavior that no longer supports semicolons
// as query parameter separators (see golang.org/issue/25192).
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
