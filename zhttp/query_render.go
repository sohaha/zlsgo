package zhttp

import (
	"bytes"
	"github.com/sohaha/zlsgo/zstring"
	"golang.org/x/net/html"
)

func (r QueryHTML) Exist() bool {
	return r.node != nil && r.node.Data != ""
}

func (r QueryHTML) Attr(key string) string {
	return r.Attrs()[key]
}

func (r QueryHTML) Attrs() (attrs map[string]string) {
	node := r.getNode()
	if node.Type != html.ElementNode || len(node.Attr) == 0 {
		return
	}
	return getAttrValue(node.Attr)
}

func (r QueryHTML) Name() string {
	return r.getNode().Data
}

func (r QueryHTML) Text(trimSpace ...bool) string {
	text := getElText(r, false)
	if len(trimSpace) > 0 && trimSpace[0] {
		text = zstring.TrimSpace(text)
	}
	return text
}

func (r QueryHTML) FullText(trimSpace ...bool) string {
	text := getElText(r, true)
	if len(trimSpace) > 0 && trimSpace[0] {
		text = zstring.TrimSpace(text)
	}
	return text
}

func (r QueryHTML) HTML(trimSpace ...bool) string {
	var b bytes.Buffer
	if err := html.Render(&b, r.getNode()); err != nil {
		return ""
	}
	text := b.String()
	if len(trimSpace) > 0 && trimSpace[0] {
		text = zstring.TrimSpace(text)
	}
	return text
}
