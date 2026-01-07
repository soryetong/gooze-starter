package gzauth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var aesKey = []byte("")

func SetAESKey(key string) {
	aesKey = []byte(key)
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	unPadding := int(data[length-1])
	return data[:(length - unPadding)]
}

func AesEncrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	plainBytes := pkcs7Padding([]byte(plainText), block.BlockSize())
	cipherText := make([]byte, len(plainBytes))

	iv := aesKey[:block.BlockSize()]

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, plainBytes)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func AesDecrypt(cipherBase64 string) (string, error) {
	cipherBytes, _ := base64.StdEncoding.DecodeString(cipherBase64)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	iv := aesKey[:block.BlockSize()]
	mode := cipher.NewCBCDecrypter(block, iv)

	plain := make([]byte, len(cipherBytes))
	mode.CryptBlocks(plain, cipherBytes)

	plain = pkcs7UnPadding(plain)
	return string(plain), nil
}
