package zfile

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CopyDir recursively copies the source directory to the destination directory.
// It preserves file permissions and can filter files using the optional filterFn function.
// If the filter function returns false for a file, that file will not be copied.
func CopyDir(source string, dest string, filterFn ...func(srcFilePath, destFilePath string) bool) (err error) {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dest, info.Mode())
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

// ReadFile reads the entire contents of a file into memory.
// It returns the file contents as a byte slice and any error encountered.
func ReadFile(path string) ([]byte, error) {
	path = RealPath(path)
	return ioutil.ReadFile(path)
}

// ReadLineFile reads a file line by line and calls the provided handle function for each line.
// The handle function receives the line number and the line content as parameters.
// If the handle function returns an error, reading stops and that error is returned.
func ReadLineFile(path string, handle func(line int, data []byte) error) (err error) {
	var f *os.File
	f, err = os.Open(RealPath(path))
	if err != nil {
		return
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	i := 0
	for {
		i++
		line, lerr := rd.ReadBytes('\n')
		if err = handle(i, line); err != nil {
			break
		}
		if lerr != nil || io.EOF == lerr {
			break
		}
	}
	return
}

// WriteFile writes data to a file, creating the file if it doesn't exist.
// If isAppend is true, data is appended to the file; otherwise, the file is overwritten.
// It creates any necessary parent directories automatically.
func WriteFile(path string, b []byte, isAppend ...bool) (err error) {
	var file *os.File
	path = RealPath(path)
	if FileExist(path) {
		if len(isAppend) > 0 && isAppend[0] {
			file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		} else {
			file, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, os.ModeExclusive)
		}
	} else {
		_ = RealPathMkdir(filepath.Dir(path))
		file, err = os.Create(path)
	}
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	file.Close()
	return err
}

// PutOffset writes data to a file at the specified offset position.
// If the file doesn't exist, it will be created.
// This is useful for modifying specific portions of a file without rewriting the entire content.
func PutOffset(path string, b []byte, offset int64) (err error) {
	var file *os.File
	path = RealPath(path)
	if FileExist(path) {
		file, err = os.OpenFile(path, os.O_WRONLY, os.ModeAppend)
	} else {
		file, err = os.Create(path)
	}
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteAt(b, offset)
	return err
}

// PutAppend appends data to the end of a file.
// If the file doesn't exist, it will be created along with any necessary parent directories.
// This is a convenience wrapper around WriteFile with append mode.
func PutAppend(path string, b []byte) (err error) {
	var file *os.File
	path = RealPath(path)
	if FileExist(path) {
		file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	} else {
		_ = RealPathMkdir(filepath.Dir(path))
		file, err = os.Create(path)
	}
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(b)
	return err
}
