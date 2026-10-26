package zcli

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHandleExecDownloadFileUsesCustomHandler(t *testing.T) {
	source := filepath.Join(t.TempDir(), "artifact.bin")
	dest := t.TempDir()
	if err := os.WriteFile(source, []byte("fixture"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var called string
	var calledDir string
	err := handleExecDownloadFile(&ExecOptions{
		DownloadHandlerWithDir: func(file string, dir string) error {
			called = file
			calledDir = dir
			return nil
		},
	}, source, dest)
	if err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	if got, want := called, source; got != want {
		t.Fatalf("DownloadHandler file = %q, want %q", got, want)
	}
	if got, want := filepath.Clean(calledDir), filepath.Clean(dest); got != want {
		t.Fatalf("DownloadHandler dir = %q, want %q", got, want)
	}
}

func TestHandleExecDownloadFileHandlerCanInstallTargetBinary(t *testing.T) {
	source := filepath.Join(t.TempDir(), "artifact.bin")
	dest := t.TempDir()
	target := filepath.Join(dest, "tool")
	if err := os.WriteFile(source, []byte("fixture"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := handleExecDownloadFile(&ExecOptions{
		DownloadHandlerWithDir: func(file string, dir string) error {
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, "tool"), data, 0o755)
		},
	}, source, dest)
	if err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "fixture" {
		t.Fatalf("installed content = %q, want fixture", got)
	}
}

func TestHandleExecDownloadFileUsesLegacyCustomHandler(t *testing.T) {
	source := filepath.Join(t.TempDir(), "artifact.bin")
	if err := os.WriteFile(source, []byte("fixture"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var called string
	err := handleExecDownloadFile(&ExecOptions{
		DownloadHandler: func(file string) error {
			called = file
			return nil
		},
	}, source, t.TempDir())
	if err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	if got, want := called, source; got != want {
		t.Fatalf("legacy DownloadHandler file = %q, want %q", got, want)
	}
}

func TestHandleExecDownloadFilePrefersHandlerWithDir(t *testing.T) {
	source := filepath.Join(t.TempDir(), "artifact.bin")
	if err := os.WriteFile(source, []byte("fixture"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	calledLegacy := false
	calledNew := false
	err := handleExecDownloadFile(&ExecOptions{
		DownloadHandler: func(file string) error {
			calledLegacy = true
			return nil
		},
		DownloadHandlerWithDir: func(file string, dir string) error {
			calledNew = true
			return nil
		},
	}, source, t.TempDir())
	if err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	if !calledNew {
		t.Fatal("DownloadHandlerWithDir should be preferred")
	}
	if calledLegacy {
		t.Fatal("legacy DownloadHandler should not run when DownloadHandlerWithDir is set")
	}
}

func TestHandleExecDownloadFileRejectsZipSlip(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "payload.zip")
	writeExecZipArchive(t, archive, map[string]string{
		"../escape.txt": "blocked",
	})

	err := handleExecDownloadFile(&ExecOptions{}, archive, t.TempDir())
	if err == nil {
		t.Fatal("handleExecDownloadFile() expected zip slip error")
	}
	if !strings.Contains(err.Error(), "invalid archive path") {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}
}

func TestHandleExecDownloadFileExtractsZipNestedFile(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "payload.zip")
	dest := t.TempDir()
	writeExecZipArchive(t, archive, map[string]string{
		"bin/tool.txt": "payload",
	})

	if err := handleExecDownloadFile(&ExecOptions{}, archive, dest); err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "bin", "tool.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "payload" {
		t.Fatalf("extracted zip content = %q, want payload", got)
	}
}

func TestHandleExecDownloadFileCopiesPlainFileToBinFilePath(t *testing.T) {
	source := filepath.Join(t.TempDir(), "artifact.bin")
	dest := t.TempDir()
	if err := os.WriteFile(source, []byte("fixture"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := handleExecDownloadFile(&ExecOptions{BinFile: "bin/tool"}, source, dest); err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "bin", "tool"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "fixture" {
		t.Fatalf("copied content = %q, want fixture", got)
	}
}

func TestHandleExecDownloadFileGunzipsToBinFilePath(t *testing.T) {
	source := filepath.Join(t.TempDir(), "artifact.gz")
	dest := t.TempDir()
	writeExecGzipFile(t, source, "payload")

	if err := handleExecDownloadFile(&ExecOptions{BinFile: "bin/tool"}, source, dest); err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "bin", "tool"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "payload" {
		t.Fatalf("gunzip content = %q, want payload", got)
	}
}

func TestHandleExecDownloadFileExtractsTarGz(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "payload.tgz")
	dest := t.TempDir()
	writeExecTarGzArchive(t, archive, map[string]string{
		"bin/tool.txt": "payload",
	})

	if err := handleExecDownloadFile(&ExecOptions{}, archive, dest); err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "bin", "tool.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "payload" {
		t.Fatalf("extracted content = %q, want payload", got)
	}
}

func TestHandleExecDownloadFileExtractsTarHardLink(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "payload.tgz")
	dest := t.TempDir()
	writeExecTarGzEntries(t, archive, []execTarEntry{
		{name: "bin/tool", mode: 0o755, content: "payload"},
		{name: "bin/tool-hard", mode: 0o755, typeflag: tar.TypeLink, linkname: "bin/tool"},
	})

	if err := handleExecDownloadFile(&ExecOptions{}, archive, dest); err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "bin", "tool-hard"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "payload" {
		t.Fatalf("hard link content = %q, want payload", got)
	}
}

func TestHandleExecDownloadFileExtractsTarSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks on Windows requires extra privileges")
	}

	archive := filepath.Join(t.TempDir(), "payload.tgz")
	dest := t.TempDir()
	writeExecTarGzEntries(t, archive, []execTarEntry{
		{name: "lib/tool", mode: 0o755, content: "payload"},
		{name: "lib/tool-link", typeflag: tar.TypeSymlink, linkname: "tool"},
	})

	if err := handleExecDownloadFile(&ExecOptions{}, archive, dest); err != nil {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}

	link := filepath.Join(dest, "lib", "tool-link")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink() error = %v", err)
	}
	if target != "tool" {
		t.Fatalf("symlink target = %q, want tool", target)
	}

	got, err := os.ReadFile(link)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "payload" {
		t.Fatalf("symlink content = %q, want payload", got)
	}
}

