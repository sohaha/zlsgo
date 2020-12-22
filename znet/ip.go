package znet

import (
	"errors"
	"math"
	"net"
	"net/http"
	"strings"
)

var localNetworks []*net.IPNet

func init() {
	localNetworks = make([]*net.IPNet, 4)
	for i, sNetwork := range []string{
		"10.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.168.0.0/16",
	} {
		_, network, _ := net.ParseCIDR(sNetwork)
		localNetworks[i] = network
	}
}

// IsLocalAddrIP IsLocalAddrIP
func IsLocalAddrIP(ip string) bool {
	return IsLocalIP(net.ParseIP(ip))
}

// IsLocalIP IsLocalIP
func IsLocalIP(ip net.IP) bool {
	for _, network := range localNetworks {
		if network.Contains(ip) {
			return true
		}
	}
	if ip.String() == "0.0.0.0" {
		return true
	}
	return ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast()
}

// ClientIP ClientIP
func ClientIP(r *http.Request) (ip string) {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip = strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return
	}
	ip, _, _ = net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	return
}

// ClientPublicIP ClientPublicIP
func ClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" && !IsLocalAddrIP(ip) {
			return ip
		}
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" && !IsLocalAddrIP(ip) {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil && !IsLocalAddrIP(ip) {
		return ip

	}

	return ""
}

// RemoteIP RemoteIP
func RemoteIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// IPString2Long IPString2Long
func IPString2Long(ip string) (i uint, err error) {
	b := net.ParseIP(ip).To4()
	if b == nil {
		err = errors.New("invalid ipv4 format")
		return
	}
	i = uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
	return
}

// Long2IPString Long2IPString
func Long2IPString(i uint) (IP string, err error) {
	if i > math.MaxUint32 {
		err = errors.New("beyond the scope of ipv4")
		return
	}

	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)

	IP = ip.String()
	return
}

// IP2Long IP2Long
func IP2Long(ip net.IP) (i uint, err error) {
	b := ip.To4()
	if b == nil {
		err = errors.New("invalid ipv4 format")
		return
	}

	i = uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
	return
}

// Long2IP Long2IP
func Long2IP(i uint) (ip net.IP, err error) {
	if i > math.MaxUint32 {
		err = errors.New("beyond the scope of ipv4")
		return
	}

	ip = make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)

	return
}
