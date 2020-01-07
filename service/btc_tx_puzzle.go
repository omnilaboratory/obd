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

// 解密 Transaction puzzle https://en.bitcoin.it/wiki/Script#Transaction_puzzle
func CreateCustomTx() {
	tx := wire.NewMsgTx(2)

	utxoHash, _ := chainhash.NewHashFromStr("5b593a052edd29f1e1aa94bb8493d0a6bf51385098a7af84660f4115274323af")
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

	//第二个地址
	address := "mtRoRNpVYhMRYPjoz8u9Eqnmm5LqyDzXgh"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	//pubKeyHash := addr.ScriptAddress()

	lock, _ := txscript.NewScriptBuilder().
		AddInt64(1).
		AddOp(txscript.OP_EQUAL).
		Script()
	//0.0002
	tx.AddTxOut(wire.NewTxOut(2000, lock))

	// 3. 签名
	prevPkScriptHex := "76a9145d48e1d03e4f8c690bf43d97d25f68ef6f36896d88ac"
	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript

	privKey := "cRuSwcDrc1gwoeaCvr4qFR9sgHB8wFxHzeqW1Bueo885S6RSYYxH" // 私钥
	sign(tx, privKey, prevPkScripts)
	printTx(tx)
	txHex, _ := getTxHex(tx)
	result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//	0979f9bea38d2e9b425147ed16585567d85b3f4b4e682044bbeb593e0f6c718d
}

func CreateCustomSpendTx() {
	tx := wire.NewMsgTx(2)

	utxoHash, _ := chainhash.NewHashFromStr("0979f9bea38d2e9b425147ed16585567d85b3f4b4e682044bbeb593e0f6c718d")
	point := wire.OutPoint{Hash: *utxoHash, Index: 1}
	unlock, _ := txscript.NewScriptBuilder().
		AddInt64(1).
		Script()
	tx.AddTxIn(wire.NewTxIn(&point, unlock, nil))

	//第二个地址
	address := "mtRoRNpVYhMRYPjoz8u9Eqnmm5LqyDzXgh"
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

	printTx(tx)
	txHex, _ := getTxHex(tx)
	result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//	67a7903399d70b6898a1e001ede096e3a3645d79a52447a7655c4d8e2e50076c
}
