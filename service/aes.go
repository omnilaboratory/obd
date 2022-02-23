package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
)

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func AesEncrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func AesDecrypt(data []byte, passphrase string) ([]byte,error) {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		//panic(err.Error())
		return nil ,err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		//panic(err.Error())
		return nil ,err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		//panic(err.Error())
		return nil ,err
	}
	return plaintext,err
}

func encryptHexFile(filename string, data []byte, passphrase string) error {
	f, err := os.Create(filename)
	defer f.Close()
	if err == nil {
		_,err=f.WriteString(hex.EncodeToString(AesEncrypt(data, passphrase)))
	}
	return err
}

func decryptHexFile(filename string, passphrase string) ([]byte,error) {
	data, _ := ioutil.ReadFile(filename)
	bs,err:=hex.DecodeString(string(data))
	if err == nil {
		return AesDecrypt(bs, passphrase)
	}
	return nil,err
}