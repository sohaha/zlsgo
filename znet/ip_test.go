package znet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo/ztype"

	"github.com/sohaha/zlsgo"
)

func TestNetIP(t *testing.T) {
	T := zlsgo.NewTest(t)
	r := newServer()
	r.GET("/ip", func(c *Context) {
		T.Equal("", c.GetClientIP())
		ip := "127.0.0.1"
		ipb := uint(2130706433)
		_, _ = IPString2Long("127")
		l, _ := IPString2Long(ip)
		T.Equal(ipb, l)
		i, _ := Long2IPString(l)
		T.Equal(ip, i)
		ip2P, _ := Long2IP(l)
		T.Equal(ip, ztype.ToString(ip2P))
		ip2L, _ := IP2Long(ip2P)
		T.Equal(ipb, ip2L)
		T.Equal(true, IsLocalAddrIP(ip))
		t.Log(RemoteIP(c.Request))
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ip", nil)
	r.ServeHTTP(w, req)
}

func TestLocalAddrIP(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Log(IsLocalAddrIP("127.0.0.1"))
	tt.Log(IsLocalAddrIP("192.168.3.1"))
	tt.Log(IsLocalAddrIP("172.31.255.255"))
	tt.Log(IsLocalAddrIP("0.0.0.0"))
	tt.Log(IsLocalAddrIP("58.247.214.47"))
	tt.Log(IsLocalAddrIP("11.11.11.11"))
}
