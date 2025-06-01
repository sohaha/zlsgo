// Package znet provides a lightweight and high-performance HTTP web framework.
package znet

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/zutil"
	"github.com/sohaha/zlsgo/zutil/daemon"

	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zshell"
)

type (
	// Context represents the HTTP request and response context.
	// It provides methods for accessing request data, setting response data,
	// and managing the request lifecycle.
	Context struct {
		startTime     time.Time
		render        Renderer
		Writer        http.ResponseWriter
		injector      zdi.Injector
		cacheForm     url.Values
		Log           *zlog.Logger
		customizeData map[string]interface{}
		header        map[string][]string
		Request       *http.Request
		cacheJSON     *zjson.Res
		stopHandle    *zutil.Bool
		done          *zutil.Bool
		Engine        *Engine
		prevData      *PrevData
		Cache         *zcache.Table
		renderError   ErrHandlerFunc
		cacheQuery    url.Values
		ip            string
		rawData       []byte
		middleware    []handlerFn
		mu            zsync.RBMutex
	}
	// Engine is the core of the web framework, providing HTTP routing and server functionality.
	// It manages routes, middleware, templates, and server configuration.
	Engine struct {
		pool                 sync.Pool
		injector             zdi.Injector
		preHandler           Handler
		views                Template
		template             *tpl
		Log                  *zlog.Logger
		templateFuncMap      template.FuncMap
		router               *router
		BindTag              string
		webModeName          string
		BindStructDelimiter  string
		BindStructCase       func(string) string
		BindStructSuffix     string
		customMethodType     string
		addr                 []addrSt
		shutdowns            []func()
		MaxMultipartMemory   int64
		MaxRequestBodySize   int64
		webMode              int
		writeTimeout         time.Duration
		readTimeout          time.Duration
		ShowFavicon          bool
		AllowQuerySemicolons bool
		customRenderings     []reflect.Type
	}
	// TlsCfg holds TLS configuration for secure HTTP connections.
	TlsCfg struct {
		HTTPProcessing interface{}
		Config         *tls.Config
		Cert           string
		Key            string
		HTTPAddr       string
	}
	// tpl is an internal structure for template management.
	tpl struct {
		tpl             *template.Template
		templateFuncMap template.FuncMap
		pattern         string
	}
	// addrSt represents a server address with optional TLS configuration.
	addrSt struct {
		TlsCfg
		addr string
	}
	// router manages the HTTP route trees and middleware stack.
	router struct {
		trees      map[string]*Tree
		notFound   handlerFn
		prefix     string
		parameters Parameters
		middleware []handlerFn
	}
	// Handler is the interface for HTTP request handlers.
	// It can be a function with various signatures that the framework adapts to.
	Handler interface{}
	// firstHandler is a specialized array type for middleware insertion at the beginning.
	firstHandler [1]Handler
	// HandlerFunc is the legacy handler function signature.
	// It receives a context pointer but doesn't return an error.
	HandlerFunc func(c *Context)
	// handlerFn is the internal handler function signature that supports error returns.
	handlerFn func(c *Context) error
	// MiddlewareFunc defines the middleware function signature.
	// It receives both the context and the next handler in the chain.
	MiddlewareFunc func(c *Context, fn Handler)
	// ErrHandlerFunc defines the error handler function signature.
	// It receives both the context and the error that occurred.
	ErrHandlerFunc func(c *Context, err error)
	// MiddlewareType is a public type alias for Handler used in middleware contexts.
	MiddlewareType Handler
	// Parameters stores route-related information during request processing.
	Parameters struct {
		routeName string
	}
	// serverMap associates an Engine instance with its HTTP server.
	serverMap struct {
		engine *Engine
		srv    *http.Server
	}
)

const (
	// defaultMultipartMemory defines the default maximum memory for parsing multipart forms (32 MB).
	defaultMultipartMemory = 32 << 20 // 32 MB
	// DebugMode indicates development mode with verbose logging.
	DebugMode = "dev"
	// ProdMode indicates production mode with minimal logging.
	ProdMode = "prod"
	// TestMode indicates testing mode.
	TestMode = "test"
	// QuietMode indicates a mode with no logging output.
	QuietMode         = "quiet"
	defaultServerName = ""
	defaultBindTag    = "json"
	quietCode         = -1
	prodCode          = 0
	debugCode         = iota
	testCode
)

