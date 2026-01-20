package zstring

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestPKCS7UnPaddingValidation(t *testing.T) {
	tt := zlsgo.NewTest(t)

	data := []byte("abc")
	padded := PKCS7Padding(data, 8)
	out, err := PKCS7UnPadding(padded)
	tt.EqualNil(err)
	tt.Equal(data, out)

	_, err = PKCS7UnPadding([]byte{})
	tt.EqualTrue(err != nil)

	_, err = PKCS7UnPadding([]byte{1, 2, 3, 0})
	tt.EqualTrue(err != nil)

	_, err = PKCS7UnPadding([]byte{1, 2, 3, 5})
	tt.EqualTrue(err != nil)

	_, err = PKCS7UnPadding([]byte{1, 2, 3, 2})
	tt.EqualTrue(err != nil)
}
