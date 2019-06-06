/*
 * @Author: seekwe
 * @Date:   2019-05-29 11:53:43
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-29 17:43:56
 */

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
	localNetworks = make([]*net.IPNet, 19)
	for i, sNetwork := range []string{
		"10.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"172.17.0.0/12",
		"172.18.0.0/12",
		"172.19.0.0/12",
		"172.20.0.0/12",
		"172.21.0.0/12",
		"172.22.0.0/12",
		"172.23.0.0/12",
		"172.24.0.0/12",
		"172.25.0.0/12",
		"172.26.0.0/12",
		"172.27.0.0/12",
		"172.28.0.0/12",
		"172.29.0.0/12",
		"172.30.0.0/12",
		"172.31.0.0/12",
		"192.168.0.0/16",
	} {
		_, network, _ := net.ParseCIDR(sNetwork)
		localNetworks[i] = network
	}
}

func HasLocalIPddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

func HasLocalIP(ip net.IP) bool {
	for _, network := range localNetworks {
		if network.Contains(ip) {
			return true
		}
	}

	return ip.IsLoopback()
}

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

func ClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" && !HasLocalIPddr(ip) {
			return ip
		}
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" && !HasLocalIPddr(ip) {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		if !HasLocalIPddr(ip) {
			return ip
		}
	}

	return ""
}

func RemoteIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

func IPString2Long(ip string) (i uint, err error) {
	b := net.ParseIP(ip).To4()
	if b == nil {
		err = errors.New("invalid ipv4 format")
		return
	}
	i = uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
	return
}

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

func IP2Long(ip net.IP) (i uint, err error) {
	b := ip.To4()
	if b == nil {
		err = errors.New("invalid ipv4 format")
		return
	}

	i = uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
	return
}

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
