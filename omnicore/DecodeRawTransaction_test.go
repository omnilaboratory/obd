package omnicore

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func TestDecodeRawTransaction(t *testing.T) {
	tx_hex := "0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff036a5a0d00000000001976a914c6734676a08e3c6438bd95fa62c57939c988a17b88ac0000000000000000166a146f6d6e690000000000000002000000000098968022020000000000001976a914ee692ea81da1b12d3dd8f53fd504865c9d843f5288ac00000000"
	fmt.Println(DecodeRawTransaction(tx_hex, &chaincfg.MainNetParams))
}

func TestVerfyOpreturnPayload(t *testing.T) {
	property_id_expected := "2"
	amount_expected := "0.1"
	devisible_expected := true

	// This testing hex equals to "OP_RETURN 6f6d6e6900000000000000020000000000989680",
	// which is encoded from payload of property id 2 and amount 0.1
	opreturn_hex := "6a146f6d6e6900000000000000020000000000989680"
	if VerfyOpreturnPayload(opreturn_hex, property_id_expected, amount_expected, devisible_expected) {
		fmt.Println("Match!")
	} else {
		fmt.Println("Please check opreturn string and expected data, and try again.")
	}

}
