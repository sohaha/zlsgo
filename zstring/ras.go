package zstring

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// RSAEncrypt RSA Encrypt
func RSAEncrypt(plainText []byte, publicKey string) ([]byte, error) {
	block, _ := pem.Decode(String2Bytes(publicKey))
	if block == nil {
		return nil, errors.New("public key is illegal")
	}
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, _ := publicKeyInterface.(*rsa.PublicKey)
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, k, plainText)
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

// RSADecrypt RSA Decrypt
func RSADecrypt(cipherText []byte, privateKey string) ([]byte, error) {
	block, _ := pem.Decode(String2Bytes(privateKey))
	if block == nil {
		return nil, errors.New("private key is illegal")
	}
	k, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	cipherText, _ = Base64Decode(cipherText)
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, k, cipherText)
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
