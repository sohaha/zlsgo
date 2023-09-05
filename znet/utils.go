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

type utils struct {
	ContextKey contextKeyType
}

var Utils = utils{
	ContextKey: contextKeyType{},
}

const (
	defaultPattern = `[^\/.]+`
	idPattern      = `[\d]+`
	idKey          = `id`
	allPattern     = `.*`
	allKey         = `*`
)

var matchCache = zcache.NewFast(func(o *zcache.Options) {
	o.LRU2Cap = 100
})

// URLMatchAndParse checks if the request matches the route path and returns a map of the parsed
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

func getHostname(addr string, isTls bool) string {
	hostname := "http://"
	if isTls {
		hostname = "https://"
	}
	return hostname + resolveHostname(addr)
}

func (u utils) TreeFind(t *Tree, path string) (handlerFn, []handlerFn, bool) {
	nodes := t.Find(path, false)
	for i := range nodes {
		node := nodes[i]
		if node.handle != nil {
			if node.path == path {
				return node.handle, node.middleware, true
			}
		}
	}

	if len(nodes) == 0 {
		res := strings.Split(path, "/")
		p := ""
		if len(res) == 1 {
			p = res[0]
		} else {
			p = res[1]
		}
		nodes := t.Find(p, true)
		for _, node := range nodes {
			if handler := node.handle; handler != nil && node.path != path {
				if matchParamsMap, ok := u.URLMatchAndParse(path, node.path); ok {
					return func(c *Context) error {
						req := c.Request
						ctx := context.WithValue(req.Context(), u.ContextKey, matchParamsMap)
						c.Request = req.WithContext(ctx)
						return node.Handle()(c)
					}, node.middleware, true
				}
			}
		}
	}
	return nil, nil, false
}

func (_ utils) CompletionPath(p, prefix string) string {
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

// func (utils) IsAbort(c *Context) bool {
// 	return c.stopHandle.Load()
// }

// AppendHandler append handler to context, Use caution
func (utils) AppendHandler(c *Context, handlers ...Handler) {
	hl := len(handlers)
	if hl == 0 {
		return
	}

	for i := range handlers {
		c.middleware = append(c.middleware, Utils.ParseHandlerFunc(handlers[i]))
	}
}

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

func resolveHostname(addrString string) string {
	if strings.Index(addrString, ":") == 0 {
		return "127.0.0.1" + addrString
	}
	return addrString
}

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

type tlsRedirectHandler struct {
	Domain string
}

func (t *tlsRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, t.Domain+r.URL.String(), http.StatusMovedPermanently)
}

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

func (c *Context) clone(w http.ResponseWriter, r *http.Request) {
	c.Request = r
	c.Writer = w
	c.injector = zdi.New(c.Engine.injector)
	c.injector.Maps(c)
	c.startTime = time.Now()
	c.renderError = defErrorHandler()
	c.stopHandle.Store(false)
	c.done.Store(false)
}

func (e *Engine) acquireContext() *Context {
	return e.pool.Get().(*Context)
}

func (e *Engine) releaseContext(c *Context) {
	c.prevData.Code.Store(0)
	c.mu.Lock()
	c.middleware = c.middleware[0:0]
	c.customizeData = map[string]interface{}{}
	c.header = map[string][]string{}
	c.render = nil
	c.renderError = nil
	c.cacheJSON = nil
	c.cacheQuery = nil
	c.cacheForm = nil
	c.injector = nil
	c.rawData = nil
	c.prevData.Content = c.prevData.Content[0:0]
	c.prevData.Type = ContentTypePlain
	c.mu.Unlock()
	e.pool.Put(c)
}

func (s *serverMap) GetAddr() string {
	return s.srv.Addr
}
