package omnicore

import (
	"btcd/btcd/txscript"
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

func CreateMultiSigAddr(addr1_pubkey_str string, addr2_pubkey_str string, defaultNet *chaincfg.Params) (string, string) {

	addr1_pubkey_byte_arr, _ := hex.DecodeString(addr1_pubkey_str)
	addr2_pubkey_byte_arr, _ := hex.DecodeString(addr2_pubkey_str)

	address1_pubkey, _ := btcutil.NewAddressPubKey(addr1_pubkey_byte_arr, defaultNet)
	address2_pubkey, _ := btcutil.NewAddressPubKey(addr2_pubkey_byte_arr, defaultNet)
	//fmt.Println(address1_pubkey.EncodeAddress())
	//fmt.Println(address2_pubkey.EncodeAddress())

	pkScript, _ := txscript.MultiSigScript([]*btcutil.AddressPubKey{address1_pubkey, address2_pubkey}, 2)

	scriptAddr, _ := btcutil.NewAddressScriptHash(pkScript, defaultNet)

	return scriptAddr.EncodeAddress(), hex.EncodeToString(pkScript)
	//scriptPkScript, _ := txscript.PayToAddrScript(scriptAddr)

	//return hex.EncodeToString(scriptPkScript)
}
