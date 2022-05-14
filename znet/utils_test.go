package znet

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestCompletionPath(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal("/", CompletionPath("/", "/"))
	tt.Equal("/", CompletionPath("", "/"))
	tt.Equal("/", CompletionPath(" ", "/"))
	tt.Equal("/a", CompletionPath("a", "/"))
	tt.Equal("/a", CompletionPath("/a ", "/"))
	tt.Equal("/a/", CompletionPath("/a/", "/"))
	tt.Equal("/a b", CompletionPath("a b  ", "/"))
	tt.Equal("/a b/", CompletionPath("a b/", "/"))
	tt.Equal("/d/:id", CompletionPath(":id", "/d//"))
	tt.Equal("/d/:id", CompletionPath(":id", "d/////"))
	tt.Equal("/g", CompletionPath("", "/g"))
	tt.Equal("/g/", CompletionPath("/", "/g"))
	tt.Equal("/xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", CompletionPath("xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", "/"))
	tt.Equal("/aaa/xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", CompletionPath("xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", "aaa"))
}
