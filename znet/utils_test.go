package znet

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestCompletionPath(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal("/", completionPath("/", "/"))
	tt.Equal("/", completionPath("", "/"))
	tt.Equal("/", completionPath(" ", "/"))
	tt.Equal("/a", completionPath("a", "/"))
	tt.Equal("/a", completionPath("/a ", "/"))
	tt.Equal("/a/", completionPath("/a/", "/"))
	tt.Equal("/a b", completionPath("a b  ", "/"))
	tt.Equal("/a b/", completionPath("a b/", "/"))
	tt.Equal("/d/:id", completionPath(":id", "/d//"))
	tt.Equal("/g", completionPath("", "/g"))
	tt.Equal("/g/", completionPath("/", "/g"))
	tt.Equal("/xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", completionPath("xxx/{name:[\\w\\d-]+}.{ext:[\\w]+}", "/"))
}
