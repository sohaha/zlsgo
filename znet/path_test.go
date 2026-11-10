package znet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
)

func TestResolvePath(t *testing.T) {
	tt := zlsgo.NewTest(t)

	t.Run("absolute path is untouched", func(t *testing.T) {
		abs := filepath.Join(t.TempDir(), "index.html")
		got, fromCwd := resolvePath(abs)
		tt.Equal(abs, got)
		tt.Equal(false, fromCwd)
	})

	t.Run("relative path resolves against project directory", func(t *testing.T) {
		dir := zfile.RealPathMkdir("tmpResolvePath")
		defer zfile.Rmdir(dir)

		got, fromCwd := resolvePath("tmpResolvePath")
		tt.Equal(dir, got)
		tt.Equal(false, fromCwd)

		// Glob patterns are probed through their parent directory.
		got, fromCwd = resolvePath("tmpResolvePath/*.html")
		tt.Equal(dir+"/*.html", got)
		tt.Equal(false, fromCwd)
	})

	t.Run("falls back to the working directory", func(t *testing.T) {
		cwd := t.TempDir()
		_ = zfile.RealPathMkdir(filepath.Join(cwd, "cwdOnly"))
		t.Chdir(cwd)

		got, fromCwd := resolvePath("cwdOnly")
		tt.Equal(true, fromCwd)
		tt.Equal(true, zfile.DirExist(got))
		tt.Equal(true, strings.HasSuffix(got, "cwdOnly"))
	})

	t.Run("missing path stays canonical", func(t *testing.T) {
		got, fromCwd := resolvePath("tmpResolvePathMissing")
		tt.Equal(zfile.RealPath("tmpResolvePathMissing"), got)
		tt.Equal(false, fromCwd)
	})

	t.Run("upward path stays with the working directory", func(t *testing.T) {
		cwd, err := os.Getwd()
		tt.NoError(err)

		got, fromCwd := resolvePath("../")
		tt.Equal(true, fromCwd)
		tt.Equal(normalizePath(filepath.Dir(cwd)), got)
	})
}

// TestLoadHTMLGlobRelative covers the regression where a pattern without a
// wildcard resolved against the working directory while the glob form resolved
// against the project directory.
func TestLoadHTMLGlobRelative(t *testing.T) {
	tt := zlsgo.NewTest(t)
	dir := zfile.RealPathMkdir("tmpRelativeTemplate")
	defer zfile.Rmdir(dir)
	tt.NoError(zfile.WriteFile(dir+"/index.html", []byte(`<p>{{.Title}}</p>`)))

	patterns := []string{"tmpRelativeTemplate", "tmpRelativeTemplate/*.html"}
	for i, pattern := range patterns {
		r := New(fmt.Sprintf("path-test-%d", i+1))
		r.SetMode(DebugMode)
		r.LoadHTMLGlob(pattern)

		w := newRequest(r, "GET", "/relative", "/relative", func(c *Context) {
			c.Template(200, "index.html", Data{"Title": "Relative"})
		})
		tt.Equal(200, w.Code)
		tt.Equal(`<p>Relative</p>`, w.Body.String())
	}
}

// TestLoadHTMLGlobWorkingDirFallback keeps working-directory setups alive.
func TestLoadHTMLGlobWorkingDirFallback(t *testing.T) {
	tt := zlsgo.NewTest(t)

	cwd := t.TempDir()
	tt.NoError(zfile.WriteFile(filepath.Join(cwd, "cwdTpl", "index.html"), []byte(`<p>cwd</p>`)))
	t.Chdir(cwd)

	r := New("path-test-cwd")
	r.SetMode(DebugMode)
	r.LoadHTMLGlob("cwdTpl")

	w := newRequest(r, "GET", "/cwd", "/cwd", func(c *Context) {
		c.Template(200, "index.html", nil)
	})
	tt.Equal(200, w.Code)
	tt.Equal(`<p>cwd</p>`, w.Body.String())
}

// TestStaticRelativeRoot checks that a relative static root follows the same
// resolution as templates.
func TestStaticRelativeRoot(t *testing.T) {
	tt := zlsgo.NewTest(t)

	dir := zfile.RealPathMkdir("tmpStaticRoot")
	defer zfile.Rmdir(dir)
	tt.NoError(zfile.WriteFile(dir+"/app.js", []byte("console.log(1)")))

	r := New("path-test-static")
	r.SetMode(DebugMode)
	r.Static("/assets", "tmpStaticRoot")

	w := newRequest(r, "GET", "/assets/app.js", "")
	tt.Equal(200, w.Code)
	tt.Equal("console.log(1)", w.Body.String())
}
