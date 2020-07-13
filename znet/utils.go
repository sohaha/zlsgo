package znet

import (
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
)

func completionPath(path, prefix string) string {
	if path != "" {
		tmp := zstring.Buffer()
		// prefixHasSuffix := strings.HasSuffix(prefix, "/")
		pathHasPrefix := strings.HasPrefix(path, "/")
		if prefix == "" && !pathHasPrefix {
			prefix = "/"
		} else if pathHasPrefix && strings.HasSuffix(prefix, "/") {
			prefix = strings.TrimSuffix(prefix, "/")
		}
		if prefix != "" && !strings.HasPrefix(prefix, "/") {
			tmp.WriteString("/")
		}
		tmp.WriteString(prefix)
		if !strings.HasSuffix(prefix, "/") {
			tmp.WriteString("/")
		}
		tmp.WriteString(strings.TrimPrefix(path, "/"))
		path = tmp.String()
	} else {
		path = prefix
	}

	return path
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
	if zfile.FileExist(file) {
		t, err = template.ParseFiles(templateFile...)
		if err == nil && funcMap != nil {
			t.Funcs(funcMap)
		}
	} else {
		t = template.New(file)
		if funcMap != nil {
			t.Funcs(funcMap)
		}
		t, err = t.Parse(file)
	}
	return
}

func parsPattern(res []string) (string, []string) {
	var (
		matchName []string
		pattern   = zstring.Buffer()
	)
	for _, str := range res {
		if str == "" {
			continue
		}
		strLen := len(str)
		firstChar := string(str[0])
		lastChar := string(str[strLen-1])
		// todo Need to optimize
		if firstChar == "{" && lastChar == "}" {
			matchStr := string(str[1 : strLen-1])
			res := strings.Split(matchStr, ":")
			matchName = append(matchName, res[0])
			pattern.WriteString("/(")
			pattern.WriteString(res[1])
			pattern.WriteString(")")
		} else if firstChar == ":" {
			matchStr := str
			res := strings.Split(matchStr, ":")
			key := res[1]
			if key == "full" {
				key = allKey
			}
			matchName = append(matchName, key)
			if key == idKey {
				pattern.WriteString("/(")
				pattern.WriteString(idPattern)
				pattern.WriteString(")")
			} else if key == allKey {
				pattern.WriteString("/(")
				pattern.WriteString(allPattern)
				pattern.WriteString(")")
			} else {
				pattern.WriteString("/(")
				pattern.WriteString(defaultPattern)
				pattern.WriteString(")")
			}
		} else if firstChar == "*" {
			pattern.WriteString("/(")
			pattern.WriteString(allPattern)
			pattern.WriteString(")")
			matchName = append(matchName, allKey)
		} else {
			pattern.WriteString("/")
			pattern.WriteString(str)
		}
	}
	return pattern.String(), matchName
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
		Code:          http.StatusOK,
		startTime:     time.Time{},
		header:        map[string]string{},
		customizeData: map[string]interface{}{},
	}
}

func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
	c.Lock()
	c.Request = r
	c.Writer = w
	c.Code = http.StatusOK
	c.startTime = time.Now()
	c.stopHandle = false
	c.middleware = []HandlerFunc{}[:0]
	c.customizeData = map[string]interface{}{}
	c.header = map[string]string{}
	c.render = nil
	c.rawData = c.rawData[:0]
	c.Unlock()
}

func (e *Engine) acquireContext() *Context {
	return e.pool.Get().(*Context)
}

func (e *Engine) releaseContext(c *Context) {
	e.pool.Put(c)
}
