package znet

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
)

// anyMethod is a special method name that matches all HTTP methods.
const anyMethod = "ANY"

var (
	// ErrGenerateParameters is returned when generating a route withRequestLog wrong parameters.
	ErrGenerateParameters = errors.New("params contains wrong parameters")

	// ErrNotFoundRoute is returned when generating a route that can not find route in tree.
	ErrNotFoundRoute = errors.New("cannot find route in tree")

	// ErrNotFoundMethod is returned when generating a route that can not find method in tree.
	ErrNotFoundMethod = errors.New("cannot find method in tree")

	// ErrPatternGrammar is returned when generating a route that pattern grammar error.
	ErrPatternGrammar = errors.New("pattern grammar error")

	methods = map[string]struct{}{
		http.MethodGet:     {},
		http.MethodPost:    {},
		http.MethodPut:     {},
		http.MethodDelete:  {},
		http.MethodPatch:   {},
		http.MethodHead:    {},
		http.MethodOptions: {},
		http.MethodConnect: {},
		http.MethodTrace:   {},
		anyMethod:          {},
	}
	methodsKeys = make([]string, 0, len(methods))
)

// init initializes the methodsKeys slice with all supported HTTP methods.
// This is used for method validation and iteration over supported methods.
func init() {
	for k := range methods {
		methodsKeys = append(methodsKeys, k)
	}
}

type (
	// contextKeyType Private Value Structure for Each Request
	contextKeyType struct{}
)

// temporarilyTurnOffTheLog temporarily disables logging and returns a function
// that restores the previous log level when called.
func temporarilyTurnOffTheLog(e *Engine, msg string) func() {
	mode := e.webMode
	e.webMode = prodCode
	return func() {
		e.webMode = mode
		if e.IsDebug() {
			e.Log.Debug(msg)
		}
	}
}

// toHTTPError converts file system errors to appropriate HTTP responses.
// It handles common errors like file not found and permission denied.
func (c *Context) toHTTPError(err error) {
	if errors.Is(err, fs.ErrNotExist) {
		c.String(http.StatusNotFound, "404 page not found")
		return
	}
	if errors.Is(err, fs.ErrPermission) {
		c.String(http.StatusForbidden, "403 Forbidden")
		return
	}
	c.String(http.StatusInternalServerError, "500 Internal Server Error")
}

// StaticFS serves files from the given file system at the specified path.
// It registers GET, HEAD, and OPTIONS handlers for the specified path and its subdirectories.
func (e *Engine) StaticFS(relativePath string, fs http.FileSystem, moreHandler ...Handler) {
	var urlPattern string

	ap := Utils.CompletionPath(relativePath, e.router.prefix)
	f := fmt.Sprintf("%%s %%-40s -> %s/", zfile.SafePath(fmt.Sprintf("%s", fs)))
	if e.webMode == testCode {
		f = "%s %-40s"
	}
	log := temporarilyTurnOffTheLog(e, routeLog(e.Log, f, "FILE", ap))
	handler := func(c *Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, relativePath)
		f, err := fs.Open(p)
		if err != nil {
			c.toHTTPError(err)
			return
		}

		defer f.Close()

		fileInfo, err := f.Stat()
		if err != nil {
			c.toHTTPError(err)
			return
		}

		if !isModified(c, fileInfo.ModTime()) {
			return
		}

		c.prevData.Content, err = io.ReadAll(f)
		if err != nil {
			c.toHTTPError(err)
			return
		}

		c.prevData.Type = zfile.GetMimeType(p, c.prevData.Content)
		c.SetContentType(c.prevData.Type)
		c.prevData.Code.Store(http.StatusOK)
	}
	if strings.HasSuffix(relativePath, "/") {
		urlPattern = path.Join(relativePath, "*")
		e.GET(relativePath, handler, moreHandler...)
	} else {
		urlPattern = path.Join(relativePath, "/*")
		e.GET(relativePath, func(c *Context) {
			c.Redirect(relativePath + "/")
		}, moreHandler...)
		e.GET(relativePath+"/", handler, moreHandler...)
	}
	e.GET(urlPattern, handler, moreHandler...)
	e.HEAD(urlPattern, handler, moreHandler...)
	e.OPTIONS(urlPattern, handler, moreHandler...)
	log()
}

// Static serves files from the given root directory at the specified path.
// This is a convenience wrapper around StaticFS with http.Dir.
func (e *Engine) Static(relativePath, root string, moreHandler ...Handler) {
	e.StaticFS(relativePath, http.Dir(root), moreHandler...)
}

