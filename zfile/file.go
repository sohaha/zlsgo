// Package zfile file and path operations in daily development
package zfile

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var (
	ProjectPath = "."
)

func init() {
	// abs, _ := filepath.Abs(".")
	// ProjectPath = RealPath(abs)
	ProjectPath = ProgramPath()
	if strings.Contains(ProjectPath, TmpPath("")) {
		ProjectPath = RootPath()
	}
}

// PathExist PathExist
// 1 exists and is a directory path, 2 exists and is a file path, 0 does not exist
func PathExist(path string) (int, error) {
	path = RealPath(path)
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
	return SizeFormat(FileSizeUint(file))
}

// FileSizeUint file size to uint64
func FileSizeUint(file string) (size uint64) {
	file = RealPath(file)
	fileInfo, err := os.Stat(file)
	if err != nil {
		return 0
	}
	return uint64(fileInfo.Size())
}

// SizeFormat Format file size
func SizeFormat(s uint64) string {
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

// RootPath Project Launch Path
func RootPath() string {
	path, _ := filepath.Abs(".")
	return RealPath(path, true)
}

func TmpPath(pattern ...string) string {
	p := ""
	if len(pattern) > 0 {
		p = pattern[0]
	}
	path, _ := ioutil.TempDir("", p)
	if p == "" {
		path, _ = filepath.Split(path)
	}
	path, _ = filepath.EvalSymlinks(path)
	return RealPath(path)
}

// SafePath get an safe absolute path
func SafePath(path string, pathRange ...string) string {
	base := ""
	if len(pathRange) == 0 {
		base = RootPath()
	} else {
		base = RealPath(pathRange[0], true)
	}
	return strings.TrimPrefix(RealPath(path, false), base)
}

// RealPath get an absolute path
func RealPath(path string, addSlash ...bool) (realPath string) {
	if len(path) == 0 || (path[0] != '/' && !filepath.IsAbs(path)) {
		path = ProjectPath + "/" + path
	}
	realPath, _ = filepath.Abs(path)
	realPath = strings.Replace(realPath, "\\", "/", -1)
	realPath = pathAddSlash(realPath, addSlash...)

	return
}

// RealPathMkdir get an absolute path, create it if it doesn't exist
func RealPathMkdir(path string, addSlash ...bool) string {
	realPath := RealPath(path, addSlash...)
	if DirExist(realPath) {
		return realPath
	}
	_ = os.MkdirAll(realPath, os.ModePerm)
	return realPath
}

// Rmdir support to keep the current directory
func Rmdir(path string, notIncludeSelf ...bool) (ok bool) {
	realPath := RealPath(path)
	err := os.RemoveAll(realPath)
	ok = err == nil
	if ok && len(notIncludeSelf) > 0 && notIncludeSelf[0] {
		_ = os.Mkdir(path, os.ModePerm)
	}
	return
}

// Remove removes the named file or (empty) directory
func Remove(path string) error {
	realPath := RealPath(path)
	return os.Remove(realPath)
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
		var sourceinfo os.FileInfo
		sourceinfo, err = os.Stat(source)
		if err == nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}
	return
}

// ProgramPath program directory path
func ProgramPath(addSlash ...bool) (path string) {
	ePath, err := os.Executable()
	if err != nil {
		ePath = ProjectPath
	} else {
		ePath = filepath.Dir(ePath)
	}
	realPath, err := filepath.EvalSymlinks(ePath)
	if err == nil {
		ePath = realPath
	}
	path = RealPath(ePath, addSlash...)

	return
}

func pathAddSlash(path string, addSlash ...bool) string {
	if len(addSlash) > 0 && addSlash[0] && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
