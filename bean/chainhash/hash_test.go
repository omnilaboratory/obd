package chainhash

import (
	"log"
	"testing"
)

func TestNewHashFromStr(t *testing.T) {
	str := "4c0e8d55056c5bd3cb4404acfcfb6e3b51127fb9131d9260892ef207c8ab4380"
	//str := "4"
	log.Println(len(str))
	hashes, e := NewHashFromStr(str)
	if e != nil {
		log.Println(e)
	}
	log.Println(*hashes)
}
