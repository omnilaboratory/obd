package tool

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/omnilaboratory/obd/config"
	"golang.org/x/crypto/ripemd160"
	"io"
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
		return "", errors.New("invalid pubkey")
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
		return "", errors.New("invalid pubkey")
	}
	netAddr.SetFormat(btcutil.PKFCompressed)
	address = netAddr.EncodeAddress()
	log.Println("BitCoin Address:", address)

	return address, nil
}

func GetPubKeyFromWifAndCheck(privKeyHex string, pubKey string) (pubKeyFromWif string, err error) {
	if CheckIsString(&privKeyHex) == false {
		return "", errors.New("wrong private key")
	}
	if CheckIsString(&pubKey) == false {
		return "", errors.New("wrong pubKey")
	}

	wif, err := btcutil.DecodeWIF(privKeyHex)
	if err != nil {
		return "", errors.New("wrong private key")
	}
	pubKeyFromWif = hex.EncodeToString(wif.PrivKey.PubKey().SerializeCompressed())
	if pubKeyFromWif != pubKey {
		return "", errors.New("private key and pubkey are not parent relationship")
	}
	return pubKeyFromWif, nil
}

func RandBytes(size int) (string, error) {
	arr := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, arr); err != nil {
		return "", err
	}
	log.Println(arr)
	return base64.StdEncoding.EncodeToString(arr), nil
}

// 判断文件夹是否存在,如果不存在，则创建
func PathExistsAndCreate(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		return nil
	}
	return errors.New("fail to create")
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

// get obd node id
func GetObdNodeId() string {
	source := GetMacAddrs() + ":" + strconv.Itoa(config.ServerPort)
	return SignMsgWithSha256([]byte(source))
}

// get obd node id
func GetTrackerNodeId() string {
	source := GetMacAddrs() + ":" + strconv.Itoa(config.TrackerServerPort)
	return SignMsgWithSha256([]byte(source))
}

func GetIPs() (ips string) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("fail to get net interface addrs: %v", err)
		return ips
	}
	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ips
}
