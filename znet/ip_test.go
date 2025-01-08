package znet

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztype"
)

func TestNetIP(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newServer()
	r.GET("/ip", func(c *Context) {
		t.Equal("", c.GetClientIP())
		ip := "127.0.0.1"
		ipb := uint(2130706433)
		_, _ = IPToLong("127")
		l, _ := IPToLong(ip)
		t.Equal(ipb, l)
		i, _ := LongToIP(l)
		t.Equal(ip, i)
		ip2P, _ := LongToNetIP(l)
		t.Equal(ip, ztype.ToString(ip2P))
		ip2L, _ := NetIPToLong(ip2P)
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

func TestGetIPV(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.EqualExit(4, GetIPv("127.0.0.1"))
	t.EqualExit(4, GetIPv("172.31.255.255"))
	t.EqualExit(6, GetIPv("2001:db8:1:2::1"))
	t.EqualTrue(GetIPv("2001:db8:1:2::1") != 4)
}

func TestIP2(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	ip4Str := "127.0.0.1"
	ip4 := net.ParseIP(ip4Str)
	ipLong, _ := NetIPToLong(ip4)
	ip4Raw, _ := LongToNetIP(ipLong)
	t.EqualExit(ip4Str, ip4Raw.String())

	ip6Str := "2001:db8:1:2::1"
	ip6 := net.ParseIP(ip6Str)
	ipv6Long, _ := NetIPv6ToLong(ip6)
	ip6Raw, _ := LongToNetIPv6(ipv6Long)
	t.EqualExit(ip6Str, ip6Raw.String())

	t.Log(NetIPToLong(net.ParseIP("127.0.0.1")))
	t.Log(NetIPToLong(net.ParseIP("::1")))
}

func TestInNetwork(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	sNetwork := "120.85.5.131/24"
	for k, v := range map[string]bool{
		"120.85.5.1":   true,
		"120.85.5.255": true,
		"120.85.5.256": false,
		"120.85.8.131": false,
	} {
		t.Equal(v, InNetwork(k, sNetwork))
	}
}

func TestGetPort(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	p := 3780

	port, err := Port(0, true)
	t.EqualNil(err)
	t.EqualTrue(port != p)
	tt.Log(port)

	port, err = Port(p, true)
	t.EqualNil(err)

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

func Test_parseHeadersIP(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tests := []struct {
		args string
		want []string
	}{
		{"", []string{}},
		{"11.11.11.11,1.1.1.1, 2.2.2.2", []string{
			"2.2.2.2",
			"1.1.1.1",
			"11.11.11.11",
		}},
	}
	for _, v := range tests {
		tt.EqualExit(v.want, parseHeadersIP(v.args))
	}
}

func TestIsLocalAddrIP(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tests := []struct {
		args string
		want bool
	}{
		{"127.0.0.1", true},
		{"192.168.3.199", true},
		{"18.22.1.3", false},
	}
	for _, v := range tests {
		tt.EqualExit(v.want, IsLocalAddrIP(v.args))
	}

	request, _ := http.NewRequest("POST", "/", nil)
	request.Header.Set("X-Forwarded-For", "  20.20.20.20, 30.30.30.30,10.10.10.10")
	remoteIP := "10.10.10.10"
	t.Log(getTrustedIP(request, remoteIP))
	t.Log(RemoteIP(request))
}
