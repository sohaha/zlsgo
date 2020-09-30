package zstring

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"strings"
)

func assKeyPadding(key string) []byte {
	k := String2Bytes(key)
	l := len(k)
	switch l {
	case 16, 24, 32:
		return k
	default:
		if l < 16 {
			return append(k, String2Bytes(strings.Repeat(" ", 16-l))...)
		} else if l < 24 {
			return append(k, String2Bytes(strings.Repeat(" ", 24-l))...)
		} else if l < 32 {
			return append(k, String2Bytes(strings.Repeat(" ", 32-l))...)
		}
		return k[0:32]
	}
}

// PKCS7 fill mode
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding Reverse operation of padding to delete the padding string
func PKCS7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("encryption string error")
	} else {
		unpadding := int(origData[length-1])
		return origData[:(length - unpadding)], nil
	}
}

// AesEnCrypt aes encryption
func AesEnCrypt(origData []byte, key string) (crypted []byte, err error) {
	k := assKeyPadding(key)
	var block cipher.Block
	block, err = aes.NewCipher(k)
	if err == nil {
		blockSize := block.BlockSize()
		origData = PKCS7Padding(origData, blockSize)
		blocMode := cipher.NewCBCEncrypter(block, k[:blockSize])
		crypted = make([]byte, len(origData))
		blocMode.CryptBlocks(crypted, origData)
	}
	return
}

// AesDeCrypt aes decryption
func AesDeCrypt(cypted []byte, key string) (origData []byte, err error) {
	k := assKeyPadding(key)
	var block cipher.Block
	block, err = aes.NewCipher(k)
	if err == nil {
		blockSize := block.BlockSize()
		blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
		origData = make([]byte, len(cypted))
		defer func() {
			if e := recover(); e != nil {
				var ok bool
				err, ok = e.(error)
				if !ok {
					err = fmt.Errorf("%s", e)
				}
			}
		}()
		blockMode.CryptBlocks(origData, cypted)
		if err == nil {
			origData, err = PKCS7UnPadding(origData)
		}
	}
	return
}

func AesEnCryptString(origData string, key string) (string, error) {
	str := ""
	cypted, err := AesEnCrypt(String2Bytes(origData), key)
	if err == nil {
		str = Bytes2String(Base64Encode(cypted))
	}
	return str, nil
}

func AesDeCryptString(cypted string, key string) (string, error) {
	base64Byte, _ := Base64Decode(String2Bytes(cypted))
	origData, err := AesDeCrypt(base64Byte, key)
	if err != nil {
		return "", err
	}
	return Bytes2String(origData), nil
}
