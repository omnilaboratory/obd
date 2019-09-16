package tool

import (
	"fmt"
	"log"
	"testing"
)

func TestGetAddress(t *testing.T) {
	//GetAddressFromPubKey("03870f2aebd7ac762bf26de14bf4624781cd4e4ed3ca4ada16c883f1d7a492ec0a")

	s, _ := RandBytes(16)
	log.Println(s)

	SetAesAndIV("", "03a492b58a0df0a33119d308e263f676"[:16])
	se, err := AesEncrypt("aes-20170416-30-1000")
	fmt.Println(se, err)
	sd, err := AesDecrypt2(se)
	fmt.Println(sd)
}
