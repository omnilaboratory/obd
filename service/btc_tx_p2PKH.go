package service

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"log"
)

// 解密 pay-to-pubkey-hash puzzle https://en.bitcoin.it/wiki/Script#Standard_Transaction_to_Bitcoin_address_.28pay-to-pubkey-hash.29

//新的地址 用addrtool_mnemonic_test.go的Test_Demo2生成，index = 1
//{"index":1,"address":"n2Lo7zwvek2NHmvHSD85WARKSfYwJxofve","pub_key":"0288d87affee96b8fd0fc5eb6fa4f4dae0c2e69145deda8d0768a3c2aa3a77a0b5","wif":"cVEQD3Wm9pdpmAjHz3AuA5uGqZVgEt2kKbVrwUwRTsyLx9z12KvT"}
func CreateP2PKHTx() {
	tx := wire.NewMsgTx(2)

	utxoHash, _ := chainhash.NewHashFromStr("8b2d1636c0fa3825428a651f93e38ee4b862fa3bcc7cac2e18d957edf41ab7ae")
	point := wire.OutPoint{Hash: *utxoHash, Index: 0}
	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil
	tx.AddTxIn(wire.NewTxIn(&point, nil, nil))

	//一共有0.00100000，也即：89000  给新的地址5000 给自己留下 80000 3000 miner fee
	//一共有0.00100000，也即：80000  给新的地址5000 给自己留下 73000 3000 miner fee
	//一共有0.00100000，也即：73000  给新的地址5000 给自己留下 70000 1000 miner fee
	//一共有0.00100000，也即：70000  给新的地址2000 给自己留下 68000 1000 miner fee
	//一共有0.00100000，也即：66000  给新的地址2000 给自己留下 63000 1000 miner fee
	changeAddr := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"
	// 1.1 输出1, 给自己转剩下的钱
	addr, _ := btcutil.DecodeAddress(changeAddr, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	tx.AddTxOut(wire.NewTxOut(63000, pkScript))

	//第二个地址 "wif":"cVEQD3Wm9pdpmAjHz3AuA5uGqZVgEt2kKbVrwUwRTsyLx9z12KvT"
	address := "n2Lo7zwvek2NHmvHSD85WARKSfYwJxofve"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pubKeyHash := addr.ScriptAddress()

	lock, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
	//0.0002
	tx.AddTxOut(wire.NewTxOut(2000, lock))

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
	result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//	c698824479271d4355d36f8379b15f792acfb22cd6922ab5fb21ef7b96728631
}

func CreateP2PKHSpendTx() {
	tx := wire.NewMsgTx(2)

	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil
	utxoHash, _ := chainhash.NewHashFromStr("c698824479271d4355d36f8379b15f792acfb22cd6922ab5fb21ef7b96728631")
	point := wire.OutPoint{Hash: *utxoHash, Index: 1}
	tx.AddTxIn(wire.NewTxIn(&point, nil, nil))

	//第一个输出
	address := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"
	addr, _ := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pubKeyHash := addr.ScriptAddress()
	lock, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
	tx.AddTxOut(wire.NewTxOut(1500, lock))

	prevPkScriptHex := "76a914e46ed306720c40a495797f02c2ae15852e8cec4088ac"
	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript
	// 3. 签名
	privKey := "cVEQD3Wm9pdpmAjHz3AuA5uGqZVgEt2kKbVrwUwRTsyLx9z12KvT" // 私钥
	sign(tx, privKey, prevPkScripts)

	txHex, _ := getTxHex(tx)
	log.Println(txHex)
	result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//	e701c7ac8c8036e6c9fb0517b1f1416712a38380bac9f89c88a39ba14b115042
}
