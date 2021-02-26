package tool

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/omnilaboratory/obd/config"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/ripemd160"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func CheckIsString(str *string) bool {
	if str == nil {
		return false
	}
	*str = strings.Trim(*str, " ")
	if len(*str) == 0 {
		return false
	}
	return true
}
func CheckIsAddress(address string) bool {
	if CheckIsString(&address) == false {
		return false
	}
	_, err := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	if err != nil {
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
	serializedPubKey, err := hex.DecodeString(pubKey)
	if err != nil {
		log.Println(err)
		return "", errors.New("invalid pubKey")
	}
	net := GetCoreNet()
	netAddr, err := btcutil.NewAddressPubKey(serializedPubKey, net)
	if err != nil {
		log.Println(err)
		return "", errors.New("invalid pubKey")
	}
	netAddr.SetFormat(btcutil.PKFCompressed)
	address = netAddr.EncodeAddress()

	return address, nil
}

// 判断文件夹是否存在,如果不存在，则创建
func PathExistsAndCreate(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Println(err)
			return err
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func FloatToString(input_num float64, prec int) string {
	return strconv.FormatFloat(input_num, 'f', prec, 64)
}

func CheckPsw(psw string) (flag bool) {
	reg := regexp.MustCompile("^[a-zA-Z0-9]{6,32}$")
	return reg.MatchString(psw)
}

func GetMacAddrs() (macAddrs string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddrs := netInterface.HardwareAddr.String()
		if len(macAddrs) == 0 {
			continue
		}
		return macAddrs
	}
	return macAddrs
}

func GetUserPeerId(mnemonic string) string {
	source := mnemonic + "@" + GetMacAddrs() + ":" + strconv.Itoa(config.ServerPort) + "in" + config.ChainNodeType
	return SignMsgWithSha256([]byte(source))
}

// get obd node id
func GetObdNodeId() string {
	source := "obd:" + GetMacAddrs() + ":" + strconv.Itoa(config.ServerPort)
	return SignMsgWithSha256([]byte(source)) + config.ChainNodeType
}

func GetCoreNet() *chaincfg.Params {
	chainNet := &chaincfg.MainNetParams
	if strings.Contains(config.ChainNodeType, "main") {
		chainNet = &chaincfg.MainNetParams
	}
	if strings.Contains(config.ChainNodeType, "test") {
		chainNet = &chaincfg.TestNet3Params
	}
	if strings.Contains(config.ChainNodeType, "reg") {
		chainNet = &chaincfg.RegressionNetParams
	}
	return chainNet
}

func GetBtcMinerAmount(total float64) float64 {
	out, _ := decimal.NewFromFloat(total).Div(decimal.NewFromFloat(4.0)).Sub(decimal.NewFromFloat(GetOmniDustBtc())).Round(8).Float64()
	return out
}

func GetOmniDustBtc() float64 {
	return 0.00000546
}

func GenerateInitHashCode() string {
	key, _ := btcec.NewPrivateKey(btcec.S256())
	wif, _ := btcutil.NewWIF(key, GetCoreNet(), true)
	pubKeyHash := btcutil.Hash160(wif.SerializePubKey())
	addr, _ := btcutil.NewAddressPubKeyHash(pubKeyHash, GetCoreNet())
	return addr.String()
}

var grpcSession string

func GetGRpcSession() string {
	if len(grpcSession) == 0 {
		grpcSession = GenerateInitHashCode()
	}
	return grpcSession
}
