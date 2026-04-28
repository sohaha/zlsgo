package znet

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFileAttachmentUsesProvidedPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "attachment.txt")
	if err := os.WriteFile(filePath, []byte("attachment body"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := New()
	r.GET("/download", func(c *Context) {
		c.FileAttachment(filePath, "download.txt")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/download", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != "attachment body" {
		t.Fatalf("expected file body, got %q", got)
	}
	if got := w.Header().Get("Content-Disposition"); got != `attachment; filename="download.txt"` {
		t.Fatalf("unexpected content disposition: %q", got)
	}
}
