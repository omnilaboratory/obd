package bean

import (
	"LightningOnOmni/bean/chainhash"
	"log"
	"testing"
)

func TestNewChanIDFromOutPoint(t *testing.T) {

	hash, _ := chainhash.NewHashFromStr("4c0e8d55056c5bd3cb4404acfcfb6e3b51127fb9131d9260892ef207c8ab4380")
	op := &OutPoint{
		Hash:  *hash,
		Index: 0,
	}
	log.Println(*hash)
	point := NewChanIDFromOutPoint(op)
	log.Println(string(point[:]))
}
