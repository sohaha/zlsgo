package znet

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestCompletionPath(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal("/", Utils.CompletionPath("/", "/"))
	tt.Equal("/", Utils.CompletionPath("//", "///"))
	tt.Equal("/", Utils.CompletionPath("", "/"))
	tt.Equal("/", Utils.CompletionPath(" ", "/"))
	tt.Equal("/a", Utils.CompletionPath("a", "/"))
	tt.Equal("/a", Utils.CompletionPath("/a ", "/"))
	tt.Equal("/a/", Utils.CompletionPath("/a/", "/"))
	tt.Equal("/a b", Utils.CompletionPath("a b  ", "/"))
	tt.Equal("/a b/", Utils.CompletionPath("a b/", "/"))
	tt.Equal("/d/:id", Utils.CompletionPath(":id", "/d//"))
	tt.Equal("/d/:id", Utils.CompletionPath(":id", "d/////"))
	tt.Equal("/g", Utils.CompletionPath("", "/g"))
	tt.Equal("/g/", Utils.CompletionPath("/", "/g"))
	tt.Equal("/xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", Utils.CompletionPath("xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", "/"))
	tt.Equal("/aaa/xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", Utils.CompletionPath("xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", "aaa"))
}

func TestURLMatchAndParse(t *testing.T) {
	tt := zlsgo.NewTest(t)

	match, ok := Utils.URLMatchAndParse("/", "/")
	t.Log(match)
	tt.EqualTrue(!ok)
	tt.Equal(0, len(match))

	match, ok = Utils.URLMatchAndParse("/123", "/:id")
	t.Log(match)
	tt.EqualTrue(ok)
	tt.Equal(1, len(match))

	match, ok = Utils.URLMatchAndParse("/hi/1.1/123", "/:name/:code/:id")
	t.Log(match)
	tt.EqualTrue(ok)
	tt.Equal(3, len(match))

	match, ok = Utils.URLMatchAndParse("/aaa/hi", "/:name/:*")
	t.Log(match)
	tt.EqualTrue(ok)
	tt.Equal(2, len(match))
}

func Test_parsPattern(t *testing.T) {
	tt := zlsgo.NewTest(t)

	p, s := parsePattern([]string{`{name:[\w\p\-]+}.{ext:[\w]+}`}, "")
	tt.Log(p, s)
	tt.EqualExit(`([\w\p\-]+).([\w]+)`, p)
	tt.EqualExit([]string{"name", "ext"}, s)

	p, s = parsePattern([]string{"{name:[\\w\\d-]+}"}, "")
	tt.Log(p, s)
	tt.EqualExit("([\\w\\d-]+)", p)
	tt.EqualExit([]string{"name"}, s)

	p, s = parsePattern([]string{":name", ":id"}, "/")
	tt.Log(p, s)
	tt.EqualExit(`/([^/]+)/([\d]+)`, p)
	tt.EqualExit([]string{"name", "id"}, s)

	p, s = parsePattern([]string{"{p:[\\w\\d-]+}.pth"}, "")
	tt.Log(p, s)
	tt.EqualExit("([\\w\\d-]+).pth", p)
	tt.EqualExit([]string{"p"}, s)

	p, s = parsePattern([]string{`{key:[^\/.]+}.{ext:[^/.]+}`}, "")
	tt.Log(p, s)
	tt.EqualExit(`([^\/.]+).([^/.]+)`, p)
	tt.EqualExit([]string{"key", "ext"}, s)

	p, s = parsePattern([]string{"{name:[^\\", ".]+}"}, "")
	tt.Log(p, s)
	tt.EqualExit("([^/.]+)", p)
	tt.EqualExit([]string{"name"}, s)

	p, s = parsePattern([]string{`xxx-(?P<name>[\w\p{Han}\-]+).(?P<ext>[a-zA-Z]+)`}, "")
	tt.Log(p, s)
	tt.EqualExit("xxx-(?P<name>[\\w\\p{Han}\\-]+).(?P<ext>[a-zA-Z]+)", p)
	tt.EqualExit([]string{"name", "ext"}, s)
}
