package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
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
	log.Println(serializedPubKey)
	// test TestNet3Params
	// main MainNetParams
	netAddr, err := btcutil.NewAddressPubKey(serializedPubKey, &chaincfg.TestNet3Params)
	if err != nil {
		log.Println(err)
		return "", err
	}
	netAddr.SetFormat(btcutil.PKFCompressed)
	address = netAddr.EncodeAddress()
	fmt.Println("Bitcoin Address:", address)
	return address, nil
}
