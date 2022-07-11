package omnicore

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/txscript"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/wire"
)

func VerifySignature(pubkey_hex string, signature_hex string, scriptSig string, signed_tx_hex string) bool {

	signed_tx := wire.NewMsgTx(wire.TxVersion)

	bytes_arr, _ := hex.DecodeString(signed_tx_hex)

	if err := signed_tx.Deserialize(bytes.NewReader(bytes_arr)); err != nil {
		fmt.Println("VerifySignature(...): error deserializing unsigned tx_hex")
	}
	// Decode hex-encoded serialized public key.
	pubKeyBytes, _ := hex.DecodeString(pubkey_hex)
	pubKey, _ := btcec.ParsePubKey(pubKeyBytes, btcec.S256())

	// Decode hex-encoded serialized signature.
	sig_bytes, _ := hex.DecodeString(signature_hex)
	signature, _ := btcec.ParseSignature(sig_bytes, btcec.S256())

	// Verify the signature for the message using the public key.

	messageHash, error := txscript.CalcSignatureHash([]byte(scriptSig), txscript.SigHashAll, signed_tx, 0)
	//messageHash, error := txscript.CalcSignatureHash([]byte(scriptSig_true.Hex), 1, &tx, 0)
	if error != nil {
		fmt.Println("VerifySignature(...)  can not calculate signature hash.")
		return false
	}

	return signature.Verify(messageHash, pubKey)

}

func VerifySignatureFromTxHex(unsigned_tx_hex string, signed_tx_hex string) bool {
	unsigned_tx := wire.NewMsgTx(wire.TxVersion)

	u_bytes_arr, _ := hex.DecodeString(unsigned_tx_hex)

	if err := unsigned_tx.Deserialize(bytes.NewReader(u_bytes_arr)); err != nil {
		fmt.Println("VerifySignatureFromTxHex(...): error deserializing unsigned tx_hex")
	}

	signed_tx := wire.NewMsgTx(wire.TxVersion)

	bytes_arr, _ := hex.DecodeString(signed_tx_hex)

	if err := signed_tx.Deserialize(bytes.NewReader(bytes_arr)); err != nil {
		fmt.Println("VerifySignatureFromTxHex(...): error deserializing signed tx_hex")
	}

	// Prove that the transaction has been validly signed by executing the
	// script pair.
	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
		txscript.ScriptStrictMultiSig |
		txscript.ScriptDiscourageUpgradableNops
	vm, err := txscript.NewEngine(unsigned_tx.TxOut[0].PkScript, signed_tx, 0,
		flags, nil, nil, -1)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if err := vm.Execute(); err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("Transaction successfully signed")

	return true
}
