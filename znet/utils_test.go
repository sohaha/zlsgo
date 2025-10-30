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

	p, s := ParsePattern([]string{`{name:[\w\p\-]+}.{ext:[\w]+}`}, "")
	tt.Log(p, s)
	tt.EqualExit(`([\w\p\-]+).([\w]+)`, p)
	tt.EqualExit([]string{"name", "ext"}, s)

	p, s = ParsePattern([]string{"{name:[\\w\\d-]+}"}, "")
	tt.Log(p, s)
	tt.EqualExit("([\\w\\d-]+)", p)
	tt.EqualExit([]string{"name"}, s)

	p, s = ParsePattern([]string{":name", ":id"}, "/")
	tt.Log(p, s)
	tt.EqualExit(`/([^/]+)/([\d]+)`, p)
	tt.EqualExit([]string{"name", "id"}, s)

	p, s = ParsePattern([]string{"{p:[\\w\\d-]+}.pth"}, "")
	tt.Log(p, s)
	tt.EqualExit("([\\w\\d-]+).pth", p)
	tt.EqualExit([]string{"p"}, s)

	p, s = ParsePattern([]string{`{key:[^\/.]+}.{ext:[^/.]+}`}, "")
	tt.Log(p, s)
	tt.EqualExit(`([^\/.]+).([^/.]+)`, p)
	tt.EqualExit([]string{"key", "ext"}, s)

	p, s = ParsePattern([]string{"{name:[^\\", ".]+}"}, "")
	tt.Log(p, s)
	tt.EqualExit("([^/.]+)", p)
	tt.EqualExit([]string{"name"}, s)

	p, s = ParsePattern([]string{`xxx-(?P<name>[\w\p{Han}\-]+).(?P<ext>[a-zA-Z]+)`}, "")
	tt.Log(p, s)
	tt.EqualExit("xxx-(?P<name>[\\w\\p{Han}\\-]+).(?P<ext>[a-zA-Z]+)", p)
	tt.EqualExit([]string{"name", "ext"}, s)
}

func TestURLMatchAndParse_SafeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	// 1) Edge case: colon without name should not panic and should match a segment
	if params, ok := Utils.URLMatchAndParse("/a/b", "/a/:"); !ok {
		t.Fatalf("expected ok for route /a/: with /a/b, got params=%v ok=%v", params, ok)
	}

	// 2) Braces without explicit expr defaults to defaultPattern
	if params, ok := Utils.URLMatchAndParse("/file/readme", "/file/{name}"); !ok {
		t.Fatalf("expected ok for /file/{name}, got params=%v ok=%v", params, ok)
	} else if params["name"] != "readme" {
		t.Fatalf("expected name=readme, got %v", params)
	}

	// 3) {id} defaults to digits
	if params, ok := Utils.URLMatchAndParse("/user/123", "/user/{id}"); !ok {
		t.Fatalf("expected ok for /user/{id}, got params=%v ok=%v", params, ok)
	} else if params["id"] != "123" {
		t.Fatalf("expected id=123, got %v", params)
	}

	// 4) wildcard * captures rest
	if params, ok := Utils.URLMatchAndParse("/static/css/app.css", "/static/*"); !ok {
		t.Fatalf("expected ok for /static/*, got params=%v ok=%v", params, ok)
	} else if params["*"] != "css/app.css" {
		t.Fatalf("expected *=css/app.css, got %v", params)
	}

	// 5) named group regex works and maps by name
	if params, ok := Utils.URLMatchAndParse("/r/abc", "/r/(?P<slug>[^/]+)"); !ok {
		t.Fatalf("expected ok for named group, got params=%v ok=%v", params, ok)
	} else if params["slug"] != "abc" {
		t.Fatalf("expected slug=abc, got %v", params)
	}

	// 6) more captures than names are safely ignored
	if _, ok := Utils.URLMatchAndParse("/x/ab12", "/x/{name:([a-z]+)([0-9]+)}"); !ok {
		t.Fatalf("expected ok for extra-capture case")
	}

	tt.EqualTrue(true)
}