// StaticFile serves a single file at the specified path.
// It registers GET, HEAD, and OPTIONS handlers for the specified path.
func (e *Engine) StaticFile(relativePath, filepath string, moreHandler ...Handler) {
	handler := func(c *Context) {
		c.File(filepath)
	}

	tip := routeLog(e.Log, "%s %-40s -> "+zfile.SafePath(filepath)+"/", "FILE", relativePath)
	if e.webMode == testCode {
		tip = routeLog(e.Log, "%s %-40s", "FILE", relativePath)
	}
	log := temporarilyTurnOffTheLog(e, tip)
	e.GET(relativePath, handler, moreHandler...)
	e.HEAD(relativePath, handler, moreHandler...)
	e.OPTIONS(relativePath, handler, moreHandler...)
	log()
}

// Any registers a handler for all HTTP methods on the given path.
// This is a shortcut for registering the same handler under all methods.
func (e *Engine) Any(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(anyMethod, path, action, moreHandler...)
}

// Customize registers a handler for a custom HTTP method on the given path.
// The method string is converted to uppercase before registration.
func (e *Engine) Customize(method, path string, action Handler, moreHandler ...Handler) *Engine {
	method = strings.ToUpper(method)
	return e.Handle(method, path, action, moreHandler...)
}

// GET registers a handler for HTTP GET requests on the given path.
func (e *Engine) GET(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodGet, path, action, moreHandler...)
}

// POST registers a handler for HTTP POST requests on the given path.
func (e *Engine) POST(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodPost, path, action, moreHandler...)
}

// DELETE registers a handler for HTTP DELETE requests on the given path.
func (e *Engine) DELETE(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodDelete, path, action, moreHandler...)
}

// PUT registers a handler for HTTP PUT requests on the given path.
func (e *Engine) PUT(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodPut, path, action, moreHandler...)
}

// PATCH registers a handler for HTTP PATCH requests on the given path.
func (e *Engine) PATCH(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodPatch, path, action, moreHandler...)
}

// HEAD registers a handler for HTTP HEAD requests on the given path.
func (e *Engine) HEAD(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodHead, path, action, moreHandler...)
}

// OPTIONS registers a handler for HTTP OPTIONS requests on the given path.
func (e *Engine) OPTIONS(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodOptions, path, action, moreHandler...)
}

// CONNECT registers a handler for HTTP CONNECT requests on the given path.
func (e *Engine) CONNECT(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodConnect, path, action, moreHandler...)
}

// TRACE registers a handler for HTTP TRACE requests on the given path.
func (e *Engine) TRACE(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodTrace, path, action, moreHandler...)
}

// GETAndName registers a named handler for HTTP GET requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) GETAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.GET(path, action)
}

// POSTAndName registers a named handler for HTTP POST requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) POSTAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.POST(path, action)
}

// DELETEAndName registers a named handler for HTTP DELETE requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) DELETEAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.DELETE(path, action)
}

// PUTAndName registers a named handler for HTTP PUT requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) PUTAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.PUT(path, action)
}

// PATCHAndName registers a named handler for HTTP PATCH requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) PATCHAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.PATCH(path, action)
}

// HEADAndName registers a named handler for HTTP HEAD requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) HEADAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.HEAD(path, action)
}

// OPTIONSAndName registers a named handler for HTTP OPTIONS requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) OPTIONSAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.OPTIONS(path, action)
}

// CONNECTAndName registers a named handler for HTTP CONNECT requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) CONNECTAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.CONNECT(path, action)
}

// TRACEAndName registers a named handler for HTTP TRACE requests on the given path.
// The route name can be used later with GenerateURL to generate URLs for this route.
func (e *Engine) TRACEAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.TRACE(path, action)
}

// Group creates a new router group with the given prefix.
// All routes registered within the group will have the prefix prepended.
// This is useful for organizing routes by feature or area of responsibility.
func (e *Engine) Group(prefix string, groupHandle ...func(e *Engine)) (engine *Engine) {
	if prefix == "" {
		return e
	}
	rprefix := e.router.prefix
	if rprefix != "" {
		prefix = Utils.CompletionPath(prefix, rprefix)
	}
	middleware := make([]handlerFn, len(e.router.middleware))
	copy(middleware, e.router.middleware)
	route := &router{
		prefix:     prefix,
		trees:      e.router.trees,
		middleware: middleware,
		notFound:   e.router.notFound,
	}
	engine = &Engine{
		router:              route,
		views:               e.views,
		webMode:             e.webMode,
		webModeName:         e.webModeName,
		MaxMultipartMemory:  e.MaxMultipartMemory,
		customMethodType:    e.customMethodType,
		Log:                 e.Log,
		BindStructCase:      e.BindStructCase,
		BindStructDelimiter: e.BindStructDelimiter,
		BindStructSuffix:    e.BindStructSuffix,
		templateFuncMap:     e.templateFuncMap,
		template:            e.template,
		injector:            e.injector,
	}
	engine.pool.New = func() interface{} {
		return e.NewContext(nil, nil)
	}
	if len(groupHandle) > 0 {
		groupHandle[0](engine)
	}
	return
}

