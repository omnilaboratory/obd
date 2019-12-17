package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"log"
)

//https://ibz.bz/2018/01/29/76b8e2e8c60716e70db281c6347bd9b9.html  使用golang一步步教你在比特币上刻字

func GenerateBTC() (string, string, error) {
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}

	privKeyWif, err := btcutil.NewWIF(privKey, &chaincfg.MainNetParams, false)
	if err != nil {
		return "", "", err
	}
	pubKeySerial := privKey.PubKey().SerializeUncompressed()

	pubKeyAddress, err := btcutil.NewAddressPubKey(pubKeySerial, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", err
	}

	return privKeyWif.String(), pubKeyAddress.EncodeAddress(), nil
}

func GenerateBTCTest() (string, string, error) {
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}

	privKeyWif, err := btcutil.NewWIF(privKey, &chaincfg.TestNet3Params, false)
	if err != nil {
		return "", "", err
	}
	pubKeySerial := privKey.PubKey().SerializeUncompressed()

	pubKeyAddress, err := btcutil.NewAddressPubKey(pubKeySerial, &chaincfg.TestNet3Params)
	if err != nil {
		return "", "", err
	}

	return privKeyWif.String(), pubKeyAddress.EncodeAddress(), nil
}

func createTx() {

	address := "n3Tiwiq6nEWYMXFTNZ1GGUZcGLjSasaaEr"
	var balance int64 = 3849434  // 余额
	var fee int64 = 0.001 * 1e8  // 交易费
	var leftToMe = balance - fee // 余额-交易费就是剩下再给我的

	// 1. 构造输出
	outputs := []*wire.TxOut{}

	// 1.1 输出1, 给自己转剩下的钱
	addr, _ := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	outputs = append(outputs, wire.NewTxOut(leftToMe, pkScript))

	// 1.2 输出2, 添加文字
	comment := "这是一个留言, 哈哈"
	pkScript, _ = txscript.NullDataScript([]byte(comment))
	outputs = append(outputs, wire.NewTxOut(int64(0), pkScript))

	// 2. 构造输入
	prevTxHash := "7c07ab6c3a0201e6189a3ec3fd3cea67d92125de2a3b69558a0549cd12f3ec11"
	prevPkScriptHex := "76a914f0b65ff8057ad051c768c7d2d0f7ca9cb480349488ac"
	prevTxOutputN := uint32(0)

	hash, _ := chainhash.NewHashFromStr(prevTxHash) // tx hash
	//lockin, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddInt64(100).AddData(pkScript).AddOp(txscript.OP_0).Script()
	outPoint := wire.NewOutPoint(hash, prevTxOutputN) // 第几个输出
	txIn := wire.NewTxIn(outPoint, []byte{txscript.OP_0, txscript.OP_0}, nil)

	temp := "483045022100829e7bd378cffbaf31701499432476a408492084ba549909f6b784c5bb6d6be002202963508a6d6abd346bf9eb031fd00dd8b5048cd0375a12f6d7c634029a83f54b01410483307216ef2b4fa4a149573fca3bd10fe6291257a1857ef03bc4dda3018c602c3415374316e33abbe35880d2532bcde1d2d827528ca5f82d1d0389e9cfda1bf7"
	result, _ := rpcClient.DecodeScript(temp)
	log.Println(result)
	temp = "483045022100829e7bd378cffbaf31701499432476a408492084ba549909f6b784c5bb6d6be002202963508a6d6abd346bf9eb031fd00dd8b5048cd0375a12f6d7c634029a83f54b01410483307216ef2b4fa4a149573fca3bd10fe6291257a1857ef03bc4dda3018c602c3415374316e33abbe35880d2532bcde1d2d827528ca5f82d1d0389e9cfda1bf7"
	result, _ = rpcClient.DecodeScript(temp)
	log.Println(result)

	inputs := []*wire.TxIn{txIn}

	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript

	tx := &wire.MsgTx{
		Version:  wire.TxVersion,
		TxIn:     inputs,
		TxOut:    outputs,
		LockTime: 0,
	}

	pubKeyHash, _ := hex.DecodeString("b5407cec767317d41442aab35bad2712626e17ca")
	log.Println(pubKeyHash)

	lock, _ := txscript.NewScriptBuilder().
		AddInt64(1).
		AddOp(txscript.OP_ADD).
		AddInt64(1).
		AddOp(txscript.OP_EQUAL).
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()

	//构建第一个Output，是找零0.2991024 BTC
	tx.AddTxOut(wire.NewTxOut(29910240, lock))

	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
	}
	txHex := hex.EncodeToString(buf.Bytes())
	fmt.Println("hex", txHex)

	decodeHex, err := rpcClient.DecodeRawTransaction(txHex)
	log.Println(decodeHex)

	// 3. 签名
	privKey := "91pBTxREWnCHkQkRNJLVxHP2JHg6TSwuP9natL4QEKTmaKkZqZ7" // 私钥
	sign(tx, privKey, prevPkScripts)

	// 4. 输出Hex
	buf = bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
	}
	txHex = hex.EncodeToString(buf.Bytes())
	fmt.Println("hex", txHex)

	decodeHex, err = rpcClient.DecodeRawTransaction(txHex)
	log.Println(decodeHex)

	//result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	//log.Println(result)
}

