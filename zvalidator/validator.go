package zvalidator

import (
	"github.com/sohaha/zlsgo/zjson"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// PatternEmail is email
	PatternEmail = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	// PatternIP is ip
	PatternIP = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	// PatternURLSchema is URL schema
	PatternURLSchema = `((ftp|tcp|udp|wss?|https?):\/\/)`
	// PatternURLUsername is URL username
	PatternURLUsername = `(\S+(:\S*)?@)`
	// PatternURLPath is URL path
	PatternURLPath = `((\/|\?|#)[^\s]*)`
	// PatternURLPort is URL port
	PatternURLPort = `(:(\d{1,5}))`
	// PatternURLIP is URL ip
	PatternURLIP = `([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))`
	// PatternURLSubdomain is URL subdomain
	PatternURLSubdomain = `((www\.)|([a-zA-Z0-9]([-\.][-\._a-zA-Z0-9]+)*))`
	// PatternURL is URL
	PatternURL = `^` + PatternURLSchema + `?` + PatternURLUsername + `?` + `((` + PatternURLIP + `|(\[` + PatternIP + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + PatternURLSubdomain + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + PatternURLPort + `?` + PatternURLPath + `?$`
)

// IsIP checks if the string is valid IP.
func IsIP(str string) bool {
	return net.ParseIP(str) != nil
}

// IsIPv4 checks if the string is valid IPv4.
func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	if ip == nil {
		return false
	}
	return strings.Contains(str, ".")
}

// IsIPv6 checks if the string is valid IPv6.
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	if ip == nil {
		return false
	}
	return strings.Contains(str, ":")
}

// IsURL checks if the string is URL.
func IsURL(str string) bool {
	return regexp.MustCompile(PatternURL).MatchString(str)
}

// IsPort checks if a string represents a valid port.
func IsPort(str string) bool {
	if n, err := strconv.Atoi(str); err == nil && n > 0 && n < 65536 {
		return true
	}
	return false
}

// IsEmail checks if the string is email.
func IsEmail(str string) bool {
	return regexp.MustCompile(PatternEmail).MatchString(str)
}

// IsTime check if string is valid according to given format
func IsTime(str string, format string) bool {
	_, err := time.Parse(format, str)
	return err == nil
}

// IsJSON check if the string is valid JSON (note: uses json.Unmarshal).
func IsJSON(str string) bool {
	return zjson.Valid(str)
}
