package service

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
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
		AddInt64(3).
		AddOp(txscript.OP_EQUAL).
		AddOp(txscript.OP_2).
		AddData(apub).
		AddData(bpub).
		AddOp(txscript.OP_2).
		AddOp(txscript.OP_CHECKMULTISIG).
		Script()
	log.Println(hex.EncodeToString(bytes))
	decodeScript, _ := rpcClient.DecodeScript(hex.EncodeToString(bytes))
	log.Println(decodeScript)
	//create multisig address
	hash, err := btcutil.NewAddressScriptHash(bytes, &chaincfg.TestNet3Params)
	log.Println(hash)
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
	builder := txscript.NewScriptBuilder()

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

	pk0 := (*btcec.PublicKey)(&wif0.PrivKey.PublicKey).SerializeCompressed()
	addr0, err := btcutil.NewAddressPubKey(pk0, &chaincfg.TestNet3Params)
	pk1 := (*btcec.PublicKey)(&wif1.PrivKey.PublicKey).SerializeCompressed()
	addr1, err := btcutil.NewAddressPubKey(pk1, &chaincfg.TestNet3Params)

	//生成多签地址 得到redeemscript
	redeemScriptStr := "538752210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae"
	redeemScript, _ := hex.DecodeString(redeemScriptStr)
	log.Println(hex.EncodeToString(redeemScript))

	//得到多签地址
	address = "2N6h1Vz8Ate9zH2LvpBi5EQUy5b42bf9RGk"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	//得到输出脚本
	scriptPkScriptOutput, err := txscript.PayToAddrScript(addr)
	if err != nil {
	}

	log.Println(scriptPkScriptOutput)

	prevPkScriptHex := "a914937a069c34e2d88983f890999a4699170c11f4e587"
	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript
	// 3. 签名
	privKey := "cRaLt58h5xv5nfw9vqNJ6jtopsa5SWMhA7eCmhBuyTBviiZ8dx6N" // 私钥
	signMulti(tx, privKey, prevPkScripts)
	sig0 := tx.TxIn[0].SignatureScript
	privKey = "cV1Sot9e8pb2Ern5DCtbMRDfd74dFbRhM6dhxDtg7CmdEVu7CtyD" // 私钥
	signMulti(tx, privKey, prevPkScripts)
	sig1 := tx.TxIn[0].SignatureScript

	pkScriptUnlock, _ := builder.
		AddOp(txscript.OP_0).
		AddData(sig1).
		AddData(sig0).
		AddInt64(3).
		AddData(redeemScript).
		Script()

	sigScript, err := txscript.SignTxOutput(&chaincfg.TestNet3Params, tx, 0, scriptPkScriptOutput, txscript.SigHashAll, mkGetKey(map[string]addressToKey{
		addr0.EncodeAddress(): {wif0.PrivKey, true},
		addr1.EncodeAddress(): {wif1.PrivKey, true},
	}), mkGetScript(map[string][]byte{
		address: pkScriptUnlock,
	}), nil)
	if err != nil {
		log.Println(err)
	}
	log.Println(sigScript)
	tx.TxIn[0].SignatureScript = pkScriptUnlock

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