// 签名
func sign(tx *wire.MsgTx, privKeyStr string, prevPkScripts [][]byte) {
	inputs := tx.TxIn
	wif, err := btcutil.DecodeWIF(privKeyStr)

	fmt.Println("wif err", err)
	privKey := wif.PrivKey

	for i := range inputs {
		pkScript := prevPkScripts[i]
		var script []byte
		script, err = txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, privKey, false)
		inputs[i].SignatureScript = script
	}
}

// https://www.cnblogs.com/studyzy/p/how-bitcoin-sign-a-tx.html
func buildRawTx() *wire.MsgTx {

	//https://testnet.blockchain.info/tx/f0d9d482eb122535e32a3ae92809dd87839e63410d5fd52816fc9fc6215018cc?show_adv=true

	tx := wire.NewMsgTx(wire.TxVersion)

	//https://testnet.blockchain.info/tx-index/239152566/1 0.4BTC

	utxoHash, _ := chainhash.NewHashFromStr("1dda832890f85288fec616ef1f4113c0c86b7bf36b560ea244fd8a6ed12ada52")

	point := wire.OutPoint{Hash: *utxoHash, Index: 1}

	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil

	tx.AddTxIn(wire.NewTxIn(&point, nil, nil))

	//https://testnet.blockchain.info/tx-index/239157459/1 1.1BTC

	utxoHash2, _ := chainhash.NewHashFromStr("24f284aed2b9dbc19f0d435b1fe1ee3b3ddc763f28ca28bad798d22b6bea0c66")

	point2 := wire.OutPoint{Hash: *utxoHash2, Index: 1}

	//构建第二个Input，指向一个1.1BTC的UTXO，第二个参数是解锁脚本，现在是nil

	tx.AddTxIn(wire.NewTxIn(&point2, nil, nil))

	//找零的地址（这里是16进制形式，变成Base58格式就是 mx3KrUjRzzqYTcsyyvWBiHBncLrrTPXnkV ）

	pubKeyHash, _ := hex.DecodeString("b5407cec767317d41442aab35bad2712626e17ca")
	log.Println(pubKeyHash)
	address := "mx3KrUjRzzqYTcsyyvWBiHBncLrrTPXnkV"
	addr, _ := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	log.Println(addr.ScriptAddress())
	log.Println(hex.EncodeToString(addr.ScriptAddress()))

	lock, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()

	//构建第一个Output，是找零0.2991024 BTC

	tx.AddTxOut(wire.NewTxOut(29910240, lock))

	//支付给了某个地址，仍然是16进制形式，Base58形式是：mxqnGTekzKqnMqNFHKYi8FhV99WcvQGhfH  be09abcbfda1f2c26899f062979ab0708731235a。

	pubKeyHash, _ = hex.DecodeString("be09abcbfda1f2c26899f062979ab0708731235a")

	lock2, _ := txscript.
		NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()

	//构建第二个Output，支付1.2 BTC出去

	tx.AddTxOut(wire.NewTxOut(120000000, lock2))

	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	err := tx.Serialize(buf)
	if err != nil {
		log.Println(err)
		return nil
	}
	txHex := hex.EncodeToString(buf.Bytes())

	fmt.Println("hex", txHex)
	decodeHex, err := rpcClient.DecodeRawTransaction(txHex)
	log.Println(decodeHex)
	log.Println(err)
	return tx
}

