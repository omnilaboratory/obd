package service

import (
	"LightningOnOmni/tool"
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

func GenMultiSigScript(aPub, bPub []byte) ([]byte, error) {
	if len(aPub) != 33 || len(bPub) != 33 {
		return nil, fmt.Errorf("Pubkey size error. Compressed pubkeys only")
	}

	// Swap to sort pubkeys if needed. Keys are sorted in lexicographical
	// order. The signatures within the scriptSig must also adhere to the
	// order, ensuring that the signatures for each public key appears in
	// the proper order on the stack.
	if bytes.Compare(aPub, bPub) == 1 {
		aPub, bPub = bPub, aPub
	}

	//ripemd160 := tool.SignMsgWithRipemd160([]byte("abc"))
	//bytes, _ := hex.DecodeString(ripemd160)

	bldr := txscript.NewScriptBuilder()
	bldr.AddOp(txscript.OP_2)
	bldr.AddData(aPub) // Add both pubkeys (sorted).
	bldr.AddData(bPub)
	bldr.AddOp(txscript.OP_2)
	bldr.AddOp(txscript.OP_CHECKMULTISIG)
	//bldr.AddData(bytes)
	return bldr.Script()
}

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
	prevPkScriptHex := "528776a914b5407cec767317d41442aab35bad2712626e17ca88ac"
	prevTxOutputN := uint32(0)

	hash, _ := chainhash.NewHashFromStr(prevTxHash) // tx hash
	//lockin, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddInt64(100).AddData(pkScript).AddOp(txscript.OP_0).Script()
	outPoint := wire.NewOutPoint(hash, prevTxOutputN) // 第几个输出
	script, _ := txscript.NewScriptBuilder().AddInt64(2).Script()
	txIn := wire.NewTxIn(outPoint, script, nil)

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
		//AddInt64(2).
		//AddOp(txscript.OP_EQUAL).
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
		script, err = txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, privKey, true)
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

	script, _ := txscript.NewScriptBuilder().AddInt64(2).Script()

	tx.AddTxIn(wire.NewTxIn(&point2, script, nil))
	//tx.AddTxIn(wire.NewTxIn(&point2, script, nil))

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

	signHex, err := rpcClient.SignRawTransactionWithKey(txHex, []string{"91pBTxREWnCHkQkRNJLVxHP2JHg6TSwuP9natL4QEKTmaKkZqZ7"}, nil, "ALL")
	log.Println(signHex)
	decodeHex, err = rpcClient.DecodeRawTransaction(signHex)
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

//第二次 ab49456042bc098d16ed760623b70e45e2a457b529592a1542d4f640b758c924
func createMyTx() {
	tx := wire.NewMsgTx(2)

	utxoHash, _ := chainhash.NewHashFromStr("0979f9bea38d2e9b425147ed16585567d85b3f4b4e682044bbeb593e0f6c718d")
	point := wire.OutPoint{Hash: *utxoHash, Index: 0}
	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil
	tx.AddTxIn(wire.NewTxIn(&point, nil, nil))

	//一共有0.00100000，也即：89000  给新的地址5000 给自己留下 80000 3000 miner fee
	//一共有0.00100000，也即：80000  给新的地址5000 给自己留下 73000 3000 miner fee
	//一共有0.00100000，也即：73000  给新的地址5000 给自己留下 70000 1000 miner fee
	//一共有0.00100000，也即：70000  给新的地址5000 给自己留下 68000 1000 miner fee
	//一共有0.00100000，也即：67000  给新的地址5000 给自己留下 66000 500 miner fee
	changeAddr := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"
	// 1.1 输出1, 给自己转剩下的钱
	addr, _ := btcutil.DecodeAddress(changeAddr, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	tx.AddTxOut(wire.NewTxOut(66000, pkScript))

	//第二个地址
	address := "mtRoRNpVYhMRYPjoz8u9Eqnmm5LqyDzXgh"
	address = "2N5NiNgsyLaQrXGJb5g6zXLnarTqfKt5Znb"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pubKeyHash := addr.ScriptAddress()

	//pubKeyHash ,_= hex.DecodeString("705655d302793524d7907c29ad715b999dddc587")
	lock, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUAL).
		//AddOp(txscript.OP_DUP).
		//AddOp(txscript.OP_HASH160).
		//AddData(pubKeyHash).
		//AddOp(txscript.OP_EQUALVERIFY).
		//AddOp(txscript.OP_CHECKSIG).
		Script()

	tx.AddTxOut(wire.NewTxOut(500, lock))

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
}

func WitnessScriptHash(witnessScript []byte) ([]byte, error) {
	bldr := txscript.NewScriptBuilder()
	bldr.AddOp(txscript.OP_0)
	scriptHash := sha256.Sum256(witnessScript)
	bldr.AddData(scriptHash[:])
	return bldr.Script()
}

