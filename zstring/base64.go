package zstring

import (
	"encoding/base64"
)

func Base64Encode(value []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(value)))
	base64.StdEncoding.Encode(dst, value)
	return dst
}

func Base64EncodeString(value string) string {
	data := String2Bytes(value)
	return base64.StdEncoding.EncodeToString(data)
}

func Base64Decode(data []byte) (value []byte, err error) {
	src := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(src, data)
	return src[:n], err
}

func Base64DecodeString(data string) (value string, err error) {
	var dst []byte
	dst, err = base64.StdEncoding.DecodeString(data)
	if err == nil {
		value = Bytes2String(dst)
	}
	return
}
