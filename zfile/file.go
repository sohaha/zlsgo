// Package zfile file and path operations in daily development
package zfile

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ProjectPath = "./"

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
	return RealPath(path)
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

	if p, err := filepath.EvalSymlinks(path); err == nil {
		path = p
	}
	return RealPath(path)
}

// SafePath get an safe absolute path
func SafePath(path string, pathRange ...string) string {
	base := ""
	if len(pathRange) == 0 {
		base = ProjectPath
	} else {
		base = RealPath(pathRange[0], false)
	}
	return strings.TrimPrefix(RealPath(path, false), base)
}

// RealPath get an absolute path
func RealPath(path string, addSlash ...bool) (realPath string) {
	if len(path) > 2 && path[1] == ':' {
		realPath = path
	} else {
		if len(path) == 0 || (path[0] != '/' && !filepath.IsAbs(path)) {
			path = ProjectPath + "/" + path
		}
		realPath, _ = filepath.Abs(path)
	}

	realPath = strings.Replace(realPath, "\\", "/", -1)
	realPath = pathAddSlash(realPath, addSlash...)

	return
}

// RealPathMkdir get an absolute path, create it if it doesn't exist
// If you want to ensure that the directory can be created successfully, please use HasReadWritePermission to check the permission first
func RealPathMkdir(path string, addSlash ...bool) string {
	realPath := RealPath(path, addSlash...)
	if DirExist(realPath) {
		return realPath
	}

	_ = os.MkdirAll(realPath, os.ModePerm)
	return realPath
}

// IsSubPath Is the subPath under the path
func IsSubPath(subPath, path string) bool {
	subPath = RealPath(subPath)
	path = RealPath(path)
	return strings.HasPrefix(subPath, path)
}

// Rmdir support to keep the current directory
func Rmdir(path string, notIncludeSelf ...bool) (ok bool) {
	realPath := RealPath(path)
	err := os.RemoveAll(realPath)
	ok = err == nil
	if ok && len(notIncludeSelf) > 0 && notIncludeSelf[0] {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			ok = false
		}
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

// ExecutablePath executable path
func ExecutablePath() string {
	ePath, err := os.Executable()
	if err != nil {
		ePath = os.Args[0]
	}
	realPath, err := filepath.EvalSymlinks(ePath)
	if err != nil {
		return ePath
	}
	return realPath
}

// ProgramPath program directory path
func ProgramPath(addSlash ...bool) (path string) {
	ePath := ExecutablePath()
	if ePath == "" {
		ePath = ProjectPath
	} else {
		ePath = filepath.Dir(ePath)
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

// GetMimeType get file mime type
func GetMimeType(filename string, content []byte) (ctype string) {
	if len(content) > 0 {
		ctype = http.DetectContentType(content)
	}

	if filename != "" && (ctype == "" || strings.HasPrefix(ctype, "text/plain")) {
		ntype := mime.TypeByExtension(filepath.Ext(filename))
		if ntype != "" {
			ctype = ntype
		}
	}
	return ctype
}

// HasPermission check file or directory permission
func HasPermission(path string, perm os.FileMode, noUp ...bool) bool {
	path = RealPath(path)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if len(noUp) == 0 || !noUp[0] {
				return HasPermission(filepath.Dir(path), perm, false)
			}
			return false
		}

		return !os.IsPermission(err)
	}

	return info.Mode()&perm == perm
}

// HasReadWritePermission check file or directory read and write permission
func HasReadWritePermission(path string) bool {
	return HasPermission(path, fs.FileMode(0o600))
}

type fileInfo struct {
	modTime time.Time
	path    string
	size    uint64
}

type fileInfos []fileInfo

func (f fileInfos) Len() int           { return len(f) }
func (f fileInfos) Less(i, j int) bool { return f[i].modTime.Before(f[j].modTime) }
func (f fileInfos) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

type DirStatOptions struct {
	MaxSize  uint64 // max size
	MaxTotal uint64 // max total files
}

// StatDir get directory size and total files
func StatDir(path string, options ...DirStatOptions) (size, total uint64, err error) {
	var (
		totalSize   uint64
		totalFiles  uint64
		files       = make([]fileInfo, 0, 1024)
		needCollect = len(options) > 0 && (options[0].MaxSize > 0 || options[0].MaxTotal > 0)
	)

	err = filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}

			fileSize := uint64(info.Size())
			totalSize += fileSize
			totalFiles++

			if needCollect {
				files = append(files, fileInfo{
					path:    filePath,
					size:    fileSize,
					modTime: info.ModTime(),
				})
			}
		}
		return nil
	})
	if err != nil {
		return totalSize, totalFiles, err
	}

	if needCollect {
		needCleanup := totalSize > options[0].MaxSize
		if options[0].MaxTotal > 0 && totalFiles > options[0].MaxTotal {
			needCleanup = true
		}

		if needCleanup {
			sort.Sort(fileInfos(files))
			for i := 0; i < len(files); i++ {
				if (options[0].MaxSize == 0 || totalSize <= options[0].MaxSize) &&
					(options[0].MaxTotal == 0 || totalFiles <= options[0].MaxTotal) {
					break
				}

				if err := os.Remove(files[i].path); err != nil {
					return totalSize, totalFiles, fmt.Errorf("failed to delete file %s: %v", files[i].path, err)
				}
				totalSize -= files[i].size
				totalFiles--
			}
		}
	}

	return totalSize, totalFiles, nil
}