//第一次交易 txid 3f09645ed5781a3126e1fee968a94e3e8f2c921ce50c87efb113624b2e6f640f 2019年12月18日 16:07:29
func createMyTx2() {

	var balance int64 = 100000               // 余额
	var fee int64 = 0.00001 * 1e8            // 交易费
	var outAmount int64 = 0.0001 * 1e8       // 交易费
	var leftToMe = balance - fee - outAmount // 余额-交易费就是剩下再给我的

	// 1. 构造输出
	outputs := []*wire.TxOut{}

	// 1.1 输出1, 给自己转剩下的钱
	address := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"
	addr, _ := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	outputs = append(outputs, wire.NewTxOut(leftToMe, pkScript))

	// 1.2 输出2, 给新的地址 mtRoRNpVYhMRYPjoz8u9Eqnmm5LqyDzXgh
	address = "mtRoRNpVYhMRYPjoz8u9Eqnmm5LqyDzXgh"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pkScript, _ = txscript.PayToAddrScript(addr)
	outputs = append(outputs, wire.NewTxOut(outAmount, pkScript))

	// 2. 构造输入
	prevTxHash := "2deec7c6121500d6877570f31d6f477a35d120776f51b9fc8bb0741d67f97b8b"
	prevTxOutputN := uint32(1)
	hash, _ := chainhash.NewHashFromStr(prevTxHash) // tx hash
	//lockin, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddInt64(100).AddData(pkScript).AddOp(txscript.OP_0).Script()
	outPoint := wire.NewOutPoint(hash, prevTxOutputN) // 第几个输出
	txIn := wire.NewTxIn(outPoint, nil, nil)

	inputs := []*wire.TxIn{txIn}

	prevPkScriptHex := "76a9145d48e1d03e4f8c690bf43d97d25f68ef6f36896d88ac"
	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript

	tx := &wire.MsgTx{
		Version:  wire.TxVersion,
		TxIn:     inputs,
		TxOut:    outputs,
		LockTime: 0,
	}

	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
	}
	txHex := hex.EncodeToString(buf.Bytes())
	fmt.Println("hex", txHex)

	decodeHex, err := rpcClient.DecodeRawTransaction(txHex)
	log.Println(decodeHex)

	// 3. 签名
	privKey := "cRuSwcDrc1gwoeaCvr4qFR9sgHB8wFxHzeqW1Bueo885S6RSYYxH" // 私钥
	sign(tx, privKey, prevPkScripts)

	// 4. 输出Hex
	buf = bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
	}
	txHex = hex.EncodeToString(buf.Bytes())
	fmt.Println("hex", txHex)

	decodeHex, err = rpcClient.DecodeRawTransaction(txHex)
	log.Println(decodeHex)

	result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
}

func printTx(tx *wire.MsgTx) {
	txHex, err := getTxHex(tx)
	decodeHex, err := rpcClient.DecodeRawTransaction(txHex)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(decodeHex)
	}
}

func getTxHex(tx *wire.MsgTx) (string, error) {
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	err := tx.Serialize(buf)
	if err != nil {
		log.Println(err)
		return "", err
	}
	txHex := hex.EncodeToString(buf.Bytes())
	return txHex, err
}

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
	changeAddr := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"
	// 1.1 输出1, 给自己转剩下的钱
	addr, _ := btcutil.DecodeAddress(changeAddr, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	tx.AddTxOut(wire.NewTxOut(67000, pkScript))

	//第二个地址
	address := "mtRoRNpVYhMRYPjoz8u9Eqnmm5LqyDzXgh"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	//pubKeyHash := addr.ScriptAddress()

	lock, _ := txscript.NewScriptBuilder().
		AddInt64(1).
		AddOp(txscript.OP_EQUAL).
		Script()

	//lockScript, err := WitnessScriptHash(lock)
	//if err != nil {
	//	log.Println(err)
	//}
	//tx.AddTxOut(wire.NewTxOut(2000, lockScript))
	//0.0003
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

	// 3. 签名
	prevPkScriptHex := "5187"
	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript

	//privKey := "cRuSwcDrc1gwoeaCvr4qFR9sgHB8wFxHzeqW1Bueo885S6RSYYxH" // 私钥
	//sign(tx, privKey, prevPkScripts)
	printTx(tx)
	txHex, _ := getTxHex(tx)
	result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//	67a7903399d70b6898a1e001ede096e3a3645d79a52447a7655c4d8e2e50076c
}
func CreateCustomSpendTxForScriptHash() {
	tx := wire.NewMsgTx(2)

	utxoHash, _ := chainhash.NewHashFromStr("8b2d1636c0fa3825428a651f93e38ee4b862fa3bcc7cac2e18d957edf41ab7ae")
	point := wire.OutPoint{Hash: *utxoHash, Index: 1}

	ripemd160 := tool.SignMsgWithRipemd160([]byte("abc"))
	bytes, _ := hex.DecodeString(ripemd160)

	unlock, _ := txscript.NewScriptBuilder().
		AddData(bytes).
		Script()
	tx.AddTxIn(wire.NewTxIn(&point, unlock, nil))

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

	tx.AddTxOut(wire.NewTxOut(300, lock))

	// 3. 签名
	prevPkScriptHex := "a914850c1de782e157e9b23f069209cce66341350ba387"
	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript

	//privKey := "cRuSwcDrc1gwoeaCvr4qFR9sgHB8wFxHzeqW1Bueo885S6RSYYxH" // 私钥
	//sign(tx, privKey, prevPkScripts)
	printTx(tx)
	txHex, _ := getTxHex(tx)
	result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//	67a7903399d70b6898a1e001ede096e3a3645d79a52447a7655c4d8e2e50076c
}
