package znet

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

// utils provides utility functions for the znet package.
// It contains methods for URL matching, context handling, and other common operations.
type utils struct {
	// ContextKey is used to store and retrieve values from request context.
	ContextKey contextKeyType
}

// Utils is a global instance of the utils struct that provides
// utility functions for routing, context handling, and other common operations.
var Utils = utils{
	ContextKey: contextKeyType{},
}

// Pattern constants used for URL matching and parameter extraction
const (
	// defaultPattern matches any non-slash character sequence
	defaultPattern = `[^/]+`
	// idPattern matches numeric IDs
	idPattern = `[\d]+`
	// idKey is the parameter name for ID segments
	idKey = `id`
	// allPattern matches any character sequence including slashes
	allPattern = `.*`
	// allKey is the parameter name for wildcard segments
	allKey = `*`
)

// matchCache caches compiled route patterns to improve performance.
// It uses an LRU cache with a capacity of 100 entries.
var matchCache = zcache.NewFast(func(o *zcache.Options) {
	o.LRU2Cap = 100
})

// URLMatchAndParse checks if the request URL matches the route pattern and returns
// a map of the parsed path parameters. It uses a cache to improve performance for
// frequently accessed routes.
func (_ utils) URLMatchAndParse(requestURL string, path string) (matchParams map[string]string, ok bool) {
	var (
		pattern   string
		matchName []string
	)
	matchParams, ok = make(map[string]string), true
	if v, ok := matchCache.Get(path); ok {
		m := v.([]string)
		pattern = m[0]
		matchName = m[1:]
	} else {
		res := strings.Split(path, "/")
		pattern, matchName = parsePattern(res, "/")
		matchCache.Set(path, append([]string{pattern}, matchName...))
	}

	if pattern == "" {
		return nil, false
	}

	rr, err := zstring.RegexExtract(pattern, requestURL)
	if err != nil || len(rr) == 0 {
		return nil, false
	}

	if rr[0] == requestURL {
		rr = rr[1:]
		if len(matchName) != 0 {
			for k, v := range rr {
				if key := matchName[k]; key != "" {
					matchParams[key] = v
				}
			}
		}
		return
	}

	return nil, false
}

// parsePattern converts a path pattern into a regular expression and extracts
// parameter names. It handles various parameter formats including :param, *wildcard,
// and {name:pattern} syntax.
func parsePattern(res []string, prefix string) (string, []string) {
	var (
		matchName []string
		pattern   string
	)
	l := len(res)
	for i := 0; i < l; i++ {
		str := res[i]
		if str == "" {
			continue
		}
		if strings.HasSuffix(str, "\\") && i < l-1 {
			res[i+1] = str[:len(str)-1] + "/" + res[i+1]
			continue
		}
		pattern = pattern + prefix
		l := len(str) - 1
		i := strings.IndexRune(str, ')')
		i2 := strings.IndexRune(str, '(')
		firstChar := str[0]
		// TODO Need to optimize
		if i2 != -1 && i != -1 {
			r, err := regexp.Compile(str)
			if err != nil {
				return "", nil
			}
			names := r.SubexpNames()
			matchName = append(matchName, names[1:]...)
			pattern = pattern + str
		} else if firstChar == ':' {
			matchStr := str
			res := strings.Split(matchStr, ":")
			key := res[1]
			if key == "full" {
				key = allKey
			}
			matchName = append(matchName, key)
			if key == idKey {
				pattern = pattern + "(" + idPattern + ")"
			} else if key == allKey {
				pattern = pattern + "(" + allPattern + ")"
			} else {
				pattern = pattern + "(" + defaultPattern + ")"
			}
		} else if firstChar == '*' {
			pattern = pattern + "(" + allPattern + ")"
			matchName = append(matchName, allKey)
		} else {
			i := strings.IndexRune(str, '}')
			i2 := strings.IndexRune(str, '{')
			if i2 != -1 && i != -1 {
				if i == l && i2 == 0 {
					matchStr := str[1:l]
					res := strings.Split(matchStr, ":")
					matchName = append(matchName, res[0])
					pattern = pattern + "(" + res[1] + ")"
				} else {
					if i2 != 0 {
						p, m := parsePattern([]string{str[:i2]}, "")
						if p != "" {
							pattern = pattern + p
							matchName = append(matchName, m...)
						}
						str = str[i2:]
					}
					if i >= 0 {
						ni := i - i2
						if ni < 0 {
							return "", nil
						}
						matchStr := str[1:ni]
						res := strings.Split(matchStr, ":")
						matchName = append(matchName, res[0])
						pattern = pattern + "(" + res[1] + ")"
						p, m := parsePattern([]string{str[ni+1:]}, "")
						if p != "" {
							pattern = pattern + p
							matchName = append(matchName, m...)
						}
					} else {
						pattern = pattern + str
					}
				}
			} else {
				pattern = pattern + str
			}
		}
	}

	return pattern, matchName
}

