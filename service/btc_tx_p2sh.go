package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
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
	//	2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq
	// 52210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae
	keys := make([]string, 0)
	keys = append(keys, pubA)
	keys = append(keys, pubB)
	log.Println(keys)
	result, err := rpcClient.AddMultiSigAddress(len(keys), keys)
	log.Println(err)
	log.Println(result)
	//	2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq
}

func CreateP2SHTx() {
	tx := wire.NewMsgTx(2)

	utxoHash, _ := chainhash.NewHashFromStr("2c4b2230c9d39f80d46bef5bf01025ce9ec9f8f2e8c2b8574c3d7ee55430b1ca")
	point := wire.OutPoint{Hash: *utxoHash, Index: 0}
	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil
	tx.AddTxIn(wire.NewTxIn(&point, nil, nil))

	//一共有0.00100000，也即：89000  给新的地址5000 给自己留下 80000 3000 miner fee
	//一共有0.00100000，也即：80000  给新的地址5000 给自己留下 73000 3000 miner fee
	//一共有0.00100000，也即：73000  给新的地址5000 给自己留下 70000 1000 miner fee
	//一共有0.00100000，也即：70000  给新的地址2000 给自己留下 68000 1000 miner fee
	//一共有0.00100000，也即：63000  给新的地址2000 给自己留下 62000 1000 miner fee
	//一共有0.00100000，也即：60000  给新的地址2000 给自己留下 57000 1000 miner fee
	amount := 40000
	changeAddr := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"
	// 1.1 输出1, 给自己转剩下的钱
	addr, _ := btcutil.DecodeAddress(changeAddr, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	tx.AddTxOut(wire.NewTxOut(int64(amount-11000), pkScript))

	//第二个地址 "wif":"cVEQD3Wm9pdpmAjHz3AuA5uGqZVgEt2kKbVrwUwRTsyLx9z12KvT"
	address := "2N68VtKEQZLaot4Q97Q2EW5wSiyYZbVoSBq"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pubKeyHash := addr.ScriptAddress()

	lock, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUAL).
		Script()
	//0.0002
	tx.AddTxOut(wire.NewTxOut(10000, lock))

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
}
func CreateP2SHSpendTx() {
	tx := wire.NewMsgTx(2)
	utxoHash, _ := chainhash.NewHashFromStr("eb403e007f487a1b342008168345ad5418ec356b239fcc13d3d109b7362ae3cd")
	point := wire.OutPoint{Hash: *utxoHash, Index: 1}
	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil
	redeemScriptStr := "52210316034bfadc098d3abdf9069d305576dcf70b53ab95fa4a3e911a31f4376641af2102c4483151ede561fa04e465b47db1c0309af7f1afe753baedaac46a2d2e2a73c852ae"
	redeemScript, err := hex.DecodeString(redeemScriptStr)
	log.Println(err)
	unlock, err := txscript.NewScriptBuilder().
		AddData(redeemScript).
		Script()
	log.Println(err)
	unlock = nil
	tx.AddTxIn(wire.NewTxIn(&point, unlock, nil))

	amount := 10000
	// 1.1 就转给 mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM，其他留下的都是手续费
	changeAddr := "mp2CSq75LdESK3NFUik7ZAbh1efgXYbnzM"
	addr, _ := btcutil.DecodeAddress(changeAddr, &chaincfg.TestNet3Params)
	pkScript, _ := txscript.PayToAddrScript(addr)
	tx.AddTxOut(wire.NewTxOut(int64(1000), pkScript))

	//找零
	address := "2NBaVhjCPESQohP2rmPnPSKXD5VUYEeZBbd"
	addr, _ = btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	pubKeyHash := addr.ScriptAddress()
	lock, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).
		AddOp(txscript.OP_EQUAL).
		Script()
	tx.AddTxOut(wire.NewTxOut(int64(amount-1000-3000), lock))

	prevPkScriptHex := "a914c917490524822f0988fd335de6384abf7ef9814d87"
	prevPkScript, _ := hex.DecodeString(prevPkScriptHex)
	prevPkScripts := make([][]byte, 1)
	prevPkScripts[0] = prevPkScript

	changeAddr0 := "mneg4WpS1QmZcG73AjPK3mhhkUoKHHrxnf"
	address0, _ := btcutil.DecodeAddress(changeAddr0, &chaincfg.TestNet3Params)
	wif0, _ := getWif(0)
	log.Println(wif0.Address)

	changeAddr1 := "mhs47wGp27XkY4Xv1ZYz7bE2r3wPEhyAKg"
	address1, _ := btcutil.DecodeAddress(changeAddr1, &chaincfg.TestNet3Params)
	wif1, _ := getWif(1)

	pk0 := (*btcec.PublicKey)(&wif0.PrivateKeyObj.PublicKey).SerializeCompressed()
	addr0, err := btcutil.NewAddressPubKey(pk0, &chaincfg.TestNet3Params)
	pk1 := (*btcec.PublicKey)(&wif1.PrivateKeyObj.PublicKey).SerializeCompressed()
	addr1, err := btcutil.NewAddressPubKey(pk1, &chaincfg.TestNet3Params)

	pkScript1, err := txscript.MultiSigScript([]*btcutil.AddressPubKey{addr0, addr1}, 2)
	if err != nil {
	}
	log.Println(hex.EncodeToString(pkScript1))

	scriptAddr1, err := btcutil.NewAddressScriptHash(pkScript1, &chaincfg.TestNet3Params)
	if err != nil {
	}
	log.Println(scriptAddr1.EncodeAddress())

	privKey := "cQEgroe2gVP718XSPU36FoAg1YNq9Tv92mWHQ5ACxVnWM4HY3UA3" // 私钥
	signMulti(tx, privKey, prevPkScripts)

	scriptPkScript, err := txscript.PayToAddrScript(scriptAddr1)
	if err != nil {
	}
	sigScript, err := txscript.SignTxOutput(&chaincfg.TestNet3Params, tx, 0, scriptPkScript, txscript.SigHashAll, mkGetKey(map[string]addressToKey{
		address0.EncodeAddress(): {wif0.Wifobj.PrivKey, true},
		address1.EncodeAddress(): {wif1.Wifobj.PrivKey, true},
	}), mkGetScript(map[string][]byte{
		scriptAddr1.EncodeAddress(): pkScript1,
	}), nil)

	privKey = "cVSEGBzxoDf8RSdnSjtRP1ei2ghrxKpgXybPZpBQjWG71Xrorxs8" // 私钥
	signMulti(tx, privKey, prevPkScripts)

	//signatureScript1 := tx.TxIn[0].SignatureScript

	//script, err := txscript.NewScriptBuilder().AddData(signatureScript0).AddData(redeemScript).Script()
	//if err!=nil {
	//	log.Println(err)
	//}
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

func getWif(index uint32) (wallet *Wallet, err error) {
	mnemonic := "unfold tortoise zoo hand sausage project boring corn test same elevator mansion bargain coffee brick tilt forum purpose hundred embody weapon ripple when narrow"
	userId := tool.SignMsgWithSha256([]byte(mnemonic))

	changeExtKey, _ := HDWalletService.CreateChangeExtKey(mnemonic)
	user := &bean.User{}
	user.CurrAddrIndex = 0
	user.ChangeExtKey = changeExtKey
	user.Mnemonic = mnemonic
	user.PeerId = userId
	return HDWalletService.GetAddressByIndex(user, index)
}

type addressToKey struct {
	key        *btcec.PrivateKey
	compressed bool
}

func mkGetKey(keys map[string]addressToKey) txscript.KeyDB {
	if keys == nil {
		return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
			bool, error) {
			return nil, false, errors.New("nope mkGetKey")
		})
	}
	return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
		bool, error) {
		a2k, ok := keys[addr.EncodeAddress()]
		if !ok {
			return nil, false, errors.New("nope mkGetKey")
		}
		return a2k.key, a2k.compressed, nil
	})
}

func mkGetScript(scripts map[string][]byte) txscript.ScriptDB {
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
