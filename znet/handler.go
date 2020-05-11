package znet

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zutil"
)

func Recovery(r *Engine, handler PanicFunc) HandlerFunc {
	r.router.panic = handler
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
		c.RLock()
		statusCode := zutil.GetBuff()
		defer func() {
			c.RUnlock()
			zutil.PutBuff(statusCode)
		}()
		latency := end.Sub(c.StartTime)
		code := c.Code
		statusCode.WriteString(" ")
		statusCode.WriteString(strconv.Itoa(code))
		statusCode.WriteString(" ")
		switch {
		case code >= 200 && code <= 299:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorGreen, statusCode.String())
		case code >= 300 && code <= 399:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorYellow, statusCode.String())
		default:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorRed, statusCode.String())
		}
		ft := fmt.Sprintf("%s %15s %15v %%s  %%s", status, c.GetClientIP(), latency)
		c.Log.Success(showRouteDebug(c.Log, ft, c.Request.Method, c.Request.RequestURI))
	}
}
