package znet

import (
	"github.com/sohaha/zlsgo/zstring"
	"net/http"
	"strings"
)

func completionPath(path, prefix string) string {
	if prefix != "" {
		if path != "" {
			tmp := zstring.Buffer()
			tmp.WriteString(prefix)
			tmp.WriteString("/")
			tmp.WriteString(strings.TrimPrefix(path, "/"))
			path = tmp.String()
		} else {
			path = prefix
		}
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

type tlsRedirectHandler struct {
	Domain string
}

func (t *tlsRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, t.Domain+r.URL.String(), 301)
}
