package zutil

import (
	"strings"
)

func UnescapeHTML(s string) string {
	s = strings.Replace(s, "\\u003c", "<", -1)
	s = strings.Replace(s, "\\u003e", ">", -1)
	return strings.Replace(s, "\\u0026", "&", -1)
}
