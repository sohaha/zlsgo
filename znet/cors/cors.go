/*
 * @Author: seekwe
 * @Date:   2019-05-22 17:25:18
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-05 12:10:25
 */

package cors

import (
	"net/http"

	"github.com/sohaha/zlsgo/znet"
)

func Default() znet.HandlerFunc {
	return func(c *znet.Context) {
		if !applyCors(c) {
			c.Abort(http.StatusForbidden)
		}
		c.Next()
	}
}

func applyCors(c *znet.Context) bool {
	origin := c.GetHeader("Origin")
	if len(origin) == 0 {
		return true
	}
	host := c.GetHeader("Host")
	if origin == "http://"+host || origin == "https://"+host {
		return true
	}

	if c.Request.Method == "OPTIONS" {
		c.SetHeader("Access-Control-Allow-Origin", origin)
		c.Info.Code = http.StatusNoContent
		return false
	}

	c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.SetHeader("Access-Control-Allow-Credentials", "true")
	c.SetHeader("Access-Control-Allow-Headers", "X-Requested-With")
	c.SetHeader("Access-Control-Allow-Origin", origin)
	return true
}
