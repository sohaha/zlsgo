package znet

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zcli"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zjson"

	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zshell"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	// Context context
	Context struct {
		stopHandle    bool
		rawData       string
		cacheJSON     *zjson.Res
		startTime     time.Time
		middleware    []HandlerFunc
		customizeData map[string]interface{}
		header        map[string][]string
		render        render
		prevData      *PrevData
		Writer        http.ResponseWriter
		Request       *http.Request
		Engine        *Engine
		Log           *zlog.Logger
		Cache         *zcache.Table
		l             sync.RWMutex
		cacheQuery    url.Values
	}
	// Engine is a simple HTTP route multiplexer that parses a request path
	Engine struct {
		// Log Log
		Log *zlog.Logger
		// Deprecated: 以后可能移除
		Cache               *zcache.Table
		readTimeout         time.Duration
		writeTimeout        time.Duration
		webModeName         string
		webMode             int
		ShowFavicon         bool
		MaxMultipartMemory  int64
		BindTag             string
		customMethodType    string
		addr                []addrSt
		router              *router
		preHandler          func(context *Context) bool
		pool                sync.Pool
		templateFuncMap     template.FuncMap
		BindStructDelimiter string
		BindStructSuffix    string
		template            *tpl
	}
	TlsCfg struct {
		Cert           string
		Key            string
		HTTPAddr       string
		HTTPProcessing interface{}
		Config         *tls.Config
	}
	tpl struct {
		tpl             *template.Template
		pattern         string
		templateFuncMap template.FuncMap
	}
	addrSt struct {
		addr string
		TlsCfg
	}
	router struct {
		prefix     string
		middleware []HandlerFunc
		trees      map[string]*Tree
		parameters Parameters
		notFound   HandlerFunc
		panic      PanicFunc
	}
	// HandlerFunc HandlerFunc
	HandlerFunc func(c *Context)
	// MiddlewareFunc Middleware Func
	MiddlewareFunc func(c *Context, fn HandlerFunc)
	// PanicFunc PanicFunc
	PanicFunc func(c *Context, err error)
	// MiddlewareType is a public type that is used for middleware
	MiddlewareType HandlerFunc
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
	TestMode          = "test"
	defaultServerName = "Z"
	defaultBindTag    = "json"
	prodCode          = 0
	debugCode         = iota
	testCode
)

var (
	// Log Log
	Log   = zlog.New(zlog.ColorTextWrap(zlog.ColorGreen, "[Z] "))
	Cache = zcache.New("__ZNET__")
	// Shutdown Done executed after shutting down the server
	ShutdownDone func()
	// CloseHotRestart
	CloseHotRestart bool
	fileMd5         string
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
	fileMd5, _ = zstring.Md5File(os.Args[0])
}

// New returns a newly initialized Engine object that implements the Engine
func New(serverName ...string) *Engine {
	name := defaultServerName
	if len(serverName) > 0 {
		name = serverName[0]
	}

	log := zlog.New("[" + name + "] ")
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
	}
	r.pool.New = func() interface{} {
		return r.NewContext(nil, nil)
	}
	if _, ok := zservers[name]; ok {
		Log.Fatal("serverName: [", name, "] it already exists")
	}
	zservers[name] = r
	// r.Use(withRequestLog)
	return r
}

// Server Server
func Server(serverName ...string) (engine *Engine, ok bool) {
	name := defaultServerName
	if len(serverName) > 0 {
		name = serverName[0]
	}
	if engine, ok = zservers[name]; !ok {
		Log.Warnf("serverName: %s is not", name)
		engine = New(name)
	}
	return
}

// SetShutdown Set Shutdown Func
func SetShutdown(done func()) {
	ShutdownDone = done
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

// GetMiddleware GetMiddleware
func (e *Engine) GetMiddleware() []HandlerFunc {
	return e.router.middleware
}

// SetCustomMethodField SetCustomMethodField
func (e *Engine) SetCustomMethodField(field string) {
	e.customMethodType = field
}

// CloseHotRestartFileMd5 CloseHotRestartFileMd5
func CloseHotRestartFileMd5() {
	fileMd5 = ""
}

// SetTemplateFuncMap Set Template Func
func (e *Engine) SetTemplateFuncMap(funcMap template.FuncMap) {
	e.templateFuncMap = funcMap
}

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
		templatesDebug(t)
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
	case DebugMode:
		level = zlog.LogDump
		e.webMode = debugCode
	case TestMode:
		level = zlog.LogInfo
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

// IsDebug IsDebug
func (e *Engine) IsDebug() bool {
	return e.webMode > prodCode
}

// SetTimeout setTimeout
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
			var err error
			isTls := cfg.Cert != "" || cfg.Config != nil
			addr := getAddr(cfg.addr)
			hostname := getHostname(addr, isTls)
			srv := &http.Server{
				Addr:         addr,
				Handler:      e,
				ReadTimeout:  e.readTimeout,
				WriteTimeout: e.writeTimeout,
				// MaxHeaderBytes: 1 << 20,
			}

			srvMap.Store(addr, &serverMap{e, srv})

			time.AfterFunc(time.Millisecond*100, func() {
				wrapPid := e.Log.ColorTextWrap(zlog.ColorLightGrey, fmt.Sprintf("Pid: %d", os.Getpid()))
				wrapMode := ""
				if e.webMode > 0 {
					wrapMode = e.Log.ColorTextWrap(zlog.ColorYellow, fmt.Sprintf("%s ", strings.ToUpper(e.webModeName)))
				}
				e.Log.Successf("%s %s %s%s\n", "Listen:", e.Log.ColorTextWrap(zlog.ColorLightGreen, e.Log.OpTextWrap(zlog.OpBold, hostname)), wrapMode, wrapPid)
			})
			wg.Done()
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
				err = srv.ListenAndServeTLS(cfg.Cert, cfg.Key)
			} else {
				err = srv.ListenAndServe()
			}
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

// Run run serve
func Run() {
	var (
		srvs []*serverMap
		m    sync.WaitGroup
	)

	for _, e := range zservers {
		ss := e.StartUp()
		m.Add(len(ss))
		srvs = append(srvs, ss...)
	}

	sigkill := zcli.KillSignal()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if !sigkill && !CloseHotRestart {
		runNewProcess()
	}

	for _, s := range srvs {
		r := s.engine
		if sigkill {
			r.Log.Info("Shutdown server ...")
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
		m.Done()
	}
	m.Wait()
	if ShutdownDone != nil {
		ShutdownDone()
	}
	time.Sleep(100 * time.Millisecond)
}

func runNewProcess() {
	if fileMd5 == "" {
		Log.Warn("ignore execution file md5 check")
	}
	_, err := zshell.RunNewProcess(fileMd5)
	if err != nil {
		Log.Error(err)
	}
}
