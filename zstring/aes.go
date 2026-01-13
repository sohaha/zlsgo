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

// assKeyPadding ensures the encryption key is of valid length (16, 24, or 32 bytes).
// If the key is too short, it is padded with spaces; if too long, it is truncated.
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

// PKCS7Padding implements PKCS#7 padding for block cipher encryption.
// It adds padding bytes to ensure the data length is a multiple of the block size.
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, pad...)
}

// PKCS7UnPadding removes PKCS#7 padding from decrypted data.
// It returns an error if the padding is invalid.
func PKCS7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("encryption string error")
	} else {
		u := int(origData[length-1])
		return origData[:(length - u)], nil
	}
}

// AesEncrypt encrypts data using AES in CBC mode with PKCS#7 padding.
// If an IV is provided, it is used; otherwise, the key is used as the IV.
func AesEncrypt(plainText []byte, key string, iv ...string) (ciphertext []byte,
	err error,
) {
	var k []byte
	var block cipher.Block
	if len(iv) > 0 {
		k = String2Bytes(iv[0])
		if len(k) < aes.BlockSize {
			return nil, errors.New("IV length must be at least 16 bytes for AES")
		}
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

// AesDecrypt decrypts data that was encrypted with AesEncrypt.
// If an IV is provided, it must match the IV used during encryption.
// If no IV is provided, the key is used as the IV.
func AesDecrypt(cipherText []byte, key string, iv ...string) (plainText []byte, err error) {
	var (
		block cipher.Block
		k     []byte
	)
	if len(iv) > 0 {
		k = String2Bytes(iv[0])
		if len(k) < aes.BlockSize {
			return nil, errors.New("IV length must be at least 16 bytes for AES")
		}
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

// AesEncryptString encrypts a string using AES and returns the result as a base64-encoded string.
// This is a convenience wrapper around AesEncrypt.
func AesEncryptString(plainText string, key string, iv ...string) (string, error) {
	str := ""
	c, err := AesEncrypt(String2Bytes(plainText), key, iv...)
	if err == nil {
		str = Bytes2String(Base64Encode(c))
	}
	return str, err
}

// AesDecryptString decrypts a base64-encoded string that was encrypted with AesEncryptString.
// This is a convenience wrapper around AesDecrypt.
func AesDecryptString(cipherText string, key string, iv ...string) (string,
	error,
) {
	base64Byte, _ := Base64Decode(String2Bytes(cipherText))
	origData, err := AesDecrypt(base64Byte, key, iv...)
	if err != nil {
		return "", err
	}
	return Bytes2String(origData), nil
}

// AesGCMEncrypt encrypts data using AES in GCM mode, which provides both confidentiality and authenticity.
// It generates a random nonce for each encryption operation.
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

// AesGCMDecrypt decrypts data that was encrypted with AesGCMEncrypt.
// It extracts the nonce from the beginning of the ciphertext.
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

// AesGCMEncryptString encrypts a string using AES-GCM and returns the result as a base64-encoded string.
// This is a convenience wrapper around AesGCMEncrypt.
func AesGCMEncryptString(plainText string, key string) (string, error) {
	str := ""
	c, err := AesGCMEncrypt(String2Bytes(plainText), key)
	if err == nil {
		str = Bytes2String(Base64Encode(c))
	}
	return str, err
}

// AesGCMDecryptString decrypts a base64-encoded string that was encrypted with AesGCMEncryptString.
// This is a convenience wrapper around AesGCMDecrypt.
func AesGCMDecryptString(cipherText string, key string) (string,
	error,
) {
	base64Byte, _ := Base64Decode(String2Bytes(cipherText))
	origData, err := AesGCMDecrypt(base64Byte, key)
	if err != nil {
		return "", err
	}
	return Bytes2String(origData), nil
}
