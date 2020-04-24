package znet

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zshell"
	"github.com/sohaha/zlsgo/zstring"
)

const (
	defaultMultipartMemory = 32 << 20 // 32 MB
	// DebugMode debug
	DebugMode = "debug"
	// ReleaseMode release
	ReleaseMode = "release"
	// TestMode test
	TestMode          = "test"
	defaultServerName = "Z"
	releaseCode       = iota
	debugCode
	testCode
)

type (
	// Context context
	Context struct {
		Writer  http.ResponseWriter
		Request *http.Request
		rawData string
		Engine  *Engine
		Info    *info
		Log     *zlog.Logger
		Cache   *zcache.Table
		next    HandlerFunc
	}
	// Engine is a simple HTTP route multiplexer that parses a request path
	Engine struct {
		// Log Log
		Log                *zlog.Logger
		Cache              *zcache.Table
		readTimeout        time.Duration
		writeTimeout       time.Duration
		webModeName        string
		webMode            int
		ShowFavicon        bool
		MaxMultipartMemory int64
		customMethodType   string
		addr               []addrSt
		router             *router
		preHandle          func(context *Context) bool
	}

	TlsCfg struct {
		Cert           string
		Key            string
		HTTPAddr       string
		HTTPProcessing interface{}
		Config         *tls.Config
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

	info struct {
		Code int
		sync.RWMutex
		StartTime     time.Time
		StopHandle    bool
		handlerLen    int
		middleware    []HandlerFunc
		customizeData map[string]interface{}
		heades        map[string]string
		render        render
	}

	// HandlerFunc HandlerFunc
	HandlerFunc func(c *Context)
	// MiddlewareFunc MiddlewareFunc
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
		trees: make(map[string]*Tree),
	}
	r := &Engine{
		Log:                log,
		Cache:              Cache,
		router:             route,
		readTimeout:        0 * time.Second,
		writeTimeout:       0 * time.Second,
		webModeName:        ReleaseMode,
		webMode:            releaseCode,
		addr:               []addrSt{defaultAddr},
		MaxMultipartMemory: defaultMultipartMemory,
	}

	if _, ok := zservers[name]; ok {
		Log.Fatalf("serverName: %s is has", name)
	}

	zservers[name] = r

	r.Use(withRequestLog)

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

// SetMode Setting Server Mode
func (e *Engine) SetMode(value string) {
	var level int
	switch value {
	case ReleaseMode, "":
		level = zlog.LogSuccess
		e.webMode = releaseCode
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
		value = ReleaseMode
	}
	e.webModeName = value
	e.Log.SetLogLevel(level)
}

// IsDebug IsDebug
func (e *Engine) IsDebug() bool {
	return e.webMode > releaseCode
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

// Run run serve
func Run() {
	var (
		srvMap sync.Map
		m      sync.WaitGroup
	)

	for _, e := range zservers {
		for _, cfg := range e.addr {
			go func(cfg addrSt, e *Engine) {
				m.Add(1)
				isTls := cfg.Cert != "" || cfg.Config != nil
				addr := cfg.addr
				if !strings.Contains(addr, ":") {
					addr = ":" + addr
				}
				hostname := "http://"
				if isTls {
					hostname = "https://"
				}
				hostname += resolveHostname(addr)
				srv := &http.Server{
					Addr:         addr,
					Handler:      e,
					ReadTimeout:  e.readTimeout,
					WriteTimeout: e.writeTimeout,
					// MaxHeaderBytes: 1 << 20,
				}
				srvMap.Store(addr, &serverMap{e, srv})

				wrapPid := e.Log.ColorTextWrap(zlog.ColorLightGrey, fmt.Sprintf("Pid: %d", os.Getpid()))
				time.AfterFunc(time.Millisecond*100, func() {
					e.Log.Successf("%s %s  %s\n", "Listen:", e.Log.ColorTextWrap(zlog.ColorLightGreen, e.Log.OpTextWrap(zlog.OpBold, hostname)), wrapPid)
				})
				var err error
				if isTls {
					if cfg.Config != nil {
						srv.TLSConfig = cfg.Config
					}
					if cfg.HTTPAddr != "" {
						go func(e *Engine) {
							newHostname := "http://" + resolveHostname(cfg.HTTPAddr)
							e.Log.Success(e.Log.ColorBackgroundWrap(zlog.ColorYellow, zlog.ColorDefault, e.Log.OpTextWrap(zlog.OpBold, "Listen: "+newHostname)))
							var err error
							switch processing := cfg.HTTPProcessing.(type) {
							case string:
								// e.Log.Warn(addr + " Redirect " + cfg.HTTPAddr)
								err = http.ListenAndServe(cfg.HTTPAddr, &tlsRedirectHandler{Domain: processing})
							case http.Handler:
								err = http.ListenAndServe(cfg.HTTPAddr, processing)
							default:
								err = http.ListenAndServe(cfg.HTTPAddr, e)
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
	}

	iskill := isKill()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if !iskill && !CloseHotRestart {
		runNewProcess()
	}

	srvMap.Range(func(key, value interface{}) bool {
		go func(value interface{}) {
			if s, ok := value.(*serverMap); ok {
				r := s.engine
				if iskill {
					r.Log.Info("Shutdown server ...")
				}
				err := s.srv.Shutdown(ctx)
				if err != nil {
					if iskill {
						r.Log.Error("Timeout forced close")
					}
					_ = s.srv.Close()
				} else {
					if iskill {
						r.Log.Success("Shutdown server done")
					}
				}
				m.Done()
			}
		}(value)
		return true
	})

	m.Wait()
	if ShutdownDone != nil {
		ShutdownDone()
	}
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
