/*
 * @Author: seekwe
 * @Date:   2019-05-09 16:50:25
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-04 18:57:35
 */

package znet

import (
	"fmt"
	"github.com/sohaha/zlsgo/zstring"
	"strconv"
	"time"

	"github.com/sohaha/zlsgo/zlog"
)

func withRequestLog(c *Context) {
	if c.Request.RequestURI == "/favicon.ico" {
		return
	}
	c.Next()
	requestLog(c)
}

func requestLog(c *Context) {
	var status string
	if c.Engine.IsDebug() {
		end := time.Now()
		latency := end.Sub(c.Info.StartTime)
		code := c.Info.Code
		statusCode := zstring.Buffer()
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
