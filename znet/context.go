/*
 * @Author: seekwe
 * @Date:   2019-05-10 17:05:54
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-06 16:13:28
 */

package znet

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/sohaha/zlsgo/zvalidator"
)

// Host Get the current Host
func (c *Context) Host() string {
	scheme := "http://"
	if c.Request.TLS != nil {
		scheme = "https://"
	}
	return scheme + c.Request.Host
}

// CompletionLink Complete the link and add the current domain name if it is not linked
func (c *Context) CompletionLink(link string) string {
	if isURL := zvalidator.IsURL(link); isURL {
		return link
	}
	if !strings.HasPrefix(link, "/") {
		link = "/" + link
	}
	return c.Host() + link
}

// IsWebsocket Is Websocket
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") &&
		strings.ToLower(c.GetHeader("Upgrade")) == "websocket" {
		return true
	}
	return false
}

// ClientIP Client IP
func (c *Context) ClientIP() (IP string) {
	IP = ClientPublicIP(c.Request)
	if IP == "" {
		IP = ClientIP(c.Request)
	}
	return
}

// GetHeader  Get Header
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader Set Header
func (c *Context) SetHeader(key, value string) {
	if value == "" {
		c.Writer.Header().Del(key)
	} else {
		c.Writer.Header().Set(key, value)
	}
}

// Next Execute the next Handler Func
func (c *Context) Next() (next HandlerFunc) {
	if !c.Info.StopHandle {
		middlewareLen := len(c.Info.middleware)
		if middlewareLen > 0 {
			next = c.Info.middleware[0]
			c.Info.middleware = c.Info.middleware[1:]
			next(c)
		}
	}
	return
}

// SetCookie Set Cookie
func (c *Context) SetCookie(name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   0,
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

// Referer request referer.
func (c *Context) Referer() string {
	return c.Request.Header.Get("Referer")
}

// UserAgent http request UserAgent
func (c *Context) UserAgent() string {
	return c.Request.Header.Get("User-Agent")
}
