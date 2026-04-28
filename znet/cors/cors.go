package cors

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	Config struct {
		CustomHandler Handler
		methods       string
		credentials   string
		headers       string
		exposeHeaders string
		Domains       []string
		Methods       []string
		Credentials   []string
		Headers       []string
		ExposeHeaders []string
		once          sync.Once
	}
	Handler func(conf *Config, c *znet.Context)
)

const (
	SafeHeaders = "Content-Type,Authorization,X-Requested-With,Accept,Origin,Cache-Control,X-File-Name,X-CSRF-Token"
)

func Default() znet.HandlerFunc {
	return New(&Config{})
}

func newAllowOrigins(allowAllHeaders bool) znet.HandlerFunc {
	config := &Config{
		Domains: []string{"*"},
	}
	if allowAllHeaders {
		config.Headers = []string{"*"}
	}
	return New(config)
}

func AllowAll() znet.HandlerFunc {
	return newAllowOrigins(true)
}

func AllowAllOrigins() znet.HandlerFunc {
	return newAllowOrigins(false)
}

func NewAllowHeaders() (addAllowHeader func(header string), handler znet.HandlerFunc) {
	conf := &Config{
		Headers: []string{},
		Domains: []string{"*"},
	}
	handler = New(conf)

	return func(header string) {
		if header = strings.TrimSpace(header); header != "" {
			conf.Headers = append(conf.Headers, header)
			conf.once = sync.Once{}
		}
	}, handler
}

func validateConfig(conf *Config) error {
	for _, domain := range conf.Domains {
		if domain != "*" && !strings.Contains(domain, "://") {
			return fmt.Errorf("invalid domain format: %s, should include protocol", domain)
		}
	}

	validMethods := map[string]bool{
		http.MethodGet: true, http.MethodHead: true, http.MethodPost: true,
		http.MethodPut: true, http.MethodPatch: true, http.MethodDelete: true,
		http.MethodConnect: true, http.MethodOptions: true, http.MethodTrace: true,
	}
	for _, method := range conf.Methods {
		if !validMethods[strings.ToUpper(method)] {
			return fmt.Errorf("invalid HTTP method: %s", method)
		}
	}

	return nil
}

func extractOriginFromReferer(referer string) string {
	if referer == "" {
		return ""
	}
	u, err := url.Parse(referer)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func (conf *Config) initConfig() {
	conf.once.Do(func() {
		if len(conf.Methods) == 0 {
			conf.Methods = []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodConnect,
				http.MethodOptions,
				http.MethodTrace,
			}
		}
		conf.methods = strings.Join(conf.Methods, ", ")

		if len(conf.Credentials) == 0 {
			conf.Credentials = []string{"true"}
		}
		conf.credentials = strings.Join(conf.Credentials, ", ")

		if len(conf.Headers) != 0 {
			// Check if any header is "*" (allow all)
			for _, header := range conf.Headers {
				if header == "*" {
					conf.headers = "*"
					break
				}
			}
			if conf.headers != "*" {
				conf.headers = strings.Join(conf.Headers, ", ")
			}
		} else {
			conf.headers = SafeHeaders
		}

		if len(conf.ExposeHeaders) > 0 {
			conf.exposeHeaders = strings.Join(conf.ExposeHeaders, ", ")
		}
	})
}

func New(conf *Config) znet.HandlerFunc {
	if conf == nil {
		conf = &Config{}
	}

	if err := validateConfig(conf); err != nil {
		panic(fmt.Sprintf("invalid CORS config: %v", err))
	}

	return func(c *znet.Context) {
		conf.initConfig()

		if applyCors(c, conf) {
			c.Next()
		}
	}
}

func validateOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	if len(origin) > 2048 {
		return false
	}

	if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
		return false
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	host := parsed.Hostname()
	if host == "" || strings.ContainsAny(host, " \t\n\r") {
		return false
	}

	return true
}

func isOriginAllowed(origin string, conf *Config) bool {
	if len(conf.Domains) == 0 {
		return false
	}

	for _, domain := range conf.Domains {
		if zstring.Match(origin, domain) {
			return true
		}
	}

	return false
}

func getAllowedHeaders(conf *Config, req *http.Request) string {
	if conf.headers == "*" {
		if requestedHeaders := req.Header.Get("Access-Control-Request-Headers"); requestedHeaders != "" {
			return requestedHeaders
		}
		headers := make([]string, 0, len(req.Header))
		for k := range req.Header {
			headers = append(headers, k)
		}
		return strings.Join(headers, ", ")
	}

	return conf.headers
}

func applyCors(c *znet.Context, conf *Config) bool {
	allowedOrigin := c.GetHeader("Origin")
	if allowedOrigin == "" {
		referer := c.GetHeader("Referer")
		if referer != "" {
			allowedOrigin = extractOriginFromReferer(referer)
		}
	}

	if allowedOrigin == "" {
		return true
	}

	if !validateOrigin(allowedOrigin) {
		c.Abort(http.StatusBadRequest)
		return false
	}

	if !isOriginAllowed(allowedOrigin, conf) {
		c.Abort(http.StatusForbidden)
		return false
	}

	allowHeaders := getAllowedHeaders(conf, c.Request)

	headers := map[string]string{
		"Access-Control-Allow-Methods":     conf.methods,
		"Access-Control-Allow-Credentials": conf.credentials,
		"Access-Control-Allow-Headers":     allowHeaders,
		"Access-Control-Allow-Origin":      allowedOrigin,
	}

	if conf.exposeHeaders != "" {
		headers["Access-Control-Expose-Headers"] = conf.exposeHeaders
	}

	for k, v := range headers {
		c.SetHeader(k, v, true)
	}

	if conf.CustomHandler != nil {
		conf.CustomHandler(conf, c)
	}

	if c.Request.Method == http.MethodOptions {
		c.Abort(http.StatusNoContent)
		return false
	}

	return true
}
