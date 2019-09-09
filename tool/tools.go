package tool

import (
	"crypto/sha256"
	"fmt"
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

func GetMinerFee() float64 {
	return 0.00002
}
