package bean

import (
	"log"
	"testing"
)

func TestShortChannelID_ToUint64(t *testing.T) {
	chanId := ShortChannelID{
		BlockHeight: 0xffffffFF,
		TxIndex:     0xffff,
		TxPosition:  0xffff,
	}
	log.Println(int16(0x7fff))
	log.Println(chanId.String())
	log.Println(chanId.ToUint64())
}
