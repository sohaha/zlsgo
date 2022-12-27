package znet

import (
	"bytes"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

func TestHTMLRender(t *testing.T) {
	tt := zlsgo.NewTest(t)

	engine := newGoTemplate(nil, "./testdata/html")

	err := engine.Load()
	tt.NoError(err)

	var buf bytes.Buffer
	err = engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	})
	tt.NoError(err)
	expect := `<h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2>`
	tt.Equal(expect, zstring.TrimLine(buf.String()))
}
