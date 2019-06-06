/*
 * @Author: seekwe
 * @Date:   2019-05-09 12:48:09
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-29 17:54:00
 */

package znet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestNetIp(t *testing.T) {
	T := zlsgo.NewTest(t)
	r := newServer()
	r.GET("/ip", func(c *Context) {
		T.Equal("", c.ClientIP())
		ip := "127.0.0.1"
		_, _ = IPString2Long("127")
		l, _ := IPString2Long(ip)
		T.Equal(uint(2130706433), l)
		i, _ := Long2IPString(l)
		T.Equal(ip, i)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ip", nil)
	r.ServeHTTP(w, req)
}
