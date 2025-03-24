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
	"strings"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zutil"
	"github.com/sohaha/zlsgo/zutil/daemon"

	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zshell"
)

type (
	// Context context
	Context struct {
		startTime     time.Time
		render        Renderer
		Writer        http.ResponseWriter
		injector      zdi.Injector
		stopHandle    *zutil.Bool
		prevData      *PrevData
		customizeData map[string]interface{}
		header        map[string][]string
		Request       *http.Request
		cacheJSON     *zjson.Res
		cacheForm     url.Values
		done          *zutil.Bool
		Engine        *Engine
		Log           *zlog.Logger
		// Deprecated: Please maintain your own cache
		Cache       *zcache.Table
		renderError ErrHandlerFunc
		cacheQuery  url.Values
		rawData     []byte
		middleware  []handlerFn
		mu          sync.RWMutex
	}
	// Engine is a simple HTTP route multiplexer that parses a request path
	Engine struct {
		pool                 sync.Pool
		injector             zdi.Injector
		preHandler           Handler
		views                Template
		Cache                *zcache.Table
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
		webMode              int
		writeTimeout         time.Duration
		readTimeout          time.Duration
		ShowFavicon          bool
		AllowQuerySemicolons bool
	}
	TlsCfg struct {
		HTTPProcessing interface{}
		Config         *tls.Config
		Cert           string
		Key            string
		HTTPAddr       string
	}
	tpl struct {
		tpl             *template.Template
		templateFuncMap template.FuncMap
		pattern         string
	}
	addrSt struct {
		TlsCfg
		addr string
	}
	router struct {
		trees      map[string]*Tree
		notFound   handlerFn
		prefix     string
		parameters Parameters
		middleware []handlerFn
	}
	// Handler handler func
	Handler      interface{}
	firstHandler [1]Handler
	// HandlerFunc old handler func
	HandlerFunc func(c *Context)
	handlerFn   func(c *Context) error
	// MiddlewareFunc Middleware Func
	MiddlewareFunc func(c *Context, fn Handler)
	// ErrHandlerFunc ErrHandlerFunc
	ErrHandlerFunc func(c *Context, err error)
	// MiddlewareType is a public type that is used for middleware
	MiddlewareType Handler
	// Parameters records some parameters
	Parameters struct {
		routeName string
	}
	serverMap struct {
		engine *Engine
		srv    *http.Server
	}
)

const (
	defaultMultipartMemory = 32 << 20 // 32 MB
	// DebugMode dev
	DebugMode = "dev"
	// ProdMode release
	ProdMode = "prod"
	// TestMode test
	TestMode = "test"
	// QuietMode quiet
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
	Log   = zlog.New(zlog.ColorTextWrap(zlog.ColorGreen, "[Z] "))
	Cache = zcache.New("__ZNET__")
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

// New returns a newly initialized Engine object that implements the Engine
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
		Cache:               Cache,
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

// WrapFirstMiddleware Wrapping a function in the first position of the middleware
func WrapFirstMiddleware(fn Handler) firstHandler {
	return firstHandler{fn}
}

// Server Server
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

// OnShutdown On Shutdown Func
func OnShutdown(done func()) {
	shutdownDone = done
}

// SetAddr SetAddr
func (e *Engine) SetAddr(addrString string, tlsConfig ...TlsCfg) {
	e.addr = []addrSt{
		resolveAddr(addrString, tlsConfig...),
	}
}

// AddAddr AddAddr
func (e *Engine) AddAddr(addrString string, tlsConfig ...TlsCfg) {
	e.addr = append(e.addr, resolveAddr(addrString, tlsConfig...))
}

// SetCustomMethodField Set Custom Method Field
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

// Injector Call Injector
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

// SetMode Setting Server Mode
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

// GetMode Get Mode
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

// IsDebug IsDebug
func (e *Engine) IsDebug() bool {
	return e.webMode > prodCode
}

// SetTimeout set Timeout
func (e *Engine) SetTimeout(Timeout time.Duration, WriteTimeout ...time.Duration) {
	if len(WriteTimeout) > 0 {
		e.writeTimeout = WriteTimeout[0]
		e.readTimeout = Timeout
	} else {
		e.writeTimeout = Timeout
		e.readTimeout = Timeout
	}
}

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

func Shutdown() error {
	if !isRunContext.Load() {
		return errors.New("server was started with custom context, cannot use Shutdown")
	}

	shutdown(true)
	return nil
}

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

// Run serve
func Run(cb ...func(name, addr string)) {
	RunContext(context.Background(), cb...)
}

var isRunContext = zutil.NewBool(false)

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

func runNewProcess() error {
	args := os.Args
	_, err := zshell.RunNewProcess(args[0], args)
	return err
}
