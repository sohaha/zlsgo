package zhttp

import (
	"bytes"
	"fmt"

	"golang.org/x/net/html"
)

type (
	QueryHTML struct {
		node *html.Node
	}
)

func HTMLParse(HTML []byte) (doc QueryHTML, err error) {
	var n *html.Node
	n, err = html.Parse(bytes.NewReader(HTML))
	if err != nil {
		return
	}

	for n.Type != html.ElementNode {
		switch n.Type {
		case html.DocumentNode:
			n = n.FirstChild
		case html.DoctypeNode:
			n = n.NextSibling
		case html.CommentNode:
			n = n.NextSibling
		}
	}
	doc = QueryHTML{node: n}
	return
}

func (r QueryHTML) Child() (child []QueryHTML) {
	n := r.node.FirstChild
	for {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode {
			child = append(child, QueryHTML{node: n})
		}
		n = n.NextSibling
	}
}

func (r QueryHTML) Find(el string, args ...map[string]string) QueryHTML {
	t, _ := r.MustFind(el, args...)
	return t
}

func (r QueryHTML) MustFind(el string, args ...map[string]string) (QueryHTML, error) {
	n := findEl(r.node, el, arr2Attr(args), false)
	if len(n) == 0 {
		return QueryHTML{node: &html.Node{}}, fmt.Errorf("element `%s` not found", el)
	}
	return QueryHTML{node: n[0]}, nil
}

func (r *QueryHTML) FindAll(el string, args ...map[string]string) []QueryHTML {
	t, _ := r.MustFindAll(el, args...)
	return t
}

func (r QueryHTML) MustFindAll(el string, args ...map[string]string) ([]QueryHTML, error) {
	n := findEl(r.node, el, arr2Attr(args), true)
	if len(n) == 0 {
		return []QueryHTML{{node: &html.Node{}}}, fmt.Errorf("element `%s` not found", el)
	}
	var e []QueryHTML
	for i := range n {
		e = append(e, QueryHTML{node: n[i]})
	}
	return e, nil
}
