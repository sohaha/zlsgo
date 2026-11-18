package znet

import (
	"net"
	"net/http"
	"strconv"
	"strings"
)

// ForwardedHostHeaders lists request headers that may carry the client's original Host, in order of precedence.
// By default, nginx replaces the Host header with $proxy_host (so Host becomes the upstream address).
// Configure "proxy_set_header X-Forwarded-Host $host;" on the reverse proxy to populate this header.
var ForwardedHostHeaders = []string{"X-Forwarded-Host"}

// ForwardedProtoHeaders lists request headers that may carry the client's original protocol (http/https).
var ForwardedProtoHeaders = []string{"X-Forwarded-Proto"}

// ForwardedHost returns the original Host forwarded by a trusted proxy (possibly including a port).
// To prevent Host header injection, it only returns a value when the request comes from a trusted proxy (see TrustedProxies) and the value is valid;
// otherwise it returns an empty string and callers should fall back to r.Host.
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

// ForwardedProto returns the original protocol forwarded by a trusted proxy, which is always http or https.
// It returns an empty string when no usable forwarding header is present.
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

// RequestHost returns the effective request Host, preferring the original host from forwarding headers in reverse-proxy scenarios.
func RequestHost(r *http.Request) string {
	if host := ForwardedHost(r); host != "" {
		return host
	}
	if r == nil {
		return ""
	}
	return r.Host
}

// RequestScheme returns the effective request scheme (http/https), preferring forwarding headers in reverse-proxy scenarios.
func RequestScheme(r *http.Request) string {
	if proto := ForwardedProto(r); proto != "" {
		return proto
	}
	if r != nil && r.TLS != nil {
		return "https"
	}
	return "http"
}

// isTrustedProxyRequest reports whether the request comes from a trusted proxy.
// For synthetic requests (such as unit tests) with no remote address, trust cannot be determined and the request is treated as trusted.
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

// forwardedHeaderHost returns the first valid host from a forwarding header.
// When multiple proxies are involved, the leftmost value is the original client-side value, consistent with frameworks such as Django.
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

// forwardedProtoValue returns the first valid protocol from a forwarding header.
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

// parseForwardedHeader parses an RFC 7239 Forwarded header and returns the first valid
// host and proto from left to right (leftmost elements are added by proxies closer to the client).
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

// normalizeForwardedHost sanitizes and validates a forwarded host, returning an empty string if invalid.
func normalizeForwardedHost(value string) string {
	value = strings.Trim(strings.TrimSpace(value), `"`)
	if !isValidForwardedHost(value) {
		return ""
	}
	return value
}

// isValidForwardedHost reports whether host[:port] is valid.
// It rejects whitespace, control characters, paths, query strings, and other characters that could be used for response splitting or URL injection.
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

	// Bare bracketed IPv6 address, e.g. [::1].
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return net.ParseIP(host[1:len(host)-1]) != nil
	}

	if strings.HasPrefix(host, "[") {
		_, port, err := net.SplitHostPort(host)
		return err == nil && isValidHostPort(port)
	}

	// Hostname or IPv4 with a port, e.g. example.com:8080.
	if h, port, err := net.SplitHostPort(host); err == nil {
		return h != "" && isValidHostPort(port)
	}

	// Bare IPv6 addresses must be bracketed; treat other values as hostnames.
	return !strings.ContainsRune(host, ':')
}

// isValidHostPort reports whether port is a decimal number in the range 1-65535.
func isValidHostPort(port string) bool {
	if port == "" {
		return false
	}
	n, err := strconv.Atoi(port)
	return err == nil && n > 0 && n <= 65535
}

// normalizeForwardedProto sanitizes and validates a forwarded protocol.
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
