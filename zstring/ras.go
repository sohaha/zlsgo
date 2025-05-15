package zstring

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

// GenRSAKey generates a pair of RSA private and public keys.
// The optional bits parameter specifies the key size (defaults to 1024 bits).
func GenRSAKey(bits ...int) (prvkey, pubkey []byte, err error) {
	l := 1024
	if len(bits) > 0 {
		l = bits[0]
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, l)
	if err != nil {
		return nil, nil, err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvkey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubkey = pem.EncodeToMemory(block)
	return
}

// RSAEncrypt encrypts data using RSA with a public key.
// For large data, use the bits parameter to enable chunked encryption.
func RSAEncrypt(plainText []byte, publicKey string, bits ...int) ([]byte, error) {
	pub, err := pubKey(String2Bytes(publicKey))
	if err != nil {
		return nil, err
	}
	return RSAKeyEncrypt(plainText, pub, bits...)
}

// RSAKeyEncrypt encrypts data using RSA with a public key object.
// For large data, use the bits parameter to enable chunked encryption.
func RSAKeyEncrypt(plainText []byte, publicKey *rsa.PublicKey, bits ...int) ([]byte, error) {
	if len(bits) > 0 && bits[0] > 100 {
		buf := splitBytes(plainText, bits[0]/8-11)
		buffer := bytes.NewBufferString("")
		for i := range buf {
			b, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, buf[i])
			if err != nil {
				return nil, err
			}
			buffer.Write(b)
		}
		return Base64Encode(buffer.Bytes()), nil
	}
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainText)
	if err != nil {
		return nil, err
	}
	return Base64Encode(cipherText), nil
}

// RSAEncryptString encrypts a string using RSA and returns the result as a base64-encoded string.
func RSAEncryptString(plainText string, publicKey string) (string, error) {
	c, err := RSAEncrypt(String2Bytes(plainText), publicKey)
	if err != nil {
		return "", err
	}
	return Bytes2String(c), nil
}

// RSAPriKeyEncrypt encrypts (signs) data using an RSA private key.
// This is typically used for digital signatures rather than encryption.
func RSAPriKeyEncrypt(plainText []byte, privateKey string) ([]byte, error) {
	prv, err := priKey(String2Bytes(privateKey))
	if err != nil {
		return nil, err
	}
	cipherText, err := rsa.SignPKCS1v15(nil, prv, crypto.Hash(0), plainText)
	if err != nil {
		return nil, err
	}
	return Base64Encode(cipherText), nil
}

// RSAPriKeyEncryptString encrypts (signs) a string using an RSA private key
// and returns the result as a base64-encoded string.
func RSAPriKeyEncryptString(plainText string, privateKey string) (string, error) {
	c, err := RSAPriKeyEncrypt(String2Bytes(plainText), privateKey)
	if err != nil {
		return "", err
	}
	return Bytes2String(c), nil
}

// RSADecrypt decrypts data using RSA with a private key.
// For large data encrypted in chunks, use the same bits parameter as during encryption.
func RSADecrypt(cipherText []byte, privateKey string, bits ...int) ([]byte, error) {
	prv, err := priKey(String2Bytes(privateKey))
	if err != nil {
		return nil, err
	}
	return RSAKeyDecrypt(cipherText, prv, bits...)
}

// RSAKeyDecrypt decrypts data using RSA with a private key object.
// For large data encrypted in chunks, use the same bits parameter as during encryption.
func RSAKeyDecrypt(cipherText []byte, privateKey *rsa.PrivateKey, bits ...int) ([]byte, error) {
	cipherText, _ = Base64Decode(cipherText)
	if len(bits) > 0 && bits[0] > 100 {
		buf := splitBytes(cipherText, bits[0]/8)
		buffer := bytes.NewBufferString("")
		for i := range buf {
			b, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, buf[i])
			if err != nil {
				return nil, err
			}
			buffer.Write(b)
		}
		return buffer.Bytes(), nil
	}
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
}

// RSADecryptString decrypts a base64-encoded string using RSA with a private key
// and returns the result as a string.
func RSADecryptString(cipherText string, privateKey string) (string, error) {
	p, err := RSADecrypt(String2Bytes(cipherText), privateKey)
	if err != nil {
		return "", err
	}
	return Bytes2String(p), nil
}

// RSAPubKeyDecrypt decrypts (verifies) data that was encrypted with a private key
// using the corresponding public key. This is typically used for signature verification.
func RSAPubKeyDecrypt(cipherText []byte, publicKey string) ([]byte, error) {
	pub, err := pubKey(String2Bytes(publicKey))
	if err != nil {
		return nil, err
	}
	tLen := 0
	k := (pub.N.BitLen() + 7) / 8
	if k < tLen+11 {
		return nil, fmt.Errorf("length illegal")
	}
	cipherText, _ = Base64Decode(cipherText)
	c := new(big.Int).SetBytes(cipherText)
	e := big.NewInt(int64(pub.E))
	m := c.Exp(c, e, pub.N)
	em := leftPad(m.Bytes(), k)
	return unLeftPad(em), nil
}

// RSAPubKeyDecryptString decrypts (verifies) a base64-encoded string that was encrypted
// with a private key using the corresponding public key, and returns the result as a string.
func RSAPubKeyDecryptString(cipherText string, publicKey string) (string,
	error,
) {
	p, err := RSAPubKeyDecrypt(String2Bytes(cipherText), publicKey)
	if err != nil {
		return "", err
	}
	return Bytes2String(p), nil
}

// pubKey parses a PEM encoded RSA public key.
// Returns an error if the key is invalid or in an unsupported format.
func pubKey(publicKey []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key is illegal")
	}
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, _ := publicKeyInterface.(*rsa.PublicKey)
	return pub, nil
}

// priKey parses a PEM encoded RSA private key.
// Supports both PKCS#1 and PKCS#8 formats.
func priKey(privateKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key is illegal")
	}

	switch block.Type {
	case "PRIVATE KEY":
		// pkcs8
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			parsed, ok := parsedKey.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("private key is invalid")
			}
			return parsed, nil
		}
		return nil, err
	default:
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
}

// leftPad pads a byte slice on the left to the specified size.
// Used internally for RSA encryption/decryption operations.
func leftPad(input []byte, size int) (out []byte) {
	n := len(input)
	if n > size {
		n = size
	}
	out = make([]byte, size)
	copy(out[len(out)-n:], input)
	return
}

// unLeftPad removes left padding from a byte slice.
// Used internally for RSA encryption/decryption operations.
func unLeftPad(input []byte) (out []byte) {
	n := len(input)
	t := 2
	for i := 2; i < n; i++ {
		if input[i] == 0xff {
			t = t + 1
		} else {
			if input[i] == input[0] {
				t = t + int(input[1])
			}
			break
		}
	}
	out = make([]byte, n-t)
	copy(out, input[t:])
	return
}

// splitBytes splits a byte slice into chunks of the specified size.
// Used for handling large data in RSA encryption/decryption operations.
func splitBytes(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}