// GenerateURL generates a URL for a named route with the given parameters.
// This is useful for creating links to other routes in your application.
// It returns an error if the route name doesn't exist or if required parameters are missing.
func (e *Engine) GenerateURL(method string, routeName string, params map[string]string) (string, error) {
	tree, ok := e.router.trees[method]
	if !ok {
		return "", ErrNotFoundMethod
	}

	route, ok := tree.routes[routeName]
	if !ok {
		return "", ErrNotFoundRoute
	}

	ps := strings.Split(route.path, "/")
	l := len(ps)
	segments := make([]string, 0, l)
	for i := 0; i < l; i++ {
		segment := ps[i]
		if segment != "" {
			if string(segment[0]) == ":" {
				key := params[segment[1:]]
				re := regexp.MustCompile(defaultPattern)
				if one := re.Find([]byte(key)); one == nil {
					return "", ErrGenerateParameters
				}
				segments = append(segments, key)
				continue
			}

			if string(segment[0]) == "{" {
				segmentLen := len(segment)
				if string(segment[segmentLen-1]) == "}" {
					splitRes := strings.Split(segment[1:segmentLen-1], ":")
					re := regexp.MustCompile(splitRes[1])
					key := params[splitRes[0]]
					if one := re.Find([]byte(key)); one == nil {
						return "", ErrGenerateParameters
					}
					segments = append(segments, key)
					continue
				}

				return "", ErrPatternGrammar
			}
			if string(segment[len(segment)-1]) == "}" && string(segment[0]) != "{" {
				return "", ErrPatternGrammar
			}
		}

		segments = append(segments, segment)

		continue
	}

	return strings.Join(segments, "/"), nil
}

// PreHandler sets a handler that runs before any route handler.
// This is useful for global preprocessing of all requests.
func (e *Engine) PreHandler(preHandler Handler) {
	e.preHandler = preHandler
}

// NotFoundHandler sets a custom handler for 404 Not Found responses.
// This handler is called when no route matches the request URL.
func (e *Engine) NotFoundHandler(handler Handler) {
	e.router.notFound = Utils.ParseHandlerFunc(handler)
}

// Deprecated: please use znet.Recovery(func(c *Context, err error) {})
// PanicHandler is used for handling panics
func (e *Engine) PanicHandler(handler ErrHandlerFunc) {
	e.Use(Recovery(handler))
}

// GetTrees returns the internal routing trees for all HTTP methods.
// This is primarily used for debugging and testing purposes.
func (e *Engine) GetTrees() map[string]*Tree {
	return e.router.trees
}

// Handle registers a new handler for the specified HTTP method and path.
// This is the core routing function that all other HTTP method functions use internally.
func (e *Engine) Handle(method string, path string, action Handler, moreHandler ...Handler) *Engine {
	handler, firsthandle := handlerFuncs(moreHandler)
	p, l, ok := e.addHandle(method, path, Utils.ParseHandlerFunc(action), firsthandle, handler)
	if !ok {
		return e
	}

	routeAddLog(e, method, p, action, l)
	return e
}

// addHandle is the internal implementation of route registration.
// It adds a handler to the routing tree and returns the processed path, handler count, and a success flag.
func (e *Engine) addHandle(method string, path string, handle handlerFn, beforehandle []handlerFn, moreHandler []handlerFn) (string, int, bool) {
	if _, ok := methods[method]; !ok {
		e.Log.Fatal(method + " is invalid method")
	}

	tree, ok := e.router.trees[method]
	if !ok {
		tree = NewTree()
		e.router.trees[method] = tree
	}

	path = Utils.CompletionPath(path, e.router.prefix)
	if routeName := e.router.parameters.routeName; routeName != "" {
		tree.parameters.routeName = routeName
	}

	nodes := tree.Find(path, false)
	if len(nodes) > 0 {
		node := nodes[0]
		if e.webMode != quietCode && node.path == path && node.handle != nil {
			e.Log.Track("duplicate route definition: ["+method+"]"+path, 3, 1)
			return "", 0, false
		}
	}

	middleware := make([]handlerFn, len(e.router.middleware))
	{
		copy(middleware, e.router.middleware)
		if len(moreHandler) > 0 {
			middleware = append(middleware, moreHandler...)
		}
	}

	if len(beforehandle) > 0 {
		middleware = append(beforehandle, middleware...)
	}

	tree.Add(e, path, handle, middleware...)
	tree.parameters.routeName = ""
	return path, len(middleware) + 1, true
}

