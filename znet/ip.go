package znet

import (
	"errors"
	"math"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/sohaha/zlsgo/zcache"
)

var (
	RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}
	TrustedProxies  = []string{"0.0.0.0/0"}
	LocalNetworks   = []string{"127.0.0.0/8", "10.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "172.0.0.0/8", "192.168.0.0/16"}
)

var (
	localNetworks     []*net.IPNet
	localNetworksOnce sync.Once
	proxiesCache      = zcache.NewFast(func(o *zcache.Options) {
		o.LRU2Cap = 25
	})
)

func getLocalNetworks() []*net.IPNet {
	localNetworksOnce.Do(func() {
		localNetworks = make([]*net.IPNet, 0, len(LocalNetworks))
		for _, sNetwork := range LocalNetworks {
			_, network, err := net.ParseCIDR(sNetwork)
			if err == nil {
				localNetworks = append(localNetworks, network)
			}
		}
	})
	return localNetworks
}

// IsLocalAddrIP IsLocalAddrIP
func IsLocalAddrIP(ip string) bool {
	return IsLocalIP(net.ParseIP(ip))
}

// IsLocalIP IsLocalIP
func IsLocalIP(ip net.IP) bool {
	for _, network := range getLocalNetworks() {
		if network.Contains(ip) {
			return true
		}
	}
	if ip.String() == "0.0.0.0" {
		return true
	}
	return ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast()
}

func getTrustedIP(r *http.Request, remoteIP string) string {
	ip := remoteIP
	for i := range RemoteIPHeaders {
		key := RemoteIPHeaders[i]
		val := r.Header.Get(key)
		ips := parseHeadersIP(val)
		for i := range ips {
			Log.Debug(i, ips[i])
		}
	}
	return ip
}

func getRemoteIP(r *http.Request) []string {
	ips := make([]string, 0)
	ip := RemoteIP(r)
	trusted := ip == ""
	if !trusted {
		if len(TrustedProxies) > 0 {
			netIP := net.ParseIP(ip)
			for i := range TrustedProxies {
				network, ok := proxiesCache.ProvideGet(TrustedProxies[i], func() (interface{}, bool) {
					n, err := netCIDR(TrustedProxies[i])
					if err != nil {
						return nil, false
					}
					return n, true
				})

				if ok && network.(*net.IPNet).Contains(netIP) {
					trusted = true
					break
				}
			}
		} else {
			trusted = true
		}
	}

	if trusted {
		for i := range RemoteIPHeaders {
			key := RemoteIPHeaders[i]
			val := r.Header.Get(key)
			ips = append(ips, parseHeadersIP(val)...)
		}
	}
	return append(ips, ip)
}

func parseHeadersIP(val string) []string {
	if val == "" {
		return []string{}
	}
	str := strings.Split(val, ",")
	l := len(str)
	ips := make([]string, l)
	for i := l - 1; i >= 0; i-- {
		ips[l-1-i] = strings.TrimSpace(str[i])
	}
	return ips
}

// ClientIP ClientIP
func ClientIP(r *http.Request) (ip string) {
	ips := getRemoteIP(r)
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}
	return
}

// ClientPublicIP ClientPublicIP
func ClientPublicIP(r *http.Request) string {
	var ip string
	ips := getRemoteIP(r)
	for i := range ips {
		ip = ips[i]
		if ip != "" && !IsLocalAddrIP(ip) {
			return ip
		}
	}
	return ""
}

// RemoteIP Remote IP
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
	return net.ParseIP(ip) != nil
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

func netCIDR(network string) (*net.IPNet, error) {
	_, n, err := net.ParseCIDR(network)
	if err != nil && IsIP(network) {
		_, n, err = net.ParseCIDR(network + "/24")
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

// InNetwork InNetwork
func InNetwork(ip, network string) bool {
	netIP := net.ParseIP(ip)
	n, err := netCIDR(network)
	if err != nil {
		return false
	}
	return n.Contains(netIP)
}

// Port GetPort Check if the port is available, if not, then automatically get an available
func Port(port int, change bool) (newPort int, err error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
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