var (
	// Log Log
	Log = zlog.New(zlog.ColorTextWrap(zlog.ColorGreen, "[Z] "))
	// shutdownDone Shutdown Done executed after shutting down the server
	shutdownDone func()
	// CloseHotRestart Close Hot Restart
	CloseHotRestart bool
	zservers        = map[string]*Engine{}
	defaultAddr     = addrSt{
		addr: ":3788",
	}
	// BindStructDelimiter structure route delimiter
	BindStructDelimiter = "-"
	// BindStructSuffix structure route suffix
	BindStructSuffix = ""
)

func init() {
	Log.ResetFlags(zlog.BitTime | zlog.BitLevel)
}

// New creates and initializes a new Engine instance.
// An optional serverName can be provided to identify this server in logs.
// The returned Engine is configured with default settings and ready to define routes.
func New(serverName ...string) *Engine {
	var name string
	if len(serverName) > 0 {
		name = serverName[0]
	}

	var log *zlog.Logger
	if name != "" {
		log = zlog.New("[" + name + "] ")
	} else {
		log = zlog.New("[Z] ")
	}

	log.ResetFlags(zlog.BitTime | zlog.BitLevel)
	log.SetLogLevel(zlog.LogInfo)

	route := &router{
		prefix: "/",
		trees:  make(map[string]*Tree),
	}
	r := &Engine{
		Log:                 log,
		MaxMultipartMemory:  defaultMultipartMemory,
		BindTag:             defaultBindTag,
		BindStructDelimiter: BindStructDelimiter,
		BindStructSuffix:    BindStructSuffix,
		router:              route,
		readTimeout:         0 * time.Second,
		writeTimeout:        0 * time.Second,
		webModeName:         ProdMode,
		webMode:             prodCode,
		addr:                []addrSt{defaultAddr},
		templateFuncMap:     template.FuncMap{},
		injector:            zdi.New(),
		customRenderings:    make([]reflect.Type, 0),
		shutdowns:           make([]func(), 0),
	}
	r.pool.New = func() interface{} {
		return r.NewContext(nil, nil)
	}
	if _, ok := zservers[name]; ok && name != "" {
		r.Log.Fatal("serverName: [", name, "] it already exists")
	}
	zservers[name] = r
	return r
}

// WrapFirstMiddleware wraps a handler function to be inserted at the beginning of the middleware chain.
// This is useful for middleware that must execute before any other middleware.
func WrapFirstMiddleware(fn Handler) firstHandler {
	return firstHandler{fn}
}

// Server retrieves an existing Engine instance by name.
// Returns the Engine and a boolean indicating if it was found.
func Server(serverName ...string) (engine *Engine, ok bool) {
	name := defaultServerName
	if len(serverName) > 0 {
		name = serverName[0]
	}
	if engine, ok = zservers[name]; !ok {
		engine = New(name)
		engine.Log.Warnf("serverName: %s is not", name)
	}
	return
}

// OnShutdown registers a function to be called when the server shuts down.
// This is useful for cleanup tasks that should run before the program exits.
func OnShutdown(done func()) {
	shutdownDone = done
}

// SetAddr sets the address for the server to listen on.
// Optional TLS configuration can be provided for HTTPS support.
func (e *Engine) SetAddr(addrString string, tlsConfig ...TlsCfg) {
	e.addr = []addrSt{
		resolveAddr(addrString, tlsConfig...),
	}
}

// AddAddr adds an additional address for the server to listen on.
// This allows the server to listen on multiple ports or interfaces.
func (e *Engine) AddAddr(addrString string, tlsConfig ...TlsCfg) {
	e.addr = append(e.addr, resolveAddr(addrString, tlsConfig...))
}

// SetCustomMethodField sets the field name used for HTTP method overriding.
// This allows clients to use methods like PUT/DELETE in environments that only support GET/POST.
func (e *Engine) SetCustomMethodField(field string) {
	e.customMethodType = field
}

// Deprecated: If you need to verify if a program is trustworthy, please implement it yourself.
// CloseHotRestartFileMd5 CloseHotRestartFileMd5
func CloseHotRestartFileMd5() {
}

// Deprecated: please use SetTemplate()
// SetTemplateFuncMap Set Template Func
func (e *Engine) SetTemplateFuncMap(funcMap template.FuncMap) {
	if e.views == nil {
		// compatible with the old version at present
		e.templateFuncMap = funcMap
		return
	}

	if t, ok := e.views.(*htmlEngine); ok {
		t.SetFuncMap(funcMap)
	}
}

