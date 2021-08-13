package zhttp

import (
	"bytes"

	"golang.org/x/net/html"
)

type (
	QueryHTML struct {
		node *html.Node
	}
	Els []QueryHTML
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

func (r *QueryHTML) getNode() *html.Node {
	if r.node == nil {
		r.node = &html.Node{}
	}
	return r.node
}

func (r QueryHTML) SelectChild(el string, args ...map[string]string) QueryHTML {
	var (
		node  *html.Node
		exist bool
	)
	forChild(r.getNode(), func(n *html.Node) bool {
		elArr := matchEl(n, el, arr2Attr(args))
		exist = elArr != nil
		if exist {
			node = elArr
			return false
		}
		return true
	})
	if !exist {
		return QueryHTML{node: &html.Node{}}
	}
	return QueryHTML{node: node}
}

func (r QueryHTML) SelectAllChild(el string, args ...map[string]string) (arr Els) {
	forChild(r.getNode(), func(n *html.Node) bool {
		elArr := matchEl(n, el, arr2Attr(args))
		exist := elArr != nil
		if exist {
			arr = append(arr, QueryHTML{node: elArr})
		}
		return true
	})
	return
}

// Deprecated: please use SelectAllChild("")
// Child All child elements
func (r QueryHTML) Child() (childs []QueryHTML) {
	r.ForEachChild(func(index int, child QueryHTML) bool {
		childs = append(childs, child)
		return true
	})
	return
}

func (r QueryHTML) ForEachChild(f func(index int, child QueryHTML) bool) {
	i := -1
	forChild(r.getNode(), func(n *html.Node) bool {
		i++
		return f(i, QueryHTML{node: n})
	})
}

func (r QueryHTML) NthChild(index int) QueryHTML {
	i := 0
	doc := QueryHTML{}
	forChild(r.getNode(), func(n *html.Node) bool {
		i++
		if i == index {
			doc.node = n
			return false
		}
		return true
	})
	return doc
}

func (r QueryHTML) Select(el string, args ...map[string]string) QueryHTML {
	n := findChild(r.getNode(), el, args, false)
	if len(n) == 0 {
		return QueryHTML{node: &html.Node{}}
	}
	return QueryHTML{node: n[0]}
}

func (r QueryHTML) SelectAll(el string, args ...map[string]string) (arr Els) {
	n := findChild(r.getNode(), el, args, true)
	l := len(n)
	if l == 0 {
		return
	}
	arr = make([]QueryHTML, l)
	for i := range n {
		arr[i] = QueryHTML{node: n[i]}
	}
	return arr
}

func (r QueryHTML) SelectBrother(el string, args ...map[string]string) QueryHTML {
	parent := r.SelectParent("")
	child := parent.SelectAllChild(el, args...)
	index := 0
	brother := QueryHTML{}
	for i := range child {
		q := child[i]
		if q == r {
			index = i + 1
			if len(child) > index {
				brother = child[index]
			}
			break
		}
	}
	return brother
}

func (r QueryHTML) SelectParent(el string, args ...map[string]string) QueryHTML {
	n := r.getNode()
	attr := arr2Attr(args)
	for {
		n = n.Parent
		if n == nil {
			break
		}
		p := matchEl(n, el, attr)
		if p != nil {
			return QueryHTML{node: p}
		}
	}

	return QueryHTML{node: &html.Node{}}
}

func (r QueryHTML) Find(el string) QueryHTML {
	level := parseSelector(el)
	if len(level) == 0 {
		return QueryHTML{node: &html.Node{}}
	}
	n := r
	for i := range level {
		l := level[i]
		if l.Child {
			n = n.SelectChild(l.Name, l.Attr)
		} else if l.Brother {
			n = n.SelectBrother(l.Name, l.Attr)
		} else {
			n = n.Select(l.Name, l.Attr)
		}
		if !n.Exist() {
			return QueryHTML{node: &html.Node{}}
		}
	}
	return n
}

func parseSelector(el string) []*selector {
	var (
		ss []*selector
		s  *selector
	)
	key, l := "", len(el)
	if l > 0 {
		s = &selector{i: 0, Attr: make(map[string]string, 0)}
		for i := 0; i < l; {
			v := el[i]
			add := 0
			switch v {
			case '#':
				s.appendAttr(key, el, i)
				key = "id"
			case '.':
				s.appendAttr(key, el, i)
				key = "class"
			case ' ', '>', '~':
				s.appendAttr(key, el, i)
				if s.Name != "" || len(s.Attr) != 0 {
					ss = append(ss, s)
					s = &selector{i: i + 1, Attr: make(map[string]string)}
					key = ""
				}
				if v == '>' {
					s.Child = true
				} else if v == '~' {
					s.Brother = true
				}
			}
			i = i + 1 + add
		}
	}

	if s != nil {
		s.appendAttr(key, el, l)
		ss = append(ss, s)
	}
	return ss
}
