package zstring

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func Md5(str string) string {
	h := md5.New()
	h.Write(String2Bytes(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Md5File(path string) (encrypt string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
