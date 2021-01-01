package zhttp

import (
	"bytes"

	"golang.org/x/net/html"
)

func (r QueryHTML) Exist() bool {
	return r.node.Data != ""
}

func (r QueryHTML) Attr(key string) string {
	return r.Attrs()[key]
}

func (r QueryHTML) Attrs() (attrs map[string]string) {
	if r.node.Type != html.ElementNode || len(r.node.Attr) == 0 {
		return
	}
	return getAttrValue(r.node.Attr)
}

func (r QueryHTML) Name() string {
	return r.node.Data
}

func (r QueryHTML) Text() string {
	return getElText(r, false)
}

func (r QueryHTML) FullText() string {
	return getElText(r, true)
}

func (r QueryHTML) HTML() string {
	var b bytes.Buffer
	if err := html.Render(&b, r.node); err != nil {
		return ""
	}
	return b.String()
}
