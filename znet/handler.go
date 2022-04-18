package znet

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zutil"
)

// Recovery Recovery
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
		c.l.RLock()
		statusCode := zutil.GetBuff()
		defer func() {
			c.l.RUnlock()
			zutil.PutBuff(statusCode)
		}()
		latency := end.Sub(c.startTime)
		code := c.prevData.Code
		statusCode.WriteString(" ")
		statusCode.WriteString(strconv.Itoa(code))
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
		c.Log.Success(showRouteDebug(c.Log, ft, c.Request.Method, c.Request.RequestURI))
	}
}
