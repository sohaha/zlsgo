package znet

import (
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/sohaha/zlsgo/zdi"
)

// Host returns the current host with scheme (http/https).
// If full is true, it includes the request URL path.
func (c *Context) Host(full ...bool) string {
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "https"
		if c.Request.TLS == nil {
			scheme = "http"
		}
	}
	host := c.Request.Host
	if len(full) > 0 && full[0] {
		host += c.Request.URL.String()
	}
	return scheme + "://" + host
}

// CompletionLink ensures a URL is absolute by prepending the current host
// if the provided link is relative.
func (c *Context) CompletionLink(link string) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}
	finalLink := c.Host()
	if !strings.HasPrefix(link, "/") {
		finalLink = finalLink + "/"
	}
	finalLink = finalLink + link
	return finalLink
}

// IsWebsocket determines if the current request is a WebSocket upgrade request
// by checking the Connection and Upgrade headers.
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") &&
		strings.ToLower(c.GetHeader("Upgrade")) == "websocket" {
		return true
	}
	return false
}

// IsSSE determines if the current request is expecting Server-Sent Events
// by checking if the Accept header is 'text/event-stream'.
func (c *Context) IsSSE() bool {
	return strings.ToLower(c.GetHeader("Accept")) == "text/event-stream"
}

// IsAjax determines if the current request is an AJAX request
// by checking for the X-Requested-With header with value XMLHttpRequest.
func (c *Context) IsAjax() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

// GetClientIP returns the client's IP address by checking various headers
// and connection information. It attempts to determine the most accurate
// client IP, even when behind proxies.
func (c *Context) GetClientIP() string {
	r := c.mu.RLock()
	ip := c.ip
	c.mu.RUnlock(r)
	if ip == "" {
		c.mu.Lock()
		ips := getRemoteIP(c.Request)
		ip = clientPublicIP(c.Request, ips)
		if ip == "" {
			ip = clientIP(c.Request, ips)
		}
		if ip == "" {
			ip = RemoteIP(c.Request)
		}
		c.ip = ip
		c.mu.Unlock()
	}

	return ip
}

// GetHeader returns the value of the specified request header.
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header with the given key and value.
// If value is empty, the header will be removed.
func (c *Context) SetHeader(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	c.mu.Lock()
	if value == "" {
		delete(c.header, key)
	} else {
		c.header[key] = append(c.header[key], value)
	}
	c.mu.Unlock()
}

// write finalizes the response by writing headers and body data to the response writer.
// It handles content negotiation, status codes, and ensures headers are properly set.
func (c *Context) write() {
	if !c.done.CAS(false, true) {
		return
	}

	c.Next()

	data := c.PrevContent()
	// data.Code.CAS(0, http.StatusInternalServerError)

	for key, value := range c.header {
		for i := range value {
			header := value[i]
			if i == 0 {
				c.Writer.Header().Set(key, header)
			} else {
				c.Writer.Header().Add(key, header)
			}
		}
	}

	if c.Request == nil || c.Request.Context().Err() != nil {
		return
	}

	defer func() {
		if c.Engine.IsDebug() {
			requestLog(c)
		}
	}()

	code := int(data.Code.Load())
	if code == 0 {
		code = http.StatusOK
		data.Code.Store(int32(code))
	}
	size := len(data.Content)
	if size > 0 {
		c.Writer.Header().Set("Content-Length", strconv.Itoa(size))
		c.Writer.WriteHeader(code)
		_, err := c.Writer.Write(data.Content)
		if err != nil {
			c.Log.Error(err)
		}
		return
	}
	if code != 200 {
		c.Writer.WriteHeader(code)
	}
}

// Next executes all remaining middleware in the chain.
// Returns false if the middleware chain has been stopped with Abort().
func (c *Context) Next() bool {
	for {
		if c.stopHandle.Load() {
			return false
		}
		r := c.mu.RLock()
		n := len(c.middleware) > 0
		c.mu.RUnlock(r)
		if !n {
			return true
		}
		c.next()
	}
}

// next is an internal method that executes the next middleware in the chain.
// It's called by Next() and handles the middleware execution flow.
func (c *Context) next() {
	if c.stopHandle.Load() {
		return
	}
	c.mu.Lock()
	n := c.middleware[0]
	c.middleware = c.middleware[1:]
	c.mu.Unlock()
	err := n(c)
	if err != nil {
		c.renderError(c, err)
		c.Abort()
	}
}

// SetCookie sets an HTTP cookie with the given name and value.
// Optional maxAge parameter specifies the cookie's max age in seconds (0 = session cookie).
func (c *Context) SetCookie(name, value string, maxAge ...int) {
	a := 0
	if len(maxAge) > 0 {
		a = maxAge[0]
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   a,
	}
	c.Writer.Header().Add("Set-Cookie", cookie.String())
}

// GetCookie returns the value of the cookie with the given name.
// Returns an empty string if the cookie doesn't exist.
func (c *Context) GetCookie(name string) string {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return ""
	}
	v, _ := url.QueryUnescape(cookie.Value)
	return v
}

// GetReferer returns the Referer header of the request, which contains
// the URL of the page that linked to the current page.
func (c *Context) GetReferer() string {
	return c.Request.Header.Get("Referer")
}

// GetUserAgent returns the User-Agent header of the request, which identifies
// the client software originating the request.
func (c *Context) GetUserAgent() string {
	return c.Request.Header.Get("User-Agent")
}

// ContentType returns or sets the Content-Type header.
// If contentText is provided, it sets the Content-Type header.
// Otherwise, it returns the current Content-Type of the request.
func (c *Context) ContentType(contentText ...string) string {
	var content string
	if len(contentText) > 0 {
		content = contentText[0]
	} else {
		content = c.GetHeader("Content-Type")
	}
	for i := 0; i < len(content); i++ {
		char := content[i]
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

// WithValue stores a key-value pair in the context for sharing data
// between middleware and handlers. Returns the context for chaining.
func (c *Context) WithValue(key string, value interface{}) *Context {
	c.mu.Lock()
	c.customizeData[key] = value
	c.mu.Unlock()
	return c
}

// Value retrieves data stored in the context by key.
// It returns the value and a boolean indicating if the key exists.
// If the key doesn't exist and default values are provided, the first default is returned.
func (c *Context) Value(key string, def ...interface{}) (value interface{}, ok bool) {
	r := c.mu.RLock()
	value, ok = c.customizeData[key]
	if !ok && (len(def) > 0) {
		value = def[0]
	}
	c.mu.RUnlock(r)
	return
}

// MustValue retrieves data stored in the context by key, with simplified return.
// If the key doesn't exist and default values are provided, the first default is returned.
// Unlike Value(), this method only returns the value without the existence flag.
func (c *Context) MustValue(key string, def ...interface{}) (value interface{}) {
	value, _ = c.Value(key, def...)
	return
}

// Injector returns the dependency injection container for this context.
// This allows handlers to access shared services and dependencies.
func (c *Context) Injector() zdi.Injector {
	return c.injector
}

// FileAttachment serves a file as an attachment with the specified filename.
// This will prompt the browser to download the file rather than display it.
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+strings.Replace(filename, "\"", "\\\"", -1)+`"`)
	} else {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.Writer, c.Request, filepath)
}

// isASCII checks if a string contains only ASCII characters.
// Source: https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
