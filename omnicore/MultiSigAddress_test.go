package omnicore

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func TestCreateMultiSigAddr(t *testing.T) {
	addr1_pubkey := "02db204a46ab45d872a420fef8abd935f2fe5347684f43ea58599c430f80aa82e5"
	addr2_pubkey := "02c716d071a76cbf0d29c29cacfec76e0ef8116b37389fb7a3e76d6d32cf59f4d3"

	//this pubkey produces error.
	//addr3_pubkey := "03073d3cf516dceeffaa53a84059fb8701ff5e291b9537457137be851bbc4e5525"

	fmt.Println("2-2 multisig address: ")
	address, redeemScriptHex := CreateMultiSigAddr(addr1_pubkey, addr2_pubkey, &chaincfg.MainNetParams)
	fmt.Println(address)
	fmt.Println("and pubkey script hex is:")
	fmt.Println(redeemScriptHex)

}
