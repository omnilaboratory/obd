package service

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/tidwall/gjson"
	"log"
)

// addrtool_mnemonic_test.go的Test_Demo2生成
// unfold tortoise zoo hand sausage project boring corn test same elevator mansion bargain coffee brick tilt forum purpose hundred embody weapon ripple when narrow
//{
//"index":0,
//"address":"mneg4WpS1QmZcG73AjPK3mhhkUoKHHrxnf",
//"pub_key":"0316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af",
//"private_key":"772a56ed5d0dcc444039bcd0a3a56ebf5cb404b7f8e0314ecacdfe65cf4b0ea2",
//"wif":"cRaLt58h5xv5nfw9vqNJ6jtopsa5SWMhA7eCmhBuyTBviiZ8dx6N",
//"Wifobj":Object{...},
//"PrivateKeyObj":Object{...}
//}
//{
//"index":1,
//"address":"mhs47wGp27XkY4Xv1ZYz7bE2r3wPEhyAKg",
//"pub_key":"02c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c8",
//"private_key":"dd97527f2bb8c20a0b9851df382f7d72779c83e9fab1ea31f63b33655715c76d",
//"wif":"cV1Sot9e8pb2Ern5DCtbMRDfd74dFbRhM6dhxDtg7CmdEVu7CtyD",
//"Wifobj":Object{...},
//"PrivateKeyObj":Object{...}
//}

func CreateMuiltAddress() {
	pubA := "0316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af"
	pubB := "02c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c8"
	apub, err := hex.DecodeString(pubA)
	log.Println(err)
	bpub, err := hex.DecodeString(pubB)
	log.Println(err)
	bytes, err := GenMultiSigScript(apub, bpub)
	log.Println(hex.EncodeToString(bytes))
	//create multisig address
	hash, err := btcutil.NewAddressScriptHash(bytes, &chaincfg.TestNet3Params)
	log.Println(hash)

	decodeScript, _ := rpcClient.DecodeScript(hex.EncodeToString(bytes))
	log.Println(decodeScript)

	//	2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq
	// 52210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae
	// 52210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae
	keys := make([]string, 0)
	keys = append(keys, pubA)
	keys = append(keys, pubB)
	log.Println(keys)
	result, err := rpcClient.AddMultiSigAddress(len(keys), keys)
	log.Println(err)
	log.Println(result)
	parse := gjson.Parse(string(result))
	decodeScript, _ = rpcClient.DecodeScript(parse.Get("redeemScript").String())
	log.Println(decodeScript)
	//	2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq
}

func CreateP2SHTx() {
	tx := wire.NewMsgTx(2)

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
	address := "2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pkScript, _ = txscript.PayToAddrScript(addr)

	//pubKeyHash := addr.ScriptAddress()
	//lock, _ := txscript.NewScriptBuilder().
	//	AddOp(txscript.OP_HASH160).
	//	AddData(pubKeyHash).
	//	AddOp(txscript.OP_EQUAL).
	//	Script()
	//0.0002
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
	result, err := rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//	eb403e007f487a1b342008168345ad5418ec356b239fcc13d3d109b7362ae3cd
	//c964397cb990751e84b4c0e1044b026b0985b987d5a7e03fc33b28c0ac964c9e
	//	1461e784816ef87771edbdb99f6183e522d168438467aefeb6ce89266097cd61
}

func CreateP2SHSpendTx() {
	tx := wire.NewMsgTx(2)
	utxoHash, _ := chainhash.NewHashFromStr("1461e784816ef87771edbdb99f6183e522d168438467aefeb6ce89266097cd61")
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
	address := "2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pubKeyHash := addr.ScriptAddress()
	lock, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUAL).
		Script()
	tx.AddTxOut(wire.NewTxOut(int64(amount-500-500), lock))

	wif0, err := btcutil.DecodeWIF("cRaLt58h5xv5nfw9vqNJ6jtopsa5SWMhA7eCmhBuyTBviiZ8dx6N")
	wif1, err := btcutil.DecodeWIF("cV1Sot9e8pb2Ern5DCtbMRDfd74dFbRhM6dhxDtg7CmdEVu7CtyD")

	pk0 := (*btcec.PublicKey)(&wif0.PrivKey.PublicKey).SerializeCompressed()
	addr0, err := btcutil.NewAddressPubKey(pk0, &chaincfg.TestNet3Params)
	pk1 := (*btcec.PublicKey)(&wif1.PrivKey.PublicKey).SerializeCompressed()
	addr1, err := btcutil.NewAddressPubKey(pk1, &chaincfg.TestNet3Params)

	//生成多签地址 得到redeemscript
	redeemScriptStr := "52210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae"
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

	sigScript, err := txscript.SignTxOutput(&chaincfg.TestNet3Params, tx, 0, scriptPkScript, txscript.SigHashAll, mkGetKey(map[string]addressToKey{
		addr0.EncodeAddress(): {wif0.PrivKey, true},
		addr1.EncodeAddress(): {wif1.PrivKey, true},
	}), mkGetScript(map[string][]byte{
		address: redeemScript,
	}), nil)
	log.Println(err)

	decodeScript, _ := rpcClient.DecodeScript(hex.EncodeToString(sigScript))
	log.Println(decodeScript)
	tx.TxIn[0].SignatureScript = sigScript

	txHex, _ := getTxHex(tx)
	log.Println(txHex)
	result, err := rpcClient.DecodeRawTransaction(txHex)
	log.Println(err)
	log.Println(result)

	result, err = rpcClient.SendRawTransaction(txHex)
	log.Println(err)
	log.Println(result)
	//f33cf515adcf55c5b39200d06e95b0a31c49797340e41689a5ae9da9284d7855
}

func rpcSendMulti() {
	//rpc 的生成
	privkeys := make([]string, 0)
	privkeys = append(privkeys, "cRaLt58h5xv5nfw9vqNJ6jtopsa5SWMhA7eCmhBuyTBviiZ8dx6N")
	privkeys = append(privkeys, "cV1Sot9e8pb2Ern5DCtbMRDfd74dFbRhM6dhxDtg7CmdEVu7CtyD")
	outputItems := []rpc.TransactionOutputItem{
		{ToBitCoinAddress: "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM", Amount: 0.00001},
	}
	redeemScriptStr := "52210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae"
	txid, hex, err := rpcClient.BtcCreateAndSignRawTransaction("2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq", privkeys, outputItems, 0, 0, &redeemScriptStr)
	log.Println(err)
	log.Println(hex)
	log.Println(txid)
}

type addressToKey struct {
	key        *btcec.PrivateKey
	compressed bool
}

func mkGetKey(keys map[string]addressToKey) txscript.KeyDB {
	if keys == nil {
		return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey, bool, error) {
			return nil, false, errors.New("nope mkGetKey")
		})
	}
	return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey, bool, error) {
		a2k, ok := keys[addr.EncodeAddress()]
		if !ok {
			return nil, false, errors.New("nope mkGetKey")
		}
		return a2k.key, a2k.compressed, nil
	})
}

func mkGetScript(scripts map[string][]byte) txscript.ScriptDB {
	log.Println(scripts)
	if scripts == nil {
		return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
			return nil, errors.New("nope mkGetScript 10")
		})
	}
	return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
		script, ok := scripts[addr.EncodeAddress()]
		if !ok {
			return nil, errors.New("nope mkGetScript 20")
		}
		return script, nil
	})
}
