package service

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/golangcrypto/ripemd160"
	"log"
)

//{
//"index":0,
//"address":"mneg4WpS1QmZcG73AjPK3mhhkUoKHHrxnf",
//"pub_key":"0316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af",
//"wif":"cRaLt58h5xv5nfw9vqNJ6jtopsa5SWMhA7eCmhBuyTBviiZ8dx6N",
//}
//{
//"index":1,
//"address":"mhs47wGp27XkY4Xv1ZYz7bE2r3wPEhyAKg",
//"pub_key":"02c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c8",
//"wif":"cV1Sot9e8pb2Ern5DCtbMRDfd74dFbRhM6dhxDtg7CmdEVu7CtyD",
//}

//构建一个自定义的多签地址，用于锁住输入
func CreateCustomMuiltAddress() {
	pubA := "0316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af"
	pubB := "02c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c8"
	apub, err := hex.DecodeString(pubA)
	log.Println(err)
	bpub, err := hex.DecodeString(pubB)
	log.Println(err)
	bytes, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_2).
		AddData(apub).
		AddData(bpub).
		AddOp(txscript.OP_2).
		AddOp(txscript.OP_CHECKMULTISIG).
		Script()
	log.Println(hex.EncodeToString(bytes))
	decodeScript, err := rpcClient.DecodeScript(hex.EncodeToString(bytes))
	log.Println(err)
	log.Println(decodeScript)
	//create multisig address
	hash, err := btcutil.NewAddressScriptHash(bytes, &chaincfg.TestNet3Params)
	log.Println(hash)

	result, err := rpcClient.DecodeScript("5141042f90074d7a5bf30c72cf3a8dfd1381bdbd30407010e878f3a11269d5f74a58788505cdca22ea6eab7cfb40dc0e07aba200424ab0d79122a653ad0c7ec9896bdf51ae")
	log.Println(result)
	//2N6h1Vz8Ate9zH2LvpBi5EQUy5b42bf9RGk
	//	538752210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae
}

