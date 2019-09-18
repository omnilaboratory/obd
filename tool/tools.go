package tool

import (
	"LightningOnOmni/config"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"io"
	"log"
	"strings"
)

func CheckIsString(str *string) bool {
	if str == nil {
		return false
	}
	if len(strings.Trim(*str, " ")) == 0 {
		return false
	}
	return true
}

func SignMsg(msg []byte) string {
	hash := sha256.New()
	hash.Write(msg)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func GetAddressFromPubKey(pubKey string) (address string, err error) {
	log.Println(pubKey)
	serializedPubKey, err := hex.DecodeString(pubKey)
	if err != nil {
		log.Println(err)
		return "", err
	}
	// test TestNet3Params
	// main MainNetParams
	var net *chaincfg.Params
	if config.ChainNode_Type == "test" {
		net = &chaincfg.TestNet3Params
	} else {
		net = &chaincfg.MainNetParams
	}
	netAddr, err := btcutil.NewAddressPubKey(serializedPubKey, net)
	if err != nil {
		log.Println(err)
		return "", err
	}
	netAddr.SetFormat(btcutil.PKFCompressed)
	address = netAddr.EncodeAddress()
	log.Println("BitCoin Address:", address)

	return address, nil
}

func RandBytes(size int) (string, error) {
	arr := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, arr); err != nil {
		return "", err
	}
	log.Println(arr)
	return base64.StdEncoding.EncodeToString(arr), nil
}
