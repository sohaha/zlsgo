// Package zfile provides file and path operations for common development tasks.
// It includes utilities for file manipulation, path resolution, and directory management.
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

// PathExist checks if a path exists and determines if it's a directory or file.
// Returns:
//   - 1: path exists and is a directory
//   - 2: path exists and is a file
//   - 3: path exists and is a symlink
//   - 0: path does not exist
func PathExist(path string) (int, error) {
	path = RealPath(path)
	f, err := os.Stat(path)
	if err == nil {
		isFile := 2
		if f.IsDir() {
			isFile = 1
		} else if f.Mode()&os.ModeSymlink != 0 {
			isFile = 3
		}
		return isFile, nil
	}

	return 0, err
}

// DirExist checks if the specified path exists and is a directory.
// Returns true if the path exists and is a directory, false otherwise.
func DirExist(path string) bool {
	state, _ := PathExist(path)
	return state == 1
}

// FileExist checks if the specified path exists and is a file.
// Returns true if the path exists and is a file, false otherwise.
func FileExist(path string) bool {
	state, _ := PathExist(path)
	return state == 2
}

// FileSize returns the formatted size of a file (e.g., "1.5 MB").
// If the file doesn't exist, it returns an empty string.
func FileSize(file string) (size string) {
	return SizeFormat(int64(FileSizeUint(file)))
}

// FileSizeUint returns the size of a file in bytes as a uint64.
// If the file doesn't exist, it returns 0.
func FileSizeUint(file string) (size uint64) {
	file = RealPath(file)
	fileInfo, err := os.Stat(file)
	if err != nil {
		return 0
	}
	return uint64(fileInfo.Size())
}

// SizeFormat converts a size in bytes (int) to a human-readable string
// with appropriate units (B, KB, MB, GB, etc.).
func SizeFormat(s int64) string {
	if s < 0 {
		return "-" + SizeFormat(-s)
	}

	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	humanateBytes := func(s int64, base float64, sizes []string) string {
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

// RootPath returns the absolute path of the current working directory.
// This is typically the directory from which the program was launched.
func RootPath() string {
	path, _ := filepath.Abs(".")
	return RealPath(path)
}

// TmpPath returns a path to a temporary directory.
// If a pattern is provided, it creates a temporary directory with that pattern.
// Otherwise, it returns the system's temporary directory.
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

// SafePath returns a path that is guaranteed to be within the specified base directory.
// If no base directory is provided, it uses the project path as the base.
// This helps prevent directory traversal vulnerabilities.
func SafePath(path string, pathRange ...string) string {
	base := ""
	if len(pathRange) == 0 {
		base = ProjectPath
	} else {
		base = RealPath(pathRange[0], false)
	}
	return strings.TrimPrefix(RealPath(path, false), base)
}

// RealPath converts a relative path to an absolute path.
// If addSlash is true, it ensures the path ends with a slash.
// The function normalizes path separators to forward slashes.
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

// RealPathMkdir converts a path to an absolute path and creates the directory if it doesn't exist.
// If addSlash is true, it ensures the path ends with a slash.
// Note: To ensure the directory can be created successfully, use HasReadWritePermission
// to check permissions before calling this function.
func RealPathMkdir(path string, addSlash ...bool) string {
	realPath := RealPath(path, addSlash...)
	if DirExist(realPath) {
		return realPath
	}

	_ = os.MkdirAll(realPath, os.ModePerm)
	return realPath
}

// IsSubPath checks if subPath is contained within path.
// Both paths are converted to absolute paths before comparison.
func IsSubPath(subPath, path string) bool {
	subPath = RealPath(subPath)
	path = RealPath(path)
	return strings.HasPrefix(subPath, path)
}

// Rmdir removes a directory and all its contents recursively.
// If notIncludeSelf is true, it recreates the directory after deletion,
// effectively removing all contents while keeping the directory itself.
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

// Remove deletes the specified file or empty directory.
// It returns an error if the path doesn't exist or if a non-empty directory is specified.
func Remove(path string) error {
	realPath := RealPath(path)
	return os.Remove(realPath)
}

// CopyFile copies a file from source to destination, preserving file mode.
// It creates the destination file if it doesn't exist and overwrites it if it does.
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

// ExecutablePath returns the absolute path of the current executable file.
// If it cannot determine the executable path, it falls back to using os.Args[0].
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

// ProgramPath returns the directory containing the current executable.
// If addSlash is true, it ensures the path ends with a slash.
// If the executable path cannot be determined, it falls back to ProjectPath.
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

// pathAddSlash adds a trailing slash to the path if addSlash is true and
// the path doesn't already end with a slash.
func pathAddSlash(path string, addSlash ...bool) string {
	if len(addSlash) > 0 && addSlash[0] && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

// GetMimeType determines the MIME type of a file based on its content and/or filename.
// If content is provided, it uses content-based detection first.
// If that fails or returns a generic type, it falls back to extension-based detection.
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

// HasPermission checks if the specified path has the requested permission mode.
// If the path doesn't exist and noUp is false, it checks the parent directory.
// Returns true if the path has the requested permissions, false otherwise.
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

// HasReadWritePermission checks if the specified path has read and write permissions (0600).
// Returns true if the path has read and write permissions, false otherwise.
func HasReadWritePermission(path string) bool {
	return HasPermission(path, fs.FileMode(0o600))
}

// fileInfo is an internal structure used to store file metadata for sorting and cleanup operations.
type fileInfo struct {
	// modTime is the last modification time of the file
	modTime time.Time
	// path is the absolute path to the file
	path string
	// size is the file size in bytes
	size uint64
}

// fileInfos is a slice of fileInfo structures that implements sort.Interface
// to allow sorting files by modification time.
type fileInfos []fileInfo

func (f fileInfos) Len() int           { return len(f) }
func (f fileInfos) Less(i, j int) bool { return f[i].modTime.Before(f[j].modTime) }
func (f fileInfos) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// DirStatOptions provides configuration options for the StatDir function.
type DirStatOptions struct {
	// MaxSize is the maximum allowed total size in bytes for the directory.
	// Files will be deleted (oldest first) if the total size exceeds this value.
	MaxSize uint64
	// MaxTotal is the maximum allowed number of files in the directory.
	// Files will be deleted (oldest first) if the total count exceeds this value.
	MaxTotal uint64
}

// StatDir calculates the total size and number of files in a directory.
// If options are provided with MaxSize or MaxTotal values, it will delete the oldest
// files to keep the directory within the specified limits.
// Returns the final size in bytes, number of files, and any error encountered.
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
