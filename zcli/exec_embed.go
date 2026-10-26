package zcli

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
)

var execTmpDir = initExecTmpDir()

func initExecTmpDir() string {
	epath, err := os.Executable()
	if err != nil {
		epath = os.Args[0]
	}
	hash, _ := zstring.Md5File(epath)
	return zfile.RealPathMkdir(filepath.Join(zfile.TmpPath(), hash))
}

func writeExecFile(path string, binary []byte) error {
	if len(binary) == 0 {
		return errors.New("binary is empty")
	}

	if err := zfile.WriteFile(path, binary); err != nil {
		return err
	}

	_ = os.Chmod(path, 0o755)
	return nil
}

func embedExecWalkDir(fe embed.FS, root, dir string) (int, error) {
	fileTotal := 0

	err := fs.WalkDir(fe, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath := strings.TrimPrefix(path, root)
		relativePath = strings.TrimLeft(relativePath, `/\`)
		fp := filepath.Join(dir, relativePath)
		if d.IsDir() {
			if err := os.MkdirAll(fp, 0o755); err != nil && !os.IsExist(err) {
				return err
			}
			return nil
		}

		data, err := fe.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(fp, data, 0o755); err != nil {
			return err
		}

		fileTotal++
		_ = os.Chmod(fp, 0o777)

		return nil
	})

	return fileTotal, err
}

func embedToDisk(fe embed.FS, tmpDir string) (int, error) {
	files, err := fe.ReadDir(".")
	if err != nil {
		return 0, err
	}

	if len(files) == 1 {
		if files[0].IsDir() {
			return embedExecWalkDir(fe, files[0].Name(), tmpDir)
		}
	}

	return embedExecWalkDir(fe, ".", tmpDir)
}
