package zstring

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
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

// PKCS7Padding PKCS7 fill mode
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, pad...)
}

// PKCS7UnPadding Reverse operation of padding to delete the padding string
func PKCS7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("encryption string error")
	} else {
		u := int(origData[length-1])
		return origData[:(length - u)], nil
	}
}

// AesEncrypt aes encryption
func AesEncrypt(plainText []byte, key string, iv ...string) (ciphertext []byte,
	err error) {
	var k []byte
	var block cipher.Block
	if len(iv) > 0 {
		k = String2Bytes(iv[0])
		block, err = aes.NewCipher(String2Bytes(key))
	} else {
		k = assKeyPadding(key)
		block, err = aes.NewCipher(k)
	}
	if err == nil {
		blockSize := block.BlockSize()
		plainText = PKCS7Padding(plainText, blockSize)
		blocMode := cipher.NewCBCEncrypter(block, k[:blockSize])
		ciphertext = make([]byte, len(plainText))
		blocMode.CryptBlocks(ciphertext, plainText)
	}
	return
}

// AesDecrypt aes decryption
func AesDecrypt(cipherText []byte, key string, iv ...string) (plainText []byte, err error) {
	var (
		block cipher.Block
		k     []byte
	)
	if len(iv) > 0 {
		k = String2Bytes(iv[0])
		block, err = aes.NewCipher(String2Bytes(key))
	} else {
		k = assKeyPadding(key)
		block, err = aes.NewCipher(k)
	}

	if err == nil {
		blockSize := block.BlockSize()
		blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
		plainText = make([]byte, len(cipherText))
		defer func() {
			if e := recover(); e != nil {
				var ok bool
				err, ok = e.(error)
				if !ok {
					err = fmt.Errorf("%s", e)
				}
			}
		}()
		blockMode.CryptBlocks(plainText, cipherText)
		if err == nil {
			plainText, err = PKCS7UnPadding(plainText)
		}
	}
	return
}

// AesEncryptString Aes Encrypt to String
func AesEncryptString(plainText string, key string, iv ...string) (string, error) {
	str := ""
	c, err := AesEncrypt(String2Bytes(plainText), key, iv...)
	if err == nil {
		str = Bytes2String(Base64Encode(c))
	}
	return str, nil
}

// AesDecryptString Aes Decrypt to String
func AesDecryptString(cipherText string, key string, iv ...string) (string,
	error) {
	base64Byte, _ := Base64Decode(String2Bytes(cipherText))
	origData, err := AesDecrypt(base64Byte, key, iv...)
	if err != nil {
		return "", err
	}
	return Bytes2String(origData), nil
}

func AesGCMEncrypt(plaintext []byte, key string) (ciphertext []byte, err error) {
	var (
		block  cipher.Block
		aesGCM cipher.AEAD
	)

	block, err = aes.NewCipher(String2Bytes(key))
	if err != nil {
		return
	}

	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	ciphertext = aesGCM.Seal(nonce, nonce, plaintext, nil)
	return
}

func AesGCMDecrypt(ciphertext []byte, key string) (plaintext []byte, err error) {
	if len(ciphertext) == 0 {
		return nil, errors.New("ciphertext is empty")
	}

	var (
		block  cipher.Block
		aesGCM cipher.AEAD
	)
	block, err = aes.NewCipher(String2Bytes(key))
	if err != nil {
		return
	}

	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext is too short")
	}
	nonce, text := ciphertext[:nonceSize], ciphertext[nonceSize:]

	return aesGCM.Open(nil, nonce, text, nil)
}

func AesGCMEncryptString(plainText string, key string) (string, error) {
	str := ""
	c, err := AesGCMEncrypt(String2Bytes(plainText), key)
	if err == nil {
		str = Bytes2String(Base64Encode(c))
	}
	return str, err
}

func AesGCMDecryptString(cipherText string, key string) (string,
	error) {
	base64Byte, _ := Base64Decode(String2Bytes(cipherText))
	origData, err := AesGCMDecrypt(base64Byte, key)
	if err != nil {
		return "", err
	}
	return Bytes2String(origData), nil
}
