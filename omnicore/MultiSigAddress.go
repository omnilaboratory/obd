package omnicore

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

func CreateMultiSigAddr(addr1_pubkey_str string, addr2_pubkey_str string, defaultNet *chaincfg.Params) (string, string, string) {

	addr1_pubkey_byte_arr, _ := hex.DecodeString(addr1_pubkey_str)
	addr2_pubkey_byte_arr, _ := hex.DecodeString(addr2_pubkey_str)

	address1_pubkey, err := btcutil.NewAddressPubKey(addr1_pubkey_byte_arr, defaultNet)
	if err != nil {
		return "", "", ""
	}
	address2_pubkey, err := btcutil.NewAddressPubKey(addr2_pubkey_byte_arr, defaultNet)
	if err != nil {
		return "", "", ""
	}
	pkScript, err := txscript.MultiSigScript([]*btcutil.AddressPubKey{address1_pubkey, address2_pubkey}, 2)
	if err != nil {
		return "", "", ""
	}
	scriptAddr, _ := btcutil.NewAddressScriptHash(pkScript, defaultNet)

	scriptPubKey := hex.EncodeToString(scriptAddr.ScriptAddress())
	scriptPubKey = "a914" + scriptPubKey + "87"

	return scriptAddr.EncodeAddress(), hex.EncodeToString(pkScript), scriptPubKey

}
