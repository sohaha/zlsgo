package znet

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	
	// "strconv"
	// "sync"
	"time"
	
	"github.com/sohaha/zlsgo/zlog"
	// "github.com/sohaha/zlsgo/zvar"
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
	defaultLocation   = "Local" // "Asia/Shanghai"
	releaseCode       = iota
	debugCode
	testCode
)

type (
	// Context context
	Context struct {
		Writer  http.ResponseWriter
		Request *http.Request
		Engine  *Engine
		Info    *info
		Log     *zlog.Logger
		next    HandlerFunc
	}
	// Engine is a simple HTTP route multiplexer that parses a request path
	Engine struct {
		// Log Log
		Log                *zlog.Logger
		readTimeout        time.Duration
		writeTimeout       time.Duration
		webModeName        string
		webMode            int
		timeLocation       *time.Location
		ShowFavicon        bool
		MaxMultipartMemory int64
		customMethodType   string
		addr               []addrSt
		router             *router
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
		Code       int
		Mutex      sync.RWMutex
		StartTime  time.Time
		StopHandle bool
		handlerLen int
		middleware []HandlerFunc
		Data       map[string]interface{}
	}
	
	// HandlerFunc HandlerFunc
	HandlerFunc func(*Context)
	// MiddlewareFunc MiddlewareFunc
	MiddlewareFunc func(*Context, HandlerFunc)
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
	Log         = zlog.New(zlog.ColorTextWrap(zlog.ColorGreen, "[Z] "))
	zservers    = map[string]*Engine{}
	defaultAddr = addrSt{
		addr: ":3788",
	}
)

func init() {
	Log.ResetFlags(zlog.BitTime | zlog.BitLevel)
}

// New returns a newly initialized Engine object that implements the Engine
func New(serverName ...string) *Engine {
	name := defaultServerName
	if len(serverName) > 0 {
		name = serverName[0]
	}
	
	log := zlog.New("[" + name + "] ")
	log.ResetFlags(zlog.BitTime | zlog.BitLevel)
	log.SetLogLevel(zlog.LogWarn)
	
	location, _ := time.LoadLocation(defaultLocation)
	
	route := &router{
		trees: make(map[string]*Tree),
	}
	r := &Engine{
		Log:                log,
		router:             route,
		readTimeout:        0 * time.Second,
		writeTimeout:       0 * time.Second,
		webModeName:        ReleaseMode,
		webMode:            releaseCode,
		timeLocation:       location,
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

// SetTimeLocation timezone
func (e *Engine) SetTimeLocation(location string) {
	e.timeLocation, _ = time.LoadLocation(location)
}

// SetCustomMethodField SetCustomMethodField
func (e *Engine) SetCustomMethodField(field string) {
	e.customMethodType = field
}

// SetMode Setting Server Mode
func (e *Engine) SetMode(value string) {
	var level int
	switch value {
	case ReleaseMode, "":
		level = zlog.LogWarn
		e.webMode = releaseCode
	case DebugMode:
		level = zlog.LogDebug
		e.webMode = debugCode
	case TestMode:
		level = zlog.LogSuccess
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
				
				e.Log.Success(e.Log.ColorBackgroundWrap(zlog.ColorLightGreen, zlog.ColorDefault, e.Log.OpTextWrap(zlog.OpBold, "Listen "+hostname)))
				var err error
				if isTls {
					if cfg.Config != nil {
						srv.TLSConfig = cfg.Config
					}
					if cfg.HTTPAddr != "" {
						go func(e *Engine) {
							newHostname := "http://" + resolveHostname(cfg.HTTPAddr)
							e.Log.Success(e.Log.ColorBackgroundWrap(zlog.ColorYellow, zlog.ColorDefault, e.Log.OpTextWrap(zlog.OpBold, "Listen "+newHostname)))
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
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	srvMap.Range(func(key, value interface{}) bool {
		go func(value interface{}) {
			if s, ok := value.(*serverMap); ok {
				r := s.engine
				r.Log.Warn("Shutdown server ...")
				if err := s.srv.Shutdown(ctx); err != nil {
					r.Log.Error(err)
				} else {
					r.Log.Success("Shutdown server done")
				}
				if iskill {
					m.Done()
				} else {
					runNewProcess(s.srv)
					m.Done()
				}
			}
		}(value)
		return true
	})
	
	m.Wait()
}

func runNewProcess(srv *http.Server) {
	Log.Warn("In development...")
}
