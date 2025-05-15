package zstring

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
)

// Base64Encode encodes a byte slice using standard base64 encoding.
func Base64Encode(value []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(value)))
	base64.StdEncoding.Encode(dst, value)
	return dst
}

// Base64EncodeString encodes a string using standard base64 encoding.
func Base64EncodeString(value string) string {
	data := String2Bytes(value)
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode decodes a base64 encoded byte slice.
func Base64Decode(data []byte) (value []byte, err error) {
	src := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(src, data)
	return src[:n], err
}

// Base64DecodeString decodes a base64 encoded string.
func Base64DecodeString(data string) (value string, err error) {
	var dst []byte
	dst, err = base64.StdEncoding.DecodeString(data)
	if err == nil {
		value = Bytes2String(dst)
	}
	return
}

// Serialize converts a value to a byte slice using Go's gob encoding.
// The value must be gob-encodable.
func Serialize(value interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	gob.Register(value)

	err := enc.Encode(&value)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnSerialize converts a byte slice back to its original value using Go's gob decoding.
// Additional types that need to be registered with gob can be passed as registers.
func UnSerialize(valueBytes []byte, registers ...interface{}) (value interface{}, err error) {
	for _, v := range registers {
		gob.Register(v)
	}
	buf := bytes.NewBuffer(valueBytes)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&value)
	return
}

// Img2Base64 reads an image file and converts it to a base64 encoded data URL.
// The returned string can be used directly in HTML img tags.
func Img2Base64(path string) (string, error) {
	path = zfile.RealPath(path)
	imgType := "jpg"
	ext := filepath.Ext(path)
	if ext != "" {
		imgType = imgType[1:]
	}
	imgBuffer, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	imgType = strings.ToLower(imgType)
	return fmt.Sprintf("data:image/%s;base64,%s", imgType, Bytes2String(Base64Encode(imgBuffer))), nil
}
