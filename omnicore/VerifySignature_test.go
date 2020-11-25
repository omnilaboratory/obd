package omnicore

import (
	"encoding/hex"
	"fmt"

	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func TestVerifySignature(t *testing.T) {
	// Testing data is from the issue: https://github.com/btcsuite/btcd/issues/1309
	//But somehow it doesn't work.

	tx_hex := "01000000013dcd7d87904c9cb7f4b79f36b5a03f96e2e729284c09856238d5353e1182b00200000000fd5d01004730440220762ce7bca626942975bfd5b130ed3470b9f538eb2ac120c2043b445709369628022051d73c80328b543f744aa64b7e9ebefa7ade3e5c716eab4a09b408d2c307ccd701483045022100abf740b58d79cab000f8b0d328c2fff7eb88933971d1b63f8b99e89ca3f2dae602203354770db3cc2623349c87dea7a50cee1f78753141a5052b2d58aeb592bcf50f014cc9524104a882d414e478039cd5b52a92ffb13dd5e6bd4515497439dffd691a0f12af9575fa349b5694ed3155b136f09e63975a1700c9f4d4df849323dac06cf3bd6458cd41046ce31db9bdd543e72fe3039a1f1c047dab87037c36a669ff90e28da1848f640de68c2fe913d363a51154a0c62d7adea1b822d05035077418267b1a1379790187410411ffd36c70776538d079fbae117dc38effafb33304af83ce4894589747aee1ef992f63280567f52f5ba870678b4ab4ff6c8ea600bd217870a8b4f1f09f3a8e8353aeffffffff0130d90000000000001976a914569076ba39fc4ff6a2291d9ea9196d8c08f9c7ab88ac00000000"

	fmt.Println(DecodeRawTransaction(tx_hex, &chaincfg.MainNetParams))

	pubkey_hex := "04a882d414e478039cd5b52a92ffb13dd5e6bd4515497439dffd691a0f12af9575fa349b5694ed3155b136f09e63975a1700c9f4d4df849323dac06cf3bd6458cd"
	signature_hex := "30440220762ce7bca626942975bfd5b130ed3470b9f538eb2ac120c2043b445709369628022051d73c80328b543f744aa64b7e9ebefa7ade3e5c716eab4a09b408d2c307ccd701"
	scriptSig := "47304402204e63d034c6074f17e9c5f8766bc7b5468a0dce5b69578bd08554e8f21434c58e0220763c6966f47c39068c8dcd3f3dbd8e2a4ea13ac9e9c899ca1fbc00e2558cbb8b01410431393af9984375830971ab5d3094c6a7d02db3568b2b06212a7090094549701bbb9e84d9477451acc42638963635899ce91bacb451a1bb6da73ddfbcf596bddf"
	if VerifySignature(pubkey_hex, signature_hex, scriptSig, tx_hex) {
		fmt.Println("Signature is correct!")
	} else {
		fmt.Println("Wrong signature!")
	}

}

func TestVerifySignatureFromTxHex(t *testing.T) {

	// txs are created by the example in TestSignature(t *testing.T)
	unsiged_tx := "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff020000ffffffff0100e1f505000000001976a9143dee47716e3cfa57df45113473a6312ebeaef31188ac00000000"
	siged_tx := "01000000013e7b20e5029f0a4edc65dcfaa0ba145a4c8ccdc522de366515fbddbbd2609f34000000006b483045022100ed3969fcdae3d50ead3e4ba5d985335ecaf1ce497870285a619c82189a90e8d602204133c58290b0263e5a2fee5c2c05e2588fcf1b23637a7d2ff240f6e1abfb830f012102a673638cb9587cb68ea08dbef685c6f2d2a751a8b3c6f2a7e9a4999e6e4bfaf5ffffffff0100000000000000000000000000"

	if VerifySignatureFromTxHex(unsiged_tx, siged_tx) {
		fmt.Println("Signature is correct!")
	} else {
		fmt.Println("Wrong signature!")
	}
}

/*
 * https://github.com/btcsuite/btcd/blob/master/txscript/example_test.go
 * ExampleSignTxOutput()
 * This is an example from above link.
 */
func TestSignature(t *testing.T) {

	// Ordinarily the private key would come from whatever storage mechanism
	// is being used, but for this example just hard code it.
	privKeyBytes, err := hex.DecodeString("22a47fa09a223f2aa079edf85a7c2" +
		"d4f8720ee63e502ee2869afab7de234b80c")
	if err != nil {
		fmt.Println(err)
		return
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
	addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash,
		&chaincfg.MainNetParams)
	if err != nil {
		fmt.Println(err)
		return
	}

	// For this example, create a fake transaction that represents what
	// would ordinarily be the real transaction that is being spent.  It
	// contains a single output that pays to address in the amount of 1 BTC.
	originTx := wire.NewMsgTx(wire.TxVersion)
	prevOut := wire.NewOutPoint(&chainhash.Hash{}, ^uint32(0))
	txIn := wire.NewTxIn(prevOut, []byte{txscript.OP_0, txscript.OP_0}, nil)
	originTx.AddTxIn(txIn)
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	txOut := wire.NewTxOut(100000000, pkScript)
	originTx.AddTxOut(txOut)
	originTxHash := originTx.TxHash()

	fmt.Println("unsigned tx is: ")
	fmt.Println(TxToHex(originTx))
	fmt.Println("got it.")

	// Create the transaction to redeem the fake transaction.
	redeemTx := wire.NewMsgTx(wire.TxVersion)

	// Add the input(s) the redeeming transaction will spend.  There is no
	// signature script at this point since it hasn't been created or signed
	// yet, hence nil is provided for it.
	prevOut = wire.NewOutPoint(&originTxHash, 0)
	txIn = wire.NewTxIn(prevOut, nil, nil)
	redeemTx.AddTxIn(txIn)

	// Ordinarily this would contain that actual destination of the funds,
	// but for this example don't bother.
	txOut = wire.NewTxOut(0, nil)
	redeemTx.AddTxOut(txOut)

	// Sign the redeeming transaction.
	lookupKey := func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
		// Ordinarily this function would involve looking up the private
		// key for the provided address, but since the only thing being
		// signed in this example uses the address associated with the
		// private key from above, simply return it with the compressed
		// flag set since the address is using the associated compressed
		// public key.
		//
		// NOTE: If you want to prove the code is actually signing the
		// transaction properly, uncomment the following line which
		// intentionally returns an invalid key to sign with, which in
		// turn will result in a failure during the script execution
		// when verifying the signature.
		//
		// privKey.D.SetInt64(12345)
		//
		return privKey, true, nil
	}
	// Notice that the script database parameter is nil here since it isn't
	// used.  It must be specified when pay-to-script-hash transactions are
	// being signed.
	sigScript, err := txscript.SignTxOutput(&chaincfg.MainNetParams,
		redeemTx, 0, originTx.TxOut[0].PkScript, txscript.SigHashAll,
		txscript.KeyClosure(lookupKey), nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	redeemTx.TxIn[0].SignatureScript = sigScript

	fmt.Println("signed tx is: ")
	fmt.Println(TxToHex(redeemTx))
	fmt.Println("got it.")

	// Prove that the transaction has been validly signed by executing the
	// script pair.
	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
		txscript.ScriptStrictMultiSig |
		txscript.ScriptDiscourageUpgradableNops
	vm, err := txscript.NewEngine(originTx.TxOut[0].PkScript, redeemTx, 0,
		flags, nil, nil, -1)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := vm.Execute(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Transaction successfully signed")

	// Output:
	// Transaction successfully signed
}
