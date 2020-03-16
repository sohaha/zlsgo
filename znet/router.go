/*
 * @Author: seekwe
 * @Date:   2019-05-09 12:44:23
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-05 12:35:57
 */

package znet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zstring"
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

	defaultPattern = `[\w\p{Han}\.\- ]+`
	idPattern      = `[\d]+`
	idKey          = `id`
	allPattern     = `[\w\p{Han}\.\-\/ ]*`
	allKey         = `*`

	contextKey = contextKeyType{}

	methods = map[string]struct{}{
		http.MethodGet:     {},
		http.MethodPost:    {},
		http.MethodPut:     {},
		http.MethodDelete:  {},
		http.MethodPatch:   {},
		http.MethodHead:    {},
		http.MethodOptions: {},
	}
)

type (
	// contextKeyType Private Value Structure for Each Request
	contextKeyType struct{}

	// ParamsMapType Storage path parameters
	ParamsMapType map[string]string
)

func temporarilyTurnOffTheLog(e *Engine, msg string) func() {
	mode := e.webMode
	e.webMode = releaseCode
	return func() {
		e.webMode = mode
		if e.IsDebug() {
			e.Log.Debug(msg)
		}
	}
}

func (e *Engine) StaticFS(relativePath string, fs http.FileSystem) {
	urlPattern := path.Join(relativePath, "/{file:.*}")
	fileServer := http.StripPrefix(relativePath, http.FileServer(fs))
	handler := func(c *Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
	log := temporarilyTurnOffTheLog(e, showRouteDebug(e.Log, fmt.Sprintf("%%s --> %%s ->> %s", fs), "Static", relativePath))
	e.GET(urlPattern, handler)
	e.HEAD(urlPattern, handler)
	log()
}

func (e *Engine) Static(relativePath, root string) {
	e.StaticFS(relativePath, http.Dir(root))
}

func (e *Engine) StaticFile(relativePath, filepath string) {
	handler := func(c *Context) {
		c.File(filepath)
	}
	log := temporarilyTurnOffTheLog(e, showRouteDebug(e.Log, "%s --> %s ->> "+filepath, "File", relativePath))
	e.GET(relativePath, handler)
	e.HEAD(relativePath, handler)
	log()
}

func (e *Engine) Any(path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	log := temporarilyTurnOffTheLog(e, showRouteDebug(e.Log, "%s --> %s", "Any", completionPath(path, e.router.prefix)))
	e.GET(path, handle, moreHandler...)
	e.POST(path, handle, moreHandler...)
	e.PUT(path, handle, moreHandler...)
	e.DELETE(path, handle, moreHandler...)
	e.PATCH(path, handle, moreHandler...)
	e.HEAD(path, handle, moreHandler...)
	e.OPTIONS(path, handle, moreHandler...)
	log()
}

func (e *Engine) Customize(method, path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	method = strings.ToUpper(method)
	e.Handle(method, path, handle, moreHandler...)
}

func (e *Engine) GET(path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	e.Handle(http.MethodGet, path, handle, moreHandler...)
}

func (e *Engine) POST(path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	e.Handle(http.MethodPost, path, handle, moreHandler...)
}

func (e *Engine) DELETE(path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	e.Handle(http.MethodDelete, path, handle, moreHandler...)
}

func (e *Engine) PUT(path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	e.Handle(http.MethodPut, path, handle, moreHandler...)
}

func (e *Engine) PATCH(path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	e.Handle(http.MethodPatch, path, handle, moreHandler...)
}

func (e *Engine) HEAD(path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	e.Handle(http.MethodHead, path, handle, moreHandler...)
}

func (e *Engine) OPTIONS(path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	e.Handle(http.MethodOptions, path, handle, moreHandler...)
}

func (e *Engine) GETAndName(path string, handle HandlerFunc, routeName string) {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	e.GET(path, handle)
}

func (e *Engine) POSTAndName(path string, handle HandlerFunc, routeName string) {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	e.POST(path, handle)
}

func (e *Engine) DELETEAndName(path string, handle HandlerFunc, routeName string) {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	e.DELETE(path, handle)
}

func (e *Engine) PUTAndName(path string, handle HandlerFunc, routeName string) {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	e.PUT(path, handle)
}

func (e *Engine) PATCHAndName(path string, handle HandlerFunc, routeName string) {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	e.PATCH(path, handle)
}

func (e *Engine) HEADAndName(path string, handle HandlerFunc, routeName string) {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	e.HEAD(path, handle)
}

func (e *Engine) OPTIONSAndName(path string, handle HandlerFunc, routeName string) {
	e.router.parameters.routeName = routeName
	defer func() { e.router.parameters.routeName = "" }()
	e.OPTIONS(path, handle)
}

func (e *Engine) Group(prefix string, groupHandle ...func(e *Engine)) (engine *Engine) {
	rprefix := e.router.prefix
	if rprefix != "" {
		if strings.HasSuffix(rprefix, "/") && strings.HasPrefix(prefix, "/") {
			prefix = strings.TrimPrefix(prefix, "/")
		}
		prefix = rprefix + prefix
	}
	route := &router{
		prefix:     prefix,
		trees:      e.router.trees,
		middleware: e.router.middleware,
	}
	engine = &Engine{
		webMode:            e.webMode,
		webModeName:        e.webModeName,
		timeLocation:       e.timeLocation,
		MaxMultipartMemory: e.MaxMultipartMemory,
		customMethodType:   e.customMethodType,
		router:             route,
		Log:                e.Log,
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

	var segments []string
	res := strings.Split(route.path, "/")
	for _, segment := range res {
		if segment != "" {
			if string(segment[0]) == ":" {
				key := params[string(segment[1:])]
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
					splitRes := strings.Split(string(segment[1:segmentLen-1]), ":")
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
	return "/" + strings.Join(segments, "/"), nil
}

func (e *Engine) NotFoundFunc(handler HandlerFunc) {
	e.router.notFound = handler
}

func (e *Engine) PanicHandler(handler PanicFunc) {
	e.router.panic = handler
}

// GetTrees Get Trees
func (e *Engine) GetTrees() map[string]*Tree {
	return e.router.trees
}

// Handle registers new request handler
func (e *Engine) Handle(method string, path string, handle HandlerFunc, moreHandler ...HandlerFunc) {
	if _, ok := methods[method]; !ok {
		e.Log.Fatal(method + " is invalid method")
	}

	tree, ok := e.router.trees[method]
	if !ok {
		tree = NewTree()
		e.router.trees[method] = tree
	}

	path = completionPath(path, e.router.prefix)
	if routeName := e.router.parameters.routeName; routeName != "" {
		tree.parameters.routeName = routeName
	}
	if e.IsDebug() {
		e.Log.Debug(showRouteDebug(e.Log, "%s --> %s", method, path))
	}
	middleware := e.router.middleware
	moreHandlerLen := len(moreHandler)
	if moreHandlerLen > 0 {
		index := moreHandlerLen - 1
		lastHandle := moreHandler[index]
		moreHandler = moreHandler[:index]
		middleware = append(middleware, handle)
		middleware = append(middleware, moreHandler...)
		handle = lastHandle
	}

	tree.Add(path, handle, middleware...)
	tree.parameters.routeName = ""
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	requestURL := req.URL.Path
	if !e.ShowFavicon && requestURL == "/favicon.ico" {
		return
	}
	rw := &Context{
		Writer:  w,
		Request: req,
		Engine:  e,
		Log:     e.Log,
		Info: &info{
			Code:          http.StatusOK,
			StartTime:     time.Now(),
			heades:        map[string]string{},
			customizeData: map[string]interface{}{},
		},
	}
	if e.router.panic != nil {
		defer func() {
			if err := recover(); err != nil {
				rw.Info.Code = http.StatusInternalServerError
				errMsg, ok := err.(error)
				if !ok {
					errMsg = errors.New(fmt.Sprint(err))
				}
				e.router.panic(rw, errMsg)
				requestLog(rw)
				e.Log.Error(errMsg)
				e.Log.Track("Track Panic: ", 0, 2)
				// Log.Stack()
				// trace := make([]byte, 1<<16)
				// n := runtime.Stack(trace, true)
				// Log.Errorf("panic: '%v'\n, Stack Trace:\n %s", err, string(trace[:int(math.Min(float64(n), float64(7000)))]))
			}
		}()
	}

	if req.Method == "POST" && e.customMethodType != "" {
		if tmpType, ok := rw.GetPostForm(e.customMethodType); ok {
			req.Method = strings.ToUpper(tmpType)
		}
	}
	if _, ok := e.router.trees[req.Method]; !ok {
		e.HandleNotFound(rw, e.router.middleware)
		return
	}

	if e.FindHandle(rw, req, requestURL, true) {
		e.HandleNotFound(rw, e.router.middleware)
	}
}

func (e *Engine) FindHandle(rw *Context, req *http.Request, requestURL string, applyMiddleware bool) (not bool) {
	nodes := e.router.trees[req.Method].Find(requestURL, false)
	if len(nodes) > 0 {
		node := nodes[0]
		if node.handle != nil {
			if node.path == requestURL {
				if applyMiddleware {
					handle(rw, node.handle, node.middleware)
				} else {
					handle(rw, node.handle, []HandlerFunc{})
				}
				return
			}
			if node.path == requestURL[1:] {
				if applyMiddleware {
					handle(rw, node.handle, node.middleware)
				} else {
					handle(rw, node.handle, []HandlerFunc{})
				}
				return
			}
		}
	}

	if len(nodes) == 0 {
		res := strings.Split(requestURL, "/")
		prefix := res[1]
		nodes := e.router.trees[req.Method].Find(prefix, true)
		for _, node := range nodes {
			if handler := node.handle; handler != nil && node.path != requestURL {
				if matchParamsMap, ok := e.matchAndParse(requestURL, node.path); ok {
					ctx := context.WithValue(req.Context(), contextKey, matchParamsMap)
					req = req.WithContext(ctx)
					rw.Request = req
					if applyMiddleware {
						handle(rw, handler, node.middleware)
					} else {
						handle(rw, handler, []HandlerFunc{})
					}
					return
				}
			}
		}
	}
	return true
}

func (e *Engine) Use(middleware ...HandlerFunc) {
	if len(middleware) > 0 {
		e.router.middleware = append(e.router.middleware, middleware...)
	}
}

func (e *Engine) HandleNotFound(c *Context, middleware []HandlerFunc) {
	c.Info.Code = http.StatusNotFound
	if e.router.notFound != nil {
		handle(c, e.router.notFound, middleware)
		return
	}
	handle(c, func(_ *Context) {
		c.Abort(0)
		http.NotFound(c.Writer, c.Request)
	}, middleware)
}

func handle(c *Context, handler HandlerFunc, middleware []HandlerFunc) {
	c.Info.middleware = append(middleware, handler)
	c.Info.handlerLen = len(c.Info.middleware)
	c.Next()
	c.done()
}

// Match checks if the request matches the route pattern
func (e *Engine) Match(requestURL string, path string) bool {
	_, ok := e.matchAndParse(requestURL, path)
	return ok
}

// matchAndParse checks if the request matches the route path and returns a map of the parsed
func (e *Engine) matchAndParse(requestURL string, path string) (matchParams ParamsMapType, bl bool) {
	bl = true
	matchParams = make(ParamsMapType)
	res := strings.Split(path, "/")
	pattern, matchName := parsPattern(res)
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
