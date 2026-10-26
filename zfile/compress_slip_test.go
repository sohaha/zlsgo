package zfile

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/sohaha/zlsgo"
)

type archiveEntry struct {
	name    string
	content string
}

func buildEvilTarGz(t *testing.T, name string, entries []archiveEntry) {
	t.Helper()
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	for _, entry := range entries {
		hdr := &tar.Header{
			Name: entry.name,
			Mode: 0o600,
			Size: int64(len(entry.content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(entry.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func buildEvilZip(t *testing.T, name string, entries []archiveEntry) {
	t.Helper()
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for _, entry := range entries {
		w, err := zw.Create(entry.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(entry.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGzDeCompressZipSlip(t *testing.T) {
	tt := zlsgo.NewTest(t)
	gz := "tmp_evil.tar.gz"
	dest := "tmp_slip_gz"
	buildEvilTarGz(t, gz, []archiveEntry{
		{"ok.txt", "ok"},
		{"../../../../tmp_evil_escape.txt", "poc"},
		{"/tmp_evil_abs.txt", "poc"},
	})
	defer Rmdir(gz)
	defer Rmdir(dest)

	err := GzDeCompress(gz, dest)
	tt.EqualTrue(err != nil)

	_, statErr := os.Stat(RealPath("../tmp_evil_escape.txt"))
	tt.EqualTrue(os.IsNotExist(statErr))
	_, statErr = os.Stat(RealPath("./tmp_evil_abs.txt"))
	tt.EqualTrue(os.IsNotExist(statErr))
	tt.EqualTrue(FileExist(filepath.Join(RealPath(dest), "ok.txt")))
}

func TestZipDeCompressZipSlip(t *testing.T) {
	tt := zlsgo.NewTest(t)
	zipFile := "tmp_evil.zip"
	dest := "tmp_slip_zip"
	buildEvilZip(t, zipFile, []archiveEntry{
		{"ok.txt", "ok"},
		{"../../../../tmp_evil_zip_escape.txt", "poc"},
		{"/tmp_evil_zip_abs.txt", "poc"},
	})
	defer Rmdir(zipFile)
	defer Rmdir(dest)

	err := ZipDeCompress(zipFile, dest)
	tt.EqualTrue(err != nil)

	_, statErr := os.Stat(RealPath("../tmp_evil_zip_escape.txt"))
	tt.EqualTrue(os.IsNotExist(statErr))
	_, statErr = os.Stat(RealPath("./tmp_evil_zip_abs.txt"))
	tt.EqualTrue(os.IsNotExist(statErr))
	tt.EqualTrue(FileExist(filepath.Join(RealPath(dest), "ok.txt")))
}

func TestSafeExtractPath(t *testing.T) {
	tt := zlsgo.NewTest(t)
	dest := RealPath("tmp_slip_unit", true)

	target, err := safeExtractPath(dest, "a/b.txt")
	tt.EqualNil(err)
	tt.Equal(filepath.Join(filepath.Clean(dest), "a", "b.txt"), target)

	target, err = safeExtractPath(dest, "sub/../a.txt")
	tt.EqualNil(err)
	tt.Equal(filepath.Join(filepath.Clean(dest), "a.txt"), target)

	_, err = safeExtractPath(dest, "../escape.txt")
	tt.EqualTrue(err != nil)

	_, err = safeExtractPath(dest, "a/../../escape.txt")
	tt.EqualTrue(err != nil)

	_, err = safeExtractPath(dest, "/etc/passwd")
	tt.EqualTrue(err != nil)

	_, err = safeExtractPath(dest, "")
	tt.EqualTrue(err != nil)

	Rmdir("tmp_slip_unit")
}