// ServeHTTP implements the http.Handler interface.
// This is the main entry point for handling HTTP requests in the framework.
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path
	if !e.ShowFavicon && p == "/favicon.ico" {
		return
	}

	c := e.acquireContext(w, req)
	defer func() {
		c.write()
		e.releaseContext(c)
	}()

	if e.AllowQuerySemicolons {
		allowQuerySemicolons(c.Request)
	}

	// custom method type
	if req.Method == "POST" && e.customMethodType != "" {
		if tmpType := c.GetHeader(e.customMethodType); tmpType != "" {
			req.Method = strings.ToUpper(tmpType)
		}
	}
	if e.preHandler != nil {
		if preHandler, ok := e.preHandler.(func(*Context) bool); ok {
			if preHandler(c) {
				return
			}
		} else {
			err := Utils.ParseHandlerFunc(e.preHandler)(c)
			if err != nil {
				c.renderError(c, err)
				c.Abort()
				return
			}
		}
	}
	if c.stopHandle.Load() {
		return
	}

	if e.FindHandle(c, req, p, true) {
		e.handleNotFound(c, true)
	}
}

// FindHandle searches for a handler matching the request URL and executes it if found.
// It returns true if a handler was found and executed, false otherwise.
func (e *Engine) FindHandle(rw *Context, req *http.Request, requestURL string, applyMiddleware bool) (not bool) {
	var anyTrees bool
	t, ok := e.router.trees[req.Method]
	if !ok {
		t, ok = e.router.trees[anyMethod]
		anyTrees = true
	}
	if !ok {
		return true
	}

	engine, handler, middleware, ok := Utils.TreeFind(t, requestURL)
	if !ok && !anyTrees {
		t, ok = e.router.trees[anyMethod]
		if ok {
			engine, handler, middleware, ok = Utils.TreeFind(t, requestURL)
		}
	}

	if !ok {
		return true
	}

	if engine != nil {
		rw.Engine = engine
	}

	if applyMiddleware {
		handleAction(rw, handler, middleware)
	} else {
		handleAction(rw, handler, []handlerFn{})
	}
	return false
}

// Use adds global middleware to the engine.
// These middleware functions will be executed for every request before route-specific middleware.
func (e *Engine) Use(middleware ...Handler) {
	if len(middleware) > 0 {
		middleware, firstMiddleware := handlerFuncs(middleware)
		e.router.middleware = append(firstMiddleware, e.router.middleware...)
		e.router.middleware = append(e.router.middleware, middleware...)
	}
}

// handleNotFound processes a 404 Not Found response.
// If applyMiddleware is true, it applies global middleware before calling the not found handler.
func (e *Engine) handleNotFound(c *Context, applyMiddleware bool) {
	var middleware []handlerFn
	if applyMiddleware {
		middleware = e.router.middleware
	}
	c.prevData.Code.Store(http.StatusNotFound)

	if e.router.notFound != nil {
		handleAction(c, e.router.notFound, middleware)
		return
	}

	handleAction(c, func(_ *Context) error {
		c.Byte(404, []byte("404 page not found"))
		return nil
	}, middleware)
}

// HandleNotFound is a public wrapper around handleNotFound.
// It allows external code to trigger a not found response for a context.
func (e *Engine) HandleNotFound(c *Context, applyMiddleware ...bool) {
	var apply bool
	if len(applyMiddleware) > 0 {
		apply = applyMiddleware[0]
	}
	e.handleNotFound(c, apply)
	c.stopHandle.Store(true)
}

// handleAction executes a handler function with its middleware chain.
// It processes middleware in order, then calls the main handler if no middleware aborts the chain.
func handleAction(c *Context, handler handlerFn, middleware []handlerFn) {
	c.middleware = append(middleware, handler)
	c.Next()
}

// Match checks if the request URL matches the route pattern.
// This is used internally for routing but can also be used to test if a URL would match a pattern.
func (e *Engine) Match(requestURL string, path string) bool {
	_, ok := Utils.URLMatchAndParse(requestURL, path)
	return ok
}
