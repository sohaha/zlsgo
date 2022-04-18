package zstring

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func Md5(s string) string {
	return Md5Byte(String2Bytes(s))
}

func Md5Byte(s []byte) string {
	h := md5.New()
	_, _ = h.Write(s)
	return hex.EncodeToString(h.Sum(nil))
}

func Md5File(path string) (encrypt string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	// r := bufio.NewReader(f)
	h := md5.New()
	_, err = io.Copy(h, f)
	f.Close()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
