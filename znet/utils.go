package znet

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
)

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

func completionPath(path, prefix string) string {
	if path == "/" {
		if prefix == "/" {
			return prefix
		}
		return prefix + path
	} else if path == "" {
		return prefix
	}
	n := zstring.TrimSpace(strings.Join([]string{prefix, path}, "/"))
	l := len(n)
	b := zstring.Buffer(l)
	b.WriteByte('/')
	for i := 1; i < l; i++ {
		if n[i] == '/' {
			if i == 0 || i == l-1 {
				continue
			}
			if n[i-1] == '/' {
				continue
			}
		}
		b.WriteByte(n[i])
	}
	n = b.String()
	if strings.HasSuffix(path, "/") {
		n = n + "/"
	}
	return n
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

func parsPattern(res []string, prefix string) (string, []string) {
	var (
		matchName []string
		pattern   string
	)
	for _, str := range res {
		if str == "" {
			continue
		}
		pattern = pattern + prefix
		l := len(str) - 1
		i := strings.Index(str, "}")
		i2 := strings.Index(str, "{")
		firstChar := string(str[0])
		// todo Need to optimize
		if i2 != -1 && i != -1 {
			// lastChar := string(str[l])
			if i == l && i2 == 0 {
				matchStr := str[1:l]
				res := strings.Split(matchStr, ":")
				matchName = append(matchName, res[0])
				pattern = pattern + "(" + res[1] + ")"
			} else {
				if i2 != 0 {
					p, m := parsPattern([]string{str[:i2]}, "")
					if p != "" {
						pattern = pattern + p
						matchName = append(matchName, m...)
					}
					str = str[i2:]
				}
				if i >= 0 {
					ni := i - i2
					matchStr := str[1:ni]
					res := strings.Split(matchStr, ":")
					matchName = append(matchName, res[0])
					pattern = pattern + "(" + res[1] + ")"
					p, m := parsPattern([]string{str[ni+1:]}, "")
					if p != "" {
						pattern = pattern + p
						matchName = append(matchName, m...)
					}
				} else {
					pattern = pattern + str
				}
			}

		} else if firstChar == ":" {
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
		} else if firstChar == "*" {
			pattern = pattern + "(" + allPattern + ")"
			matchName = append(matchName, allKey)
		} else {
			pattern = pattern + str
		}
	}
	return pattern, matchName
}

type tlsRedirectHandler struct {
	Domain string
}

func (t *tlsRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, t.Domain+r.URL.String(), 301)
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
		prevData: &PrevData{
			Code: http.StatusOK,
			Type: ContentTypePlain,
		},
	}
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Request = r
	c.Writer = w
	c.startTime = time.Now()
}

func (e *Engine) acquireContext() *Context {
	return e.pool.Get().(*Context)
}

func (e *Engine) releaseContext(c *Context) {
	c.l.Lock()
	c.prevData.Code = http.StatusOK
	c.stopHandle = false
	c.middleware = c.middleware[0:0]
	c.customizeData = map[string]interface{}{}
	c.header = map[string][]string{}
	c.render = nil
	c.cacheJSON = nil
	c.cacheQuery = nil
	c.rawData = c.rawData[0:0]
	c.prevData.Content = c.prevData.Content[0:0]
	c.prevData.Type = ContentTypePlain
	c.l.Unlock()
	e.pool.Put(c)
}
