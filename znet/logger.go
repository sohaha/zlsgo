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

	"github.com/sohaha/zlsgo/zutil"

	"github.com/sohaha/zlsgo/zlog"
)

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
