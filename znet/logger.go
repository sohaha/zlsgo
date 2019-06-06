/*
 * @Author: seekwe
 * @Date:   2019-05-09 16:50:25
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-04 18:57:35
 */

package znet

import (
	"fmt"
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
	if c.Engine.webMode > releaseCode {
		end := time.Now()
		latency := end.Sub(c.Info.StartTime)
		code := c.Code
		status := " " + strconv.Itoa(code) + " "
		switch {
		case code >= 200 && code <= 299:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorGreen, status)
		case code >= 300 && code <= 399:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorYellow, status)
		default:
			status = c.Log.ColorBackgroundWrap(zlog.ColorBlack, zlog.ColorRed, status)
		}
		ft := fmt.Sprintf("%s %15s %15v %%s  %%s", status, c.Log.ColorTextWrap(zlog.ColorWhite, c.ClientIP()), latency)
		c.Log.Success(showRouteDebug(c.Log, ft, c.Request.Method, c.Request.RequestURI))
	}
}
