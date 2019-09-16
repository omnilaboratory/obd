package tool

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	commonKey = []byte("nanjishidu170416")
	syncMutex sync.Mutex
)

func SetAesKey(key string) {
	syncMutex.Lock()
	defer syncMutex.Unlock()
	commonKey = []byte(key)
}
func AesEncrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(commonKey)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cipher.NewCFBEncrypter(block, iv).XORKeyStream(cipherText[aes.BlockSize:],
		[]byte(plaintext))
	return hex.EncodeToString(cipherText), nil

}
func AesDecrypt(d string) (string, error) {
	ciphertext, err := hex.DecodeString(d)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(commonKey)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	fmt.Println(len(ciphertext), len(iv))
	cipher.NewCFBDecrypter(block, iv).XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}
