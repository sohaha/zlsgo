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
			matchName = append(matchName, res[1])
			if res[1] == idKey {
				pattern.WriteString("/(")
				pattern.WriteString(idPattern)
				pattern.WriteString(")")
			} else if res[1] == allKey {
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
