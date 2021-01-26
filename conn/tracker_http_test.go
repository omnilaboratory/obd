package conn2tracker

import (
	"log"
	"testing"
)

func TestTestMemPoolAccept(t *testing.T) {
	hex := ""
	hex = "0200000001ea995e9f5ab22ca075fbd2f183b3fb41d1c88562a67a8cc9d280ecca7353f61b010000006a47304402200c850e330273bffa40bad407baa39dfe5595e6427eaab6b5190f99e4b9ac1d6c022015aa7385c04659dcb44f4e552769a64d5e91941260aed9e862b3dd6d17a20e48012102241c98bb5c64f34bd414f05e722924a6946569f817190bbe9752f1c0416b20dcffffffff02502400000000000017a9147297c8404e9dc908ce10fb1db3660736dd45f7ef87caa00b00000000001976a914928f34815d1a8f54afe239ad68391fcddb505a6588ac00000000"
	accept := TestMemPoolAccept(hex)
	log.Println(accept)
	transaction, err := SendRawTransaction(hex)
	log.Println(err)
	log.Println(transaction)
}
