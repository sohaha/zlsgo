/*
 * @Author: seekwe
 * @Date:   2019-05-10 17:05:54
 * @Last Modified by:   seekwe
 * @Last Modified time: 2020-03-06 19:07:15
 */

package znet

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/sohaha/zlsgo/zstring"
)

// Host Get the current Host
func (c *Context) Host() string {
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
	}

	return scheme + "://" + c.Request.Host
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
	c.Info.Mutex.Lock()
	if value == "" {
		delete(c.Info.heades, key)
	} else {
		c.Info.heades[key] = value
	}
	c.Info.Mutex.Unlock()
}

func (c *Context) done() {
	data := c.PrevContent()
	c.Info.Mutex.RLock()
	r := c.Info.render
	for key, value := range c.Info.heades {
		c.Writer.Header().Set(key, value)
	}
	c.Info.Mutex.RUnlock()
	if r != nil {
		c.Writer.WriteHeader(data.Code)
		_, err := c.Writer.Write(data.Content)
		if err != nil {
			// panic(err)
			c.Log.Error(err)
		}
	} else if data.Code != 0 && data.Code != 200 {
		c.Writer.WriteHeader(data.Code)
	}
}

func (c *Context) Next() (next HandlerFunc) {
	c.Info.Mutex.RLock()
	StopHandle := c.Info.StopHandle
	middlewareLen := len(c.Info.middleware)
	c.Info.Mutex.RUnlock()
	if !StopHandle {
		if middlewareLen > 0 {
			next = c.Info.middleware[0]
			c.Info.middleware = c.Info.middleware[1:]
			next(c)
			c.PrevContent()
		}
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
func (c *Context) ContentType() string {
	content := c.GetHeader("Content-Type")
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

// WithValue context sharing data
func (c *Context) WithValue(key string, value interface{}) *Context {
	c.Info.Mutex.Lock()
	c.Info.customizeData[key] = value
	c.Info.Mutex.Unlock()
	return c
}

// Value get context sharing data
func (c *Context) Value(key string, def ...interface{}) (value interface{}, ok bool) {
	c.Info.Mutex.RLock()
	value, ok = c.Info.customizeData[key]
	if !ok && (len(def) > 0) {
		value = def[0]
	}
	c.Info.Mutex.RUnlock()
	return
}
