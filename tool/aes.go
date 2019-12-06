package tool

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"sync"
)

var (
	commonKey = []byte("03a492b58a0df0a3")
	iv        = "03a492b58a0df0a3"
	syncMutex sync.Mutex
)

func SetAesAndIV(key string, myIv string) {
	syncMutex.Lock()
	defer syncMutex.Unlock()
	if CheckIsString(&key) {
		commonKey = []byte(key)
	}
	if CheckIsString(&myIv) {
		iv = myIv
	}
}

// cbc mode
func AesEncrypt(encodeStr string, psw string) (string, error) {
	if CheckIsString(&encodeStr) == false {
		return "", errors.New("empty data")
	}

	md5Psw := SignMsgWithMd5([]byte(psw))
	commonKey = []byte(md5Psw)
	iv = string(commonKey[0:16])
	block, err := aes.NewCipher(commonKey)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	encodeBytes := []byte(encodeStr)
	encodeBytes = PKCS5Padding(encodeBytes, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))
	crypted := make([]byte, len(encodeBytes))

	blockMode.CryptBlocks(crypted, encodeBytes)

	return base64.StdEncoding.EncodeToString(crypted), nil
}

func AesDecrypt2(cryted string, psw string) (string, error) {
	if CheckIsString(&cryted) == false {
		return "", errors.New("empty data")
	}
	crytedByte, err := base64.StdEncoding.DecodeString(cryted)
	if err != nil {
		return "", err
	}
	md5Psw := SignMsgWithMd5([]byte(psw))
	commonKey = []byte(md5Psw)
	iv = string(commonKey[0:16])

	block, err := aes.NewCipher(commonKey)
	if err != nil {
		return "", err
	}

	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	orig := make([]byte, len(crytedByte))
	blockMode.CryptBlocks(orig, crytedByte)
	orig = PKCS5UnPadding(orig)
	return string(orig), nil
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