type Transaction struct {
	TxId               string `json:"txid"`
	SourceAddress      string `json:"source_address"`
	DestinationAddress string `json:"destination_address"`
	Amount             int64  `json:"amount"`
	UnsignedTx         string `json:"unsignedtx"`
	SignedTx           string `json:"signedtx"`
}

func CreateTransaction(secret string, destination string, amount int64, txHash string) (Transaction, error) {
	var transaction Transaction
	wif, err := btcutil.DecodeWIF(secret)
	if err != nil {
		return Transaction{}, err
	}
	addresspubkey, _ := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), &chaincfg.MainNetParams)
	sourceTx := wire.NewMsgTx(wire.TxVersion)
	sourceUtxoHash, _ := chainhash.NewHashFromStr(txHash)
	sourceUtxo := wire.NewOutPoint(sourceUtxoHash, 0)
	sourceTxIn := wire.NewTxIn(sourceUtxo, nil, nil)
	destinationAddress, err := btcutil.DecodeAddress(destination, &chaincfg.MainNetParams)
	sourceAddress, err := btcutil.DecodeAddress(addresspubkey.EncodeAddress(), &chaincfg.MainNetParams)
	if err != nil {
		return Transaction{}, err
	}
	destinationPkScript, _ := txscript.PayToAddrScript(destinationAddress)
	sourcePkScript, _ := txscript.PayToAddrScript(sourceAddress)
	sourceTxOut := wire.NewTxOut(amount, sourcePkScript)
	sourceTx.AddTxIn(sourceTxIn)
	sourceTx.AddTxOut(sourceTxOut)
	sourceTxHash := sourceTx.TxHash()
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	prevOut := wire.NewOutPoint(&sourceTxHash, 0)
	redeemTxIn := wire.NewTxIn(prevOut, nil, nil)
	redeemTx.AddTxIn(redeemTxIn)
	redeemTxOut := wire.NewTxOut(amount, destinationPkScript)
	redeemTx.AddTxOut(redeemTxOut)
	sigScript, err := txscript.SignatureScript(redeemTx, 0, sourceTx.TxOut[0].PkScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
		return Transaction{}, err
	}
	redeemTx.TxIn[0].SignatureScript = sigScript
	flags := txscript.StandardVerifyFlags
	vm, err := txscript.NewEngine(sourceTx.TxOut[0].PkScript, redeemTx, 0, flags, nil, nil, amount)
	if err != nil {
		return Transaction{}, err
	}
	if err := vm.Execute(); err != nil {
		return Transaction{}, err
	}
	var unsignedTx bytes.Buffer
	var signedTx bytes.Buffer
	sourceTx.Serialize(&unsignedTx)
	redeemTx.Serialize(&signedTx)
	transaction.TxId = sourceTxHash.String()
	transaction.UnsignedTx = hex.EncodeToString(unsignedTx.Bytes())
	transaction.Amount = amount
	transaction.SignedTx = hex.EncodeToString(signedTx.Bytes())
	transaction.SourceAddress = sourceAddress.EncodeAddress()
	transaction.DestinationAddress = destinationAddress.EncodeAddress()
	return transaction, nil
}

func createTx3() {
	// Addr: mxWG2KkzEqtfJonJmpLk5Y9MQgf2EKLTUj
	//  "pubkey": "02630fc5610b53acffc0ed1cb4e31698990fa885d8a2f460fd9a81a46848ef4ff0"
	//  "scriptPubKey": "76a914ba588588acb708b1cc8b0128de765e6836bfa74188ac"

	// preimage R is 'TestR'
	//preimage_R := "TestR"
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

	// construct an output
	builder := txscript.NewScriptBuilder()

	// Add Hash(R) to output
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(btcutil.Hash160(paymentHash[:]))
	builder.AddOp(txscript.OP_EQUALVERIFY)

	// If checking R is passed, then check address signature.
	builder.AddOp(txscript.OP_IF)
	builder.AddOp(txscript.OP_DROP)
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)

	// Add pubkey hash of an address to output
	addrPubKey := "02630fc5610b53acffc0ed1cb4e31698990fa885d8a2f460fd9a81a46848ef4ff0"
	pubKeyHash, _ := hex.DecodeString(addrPubKey)
	builder.AddData(btcutil.Hash160(pubKeyHash))

	builder.AddOp(txscript.OP_EQUAL)
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddOp(txscript.OP_ENDIF)
}