// getAddr normalizes an address string, ensuring it has a port.
// If no port is specified or the port is 0, it finds an available port.
func getAddr(addr string) string {
	var port int
	if strings.Contains(addr, ":") {
		port, _ = strconv.Atoi(strings.Split(addr, ":")[1])
	} else {
		port, _ = strconv.Atoi(addr)
		addr = ":" + addr
	}
	if port != 0 {
		return addr
	}
	port, _ = Port(port, true)
	return ":" + strconv.Itoa(port)
}

// getHostname constructs a full URL with the appropriate scheme (http/https)
// based on whether TLS is enabled, and resolves the hostname from the address.
func getHostname(addr string, isTls bool) string {
	hostname := "http://"
	if isTls {
		hostname = "https://"
	}
	return hostname + resolveHostname(addr)
}

// TreeFind searches for a handler matching the given path in the routing tree.
// It returns the engine, handler function, middleware stack, and a boolean
// indicating whether a match was found.
func (u utils) TreeFind(t *Tree, path string) (*Engine, handlerFn, []handlerFn, bool) {
	nodes := t.Find(path, false)
	for i := range nodes {
		node := nodes[i]
		if node.handle != nil {
			if node.path == path {
				return node.engine, node.handle, node.middleware, true
			}
		}
	}

	if len(nodes) == 0 || strings.HasSuffix(path, "/") {
		res := strings.Split(path, "/")
		p := ""
		if len(res) == 1 {
			p = res[0]
		} else {
			p = res[1]
		}
		nodes := t.Find(p, true)
		for i := range nodes {
			if handler := nodes[i].handle; handler != nil && nodes[i].path != path {
				if matchParamsMap, ok := u.URLMatchAndParse(path, nodes[i].path); ok {
					return nodes[i].engine, func(c *Context) error {
						req := c.Request
						ctx := context.WithValue(req.Context(), u.ContextKey, matchParamsMap)
						c.Request = req.WithContext(ctx)
						return nodes[i].Handle()(c)
					}, nodes[i].middleware, true
				}
			}
		}
	}
	return nil, nil, nil, false
}

// CompletionPath ensures a path has the correct prefix and format.
// It adds the prefix if needed and ensures the path starts with a slash.
func (utils) CompletionPath(p, prefix string) string {
	suffix := strings.HasSuffix(p, "/")
	p = strings.TrimLeft(p, "/")
	prefix = strings.TrimRight(prefix, "/")
	path := zstring.TrimSpace(path.Join("/", prefix, p))

	if path == "" {
		path = "/"
	} else if suffix && path != "/" {
		path = path + "/"
	}

	return path
}

// IsAbort checks if request handling has been aborted for the given context.
// It returns true if the context's stopHandle flag is set.
func (utils) IsAbort(c *Context) bool {
	return c.stopHandle.Load()
}

// AppendHandler appends handlers to the context's middleware stack.
// Use with caution as this modifies the middleware chain during request processing.
func (utils) AppendHandler(c *Context, handlers ...Handler) {
	hl := len(handlers)
	if hl == 0 {
		return
	}

	for i := range handlers {
		c.middleware = append(c.middleware, Utils.ParseHandlerFunc(handlers[i]))
	}
}

