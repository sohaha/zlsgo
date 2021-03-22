package znet

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/sohaha/zlsgo/ztype"

	"github.com/sohaha/zlsgo"
)

func TestNetIP(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newServer()
	r.GET("/ip", func(c *Context) {
		t.Equal("", c.GetClientIP())
		ip := "127.0.0.1"
		ipb := uint(2130706433)
		_, _ = IPString2Long("127")
		l, _ := IPString2Long(ip)
		t.Equal(ipb, l)
		i, _ := Long2IPString(l)
		t.Equal(ip, i)
		ip2P, _ := Long2IP(l)
		t.Equal(ip, ztype.ToString(ip2P))
		ip2L, _ := IP2Long(ip2P)
		t.Equal(ipb, ip2L)
		t.Equal(true, IsLocalAddrIP(ip))
		tt.Log(RemoteIP(c.Request))
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ip", nil)
	r.ServeHTTP(w, req)
}

func TestLocalAddrIP(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Log(IsLocalAddrIP("127.0.0.1"))
	t.Log(IsLocalAddrIP("192.168.3.1"))
	t.Log(IsLocalAddrIP("172.31.255.255"))
	t.Log(IsLocalAddrIP("0.0.0.0"))
	t.Log(IsLocalAddrIP("58.247.214.47"))
	t.Log(IsLocalAddrIP("11.11.11.11"))
}

func TestIsIP(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.EqualTrue(IsIP("127.0.0.1"))
	t.EqualTrue(IsIP("172.31.255.255"))
	t.EqualTrue(!IsIP("172.31.255.a"))
}

func TestGetPort(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	p := 3780
	port, err := Port(p, true)
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	t.EqualNil(err)
	defer l.Close()

	port, err = Port(p, true)
	t.EqualNil(err)
	t.EqualTrue(port != p)
	tt.Log(port)

	port, err = Port(p, false)
	t.EqualTrue(err != nil)
	tt.Log(port, err)

	port, err = MultiplePort([]int{p, 1234}, false)
	t.EqualNil(err)
	t.Equal(1234, port)

	port, err = MultiplePort([]int{p}, false)
	t.EqualTrue(err != nil)
	tt.Log(port, err)
}