func TestHandleExecDownloadFileRejectsEscapingTarSymlink(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "payload.tgz")
	writeExecTarGzEntries(t, archive, []execTarEntry{
		{name: "bin/tool-link", typeflag: tar.TypeSymlink, linkname: "../../etc/passwd"},
	})

	err := handleExecDownloadFile(&ExecOptions{}, archive, t.TempDir())
	if err == nil {
		t.Fatal("handleExecDownloadFile() expected escaping symlink error")
	}
	if !strings.Contains(err.Error(), "escapes extraction dir") {
		t.Fatalf("handleExecDownloadFile() error = %v", err)
	}
}

func TestValidateExecArchiveRejectsErrorPage(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "payload.tar.gz")
	if err := os.WriteFile(archive, []byte("not found"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := validateExecArchive(archive)
	if err == nil {
		t.Fatal("validateExecArchive() expected error for html payload")
	}
	if !strings.Contains(err.Error(), "not a valid tar.gz archive") {
		t.Fatalf("validateExecArchive() error = %v", err)
	}
}

func TestValidateExecArchiveAcceptsGzipPayload(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "payload.tar.gz")
	writeExecGzipFile(t, archive, "payload")

	if err := validateExecArchive(archive); err != nil {
		t.Fatalf("validateExecArchive() error = %v", err)
	}
}

func TestValidateExecArchiveIgnoresPlainFiles(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(file, []byte("payload"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := validateExecArchive(file); err != nil {
		t.Fatalf("validateExecArchive() error = %v", err)
	}
}

func TestExecEmbedToDiskCopiesEmbeddedFiles(t *testing.T) {
	dest := t.TempDir()

	total, err := embedToDisk(execEmbedFixtureFS, dest)
	if err != nil {
		t.Fatalf("embedToDisk() error = %v", err)
	}
	if total != 1 {
		t.Fatalf("embedToDisk() total = %d, want 1", total)
	}

	got, err := os.ReadFile(filepath.Join(dest, "embedfs", "root", "bin", "tool.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.TrimSpace(string(got)) != "embedded-tool" {
		t.Fatalf("embedded file content = %q, want embedded-tool", got)
	}
}

func writeExecZipArchive(t *testing.T, path string, files map[string]string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("Write(%q) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("zip Close() error = %v", err)
	}
}

func writeExecTarGzArchive(t *testing.T, path string, files map[string]string) {
	t.Helper()

	entries := make([]execTarEntry, 0, len(files))
	for name, content := range files {
		entries = append(entries, execTarEntry{name: name, mode: 0o644, content: content})
	}
	writeExecTarGzEntries(t, path, entries)
}

type execTarEntry struct {
	name     string
	mode     int64
	content  string
	typeflag byte
	linkname string
}

func writeExecTarGzEntries(t *testing.T, path string, entries []execTarEntry) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Mode:     entry.mode,
			Size:     int64(len(entry.content)),
			Typeflag: entry.typeflag,
			Linkname: entry.linkname,
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("WriteHeader(%q) error = %v", entry.name, err)
		}
		if entry.content == "" {
			continue
		}
		if _, err := tarWriter.Write([]byte(entry.content)); err != nil {
			t.Fatalf("Write(%q) error = %v", entry.name, err)
		}
	}
}

func writeExecGzipFile(t *testing.T, path string, content string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	if _, err := gzipWriter.Write([]byte(content)); err != nil {
		t.Fatalf("gzip Write() error = %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("gzip Close() error = %v", err)
	}
}
