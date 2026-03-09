package znet

import (
	"net/http"
	"net/url"

	"github.com/sohaha/zlsgo/zdi"
)

// Clone creates a request-scoped copy of the context for isolated handler execution.
// Mutable fields are deep-copied so the clone can be used safely in another goroutine.
func (c *Context) Clone(w http.ResponseWriter, req *http.Request) *Context {
	if req == nil {
		req = c.Request
	}
	if w == nil {
		w = c.Writer
	}

	clone := c.Engine.NewContext(w, req)
	clone.Log = c.Log
	clone.Cache = c.Cache
	clone.renderError = c.renderError
	clone.render = c.render
	clone.startTime = c.startTime
	clone.ip = c.ip
	clone.rawData = append(clone.rawData[:0], c.rawData...)
	clone.middleware = append(clone.middleware, c.middleware...)
	clone.prevData.Code.Store(c.prevData.Code.Load())
	clone.prevData.Type = c.prevData.Type
	clone.prevData.Content = append(clone.prevData.Content[:0], c.prevData.Content...)

	for k, v := range c.header {
		clone.header[k] = append([]string(nil), v...)
	}
	for k, v := range c.customizeData {
		clone.customizeData[k] = v
	}
	if c.cacheQuery != nil {
		clone.cacheQuery = cloneValues(c.cacheQuery)
	}
	if c.cacheForm != nil {
		clone.cacheForm = cloneValues(c.cacheForm)
	}
	if c.injector != nil {
		clone.injector = zdi.New(c.injector)
		clone.injector.Maps(clone)
	}

	return clone
}

func cloneValues(values url.Values) url.Values {
	cloned := make(url.Values, len(values))
	for k, v := range values {
		cloned[k] = append([]string(nil), v...)
	}
	return cloned
}

// CopyResponse copies the prepared response state from another context.
func (c *Context) CopyResponse(from *Context) {
	data := from.PrevContent()

	c.mu.Lock()
	for k := range c.header {
		delete(c.header, k)
	}
	for k, v := range from.header {
		c.header[k] = append([]string(nil), v...)
	}
	c.mu.Unlock()

	c.prevData.Code.Store(data.Code.Load())
	c.prevData.Type = data.Type
	c.prevData.Content = append(c.prevData.Content[:0], data.Content...)
}
