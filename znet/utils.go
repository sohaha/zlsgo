package znet

import (
	"github.com/sohaha/zlsgo/zstring"
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