// resolveAddr converts an address string and optional TLS configuration
// into an addrSt structure used for server configuration.
func resolveAddr(addrString string, tlsConfig ...TlsCfg) addrSt {
	cfg := addrSt{
		addr: addrString,
	}
	if len(tlsConfig) > 0 {
		cfg.Cert = tlsConfig[0].Cert
		cfg.HTTPAddr = tlsConfig[0].HTTPAddr
		cfg.HTTPProcessing = tlsConfig[0].HTTPProcessing
		cfg.Key = tlsConfig[0].Key
		cfg.Config = tlsConfig[0].Config
	}
	return cfg
}

// resolveHostname extracts or constructs a hostname from an address string.
// It handles various formats including IP addresses and port specifications.
func resolveHostname(addrString string) string {
	if strings.Index(addrString, ":") == 0 {
		return "127.0.0.1" + addrString
	}
	return addrString
}

// templateParse parses template files and applies the provided function map.
// It returns the parsed template or an error if parsing fails.
func templateParse(templateFile []string, funcMap template.FuncMap) (t *template.Template, err error) {
	if len(templateFile) == 0 {
		return nil, errors.New("template file cannot be empty")
	}
	file := templateFile[0]
	if len(file) <= 255 && zfile.FileExist(file) {
		for i := range templateFile {
			templateFile[i] = zfile.RealPath(templateFile[i])
		}
		t, err = template.ParseFiles(templateFile...)
		if err == nil && funcMap != nil {
			t.Funcs(funcMap)
		}
	} else {
		t = template.New("")
		if funcMap != nil {
			t.Funcs(funcMap)
		}
		t, err = t.Parse(file)
	}
	return
}

// tlsRedirectHandler implements http.Handler to redirect HTTP requests to HTTPS.
// tlsRedirectHandler implements http.Handler to redirect HTTP requests to HTTPS.
// It uses a 301 Moved Permanently status code for the redirection.
type tlsRedirectHandler struct {
	// Domain is the target domain for the HTTPS redirect
	Domain string
}

// ServeHTTP implements the http.Handler interface.
// It redirects HTTP requests to HTTPS using a 301 Moved Permanently status.
// ServeHTTP implements the http.Handler interface.
// It redirects HTTP requests to HTTPS using a 301 Moved Permanently status.
func (h *tlsRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.Domain+r.URL.String(), http.StatusMovedPermanently)
}

// NewContext creates a new Context instance for handling a request.
// This is used when you need to manually create a context outside the normal request flow.
func (e *Engine) NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:        w,
		Request:       req,
		Engine:        e,
		Log:           e.Log,
		Cache:         Cache,
		startTime:     time.Time{},
		header:        map[string][]string{},
		customizeData: map[string]interface{}{},
		stopHandle:    zutil.NewBool(false),
		done:          zutil.NewBool(false),
		prevData: &PrevData{
			Code: zutil.NewInt32(0),
			Type: ContentTypePlain,
		},
	}
}

// acquireContext gets a Context instance from the pool or creates a new one if the pool is empty.
// This is used internally to efficiently reuse Context objects.
func (e *Engine) acquireContext(w http.ResponseWriter, r *http.Request) *Context {
	c := e.pool.Get().(*Context)
	c.Engine = e
	c.Request = r
	c.Writer = w
	c.injector = zdi.New(c.Engine.injector)
	c.injector.Maps(c)
	c.startTime = time.Now()
	c.renderError = defErrorHandler()
	c.stopHandle.Store(false)
	c.done.Store(false)
	return c
}

// releaseContext returns a Context to the pool after it's been used.
// It resets the Context to its zero state before returning it to the pool.
func (e *Engine) releaseContext(c *Context) {
	c.prevData.Code.Store(0)
	c.mu.Lock()

	for k := range c.customizeData {
		delete(c.customizeData, k)
	}
	for k := range c.header {
		delete(c.header, k)
	}

	c.middleware = c.middleware[0:0]
	c.render = nil
	c.renderError = nil
	c.cacheJSON = nil
	c.cacheQuery = nil
	c.cacheForm = nil
	c.injector = nil
	c.rawData = nil
	c.Engine = nil
	c.ip = ""
	c.prevData.Content = c.prevData.Content[0:0]
	c.prevData.Type = ContentTypePlain
	c.mu.Unlock()
	e.pool.Put(c)
}

// GetAddr returns the address string of the server.
// This is used to display the server's listening address.
func (s *serverMap) GetAddr() string {
	return s.srv.Addr
}
