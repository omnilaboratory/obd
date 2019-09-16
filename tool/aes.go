package tool

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"sync"
)

var (
	commonKey = []byte("nanjishidu170416")
	iv        = "03a492b58a0df0a3"
	syncMutex sync.Mutex
)

func SetAesAndIV(key string, myIv string) {
	syncMutex.Lock()
	defer syncMutex.Unlock()
	commonKey = []byte(key)
	iv = myIv
}
func AesEncrypt(encodeStr string) (string, error) {
	encodeBytes := []byte(encodeStr)
	block, err := aes.NewCipher(commonKey)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	encodeBytes = PKCS5Padding(encodeBytes, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))
	crypted := make([]byte, len(encodeBytes))
	blockMode.CryptBlocks(crypted, encodeBytes)

	return base64.StdEncoding.EncodeToString(crypted), nil
}

func AesDecrypt2(cryted string) string {
	crytedByte, _ := base64.StdEncoding.DecodeString(cryted)
	block, _ := aes.NewCipher(commonKey)

	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	orig := make([]byte, len(crytedByte))
	blockMode.CryptBlocks(orig, crytedByte)
	orig = PKCS5UnPadding(orig)
	return string(orig)
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
