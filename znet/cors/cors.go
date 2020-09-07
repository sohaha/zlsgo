package cors

import (
	"net/http"

	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	// Config cors configuration
	Config struct {
		// Domains whitelist domain name
		Domains []string
	}
)

func Default() znet.HandlerFunc {
	conf := &Config{Domains: []string{}}
	return func(c *znet.Context) {
		if applyCors(c, conf) {
			c.Next()
		}
	}
}

func New(conf *Config) znet.HandlerFunc {
	return func(c *znet.Context) {
		if applyCors(c, conf) {
			c.Next()
		}
	}
}

func applyCors(c *znet.Context, conf *Config) bool {
	origin := c.GetHeader("Origin")
	if len(origin) == 0 {
		return true
	}

	domains := conf.Domains
	if len(domains) > 0 {
		adopt := false
		for k := range domains {
			if adopt = zstring.Match(origin, domains[k]); adopt {
				break
			}
		}
		if !adopt {
			c.Abort(http.StatusForbidden)
			return false
		}
	}

	c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.SetHeader("Access-Control-Allow-Credentials", "true")
	c.SetHeader("Access-Control-Allow-Headers", "X-Requested-With")
	c.SetHeader("Access-Control-Allow-Origin", origin)

	if c.Request.Method == "OPTIONS" {
		c.Abort(http.StatusNoContent)
		return false
	}

	return true
}
