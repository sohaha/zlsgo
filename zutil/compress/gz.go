package compress

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
)

// Compress use gzip to compress to tar.gz
func Compress(currentPath, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	currentPath = zfile.RealPath(currentPath, true)
	dest = zfile.RealPath(dest)
	err := filepath.Walk(currentPath,
		func(path string, info os.FileInfo, err error) error {
			if info == nil || err != nil {
				return err
			}
			if info.IsDir() || path == dest {
				return nil
			}
			header, e := tar.FileInfoHeader(info, "")
			if e != nil {
				return e
			}
			header.Name = strings.Replace(path, currentPath, "", -1)
			e = tw.WriteHeader(header)
			if e != nil {
				return e
			}
			var file *os.File
			file, e = os.Open(path)
			if e != nil {
				return e
			}
			defer file.Close()
			_, e = io.Copy(tw, file)
			if e != nil {
				return e
			}
			return nil
		})
	return err
}

// DeCompress unzip tar.gz
func DeCompress(tarFile, dest string) error {
	dest = zfile.RealPath(dest, true)
	tarFile = zfile.RealPath(tarFile)
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return err
		}
		_, _ = io.Copy(file, tr)
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}
