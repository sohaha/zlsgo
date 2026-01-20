package zstring

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sohaha/zlsgo/zfile"
)

var (
	base64EncodePool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 0, 4096)
			return &buf
		},
	}
	base64DecodePool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 0, 4096)
			return &buf
		},
	}
)

// Base64Encode encodes a byte slice using standard base64 encoding.
func Base64Encode(value []byte) []byte {
	needed := base64.StdEncoding.EncodedLen(len(value))

	bufPtr := base64EncodePool.Get().(*[]byte)
	buf := *bufPtr

	if cap(buf) < needed {
		buf = make([]byte, needed)
	}
	buf = buf[:needed]

	base64.StdEncoding.Encode(buf, value)

	result := make([]byte, len(buf))
	copy(result, buf)

	for i := range buf {
		buf[i] = 0
	}
	*bufPtr = buf[:0]
	base64EncodePool.Put(bufPtr)

	return result
}

// Base64EncodeString encodes a string using standard base64 encoding.
func Base64EncodeString(value string) string {
	data := String2Bytes(value)
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode decodes a base64 encoded byte slice.
func Base64Decode(data []byte) (value []byte, err error) {
	needed := base64.StdEncoding.DecodedLen(len(data))

	bufPtr := base64DecodePool.Get().(*[]byte)
	buf := *bufPtr

	if cap(buf) < needed {
		buf = make([]byte, needed)
	}
	buf = buf[:needed]

	n, err := base64.StdEncoding.Decode(buf, data)

	value = make([]byte, n)
	copy(value, buf[:n])

	for i := range buf {
		buf[i] = 0
	}
	*bufPtr = buf[:0]
	base64DecodePool.Put(bufPtr)

	return
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
