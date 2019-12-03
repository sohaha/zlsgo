/*
 * @Author: seekwe
 * @Date:   2019-05-17 17:08:52
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-25 14:15:27
 */

package zfile

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

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
		size = FileSizeFormat(fileInfo.Size())
	}
	return
}

// FileSizeFormat Format file size
func FileSizeFormat(s int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	humanateBytes := func(s uint64, base float64, sizes []string) string {
		if s < 10 {
			return fmt.Sprintf("%d B", s)
		}
		e := math.Floor(logn(float64(s), base))
		suffix := sizes[int(e)]
		val := float64(s) / math.Pow(base, math.Floor(e))
		f := "%.0f"
		if val < 10 {
			f = "%.1f"
		}
		return fmt.Sprintf(f+" %s", val, suffix)
	}
	return humanateBytes(uint64(s), 1024, sizes)
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

// RealPath get an absolute path
func RealPath(path string, addSlash ...bool) (realPath string) {
	realPath, _ = filepath.Abs(path)
	if len(addSlash) > 0 && addSlash[0] {
		realPath += "/"
	}
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
