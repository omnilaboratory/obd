package conn2tracker

import (
	"log"
	"testing"
)

func TestTestMemPoolAccept(t *testing.T) {
	hex := ""
	hex = "0200000002c5c7eda485497b0836886cff8128becb5b98411faa7b79135c55be1668709104020000006b483045022100a55c4417d58e1b10cf090a6c12f685628a6a02326976f416c17613ed33cfc59a0220663aba3cab3a6f93796cde025afe9516295630aae600742cc9f9c3aed78aba24012103f07e518b9d2d3758ff504598cf484c65b612c27a8c755ce53663398db97addddffffffffdbc91d57b3cac1df8f6f25ed8f005d6d3ec09bfe86900b03089f6b51b2cd1edd010000006a47304402207ad7fb830773522b8b2bc636c4e20e27eb2aa5b0d98a323588d9b1379dfa629502206691e2d9fc7a4fbdecd8f071467853ae01caed7fc563444a801c79b9f432c555012103f07e518b9d2d3758ff504598cf484c65b612c27a8c755ce53663398db97addddffffffff0318c69a3b000000001976a9144c76521934cefc4f21f8373426747ba7f89c282488ac0000000000000000166a146f6d6e690000000080000004000000174876e800220200000000000017a9146645eeee46fbc3eb5740222de384122075c17a358700000000"
	//transaction := omnicore.DecodeRawTransaction(hex, &chaincfg.RegressionNetParams)

	transaction, err := OmniDecodeTransaction(hex)
	log.Println(err)
	log.Println(transaction)
	accept := TestMemPoolAccept(hex)
	log.Println(accept)
	//transaction, err := SendRawTransaction(hex)
	//log.Println(err)
	//log.Println(transaction)
}
func Test1(t *testing.T) {
	//log.Println(GetUserP2pNodeId("5773cc7b3b2fb80453ba82663b71992a91ecc8eb4b6b76fa6a60e42a6c913fa0"))
	address, err := GetBalanceByAddress("ms6kvv4RXqiQE53HbZ7NhaPmZpngb4XRiY")
	log.Println(address)
	log.Println(err)
}