// Injector returns the dependency injection container used by this Engine.
// It can be used to register services for use in handlers.
func (e *Engine) Injector() zdi.TypeMapper {
	return e.injector
}

// Deprecated: please use SetTemplate()
// SetHTMLTemplate Set HTML Template
func (e *Engine) SetHTMLTemplate(t *template.Template) {
	val := &tpl{
		tpl:             t,
		templateFuncMap: template.FuncMap{},
	}
	e.template = val
}

// LoadHTMLGlob Load Glob HTML
// LoadHTMLGlob loads HTML templates from the specified glob pattern.
// It parses the templates and makes them available for rendering in handlers.
func (e *Engine) LoadHTMLGlob(pattern string) {
	if !strings.Contains(pattern, "*") {
		h := newGoTemplate(e, pattern)
		e.views = h
		return
	}

	// compatible with the old version at present
	pattern = zfile.RealPath(pattern)
	t, err := template.New("").Funcs(e.templateFuncMap).ParseGlob(pattern)
	if err != nil {
		e.Log.Fatalf("Template loading failed: %s\n", err)
		return
	}
	isDebug := e.IsDebug()
	val := &tpl{
		pattern:         pattern,
		tpl:             t,
		templateFuncMap: template.FuncMap{},
	}
	if isDebug {
		templatesDebug(e, t)
		val.templateFuncMap = e.templateFuncMap
	}
	e.template = val
}

// SetMode sets the server's operating mode (dev, prod, test, or quiet).
// This affects logging verbosity and other runtime behaviors.
func (e *Engine) SetMode(value string) {
	var level int
	switch value {
	case ProdMode, "":
		level = zlog.LogSuccess
		e.webMode = prodCode
	case QuietMode:
		level = zlog.LogPanic
		e.webMode = quietCode
	case DebugMode:
		level = zlog.LogDump
		e.webMode = debugCode
	case TestMode:
		level = zlog.LogDebug
		e.webMode = testCode
	default:
		e.Log.Panic("web mode unknown: " + value)
	}
	if value == "" {
		value = ProdMode
	}
	e.webModeName = value
	e.Log.SetLogLevel(level)
}

// GetMode returns the current server operating mode as a string.
func (e *Engine) GetMode() string {
	switch e.webMode {
	case prodCode:
		return ProdMode
	case quietCode:
		return QuietMode
	case debugCode:
		return DebugMode
	case testCode:
		return TestMode
	default:
		return "unknown"
	}
}

// IsDebug returns true if the server is running in debug mode.
func (e *Engine) IsDebug() bool {
	return e.webMode > prodCode
}

// SetTimeout sets the read timeout and optionally the write timeout for the HTTP server.
// These timeouts help prevent slow client attacks.
func (e *Engine) SetTimeout(Timeout time.Duration, WriteTimeout ...time.Duration) {
	if len(WriteTimeout) > 0 {
		e.writeTimeout = WriteTimeout[0]
		e.readTimeout = Timeout
	} else {
		e.writeTimeout = Timeout
		e.readTimeout = Timeout
	}
}

