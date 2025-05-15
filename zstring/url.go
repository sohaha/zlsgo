package zstring

import (
	"net/url"
	"strings"
)

// UrlEncode encodes a string for use in a URL query.
// It uses the standard encoding where spaces are converted to '+' characters.
func UrlEncode(str string) string {
	return url.QueryEscape(str)
}

// UrlDecode decodes a URL-encoded string.
// It handles the conversion of '+' characters back to spaces.
func UrlDecode(str string) (string, error) {
	return url.QueryUnescape(str)
}

// UrlRawEncode encodes a string according to RFC 3986.
// Unlike UrlEncode, it converts spaces to '%20' instead of '+'.
func UrlRawEncode(str string) string {
	return strings.Replace(url.QueryEscape(str), "+", "%20", -1)
}

// UrlRawDecode decodes a string that was encoded according to RFC 3986.
// It handles the conversion of '%20' sequences to spaces.
func UrlRawDecode(str string) (string, error) {
	return url.QueryUnescape(strings.Replace(str, "%20", "+", -1))
}
