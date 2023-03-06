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

	match, ok = Utils.URLMatchAndParse("/aaa/hi", "/:name/:*")
	t.Log(match)
	tt.EqualTrue(ok)
	tt.Equal(2, len(match))
}
