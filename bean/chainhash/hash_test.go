package chainhash

import (
	"log"
	"testing"
)

func TestNewHashFromStr(t *testing.T) {
	str := "2021222324"
	log.Println(len(str))
	log.Println([]byte(str))
	hashes, e := NewHashFromStr(str)
	if e != nil {
		log.Println(e)
	}
	log.Println(hashes)
}
