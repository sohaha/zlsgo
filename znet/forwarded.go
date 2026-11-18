package znet

import (
	"net"
	"net/http"
	"strconv"
	"strings"
)

// ForwardedHostHeaders lists the request headers that may carry the original
// client host, in priority order. nginx overwrites Host with $proxy_host by
// default, so it only sends the original domain when configured with
// `proxy_set_header X-Forwarded-Host $host`.
var ForwardedHostHeaders = []string{"X-Forwarded-Host"}

// ForwardedProtoHeaders lists the request headers that may carry the original
// client scheme (http/https), in priority order. nginx terminates TLS itself,
// so it must name the original scheme explicitly: X-Forwarded-Proto is the
// conventional choice, while X-Scheme and X-Forwarded-Scheme cover proxies
// that already emit those names.
var ForwardedProtoHeaders = []string{"X-Forwarded-Proto", "X-Forwarded-Scheme", "X-Scheme"}

// ForwardedHost returns the original host (possibly with port) forwarded by a
// trusted proxy. To prevent host-header poisoning it returns an empty string
// unless the request comes from a trusted proxy (see TrustedProxies) and the
// value is valid; callers should then fall back to r.Host.
func ForwardedHost(r *http.Request) string {
	if r == nil || !isTrustedProxyRequest(r) {
		return ""
	}

	for _, key := range ForwardedHostHeaders {
		if host := forwardedHeaderHost(r.Header.Get(key)); host != "" {
			return host
		}
	}

	return parseForwardedHeader(r.Header.Get("Forwarded")).host
}

// ForwardedProto returns the original scheme (http or https) forwarded by a
// trusted proxy. It returns an empty string when no usable header is present.
func ForwardedProto(r *http.Request) string {
	if r == nil || !isTrustedProxyRequest(r) {
		return ""
	}

	for _, key := range ForwardedProtoHeaders {
		if proto := forwardedProtoValue(r.Header.Get(key)); proto != "" {
			return proto
		}
	}

	return parseForwardedHeader(r.Header.Get("Forwarded")).proto
}

// RequestHost returns the effective host of the request, preferring the
// original host carried by proxy headers when behind a reverse proxy. A
// forwarded local/private address never replaces a public host, so a proxy
// that forwards $proxy_host (or an upstream address) cannot downgrade a
// correct Host header.
func RequestHost(r *http.Request) string {
	forwarded := ForwardedHost(r)
	if r == nil {
		return forwarded
	}
	if forwarded == "" || forwarded == r.Host {
		return r.Host
	}
	if isLocalHostPort(forwarded) && !isLocalHostPort(r.Host) {
		return r.Host
	}
	return forwarded
}

// RequestScheme returns the effective scheme (http/https) of the request,
// preferring proxy headers when behind a reverse proxy.
func RequestScheme(r *http.Request) string {
	if proto := ForwardedProto(r); proto != "" {
		return proto
	}
	if r != nil && r.TLS != nil {
		return "https"
	}
	return "http"
}

// isTrustedProxyRequest reports whether the request comes from a trusted
// proxy. Synthetic requests (for example in unit tests) have no peer address,
// so they cannot be verified and are treated as trusted.
func isTrustedProxyRequest(r *http.Request) bool {
	if len(TrustedProxies) == 0 {
		return false
	}
	ip := RemoteIP(r)
	if ip == "" {
		return true
	}
	return isProxyTrusted(net.ParseIP(ip))
}

// isLocalHostPort reports whether host[:port] points at a local, loopback or
// private address (including the name localhost).
func isLocalHostPort(hostPort string) bool {
	if hostPort == "" {
		return false
	}

	host := hostPort
	if h, _, err := net.SplitHostPort(hostPort); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && IsLocalIP(ip)
}

// forwardedHeaderHost returns the first valid host found in a forwarded
// header. In a multi-proxy chain the leftmost value is the client-side
// original, matching the behavior of frameworks such as Django.
func forwardedHeaderHost(value string) string {
	if value == "" {
		return ""
	}

	for _, part := range strings.Split(value, ",") {
		if host := normalizeForwardedHost(part); host != "" {
			return host
		}
	}

	return ""
}

// forwardedProtoValue returns the first valid scheme in a forwarded header.
func forwardedProtoValue(value string) string {
	if value == "" {
		return ""
	}

	for _, part := range strings.Split(value, ",") {
		if proto := normalizeForwardedProto(part); proto != "" {
			return proto
		}
	}

	return ""
}

type forwardedHeader struct {
	host  string
	proto string
}

// parseForwardedHeader parses an RFC 7239 Forwarded header, taking the first
// valid host and proto from left to right (leftmost elements are written by
// the proxy closest to the client).
func parseForwardedHeader(value string) forwardedHeader {
	var f forwardedHeader
	if value == "" {
		return f
	}

	for _, element := range strings.Split(value, ",") {
		for _, part := range strings.Split(element, ";") {
			key, val, ok := strings.Cut(strings.TrimSpace(part), "=")
			if !ok {
				continue
			}

			switch strings.ToLower(strings.TrimSpace(key)) {
			case "host":
				if f.host == "" {
					f.host = normalizeForwardedHost(val)
				}
			case "proto":
				if f.proto == "" {
					f.proto = normalizeForwardedProto(val)
				}
			}
		}

		if f.host != "" && f.proto != "" {
			break
		}
	}

	return f
}

// normalizeForwardedHost trims and validates a host from a forwarded header,
// returning an empty string when it is invalid.
func normalizeForwardedHost(value string) string {
	value = strings.Trim(strings.TrimSpace(value), `"`)
	if !isValidForwardedHost(value) {
		return ""
	}
	return value
}

// isValidForwardedHost checks whether host[:port] is valid. It rejects
// whitespace, control characters, paths, query strings and any other
// characters that could be used for response splitting or URL pollution.
func isValidForwardedHost(host string) bool {
	if host == "" || len(host) > 255 {
		return false
	}

	for i := 0; i < len(host); i++ {
		ch := host[i]
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z', ch >= '0' && ch <= '9':
		case ch == '-', ch == '.', ch == '_', ch == ':', ch == '[', ch == ']':
		default:
			return false
		}
	}

	// Bare IPv6 address, such as [::1].
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return net.ParseIP(host[1:len(host)-1]) != nil
	}

	if strings.HasPrefix(host, "[") {
		_, port, err := net.SplitHostPort(host)
		return err == nil && isValidHostPort(port)
	}

	// Hostname or IPv4 address with a port, such as example.com:8080.
	if h, port, err := net.SplitHostPort(host); err == nil {
		return h != "" && isValidHostPort(port)
	}

	// A bare IPv6 address must be bracketed; anything else is a hostname.
	return !strings.ContainsRune(host, ':')
}

// isValidHostPort checks whether port is a decimal number in 1-65535.
func isValidHostPort(port string) bool {
	if port == "" {
		return false
	}
	n, err := strconv.Atoi(port)
	return err == nil && n > 0 && n <= 65535
}

// normalizeForwardedProto trims and validates a scheme from a forwarded
// header, returning an empty string when it is not http or https.
func normalizeForwardedProto(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "http":
		return "http"
	case "https":
		return "https"
	default:
		return ""
	}
}
