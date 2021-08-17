package zstring

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

// RSAEncrypt RSA Encrypt
func RSAEncrypt(plainText []byte, publicKey string) ([]byte, error) {
	pub, err := pubKey(String2Bytes(publicKey))
	if err != nil {
		return nil, err
	}
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pub, plainText)
	if err != nil {
		return nil, err
	}
	return Base64Encode(cipherText), nil
}

// RSAEncryptString RSA Encrypt to String
func RSAEncryptString(plainText string, publicKey string) (string, error) {
	c, err := RSAEncrypt(String2Bytes(plainText), publicKey)
	if err != nil {
		return "", err
	}
	return Bytes2String(c), nil
}

// RSAPriKeyEncrypt RSA PriKey Encrypt
func RSAPriKeyEncrypt(plainText []byte, privateKey string) ([]byte, error) {
	pri, err := priKey(String2Bytes(privateKey))
	if err != nil {
		return nil, err
	}
	cipherText, err := rsa.SignPKCS1v15(nil, pri, crypto.Hash(0), plainText)
	if err != nil {
		return nil, err
	}
	return Base64Encode(cipherText), nil
}

// RSAPriKeyEncryptString RSA PriKey Encrypt to String
func RSAPriKeyEncryptString(plainText string, privateKey string) (string, error) {
	c, err := RSAPriKeyEncrypt(String2Bytes(plainText), privateKey)
	if err != nil {
		return "", err
	}
	return Bytes2String(c), nil
}

// RSADecrypt RSA Decrypt
func RSADecrypt(cipherText []byte, privateKey string) ([]byte, error) {
	pri, err := priKey(String2Bytes(privateKey))
	if err != nil {
		return nil, err
	}
	cipherText, _ = Base64Decode(cipherText)
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, pri, cipherText)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}

// RSADecryptString RSA Decrypt to String
func RSADecryptString(cipherText string, privateKey string) (string, error) {
	p, err := RSADecrypt(String2Bytes(cipherText), privateKey)
	if err != nil {
		return "", err
	}
	return Bytes2String(p), nil
}

// RSAPubKeyDecrypt RSA PubKey Decrypt
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

// RSAPubKeyDecryptString RSA PubKey Decrypt to String
func RSAPubKeyDecryptString(cipherText string, publicKey string) (string,
	error) {
	p, err := RSAPubKeyDecrypt(String2Bytes(cipherText), publicKey)
	if err != nil {
		return "", err
	}
	return Bytes2String(p), nil
}

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

func priKey(privateKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key is illegal")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func leftPad(input []byte, size int) (out []byte) {
	n := len(input)
	if n > size {
		n = size
	}
	out = make([]byte, size)
	copy(out[len(out)-n:], input)
	return
}

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
