package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"testing"
)

func TestDemo4(t *testing.T) {
	path := AStarPathFind{}
	path.initData(10, 10)
	path.drawMap(path.road)
	path.findPath(1, 100)
	if len(path.road) == 0 {
		log.Println("no way")
	} else {
		path.drawMap(path.road)
		log.Println(path.road)
	}
}

func TestDemo3(t *testing.T) {
	se, err := AesEncrypt("aes-20170416-30-1000", "abc")
	fmt.Println(se, err)
	sd, err := AesDecrypt2(se, "abc")
	fmt.Println(sd)
}
func TestDemo2(t *testing.T) {
	format := VerifyEmailFormat("254698@163.com")
	log.Println(format)
}

func TestGetAddress(t *testing.T) {
	//GetAddressFromPubKey("03870f2aebd7ac762bf26de14bf4624781cd4e4ed3ca4ada16c883f1d7a492ec0a")

	msg := "03c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee961"
	hash := sha256.New()
	hash.Write([]byte(msg))
	aa := fmt.Sprintf("%x", hash.Sum(nil))
	log.Println(aa)
	return
}

func TestDemo1(t *testing.T) {
	//msg :="htlctestingstring"
	//withSha256 := SignMsgWithSha256([]byte(msg))
	//log.Println(withSha256)
	//return

	s, _ := RandBytes(32)
	log.Println(s)
	log.Println(hex.EncodeToString([]byte(s)))
	temp := append([]byte("03c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee961"), s...)
	log.Println(hex.EncodeToString(temp))
	name := "alice"
	temp = append(temp, name...)
	log.Println(temp)
	log.Println(hex.EncodeToString(temp))
	ripemd160 := SignMsgWithRipemd160(temp)
	log.Println(ripemd160)
	log.Println(SignMsgWithSha256([]byte(ripemd160)))
	temp = append([]byte("03c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee962"), s...)
	log.Println(hex.EncodeToString(temp))
	temp = append(temp, name...)
	log.Println(temp)
	log.Println(hex.EncodeToString(temp))
	ripemd160 = SignMsgWithRipemd160(temp)
	log.Println(ripemd160)
	log.Println(SignMsgWithSha256([]byte(ripemd160)))

	//msg := "03c57bea53afd7c3c2d75653ca35ca968c8e9610b6448f822cfb006730870ee961"
	//publicSHA256 := sha256.Sum256([]byte(msg))
	//log.Println(publicSHA256)
	//log.Println(hex.EncodeToString(publicSHA256[:]))
	//hash := ripemd160.New()
	//n, err := hash.Write(publicSHA256[:])
	//log.Println(err)
	//log.Println(n)
	//sum := hash.Sum(nil)
	//log.Println(sum)
	//sprintf := fmt.Sprintf("%x", hash.Sum(nil))
	//log.Println(sprintf)
	//log.Println(hex.DecodeString(sprintf))
	//log.Println(hex.EncodeToString(sum))
}
