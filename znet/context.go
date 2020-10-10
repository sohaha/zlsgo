package znet

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
)

// Host Get the current Host
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

// CompletionLink Complete the link and add the current domain name if it is not linked
func (c *Context) CompletionLink(link string) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}
	finalLink := zstring.Buffer()
	finalLink.WriteString(c.Host())
	if !strings.HasPrefix(link, "/") {
		finalLink.WriteString("/")
	}
	finalLink.WriteString(link)
	return finalLink.String()
}

// IsWebsocket Is Websocket
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") &&
		strings.ToLower(c.GetHeader("Upgrade")) == "websocket" {
		return true
	}
	return false
}

// IsAjax IsAjax
func (c *Context) IsAjax() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

// GetClientIP Client IP
func (c *Context) GetClientIP() (IP string) {
	IP = ClientPublicIP(c.Request)
	if IP == "" {
		IP = ClientIP(c.Request)
	}
	return
}

// GetHeader Get Header
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader Set Header
func (c *Context) SetHeader(key, value string) {
	c.Lock()
	if value == "" {
		delete(c.header, key)
	} else {
		c.header[key] = value
	}
	c.Unlock()
}

func (c *Context) done() {
	data := c.PrevContent()
	r := c.render
	for key, value := range c.header {
		c.Writer.Header().Set(key, value)
	}
	if r != nil {
		c.Writer.WriteHeader(data.Code)
		_, err := c.Writer.Write(data.Content)
		if err != nil {
			c.Log.Error(err)
		}
	} else if data.Code != 0 && data.Code != 200 {
		c.Writer.WriteHeader(data.Code)
	}
}

func (c *Context) Next() (next HandlerFunc) {
	c.RLock()
	if !c.stopHandle && len(c.middleware) > 0 {
		next = c.middleware[0]
		c.middleware = c.middleware[1:]
		c.RUnlock()
		next(c)
		// c.PrevContent()
	} else {
		c.RUnlock()
	}

	return
}

// SetCookie Set Cookie
func (c *Context) SetCookie(name, value string, maxAge ...int) {
	_maxAge := 0
	if len(maxAge) > 0 {
		_maxAge = maxAge[0]
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   _maxAge,
	}
	c.Writer.Header().Add("Set-Cookie", cookie.String())
}

// GetCookie Get Cookie
func (c *Context) GetCookie(name string) string {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return ""
	}
	v, _ := url.QueryUnescape(cookie.Value)
	return v
}

// GetReferer request referer
func (c *Context) GetReferer() string {
	return c.Request.Header.Get("Referer")
}

// GetUserAgent http request UserAgent
func (c *Context) GetUserAgent() string {
	return c.Request.Header.Get("User-Agent")
}

// ContentType returns the Content-Type header of the request
func (c *Context) ContentType(contentText ...string) string {
	var content string
	if len(contentText) == 1 {
		content = contentText[0]
	} else {
		content = c.GetHeader("Content-Type")
	}
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

// WithValue context sharing data
func (c *Context) WithValue(key string, value interface{}) *Context {
	c.Lock()
	c.customizeData[key] = value
	c.Unlock()
	return c
}

// Value get context sharing data
func (c *Context) Value(key string, def ...interface{}) (value interface{}, ok bool) {
	c.RLock()
	value, ok = c.customizeData[key]
	if !ok && (len(def) > 0) {
		value = def[0]
	}
	c.RUnlock()
	return
}
