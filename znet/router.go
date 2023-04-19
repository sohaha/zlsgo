package znet

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
)

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
	}
)

type (
	// contextKeyType Private Value Structure for Each Request
	contextKeyType struct{}
)

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

func (e *Engine) StaticFS(relativePath string, fs http.FileSystem, moreHandler ...Handler) {
	var urlPattern string
	log := temporarilyTurnOffTheLog(e, routeLog(e.Log, fmt.Sprintf("%%s %%-40s -> %s", fs), "FILE", relativePath))
	fileServer := http.StripPrefix(relativePath, http.FileServer(fs))

	handler := func(c *Context) {
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

		fileServer.ServeHTTP(c.Writer, c.Request)
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
	log()
}

func (e *Engine) Static(relativePath, root string, moreHandler ...Handler) {
	e.StaticFS(relativePath, http.Dir(root), moreHandler...)
}

func (e *Engine) StaticFile(relativePath, filepath string) {
	handler := func(c *Context) {
		c.File(filepath)
	}
	log := temporarilyTurnOffTheLog(e, routeLog(e.Log, "%s %-40s -> "+filepath, "FILE", relativePath))
	e.GET(relativePath, handler)
	e.HEAD(relativePath, handler)
	log()
}

func (e *Engine) Any(path string, action Handler, moreHandler ...Handler) *Engine {
	middleware, firstMiddleware := handlerFuncs(moreHandler)
	_, l, ok := e.handleAny(path, Utils.ParseHandlerFunc(action), middleware, firstMiddleware)

	if ok {
		routeAddLog(e, "ANY", Utils.CompletionPath(path, e.router.prefix), action, l)
	}

	return e
}

func (e *Engine) Customize(method, path string, action Handler, moreHandler ...Handler) *Engine {
	method = strings.ToUpper(method)
	return e.Handle(method, path, action, moreHandler...)
}

func (e *Engine) GET(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodGet, path, action, moreHandler...)
}

func (e *Engine) POST(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodPost, path, action, moreHandler...)
}

func (e *Engine) DELETE(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodDelete, path, action, moreHandler...)
}

func (e *Engine) PUT(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodPut, path, action, moreHandler...)
}

func (e *Engine) PATCH(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodPatch, path, action, moreHandler...)
}

func (e *Engine) HEAD(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodHead, path, action, moreHandler...)
}

func (e *Engine) OPTIONS(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodOptions, path, action, moreHandler...)
}

func (e *Engine) CONNECT(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodConnect, path, action, moreHandler...)
}

func (e *Engine) TRACE(path string, action Handler, moreHandler ...Handler) *Engine {
	return e.Handle(http.MethodTrace, path, action, moreHandler...)
}

func (e *Engine) GETAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.GET(path, action)
}

func (e *Engine) POSTAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.POST(path, action)
}

func (e *Engine) DELETEAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.DELETE(path, action)
}

func (e *Engine) PUTAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.PUT(path, action)
}

func (e *Engine) PATCHAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.PATCH(path, action)
}

func (e *Engine) HEADAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.HEAD(path, action)
}

func (e *Engine) OPTIONSAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.OPTIONS(path, action)
}

func (e *Engine) CONNECTAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.CONNECT(path, action)
}

func (e *Engine) TRACEAndName(path string, action Handler, routeName string) *Engine {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	return e.TRACE(path, action)
}

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
	}
	engine = &Engine{
		router:              route,
		webMode:             e.webMode,
		webModeName:         e.webModeName,
		MaxMultipartMemory:  e.MaxMultipartMemory,
		customMethodType:    e.customMethodType,
		Log:                 e.Log,
		Cache:               e.Cache,
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

func (e *Engine) PreHandler(preHandler Handler) {
	e.preHandler = preHandler
}

func (e *Engine) NotFoundHandler(handler Handler) {
	e.router.notFound = Utils.ParseHandlerFunc(handler)
}

// Deprecated: please use znet.Recovery(func(c *Context, err error) {})
// PanicHandler is used for handling panics
func (e *Engine) PanicHandler(handler ErrHandlerFunc) {
	e.Use(Recovery(handler))
}

// GetTrees Load Trees
func (e *Engine) GetTrees() map[string]*Tree {
	return e.router.trees
}

// Handle registers new request handlerFn
func (e *Engine) Handle(method string, path string, action Handler, moreHandler ...Handler) *Engine {
	handler, firsthandle := handlerFuncs(moreHandler)
	p, l, ok := e.addHandle(method, path, Utils.ParseHandlerFunc(action), firsthandle, handler)
	if !ok {
		return e
	}

	routeAddLog(e, method, p, action, l)
	return e
}

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

	tree.Add(path, handle, middleware...)
	tree.parameters.routeName = ""
	return path, len(middleware) + 1, true
}

func (e *Engine) handleAny(path string, handle handlerFn, beforehandle []handlerFn, moreHandler []handlerFn) (p string, l int, ok bool) {
	for key := range methods {
		p, l, ok = e.addHandle(key, path, handle, beforehandle, moreHandler)
		if !ok {
			return p, l, false
		}
	}
	return
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path
	if !e.ShowFavicon && p == "/favicon.ico" {
		return
	}

	c := e.acquireContext()
	c.clone(w, req)
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

	if _, ok := e.router.trees[req.Method]; !ok {
		e.HandleNotFound(c)
		return
	}

	if e.FindHandle(c, req, p, true) {
		e.HandleNotFound(c)
	}
}

func (e *Engine) FindHandle(rw *Context, req *http.Request, requestURL string, applyMiddleware bool) (not bool) {
	t, ok := e.router.trees[req.Method]
	if !ok {
		return true
	}

	handler, middleware, ok := Utils.TreeFind(t, requestURL)
	if !ok {
		return true
	}

	if applyMiddleware {
		handleAction(rw, handler, middleware)
	} else {
		handleAction(rw, handler, []handlerFn{})
	}
	return false
}

func (e *Engine) Use(middleware ...Handler) {
	if len(middleware) > 0 {
		middleware, firstMiddleware := handlerFuncs(middleware)
		e.router.middleware = append(firstMiddleware, e.router.middleware...)
		e.router.middleware = append(e.router.middleware, middleware...)
	}
}

func (e *Engine) HandleNotFound(c *Context) {
	middleware := e.router.middleware
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

func handleAction(c *Context, handler handlerFn, middleware []handlerFn) {
	c.middleware = append(middleware, handler)
	c.Next()
}

// Match checks if the request matches the route pattern
func (e *Engine) Match(requestURL string, path string) bool {
	_, ok := Utils.URLMatchAndParse(requestURL, path)
	return ok
}
