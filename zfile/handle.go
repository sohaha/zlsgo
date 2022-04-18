package zfile

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CopyDir copies the source directory to the dest directory.
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

// ReadFile ReadFile
func ReadFile(path string) ([]byte, error) {
	path = RealPath(path)
	return ioutil.ReadFile(path)
}

// ReadLineFile ReadLineFile
func ReadLineFile(path string, handle func(line int,
	data []byte) error) (err error) {
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

// WriteFile WriteFile
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

// PutOffset open the specified file and write data from the specified location
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

// PutAppend open the specified file and write data at the end of the file
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
