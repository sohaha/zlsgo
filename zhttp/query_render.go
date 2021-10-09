package zhttp

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func (e Els) ForEach(f func(index int, el QueryHTML) bool) {
	for i, v := range e {
		if !f(i, v) {
			break
		}
	}
}

func (r QueryHTML) String() string {
	return r.HTML(true)
}

func (r QueryHTML) Exist() bool {
	return r.node != nil && r.node.Data != ""
}

func (r QueryHTML) Attr(key string) string {
	return r.Attrs()[key]
}

func (r QueryHTML) Attrs() map[string]string {
	node := r.getNode()
	if node.Type != html.ElementNode || len(node.Attr) == 0 {
		return make(map[string]string)
	}
	return getAttrValue(node.Attr)
}

func (r QueryHTML) Name() string {
	return r.getNode().Data
}

func (r QueryHTML) Text(trimSpace ...bool) string {
	text := getElText(r, false)
	if len(trimSpace) > 0 && trimSpace[0] {
		text = strings.TrimSpace(text)
	}
	return text
}

func (r QueryHTML) FullText(trimSpace ...bool) string {
	text := getElText(r, true)
	if len(trimSpace) > 0 && trimSpace[0] {
		text = strings.TrimSpace(text)
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
		text = strings.TrimSpace(text)
	}
	return text
}
