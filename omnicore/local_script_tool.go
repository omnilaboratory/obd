package omnicore

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"strings"
)

func VerifyOmniTxHex(hex string, propertyId int64, amount float64, toAddress string, propertyDivisible bool) (pass bool, err error) {
	decodeRawTransaction := DecodeRawTransaction(hex, tool.GetCoreNet())

	isRightAddress := false
	isRightOpreturn := false
	vouts := gjson.Parse(decodeRawTransaction).Get("vout").Array()
	for _, item := range vouts {
		scriptPubKeyType := item.Get("scriptPubKey").Get("type").Str
		if scriptPubKeyType == "nulldata" {
			opreturnHex := item.Get("scriptPubKey").Get("hex").Str
			if VerfyOpreturnPayload(opreturnHex, strconv.Itoa(int(propertyId)), tool.FloatToString(amount, 8), propertyDivisible) {
				isRightOpreturn = true
			}
		}

		value := item.Get("value").Float()
		if value <= tool.GetOmniDustBtc() {
			addresses := item.Get("scriptPubKey").Get("addresses").Array()
			for _, address := range addresses {
				if address.Str == toAddress {
					isRightAddress = true
					break
				}
			}
		}
	}
	if isRightOpreturn && isRightAddress {
		return true, nil
	}
	if isRightOpreturn && isRightAddress == false {
		return false, errors.New("wrong address")
	}
	if isRightOpreturn == false && isRightAddress {

		return false, errors.New("wrong propertyId or amount")
	}
	return false, errors.New("wrong hex")
}

func VerifyOmniTxHexOutAddress(hex string, toAddress string) (pass bool, err error) {
	decodeRawTransaction := DecodeRawTransaction(hex, tool.GetCoreNet())
	isRightAddress := false
	vouts := gjson.Parse(decodeRawTransaction).Get("vout").Array()
	for _, item := range vouts {

		value := item.Get("value").Float()
		if value <= tool.GetOmniDustBtc() {
			addresses := item.Get("scriptPubKey").Get("addresses").Array()
			for _, address := range addresses {
				if address.Str == toAddress {
					isRightAddress = true
					break
				}
			}
		}
	}
	if isRightAddress {
		return true, nil
	}
	return false, errors.New("wrong hex")
}

//SignRawTransactionWithKey(inputData.Hex, []string{inputData.Prvkey}, inputData.Inputs, "ALL")
//https://www.thepolyglotdeveloper.com/2018/03/create-sign-bitcoin-transactions-golang/
func SignRawHex(inputs []bean.RawTxInputItem, redeemHex string, privKey string) (signedHex string, err error) {
	redeemHexBytes, _ := hex.DecodeString(redeemHex)
	redeemTx := wire.MsgTx{}
	err = redeemTx.Deserialize(bytes.NewReader(redeemHexBytes))
	if err != nil {
		return "", nil
	}

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		log.Println(err)
	}

	lookupKey := func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
		return wif.PrivKey, true, nil
	}

	for index, _ := range redeemTx.TxIn {
		item := inputs[index]
		redeemScriptBytes, _ := hex.DecodeString(item.RedeemScript)
		scriptAddr, _ := btcutil.NewAddressScriptHash(redeemScriptBytes, tool.GetCoreNet())
		inputPkScriptBytes, _ := hex.DecodeString(item.ScriptPubKey)

		script, err := txscript.SignTxOutput(
			tool.GetCoreNet(),
			&redeemTx,
			index,
			inputPkScriptBytes,
			txscript.SigHashAll,
			txscript.KeyClosure(lookupKey),
			mkGetScript(map[string][]byte{
				scriptAddr.EncodeAddress(): redeemScriptBytes}),
			redeemTx.TxIn[index].SignatureScript)

		if err != nil {
			return "", err
		}

		redeemTx.TxIn[index].SignatureScript = script
	}

	toHex := TxToHex(&redeemTx)
	return toHex, nil
}

func mkGetScript(scripts map[string][]byte) txscript.ScriptDB {
	if scripts == nil {
		return txscript.ScriptClosure(func(addr btcutil.Address) (
			[]byte, error) {
			return nil, errors.New("empty redeemScript")
		})
	}
	return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte,
		error) {
		script, ok := scripts[addr.EncodeAddress()]
		if !ok {
			return nil, errors.New("wrong redeemScript")
		}
		return script, nil
	})
}

// wrong sourceHex: false stack entry at end of script execution
// emptySign: index 0 is invalid for stack size 0
// partial sign: not all signatures empty on failed checkmultisig
// all sign: nil
func VerifySignatureHex(inputs []bean.RawTxInputItem, redeemHex string) (err error) {
	redeemTx := wire.MsgTx{}
	redeemHexBytes, _ := hex.DecodeString(redeemHex)
	redeemTx.Deserialize(bytes.NewReader(redeemHexBytes))

	flags := txscript.StandardVerifyFlags

	for index, _ := range redeemTx.TxIn {
		item := inputs[index]
		inputPkScriptBytes, _ := hex.DecodeString(item.ScriptPubKey)
		vm, err := txscript.NewEngine(inputPkScriptBytes, &redeemTx, 0, flags, nil, nil, -1)
		if err != nil {
			return err
		}
		err = vm.Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

type multiSign struct {
	Address      string `json:"address"`
	RedeemScript string `json:"redeemScript"`
	ScriptPubKey string `json:"scriptPubKey"`
}

func CreateMultiSig(minSignNum int, keys []string) (result string, err error) {
	addr, redeemScript, scriptPubKey := CreateMultiSigAddr(keys[0], keys[1], tool.GetCoreNet())
	sign := multiSign{}
	sign.Address = addr
	sign.RedeemScript = redeemScript
	sign.ScriptPubKey = scriptPubKey
	marshal, _ := json.Marshal(&sign)
	return string(marshal), nil
}

func DecodeBtcRawTransaction(hex string) (result string, err error) {
	result = DecodeRawTransaction(hex, tool.GetCoreNet())
	return result, err
}

func GetMinerFee() float64 {
	price := conn2tracker.EstimateSmartFee()
	if price == 0 {
		price = 6
	} else {
		price = price / 6
	}
	if price < 4 {
		price = 4
	}
	txSize := 150 + 68 + 90
	result, _ := decimal.NewFromFloat(float64(txSize) * price).Div(decimal.NewFromFloat(100000000)).Round(8).Float64()
	return result
}

func GetTxId(hex string) string {
	testResult, err := DecodeBtcRawTransaction(hex)
	if err == nil {
		return gjson.Parse(testResult).Get("txid").Str
	}
	return ""
}

func CheckMultiSign(hex string, step int) (pass bool, err error) {
	if len(hex) == 0 {
		return false, errors.New("Empty hex")
	}
	result, err := DecodeBtcRawTransaction(hex)
	vins := gjson.Get(result, "vin").Array()
	for i := 0; i < len(vins); i++ {
		asm := vins[i].Get("scriptSig").Get("asm").Str
		asmArr := strings.Split(asm, " ")
		if step == 1 {
			if len(asmArr) != 4 || (asmArr[1] == "0" && asmArr[2] == "0") {
				return false, errors.New("err sign")
			}
		}
		if step == 2 {
			if len(asmArr) != 4 || asmArr[1] == "0" || asmArr[2] == "0" {
				return false, errors.New("err sign")
			}
		}
	}
	return true, nil
}
