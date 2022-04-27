package znet

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
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
	return []reflect.Value{reflect.ValueOf(code), reflect.ValueOf(text)}, nil
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
		defer zutil.PutBuff(statusCode)
		latency := end.Sub(c.startTime)
		code := c.prevData.Code.Load()
		statusCode.WriteString(" ")
		statusCode.WriteString(strconv.FormatInt(int64(code), 10))
		statusCode.WriteString(" ")
		s := statusCode.String()
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