func CreateCustomP2SHTx() {
	tx := wire.NewMsgTx(2)

	listUnspent, _ := rpcClient.ListUnspent("mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM")
	log.Println(listUnspent)

	utxoHash, _ := chainhash.NewHashFromStr("0a5768d06d1852776d3c9fb91d4c2c97cde95a1f30f506d0a951a28fb8e764de")
	point := wire.OutPoint{Hash: *utxoHash, Index: 0}
	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil
	tx.AddTxIn(wire.NewTxIn(&point, nil, nil))

	//一共有0.01000000，也即：1000000  给新的地址5000 给自己留下 80000 3000 miner fee
	amount := 1000000
	changeAddr := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"

	// 找零
	addr, _ := btcutil.DecodeAddress(changeAddr, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	tx.AddTxOut(wire.NewTxOut(int64(amount-3000), pkScript))
	result, _ := rpcClient.DecodeScript(hex.EncodeToString(pkScript))
	log.Println(result)

	//支出
	address := "2N6h1Vz8Ate9zH2LvpBi5EQUy5b42bf9RGk"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pkScript, _ = txscript.PayToAddrScript(addr)

	result, _ = rpcClient.DecodeScript(hex.EncodeToString(pkScript))
	log.Println(result)
	tx.AddTxOut(wire.NewTxOut(2000, pkScript))

	prevPkScriptHex := "76a9145d48e1d03e4f8c690bf43d97d25f68ef6f36896d88ac"
	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript
	// 3. 签名
	privKey := "cRuSwcDrc1gwoeaCvr4qFR9sgHB8wFxHzeqW1Bueo885S6RSYYxH" // 私钥
	sign(tx, privKey, prevPkScripts)
	printTx(tx)
	txHex, _ := getTxHex(tx)
	log.Println(txHex)

	result, err := rpcClient.DecodeRawTransaction(txHex)
	log.Println(err)
	log.Println(result)

	result, err = rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//0804696c6d0b9e10120df8cfec9909ef4223829dc50bd9ea04c02b111c7a6b77
}

func CreateCustomP2SHSpendTx() {
	tx := wire.NewMsgTx(2)

	//pubA := "0316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af"
	//pubB := "02c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c8"
	//apub, err := hex.DecodeString(pubA)
	//log.Println(err)
	//bpub, err := hex.DecodeString(pubB)
	//log.Println(err)
	utxoHash, _ := chainhash.NewHashFromStr("0804696c6d0b9e10120df8cfec9909ef4223829dc50bd9ea04c02b111c7a6b77")
	point := wire.OutPoint{Hash: *utxoHash, Index: 1}
	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil
	tx.AddTxIn(wire.NewTxIn(&point, nil, nil))

	amount := 2000
	// 支付给对方
	changeAddr := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"
	addr, _ := btcutil.DecodeAddress(changeAddr, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	tx.AddTxOut(wire.NewTxOut(int64(500), pkScript))

	//找零
	address := "2N6h1Vz8Ate9zH2LvpBi5EQUy5b42bf9RGk"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pkScript, _ = txscript.PayToAddrScript(addr)
	tx.AddTxOut(wire.NewTxOut(int64(amount-500-500), pkScript))

	wif0, err := btcutil.DecodeWIF("cRaLt58h5xv5nfw9vqNJ6jtopsa5SWMhA7eCmhBuyTBviiZ8dx6N")
	wif1, err := btcutil.DecodeWIF("cV1Sot9e8pb2Ern5DCtbMRDfd74dFbRhM6dhxDtg7CmdEVu7CtyD")

	localKey0 := (*btcec.PublicKey)(&wif0.PrivKey.PublicKey)
	localKey1 := (*btcec.PublicKey)(&wif1.PrivKey.PublicKey)
	pk0 := (*btcec.PublicKey)(&wif0.PrivKey.PublicKey).SerializeCompressed()
	addr0, err := btcutil.NewAddressPubKey(pk0, &chaincfg.TestNet3Params)
	pk1 := (*btcec.PublicKey)(&wif1.PrivKey.PublicKey).SerializeCompressed()
	addr1, err := btcutil.NewAddressPubKey(pk1, &chaincfg.TestNet3Params)

	//生成多签地址 得到redeemscript
	redeemScriptStr := "538752210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae"
	redeemScript, _ := hex.DecodeString(redeemScriptStr)
	log.Println(hex.EncodeToString(redeemScript))

	//得到多签地址
	address = "2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	//得到输出脚本
	scriptPkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
	}
	log.Println(hex.EncodeToString(scriptPkScript))

	prevPkScriptHex := "a914937a069c34e2d88983f890999a4699170c11f4e587"
	scriptPkScript, _ = hex.DecodeString(prevPkScriptHex)

	testHdSeed := chainhash.Hash{
		0xb7, 0x94, 0x38, 0x5f, 0x2d, 0x1e, 0xf7, 0xab,
		0x4d, 0x92, 0x73, 0xd1, 0x90, 0x63, 0x81, 0xb4,
		0x4f, 0x2f, 0x6f, 0x25, 0x88, 0xa3, 0xef, 0xb9,
		0x6a, 0x49, 0x18, 0x83, 0x31, 0x98, 0x47, 0x53,
	}
	revokePreimage := testHdSeed.CloneBytes()
	paymentPreimage := revokePreimage
	paymentPreimage[0] ^= 1
	paymentHash := sha256.Sum256(paymentPreimage[:])

	_, commitPoint := btcec.PrivKeyFromBytes(btcec.S256(), revokePreimage)
	revocationKey := DeriveRevocationPubkey(localKey1, commitPoint)

	htlcWitnessScript, err := SenderHTLCScript(localKey0, localKey1, revocationKey, paymentHash[:])

	htlcPkScript, err := WitnessScriptHash(htlcWitnessScript)
	if err != nil {
		log.Println("unable to create p2wsh htlc script: ", err)
	}

	sigScript, err := txscript.SignTxOutput(&chaincfg.TestNet3Params, tx, 0, scriptPkScript, txscript.SigHashAll, mkGetKey(map[string]addressToKey{
		addr0.EncodeAddress(): {wif0.PrivKey, true},
		addr1.EncodeAddress(): {wif1.PrivKey, true},
	}), mkGetScript(map[string][]byte{
		address: htlcPkScript,
	}), nil)
	if err != nil {
		log.Println(err)
	}
	log.Println(sigScript)
	tx.TxIn[0].SignatureScript = sigScript

	txHex, _ := getTxHex(tx)
	log.Println(txHex)
	result, err := rpcClient.DecodeRawTransaction(txHex)
	log.Println(err)
	log.Println(result)

	result, err = rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	if err == nil {
		log.Println(result)
	}
	//f33cf515adcf55c5b39200d06e95b0a31c49797340e41689a5ae9da9284d7855
}

func SenderHTLCScript(senderHtlcKey, receiverHtlcKey, revocationKey *btcec.PublicKey, paymentHash []byte) ([]byte, error) {

	builder := txscript.NewScriptBuilder()

	// The opening operations are used to determine if this is the receiver
	// of the HTLC attempting to sweep all the funds due to a contract
	// breach. In this case, they'll place the revocation key at the top of
	// the stack.
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(btcutil.Hash160(revocationKey.SerializeCompressed()))
	builder.AddOp(txscript.OP_EQUAL)

	// If the hash matches, then this is the revocation clause. The output
	// can be spent if the check sig operation passes.
	builder.AddOp(txscript.OP_IF)
	builder.AddOp(txscript.OP_CHECKSIG)

	// Otherwise, this may either be the receiver of the HTLC claiming with
	// the pre-image, or the sender of the HTLC sweeping the output after
	// it has timed out.
	builder.AddOp(txscript.OP_ELSE)

	// We'll do a bit of set up by pushing the receiver's key on the top of
	// the stack. This will be needed later if we decide that this is the
	// sender activating the time out clause with the HTLC timeout
	// transaction.
	builder.AddData(receiverHtlcKey.SerializeCompressed())

	// Atm, the top item of the stack is the receiverKey's so we use a swap
	// to expose what is either the payment pre-image or a signature.
	builder.AddOp(txscript.OP_SWAP)

	// With the top item swapped, check if it's 32 bytes. If so, then this
	// *may* be the payment pre-image.
	builder.AddOp(txscript.OP_SIZE)
	builder.AddInt64(32)
	builder.AddOp(txscript.OP_EQUAL)

	// If it isn't then this might be the sender of the HTLC activating the
	// time out clause.
	builder.AddOp(txscript.OP_NOTIF)

	// We'll drop the OP_IF return value off the top of the stack so we can
	// reconstruct the multi-sig script used as an off-chain covenant. If
	// two valid signatures are provided, ten then output will be deemed as
	// spendable.
	builder.AddOp(txscript.OP_DROP)
	builder.AddOp(txscript.OP_2)
	builder.AddOp(txscript.OP_SWAP)
	builder.AddData(senderHtlcKey.SerializeCompressed())
	builder.AddOp(txscript.OP_2)
	builder.AddOp(txscript.OP_CHECKMULTISIG)

	// Otherwise, then the only other case is that this is the receiver of
	// the HTLC sweeping it on-chain with the payment pre-image.
	builder.AddOp(txscript.OP_ELSE)

	// Hash the top item of the stack and compare it with the hash160 of
	// the payment hash, which is already the sha256 of the payment
	// pre-image. By using this little trick we're able save space on-chain
	// as the witness includes a 20-byte hash rather than a 32-byte hash.
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(Ripemd160H(paymentHash))
	builder.AddOp(txscript.OP_EQUALVERIFY)

	// This checks the receiver's signature so that a third party with
	// knowledge of the payment preimage still cannot steal the output.
	builder.AddOp(txscript.OP_CHECKSIG)

	// Close out the OP_IF statement above.
	builder.AddOp(txscript.OP_ENDIF)

	// Close out the OP_IF statement at the top of the script.
	builder.AddOp(txscript.OP_ENDIF)

	return builder.Script()
}

func Ripemd160H(d []byte) []byte {
	h := ripemd160.New()
	h.Write(d)
	return h.Sum(nil)
}

func DeriveRevocationPubkey(revokeBase, commitPoint *btcec.PublicKey) *btcec.PublicKey {

	// R = revokeBase * sha256(revocationBase || commitPoint)
	revokeTweakBytes := SingleTweakBytes(revokeBase, commitPoint)
	rX, rY := btcec.S256().ScalarMult(revokeBase.X, revokeBase.Y,
		revokeTweakBytes)

	// C = commitPoint * sha256(commitPoint || revocationBase)
	commitTweakBytes := SingleTweakBytes(commitPoint, revokeBase)
	cX, cY := btcec.S256().ScalarMult(commitPoint.X, commitPoint.Y,
		commitTweakBytes)

	// Now that we have the revocation point, we add this to their commitment
	// public key in order to obtain the revocation public key.
	//
	// P = R + C
	revX, revY := btcec.S256().Add(rX, rY, cX, cY)
	return &btcec.PublicKey{
		X:     revX,
		Y:     revY,
		Curve: btcec.S256(),
	}
}
func SingleTweakBytes(commitPoint, basePoint *btcec.PublicKey) []byte {
	h := sha256.New()
	h.Write(commitPoint.SerializeCompressed())
	h.Write(basePoint.SerializeCompressed())
	return h.Sum(nil)
}
