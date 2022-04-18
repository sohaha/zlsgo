package zhttp

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	selector struct {
		Name    string
		Attr    map[string]string
		i       int
		Child   bool
		Brother bool
	}
)

func (s *selector) appendAttr(key, val string, index int) {
	if key == "" {
		s.Name = val[s.i:index]
	} else {
		val = val[s.i:index]
		if v, ok := s.Attr[key]; ok && v != "" {
			s.Attr[key] = s.Attr[key] + " " + val
		} else {
			s.Attr[key] = val
		}
	}
	s.i = index + 1
}

// ConvertCookie Parse Cookie String
func ConvertCookie(cookiesRaw string) map[string]*http.Cookie {
	cookie := map[string]*http.Cookie{}
	c := strings.Split(cookiesRaw, ";")
	for _, s := range c {
		v := strings.Split(zstring.TrimSpace(s), "=")
		if len(v) == 2 {
			name := zstring.TrimSpace(v[0])
			cookie[name] = &http.Cookie{Name: name, Value: v[1]}
		}
	}
	return cookie
}

// BodyJSON make the object be encoded in json format and set it to the request body
func BodyJSON(v interface{}) *bodyJson {
	return &bodyJson{v: v}
}

// BodyXML make the object be encoded in xml format and set it to the request body
func BodyXML(v interface{}) *bodyXml {
	return &bodyXml{v: v}
}

func File(path string, field ...string) interface{} {
	var matches []string
	path = zfile.RealPath(path)
	uploads := make([]FileUpload, 0)
	fieldName := "media"
	if len(field) > 0 {
		fieldName = field[0]
	}
	s, err := os.Stat(path)
	if err == nil && !s.IsDir() {
		file, _ := os.Open(path)
		return []FileUpload{{
			File:      file,
			FileName:  filepath.Base(path),
			FieldName: fieldName,
		}}
	}
	m, err := filepath.Glob(path)
	if err != nil {
		return err
	}
	matches = append(matches, m...)
	if len(matches) == 0 {
		return ErrNoMatched
	}

	for _, match := range matches {
		if s, e := os.Stat(match); e != nil || s.IsDir() {
			continue
		}
		file, _ := os.Open(match)
		uploads = append(uploads, FileUpload{
			File:      file,
			FileName:  filepath.Base(match),
			FieldName: fieldName,
		})
	}

	return uploads
}

var UserAgentLists = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	"Mozilla/5.0 (Linux; U; Android 2.3.6; zh-cn; GT-S5660 Build/GINGERBREAD) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1 MicroMessenger/4.5.255",
	"Mozilla/5.0 (X11; OpenBSD i386) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1944.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 5.1) Gecko/20100101 Firefox/14.0 Opera/12.0",
	"Mozilla/5.0 (compatible; Googlebot/2.1;+http://www.google.com/bot.html)",
}

func RandomUserAgent() Header {
	return Header{"User-Agent": UserAgentLists[zstring.RandInt(0, len(UserAgentLists)-1)]}
}

func matchElName(n *html.Node, name string) bool {
	return name == "" || name == n.Data
}

func arr2Attr(args []map[string]string) map[string][]string {
	var attr map[string][]string
	if len(args) > 0 {
		attr = make(map[string][]string, len(args[0]))
		for i := range args[0] {
			attr[i] = strings.Fields(args[0][i])
		}
	}
	return attr
}

func getAttrValue(attributes []html.Attribute) map[string]string {
	var values = make(map[string]string)
	for i := 0; i < len(attributes); i++ {
		_, exists := values[attributes[i].Key]
		if !exists {
			values[attributes[i].Key] = attributes[i].Val
		}
	}
	return values
}

func findAttrValue(attr html.Attribute, attribute string, value []string) bool {
	if attr.Key == attribute {
		attr := strings.Fields(attr.Val)
		num := len(value)
		// todo optimization
		for i := range value {
			for a := range attr {
				if attr[a] == value[i] {
					num--
					break
				}
			}
		}
		return num == 0
	}
	return false
}

func getElText(r QueryHTML, full bool) string {
	b := zstring.Buffer()
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		if full {
			if n.Type == html.ElementNode {
				f(n.FirstChild)
			}
		}
		if n.NextSibling != nil {
			f(n.NextSibling)
		}
	}
	f(r.getNode().FirstChild)
	return b.String()
}

func forChild(node *html.Node, iterator func(n *html.Node) bool) {
	n := node.FirstChild
	for {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode {
			if !iterator(n) {
				return
			}
		}
		n = n.NextSibling
	}
}

func matchEl(n *html.Node, el string, args map[string][]string) *html.Node {
	if n.Type == html.ElementNode && matchElName(n, el) {
		if len(args) > 0 {
			for i := 0; i < len(n.Attr); i++ {
				attr := n.Attr[i]
				for name, val := range args {
					if findAttrValue(attr, name, val) {
						return n
					}
				}
			}
		} else {
			return n
		}
	}
	return nil
}

func findChild(node *html.Node, el string, args []map[string]string, multiple bool) (elArr []*html.Node) {
	attr := arr2Attr(args)
	n := matchEl(node, el, attr)
	if n != nil {
		elArr = []*html.Node{n}
		if !multiple {
			elArr = []*html.Node{n}
			return
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		p := findChild(c, el, args, multiple)
		elArr = append(elArr, p...)
		if !multiple && len(elArr) > 0 {
			return
		}
	}

	return
}
