package zfile

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GzCompress use gzip to compress to tar.gz
func GzCompress(currentPath, dest string) (err error) {
	dest = RealPath(dest)
	var d *os.File
	d, err = os.Create(dest)
	if err != nil {
		return
	}
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	currentPath = RealPath(currentPath, true)
	err = walkFile(currentPath, dest, err, func(path string, info *os.FileInfo) error {
		header, err := tar.FileInfoHeader(*info, "")
		if err != nil {
			return err
		}
		header.Name = strings.Replace(path, currentPath, "", -1)
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		var file *os.File
		file, err = os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		return err
	})
	return
}

func walkFile(currentPath string, dest string, err error, writer func(path string, info *os.FileInfo) error) error {
	err = filepath.Walk(currentPath,
		func(path string, info os.FileInfo, err error) error {
			path = RealPath(path)
			if info == nil || err != nil {
				return err
			}
			if info.IsDir() || path == dest {
				return nil
			}

			return writer(path, &info)
		})
	return err
}

// GzDeCompress unzip tar.gz
func GzDeCompress(tarFile, dest string) error {
	dest = RealPath(dest, true)
	tarFile = RealPath(tarFile)
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
		i := hdr.FileInfo()
		filename := dest + hdr.Name
		if i.IsDir() {
			_ = createDir(filename, i.Mode())
		} else {
			file, err := createFile(filename)
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tr)
			_ = file.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ZipCompress zip
func ZipCompress(currentPath, dest string) (err error) {
	dest = RealPath(dest)
	var d *os.File
	d, err = os.Create(dest)
	if err != nil {
		return
	}
	defer d.Close()
	tw := zip.NewWriter(d)
	defer tw.Close()
	currentPath = RealPath(currentPath, true)
	err = walkFile(currentPath, dest, err, func(path string, info *os.FileInfo) error {
		header, err := zip.FileInfoHeader(*info)
		if err != nil {
			return err
		}
		header.Name = strings.Replace(path, currentPath, "", -1)
		writer, err := tw.CreateHeader(header)
		if err != nil {
			return err
		}
		var file *os.File
		file, err = os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		_ = file.Close()
		return err
	})
	return err
}

func ZipDeCompress(zipFile, dest string) error {
	dest = RealPath(dest, true)
	zipFile = RealPath(zipFile)
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		filename := dest + file.Name
		if file.FileInfo().IsDir() {
			_ = createDir(filename, file.FileInfo().Mode())
			continue
		}
		w, err := createFile(filename)
		if err != nil {
			_ = rc.Close()
			return err
		}
		_, err = io.Copy(w, rc)
		_ = w.Close()
		_ = rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func createDir(dir string, perm os.FileMode) error {
	if perm == 0 {
		perm = 0755
	}
	return os.MkdirAll(dir, perm)
}

func createFile(name string) (*os.File, error) {
	dir := string([]rune(name)[0:strings.LastIndex(name, "/")])
	if !DirExist(dir) {
		err := createDir(dir, 0)
		if err != nil {
			return nil, err
		}
	}

	return os.Create(name)
}
