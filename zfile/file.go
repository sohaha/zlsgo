/*
 * @Author: seekwe
 * @Date:   2019-05-17 17:08:52
 * @Last Modified by:   seekwe
 * @Last Modified time: 2020-02-02 21:31:04
 */

package zfile

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var RootPath = RealPath(".", true)

// PathExist PathExist
// 1 exists and is a directory path, 2 exists and is a file path, 0 does not exist
func PathExist(path string) (int, error) {
	f, err := os.Stat(path)
	if err == nil {
		isFile := 2
		if f.IsDir() {
			isFile = 1
		}
		return isFile, nil
	}

	return 0, err
}

// DirExist Is it an existing directory
func DirExist(path string) bool {
	state, _ := PathExist(path)
	return state == 1
}

// FileExist Is it an existing file?
func FileExist(path string) bool {
	state, _ := PathExist(path)
	return state == 2
}

// FileSize file size
func FileSize(file string) (size string) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		size = FileSizeFormat(0)
	} else {
		size = FileSizeFormat(uint64(fileInfo.Size()))
	}
	return
}

// FileSizeFormat Format file size
func FileSizeFormat(s uint64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	humanateBytes := func(s uint64, base float64, sizes []string) string {
		if s < 10 {
			return fmt.Sprintf("%d B", s)
		}
		e := math.Floor(logSize(float64(s), base))
		suffix := sizes[int(e)]
		val := float64(s) / math.Pow(base, math.Floor(e))
		f := "%.0f"
		if val < 10 {
			f = "%.1f"
		}
		return fmt.Sprintf(f+" %s", val, suffix)
	}
	return humanateBytes(s, 1024, sizes)
}

func logSize(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

// SafePath get an safe absolute path
func SafePath(path string, addSlash ...bool) string {
	realPath := RealPath(path, addSlash...)
	return strings.TrimPrefix(realPath, RootPath)
}

// RealPath get an absolute path
func RealPath(path string, addSlash ...bool) (realPath string) {
	realPath, _ = filepath.Abs(path)
	realPath = pathAddSlash(filepath.ToSlash(realPath), addSlash...)

	return
}

// RealPathMkdir get an absolute path, create it if it doesn't exist
func RealPathMkdir(path string, addSlash ...bool) string {
	realPath := RealPath(path, addSlash...)
	if DirExist(realPath) {
		return realPath
	}
	_ = os.MkdirAll(path, os.ModePerm)
	return realPath
}

// Rmdir rmdir,support to keep the current directory
func Rmdir(path string, notIncludeSelf ...bool) (ok bool) {
	realPath := RealPath(path)
	err := os.RemoveAll(realPath)
	ok = err == nil
	if ok && len(notIncludeSelf) > 0 && notIncludeSelf[0] {
		_ = os.Mkdir(path, os.ModePerm)
	}
	return
}

// CopyFile copies the source file to the dest file.
func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()
	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()
	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}
	return nil
}

// CopyDir copies the source directory to the dest directory.
func CopyDir(source string, dest string, filterFn ...func(srcFilePath, destFilePath string) bool) (err error) {
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}
	directory, err := os.Open(source)
	if err != nil {
		return err
	}
	defer directory.Close()
	objects, err := directory.Readdir(-1)
	if err != nil {
		return err
	}
	var filter func(srcFilePath, destFilePath string) bool
	if len(filterFn) > 0 {
		filter = filterFn[0]
	}
	copySum := len(objects)
	for _, obj := range objects {
		srcFilePath := filepath.Join(source, obj.Name())
		destFilePath := filepath.Join(dest, obj.Name())
		if obj.IsDir() {
			_ = CopyDir(srcFilePath, destFilePath, filterFn...)
		} else if filter == nil || filter(srcFilePath, destFilePath) {
			_ = CopyFile(srcFilePath, destFilePath)
		} else {
			copySum--
		}
	}
	if copySum < 1 {
		Rmdir(dest)
	}
	return nil
}

// ProgramPath program directory path
func ProgramPath(addSlash ...bool) (path string) {
	ePath, err := os.Executable()
	if err != nil {
		ePath = RealPath(".")
	}
	path = pathAddSlash(filepath.Dir(ePath), addSlash...)

	return
}

func pathAddSlash(path string, addSlash ...bool) string {
	if len(addSlash) > 0 && addSlash[0] && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
