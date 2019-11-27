package tool

import (
	"LightningOnOmni/config"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
	"io"
	"log"
	"math/big"
	"regexp"
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

func VerifyEmailFormat(email string) bool {
	isString := CheckIsString(&email)
	if isString == false {
		return false
	}
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func SignMsgWithSha256(msg []byte) string {
	hash := sha256.New()
	hash.Write(msg)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
func SignMsgWithBase58(msg []byte) string {
	hash := base58.Encode(msg)
	return fmt.Sprintf("%x", hash)
}

func SignMsgWithRipemd160(msg []byte) string {
	hash := ripemd160.New()
	hash.Write(msg)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func SignMsgWithMd5(msg []byte) string {
	hash := md5.New()
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

func GetHtlcFee() float64 {
	return 1
}

// PaddedBigBytes encodes a big integer as a big-endian byte slice. The length
// of the slice is at least n bytes.
func PaddedBigBytes(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	ReadBits(bigint, ret)
	return ret
}

const (
	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
)

func ReadBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}
