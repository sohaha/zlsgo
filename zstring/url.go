package zstring

import (
	"net/url"
	"strings"
)

// UrlEncode url encode string, is + not %20
func UrlEncode(str string) string {
	return url.QueryEscape(str)
}

// UrlDecode url decode string
func UrlDecode(str string) (string, error) {
	return url.QueryUnescape(str)
}

// UrlRawEncode URL-encode according to RFC 3986.
func UrlRawEncode(str string) string {
	return strings.Replace(url.QueryEscape(str), "+", "%20", -1)
}

// UrlRawDecode Decode URL-encoded strings.
func UrlRawDecode(str string) (string, error) {
	return url.QueryUnescape(strings.Replace(str, "%20", "+", -1))
}
