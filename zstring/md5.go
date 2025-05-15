package zstring

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// ProjectMd5 calculates the MD5 hash of the current executable.
// This is useful for checking if the application binary has been modified.
func ProjectMd5() string {
	d, _ := Md5File(os.Args[0])
	return d
}

// Md5 calculates the MD5 hash of a string.
// It returns the hash as a hexadecimal encoded string.
func Md5(s string) string {
	return Md5Byte(String2Bytes(s))
}

// Md5Byte calculates the MD5 hash of a byte slice.
// It returns the hash as a hexadecimal encoded string.
func Md5Byte(s []byte) string {
	h := md5.New()
	_, _ = h.Write(s)
	return hex.EncodeToString(h.Sum(nil))
}

// Md5File calculates the MD5 hash of a file at the given path.
// It returns the hash as a hexadecimal encoded string and any error encountered.
func Md5File(path string) (encrypt string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	// r := bufio.NewReader(f)
	h := md5.New()
	_, err = io.Copy(h, f)
	_ = f.Close()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
