package znet

import (
	"errors"
	"math"
	"math/big"
	"net"
	"net/http"
	"strconv"
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

// IPToLong IPToLong
func IPToLong(ip string) (i uint, err error) {
	return NetIPToLong(net.ParseIP(ip))
}

// LongToIP LongToIP
func LongToIP(i uint) (string, error) {
	ip, err := LongToNetIP(i)
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}

// NetIPToLong NetIPToLong
func NetIPToLong(ip net.IP) (i uint, err error) {
	b := ip.To4()
	if b == nil {
		err = errors.New("invalid ipv4 format")
		return
	}

	i = uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
	return
}

// NetIPv6ToLong NetIPv6ToLong
func NetIPv6ToLong(ip net.IP) (*big.Int, error) {
	if ip == nil {
		return nil, errors.New("invalid ipv6 format")
	}
	IPv6Int := big.NewInt(0)
	IPv6Int.SetBytes(ip.To16())
	return IPv6Int, nil
}

// LongToNetIP LongToNetIP
func LongToNetIP(i uint) (ip net.IP, err error) {
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

// LongToNetIPv6 LongToNetIPv6
func LongToNetIPv6(i *big.Int) (ip net.IP, err error) {
	ip = i.Bytes()
	return
}

// IsIP IsIP
func IsIP(ip string) bool {
	address := net.ParseIP(ip)
	if address == nil {
		return false
	}
	return true
}

// GetIPv GetIPv
func GetIPv(s string) int {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return 4
		case ':':
			return 6
		}
	}
	return 0
}

// InNetwork InNetwork
func InNetwork(ip, network string) bool {
	_, n, err := net.ParseCIDR(network)
	if err != nil && IsIP(network) {
		_, n, err = net.ParseCIDR(network + "/24")
	}
	if err != nil {
		return false
	}
	netIP := net.ParseIP(ip)
	return n.Contains(netIP)
}

// Port GetPort Check if the port is available, if not, then automatically get an available
func Port(port int, change bool) (newPort int, err error) {
	host := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", host)
	if err != nil {
		if !change && port != 0 {
			return 0, err
		}
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			return 0, err
		}
	}
	defer listener.Close()
	addr := listener.Addr().String()
	_, portString, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(portString)
}

// MultiplePort Check if the multiple port is available, if not, then automatically get an available
func MultiplePort(ports []int, change bool) (int, error) {
	last := len(ports) - 1
	for k, v := range ports {
		c := false
		if last == k {
			c = change
		}
		n, err := Port(v, c)
		if err == nil {
			return n, nil
		}
	}
	return 0, errors.New("no available port")
}
