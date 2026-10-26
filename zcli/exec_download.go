package zcli

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zstring"
)

func (e *Exec) downloadToDisk() error {
	if e.options.DownloadUrl == "" {
		return errors.New("download url is empty")
	}

	downloadDir := zfile.RealPathMkdir(filepath.Join(zfile.TmpPath(), zstring.Md5(e.options.DownloadUrl)))
	downloadFile := filepath.Join(downloadDir, filepath.Base(e.options.DownloadUrl))

	if zfile.FileExist(downloadFile) {
		if err := handleExecDownloadFile(e.options, downloadFile, e.tmpDir); err == nil {
			return nil
		}
	}

	if err := e.download(downloadFile); err != nil {
		return err
	}

	return handleExecDownloadFile(e.options, downloadFile, e.tmpDir)
}

func (e *Exec) download(downloadFile string) error {
	params := []any{e.ctx}
	if e.options.DownloadProgress != nil {
		params = append(params, zhttp.DownloadProgress(func(current, total int64) {
			if current == 0 {
				e.options.DownloadProgress(0, total)
				return
			}

			e.options.DownloadProgress((float64(current)/float64(total))*100, total)
		}))
	}
	resp, err := zhttp.Get(e.options.DownloadUrl, params...)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(downloadFile), 0o755); err != nil {
		return err
	}

	return resp.ToFile(downloadFile)
}

func handleExecDownloadFile(opt *ExecOptions, file string, dir string) error {
	dir = zfile.RealPathMkdir(dir)
	if opt.DownloadHandlerWithDir != nil {
		return opt.DownloadHandlerWithDir(file, dir)
	}
	if opt.DownloadHandler != nil {
		return opt.DownloadHandler(file)
	}

	// Downloads are cached before extraction, so validate here as well:
	// a previously cached error page must not be reused.
	if err := validateExecArchive(file); err != nil {
		return err
	}

	switch execArchiveKind(file) {
	case "tar.gz":
		return unexecTarGz(opt, file, dir)
	case "zip":
		return unexecZip(opt, file, dir)
	case "gz":
		return unexecGz(opt, file, dir)
	default:
		if zfile.DirExist(file) {
			return zfile.CopyDir(file, dir)
		}

		dest, err := resolveExecOutputPath(opt, file, dir)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := zfile.CopyFile(file, dest); err != nil {
			return err
		}

		_ = os.Chmod(dest, 0o777)
		return nil
	}
}

func execArchiveKind(file string) string {
	lower := strings.ToLower(file)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	case strings.HasSuffix(lower, ".gz"):
		return "gz"
	default:
		return ""
	}
}

// validateExecArchive rejects payloads that only look like an archive by name.
// Download endpoints often answer 404 with an HTML page while still reporting
// success, which would otherwise be cached and extracted as garbage.
func validateExecArchive(file string) error {
	kind := execArchiveKind(file)
	if kind == "" {
		return nil
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	var magic [4]byte
	n, err := io.ReadFull(f, magic[:])
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return err
	}
	if n == 0 {
		return fmt.Errorf("download %s is empty", filepath.Base(file))
	}

	valid := false
	switch kind {
	case "tar.gz", "gz":
		valid = n >= 2 && magic[0] == 0x1f && magic[1] == 0x8b
	case "zip":
		valid = n == 4 && (string(magic[:]) == "PK\x03\x04" ||
			string(magic[:]) == "PK\x05\x06" ||
			string(magic[:]) == "PK\x07\x08")
	}
	if !valid {
		return fmt.Errorf("download %s is not a valid %s archive: the server may have returned an error page", filepath.Base(file), kind)
	}

	return nil
}

func resolveExecOutputPath(opt *ExecOptions, file string, dir string) (string, error) {
	target := filepath.Base(file)
	if opt != nil && opt.BinFile != "" {
		target = opt.BinFile
	}

	return safeExecExtractPath(dir, target)
}

func unexecZip(opt *ExecOptions, file string, dir string) error {
	zipReader, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	dir = zfile.RealPathMkdir(dir)
	for _, f := range zipReader.File {
		fpath, err := safeExecExtractPath(dir, f.Name)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		if _, err = io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return err
		}
		if err := rc.Close(); err != nil {
			outFile.Close()
			return err
		}
		if err := outFile.Close(); err != nil {
			return err
		}
	}
	return nil
}

func unexecTarGz(opt *ExecOptions, file string, dir string) error {
	gzipFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path, err := safeExecExtractPath(dir, header.Name)
		if err != nil {
			return err
		}
		info := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				return err
			}

			if _, err = io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		case tar.TypeLink:
			target, err := safeExecExtractPath(dir, header.Linkname)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			if err := os.Link(target, path); err != nil {
				return fmt.Errorf("link %s -> %s: %w", header.Name, header.Linkname, err)
			}
		case tar.TypeSymlink:
			if err := checkExecSymlinkTarget(dir, path, header.Linkname); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			if err := os.Symlink(header.Linkname, path); err != nil {
				return fmt.Errorf("symlink %s -> %s: %w", header.Name, header.Linkname, err)
			}
		default:
			return fmt.Errorf("unsupported file type: %v", header.Typeflag)
		}
	}

	return nil
}

func unexecGz(opt *ExecOptions, file string, dir string) error {
	gzipFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	dest, err := resolveExecOutputPath(opt, strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)), dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	outFile, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}

	if _, err = io.Copy(outFile, gzipReader); err != nil {
		outFile.Close()
		return err
	}

	if err := outFile.Close(); err != nil {
		return err
	}

	return nil
}

func safeExecExtractPath(dir, name string) (string, error) {
	cleanDir := filepath.Clean(dir)
	target := filepath.Join(cleanDir, filepath.Clean(name))
	relative, err := filepath.Rel(cleanDir, target)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid archive path: %s", name)
	}

	return target, nil
}

// checkExecSymlinkTarget rejects links that resolve outside the extraction root.
// Relative targets are resolved from the link location, matching tar semantics.
func checkExecSymlinkTarget(dir, linkPath, linkname string) error {
	target := linkname
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(linkPath), target)
	}

	relative, err := filepath.Rel(filepath.Clean(dir), filepath.Clean(target))
	if err != nil {
		return fmt.Errorf("invalid symlink target: %s -> %s", linkPath, linkname)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("symlink %s escapes extraction dir: %s", linkPath, linkname)
	}

	return nil
}