// StartUp initializes and starts the HTTP server(s) for this Engine.
// It configures all servers according to the Engine settings and begins listening
// on all configured addresses. Returns the server instances that were started.
func (e *Engine) StartUp() []*serverMap {
	var wg sync.WaitGroup
	var srvMap sync.Map
	for _, cfg := range e.addr {
		wg.Add(1)

		go func(cfg addrSt, e *Engine) {
			if e.AllowQuerySemicolons {
				e.Log.SetIgnoreLog(errURLQuerySemicolon)
			}
			errChan := make(chan error, 1)
			isTls := cfg.Cert != "" || cfg.Config != nil
			addr := getAddr(cfg.addr)
			hostname := getHostname(addr, isTls)
			srv := &http.Server{
				Addr:         addr,
				Handler:      e,
				ReadTimeout:  e.readTimeout,
				WriteTimeout: e.writeTimeout,
				// MaxHeaderBytes: 1 << 20,
				ErrorLog: log.New(e.Log, "", 0),
			}

			srvMap.Store(addr, &serverMap{e, srv})

			wg.Done()

			go func() {
				select {
				case <-errChan:
				default:
					wrapPid := e.Log.ColorTextWrap(zlog.ColorLightGrey, fmt.Sprintf("Pid: %d", os.Getpid()))
					wrapMode := ""
					if e.webMode > 0 {
						wrapMode = e.Log.ColorTextWrap(zlog.ColorYellow, fmt.Sprintf("%s ", strings.ToUpper(e.webModeName)))
					}
					e.Log.Successf("%s %s %s%s\n", "Listen:", e.Log.ColorTextWrap(zlog.ColorLightGreen, e.Log.OpTextWrap(zlog.OpBold, hostname)), wrapMode, wrapPid)
				}
			}()

			if isTls {
				if cfg.Config != nil {
					srv.TLSConfig = cfg.Config
				}
				if cfg.HTTPAddr != "" {
					httpAddr := getAddr(cfg.HTTPAddr)
					go func(e *Engine) {
						newHostname := "http://" + resolveHostname(httpAddr)
						e.Log.Success(e.Log.ColorBackgroundWrap(zlog.ColorYellow, zlog.ColorDefault, e.Log.OpTextWrap(zlog.OpBold, "Listen: "+newHostname)))
						var err error
						switch processing := cfg.HTTPProcessing.(type) {
						case string:
							err = http.ListenAndServe(httpAddr, &tlsRedirectHandler{Domain: processing})
						case http.Handler:
							err = http.ListenAndServe(httpAddr, processing)
						default:
							err = http.ListenAndServe(httpAddr, e)
						}
						e.Log.Errorf("HTTP Listen: %s\n", err)
					}(e)
				}
				errChan <- srv.ListenAndServeTLS(cfg.Cert, cfg.Key)
			} else {
				errChan <- srv.ListenAndServe()
			}

			err := <-errChan
			if err != nil && err != http.ErrServerClosed {
				e.Log.Fatalf("Listen: %s\n", err)
			} else if err != http.ErrServerClosed {
				e.Log.Info(err)
			}
		}(cfg, e)
	}

	wg.Wait()

	srvs := make([]*serverMap, 0)
	srvMap.Range(func(addr, value interface{}) bool {
		srvs = append(srvs, value.(*serverMap))
		return true
	})
	return srvs
}

// Shutdown gracefully stops all running servers.
// It waits for active connections to complete before shutting down.
// Returns an error if the shutdown process encounters any issues.
func Shutdown() error {
	if !isRunContext.Load() {
		return errors.New("server was started with custom context, cannot use Shutdown")
	}

	shutdown(true)
	return nil
}

// shutdown is the internal implementation of the shutdown process.
// If sigkill is true, it forces immediate termination rather than waiting
// for connections to complete gracefully.
func shutdown(sigkill bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for _, s := range srvs {
		r := s.engine
		if sigkill {
			r.Log.Info("Shutdown server ...")
		}
		for _, shutdown := range r.shutdowns {
			shutdown()
		}
		err := s.srv.Shutdown(ctx)
		if err != nil {
			if sigkill {
				r.Log.Error("Timeout forced close")
			}
			_ = s.srv.Close()
		} else {
			if sigkill {
				r.Log.Success("Shutdown server done")
			}
		}
		wg.Done()
	}

	wg.Wait()
	srvs = srvs[:0:0]
	if shutdownDone != nil {
		shutdownDone()
	}
}

var (
	srvs []*serverMap
	wg   sync.WaitGroup
)

// Run starts the HTTP server and begins listening for requests.
// Optional callback functions are called when each server starts, receiving the server name and address.
func Run(cb ...func(name, addr string)) {
	RunContext(context.Background(), cb...)
}

var isRunContext = zutil.NewBool(false)

// RunContext starts all configured servers with a context for cancellation.
// The provided context can be used to trigger server shutdown.
// Optional callback functions are called when each server starts.
func RunContext(ctx context.Context, cb ...func(name, addr string)) {
	isRunContext.Store(true)
	defer isRunContext.Store(false)

	for n, e := range zservers {
		ss := e.StartUp()
		wg.Add(len(ss))
		srvs = append(srvs, ss...)
		if len(cb) == 0 {
			continue
		}
		for _, v := range ss {
			cb[0](n, v.GetAddr())
		}
	}

	select {
	case <-ctx.Done():
		shutdown(true)
	case signal := <-daemon.SingleKillSignal():
		if !signal && !CloseHotRestart {
			if err := runNewProcess(); err != nil {
				Log.Error(err)
			}
		}

		shutdown(signal)
	}
}

// runNewProcess starts a new process for hot reloading.
// This is used during graceful restarts to spawn a new server process
// before shutting down the current one.
func runNewProcess() error {
	args := os.Args
	_, err := zshell.RunNewProcess(args[0], args)
	return err
}
