package zstring_test

import (
	"io"
	"strings"
	"testing"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

func TestNewTemplate(t *testing.T) {
	tt := zls.NewTest(t)

	tmpl, err := zstring.NewTemplate("hello {name}", "{", "}")
	tt.NoError(err)

	w := zutil.GetBuff()
	defer zutil.PutBuff(w)

	_, err = tmpl.Process(w, func(w io.Writer, tag string) (int, error) {
		return w.Write([]byte("Go"))
	})
	tt.NoError(err)
	tt.Equal("hello Go", w.String())

	err = tmpl.ResetTemplate("The best {n} {say}")
	tt.NoError(err)

	w.Reset()
	_, err = tmpl.Process(w, func(w io.Writer, tag string) (int, error) {
		switch tag {
		case "say":
			return w.Write([]byte("!!!"))
		}
		return w.Write([]byte("Go"))
	})
	tt.NoError(err)
	tt.Equal("The best Go !!!", w.String())
}

func BenchmarkTemplate(b *testing.B) {
	tmpl, _ := zstring.NewTemplate("hello {name}", "{", "}")

	b.Run("Buffer", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				var w strings.Builder
				_, _ = tmpl.Process(&w, func(w io.Writer, tag string) (int, error) {
					return w.Write([]byte("Go"))
				})
			}
		})
	})

	b.Run("GetBuff", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w := zutil.GetBuff()
				_, _ = tmpl.Process(w, func(w io.Writer, tag string) (int, error) {
					return w.Write([]byte("Go"))
				})
				zutil.PutBuff(w)
			}
		})
	})
}
