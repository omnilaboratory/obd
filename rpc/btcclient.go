package rpc

import (
	"LightningOnOmni/config"
	"LightningOnOmni/tool"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"math"
	"strconv"
)

//https://bitcoin.org/en/developer-reference#bitcoin-core-apis
func (client *Client) CreateMultiSig(minSignNum int, keys []string) (result string, err error) {
	for _, item := range keys {
		importAddress(item)
	}
	return client.send("createmultisig", []interface{}{minSignNum, keys})
}

func (client *Client) AddMultiSigAddress(minSignNum int, keys []string) (result string, err error) {
	for _, item := range keys {
		importAddress(item)
	}
	return client.send("addmultisigaddress", []interface{}{minSignNum, keys})
}

func (client *Client) DumpPrivKey(address string) (result string, err error) {
	importAddress(address)
	return client.send("dumpprivkey", []interface{}{address})
}

func (client *Client) GetTransactionById(txid string) (result string, err error) {
	return client.send("gettransaction", []interface{}{txid})
}

func (client *Client) GetTxOut(txid string, num int) (result string, err error) {
	return client.send("gettxout", []interface{}{txid, num})
}

func (client *Client) GetNewAddress(label string) (address string, err error) {
	return client.send("getnewaddress", []interface{}{label})
}

func (client *Client) ListUnspent(address string) (result string, err error) {
	if tool.CheckIsString(&address) == false {
		return "", errors.New("address not exist")
	}
	importAddress(address)
	return client.send("listunspent", []interface{}{0, math.MaxInt32, []string{address}})
}
func (client *Client) GetBalanceByAddress(address string) (balance decimal.Decimal, err error) {
	result, err := client.ListUnspent(address)
	balance = decimal.NewFromFloat(0)
	if err != nil {
		return balance, err
	}
	array := gjson.Parse(result).Array()
	for _, item := range array {
		amount := item.Get("amount").Float()
		balance = balance.Add(decimal.NewFromFloat(amount))
	}
	return balance, nil
}

func (client *Client) CreateRawTransaction(inputs []map[string]interface{}, outputs map[string]interface{}) (result string, err error) {
	return client.send("createrawtransaction", []interface{}{inputs, outputs})
}

func (client *Client) SignRawTransactionWithKey(hex string, privkeys []string, prevtxs []map[string]interface{}, sighashtype string) (result string, err error) {
	return client.send("signrawtransactionwithkey", []interface{}{hex, privkeys, prevtxs, sighashtype})
}

func (client *Client) SendRawTransaction(hex string) (result string, err error) {
	return client.send("sendrawtransaction", []interface{}{hex})
}

func (client *Client) DecodeRawTransaction(hex string) (result string, err error) {
	return client.send("decoderawtransaction", []interface{}{hex})
}
func (client *Client) OmniDecodeTransaction(hex string) (result string, err error) {
	return client.send("omni_decodetransaction", []interface{}{hex})
}

func (client *Client) GetBlockCount() (result int, err error) {
	height, err := client.send("getblockcount", nil)
	if err == nil {
		result, err = strconv.Atoi(height)
	}
	return result, err
}

func (client *Client) GetDifficulty() (result string, err error) {
	return client.send("getdifficulty", nil)
}

func (client *Client) GetMiningInfo() (result string, err error) {
	return client.send("getmininginfo", nil)
}

func (client *Client) GetNetworkInfo() (result string, err error) {
	return client.send("getnetworkinfo", nil)
}

func (client *Client) SignMessageWithPrivKey(privkey string, message string) (result string, err error) {
	return client.send("signmessagewithprivkey", []interface{}{privkey, message})
}

func (client *Client) VerifyMessage(address string, signature string, message string) (result string, err error) {
	importAddress(address)
	return client.send("verifymessage", []interface{}{address, signature, message})
}
func (client *Client) DecodeScript(hexString string) (result string, err error) {
	return client.send("decodescript", []interface{}{hexString})
}

func (client *Client) ValidateAddress(address string) (isValid bool, err error) {
	importAddress(address)
	result, err := client.send("validateaddress", []interface{}{address})
	if err != nil {
		return false, err
	}
	log.Println(result)
	return gjson.Get(result, "isvalid").Bool(), nil
}
func (client *Client) GetAddressInfo(address string) (json string, err error) {
	importAddress(address)
	result, err := client.send("getaddressinfo", []interface{}{address})
	if err != nil {
		return "", err
	}
	return result, nil
}

func importAddress(address string) {
	client.send("importaddress", []interface{}{address, "", false})
}

func (client *Client) ImportPrivKey(privkey string) (result string, err error) {
	return client.send("importprivkey", []interface{}{privkey, "", false})
}

func (client *Client) ImportAddress(address string) (err error) {
	result, err := client.send("importaddress", []interface{}{address, nil, false})
	if err != nil {
		return err
	}
	log.Println(result)
	return nil
}

type TransactionOutputItem struct {
	ToBitCoinAddress string
	Amount           float64
}
type TransactionInputItem struct {
	Txid         string
	ScriptPubKey string
	Vout         uint32
	Amount       float64
}

// create a transaction and just signnature , not send to the network,get the hash of signature
func (client *Client) BtcCreateAndSignRawTransaction(fromBitCoinAddress string, privkeys []string, outputItems []TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (txid string, hex string, err error) {
	if len(fromBitCoinAddress) < 1 {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return "", "", errors.New("toBitCoinAddress is empty")
	}

	if minerFee <= 0 {
		minerFee = GetMinerFee()
	}

	outTotalAmount := decimal.NewFromFloat(0)
	for _, item := range outputItems {
		outTotalAmount = outTotalAmount.Add(decimal.NewFromFloat(item.Amount))
	}

	if outTotalAmount.LessThan(decimal.NewFromFloat(config.Dust)) {
		return "", "", errors.New("wrong outTotalAmount")
	}

	if minerFee < config.Dust {
		return "", "", errors.New("minerFee too small")
	}

	result, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", "", err
	}

	array := gjson.Parse(result).Array()
	if len(array) == 0 {
		return "", "", errors.New("empty balance")
	}
	log.Println("listunspent", array)

	//out, _ := decimal.NewFromFloat(minerFee).Add(outTotalAmount).Float64()
	out, _ := outTotalAmount.Float64()

	balance := 0.0
	var inputs []map[string]interface{}
	for _, item := range array {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		if redeemScript != nil {
			node["redeemScript"] = *redeemScript
		} else {
			if item.Get("redeemScript").Exists() {
				node["redeemScript"] = item.Get("redeemScript")
			}
		}
		if sequence > 0 {
			node["sequence"] = sequence
		}
		node["scriptPubKey"] = item.Get("scriptPubKey").String()
		inputs = append(inputs, node)
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Float64()
	}
	log.Println("input list ", inputs)

	if len(inputs) == 0 || balance < out {
		return "", "", errors.New("not enough balance")
	}

	minerFeeAndOut, _ := decimal.NewFromFloat(minerFee).Add(outTotalAmount).Float64()

	subMinerFee := 0.0
	if balance <= minerFeeAndOut {
		needLessFee, _ := decimal.NewFromFloat(minerFeeAndOut).Sub(decimal.NewFromFloat(balance)).Float64()
		if needLessFee > 0 {
			var outTotalCount = 0
			for _, item := range outputItems {
				if item.Amount > 0 {
					outTotalCount++
				}
			}
			if outTotalCount > 0 {
				count, _ := decimal.NewFromString(strconv.Itoa(outTotalCount))
				subMinerFee, _ = decimal.NewFromFloat(minerFee).DivRound(count, 8).Float64()
			}
		}
	}

	drawback, _ := decimal.NewFromFloat(balance).Sub(decimal.NewFromFloat(minerFeeAndOut)).Float64()
	output := make(map[string]interface{})
	for _, item := range outputItems {
		if item.Amount > 0 {
			output[item.ToBitCoinAddress], _ = decimal.NewFromFloat(item.Amount).Sub(decimal.NewFromFloat(subMinerFee)).Float64()
		}
	}
	if drawback > 0 {
		output[fromBitCoinAddress] = drawback
	}

	//output["data"] = "59756b692077696c6c20796f75206d61727279206d65203f2054657473752e59756b692077696c6c20796f75206d61727279206d65203f2054657473752e"

	hex, err = client.CreateRawTransaction(inputs, output)
	if err != nil {
		return "", "", err
	}

	log.Println("CreateRawTransaction", hex)

	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("CreateRawTransaction DecodeRawTransaction", decodeHex)

	if privkeys == nil || len(privkeys) == 0 {
		privkeys = nil
	}
	signHex, err := client.SignRawTransactionWithKey(hex, privkeys, inputs, "ALL")
	if err != nil {
		return "", "", err
	}
	log.Println("SignRawTransactionWithKey", signHex)

	hex = gjson.Get(signHex, "hex").String()
	log.Println("SignRawTransactionWithKey", hex)
	decodeHex, _ = client.DecodeRawTransaction(hex)
	txid = gjson.Get(decodeHex, "txid").String()
	log.Println("SignRawTransactionWithKey DecodeRawTransaction", decodeHex)
	return txid, hex, err
}

func (client *Client) BtcCreateAndSignRawTransactionForUnsendInputTx(fromBitCoinAddress string, privkeys []string, inputItems []TransactionInputItem, outputItems []TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (txid string, hex string, err error) {
	if len(fromBitCoinAddress) < 1 {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return "", "", errors.New("toBitCoinAddress is empty")
	}

	if minerFee <= config.Dust {
		minerFee = GetMinerFee()
	}

	outAmount := decimal.NewFromFloat(0)
	for _, item := range outputItems {
		outAmount = outAmount.Add(decimal.NewFromFloat(item.Amount))
	}

	if outAmount.LessThan(decimal.NewFromFloat(config.Dust)) {
		return "", "", errors.New("wrong outAmount")
	}

	if minerFee < config.Dust {
		return "", "", errors.New("minerFee too small")
	}

	var outTotalCount = 0
	for _, item := range outputItems {
		if item.Amount > 0 {
			outTotalCount++
		}
	}

	subMinerFee := 0.0
	if outTotalCount > 0 {
		count, _ := decimal.NewFromString(strconv.Itoa(outTotalCount))
		subMinerFee, _ = decimal.NewFromFloat(minerFee).DivRound(count, 8).Float64()
	}
	minerFee = 0
	for i := 0; i < outTotalCount; i++ {
		minerFee, _ = decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(subMinerFee)).Float64()
	}

	balance := 0.0
	var inputs []map[string]interface{}
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		if sequence > 0 {
			node["sequence"] = sequence
		}
		if tool.CheckIsString(redeemScript) {
			node["redeemScript"] = *redeemScript
		}
		node["scriptPubKey"] = item.ScriptPubKey
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Amount)).Float64()
		inputs = append(inputs, node)
	}
	//not enough money
	if outAmount.GreaterThan(decimal.NewFromFloat(balance)) {
		return "", "", errors.New("not enough balance")
	}

	//not enough money for minerFee
	if outAmount.Equal(decimal.NewFromFloat(balance)) {
		outAmount = outAmount.Sub(decimal.NewFromFloat(minerFee))
	}

	out, _ := decimal.NewFromFloat(minerFee).Add(outAmount).Float64()
	log.Println("input list ", inputs)

	if len(inputs) == 0 || balance < out {
		return "", "", errors.New("not enough balance")
	}
	drawback, _ := decimal.NewFromFloat(balance).Sub(decimal.NewFromFloat(out)).Float64()

	output := make(map[string]interface{})
	for _, item := range outputItems {
		if item.Amount > 0 {
			output[item.ToBitCoinAddress], _ = decimal.NewFromFloat(item.Amount).Sub(decimal.NewFromFloat(subMinerFee)).Float64()
		}
	}
	if drawback > 0 {
		output[fromBitCoinAddress] = drawback
	}
	hex, err = client.CreateRawTransaction(inputs, output)
	if err != nil {
		return "", "", err
	}

	log.Println("CreateRawTransaction", hex)

	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("CreateRawTransaction DecodeRawTransaction", decodeHex)

	if privkeys == nil || len(privkeys) == 0 {
		privkeys = nil
	}
	signHex, err := client.SignRawTransactionWithKey(hex, privkeys, inputs, "ALL")
	if err != nil {
		return "", "", err
	}
	log.Println("SignRawTransactionWithKey", signHex)

	hex = gjson.Get(signHex, "hex").String()
	log.Println("SignRawTransactionWithKey", hex)
	decodeHex, _ = client.DecodeRawTransaction(hex)
	txid = gjson.Get(decodeHex, "txid").String()
	log.Println("SignRawTransactionWithKey DecodeRawTransaction", decodeHex)
	return txid, hex, err
}
func (client *Client) BtcSignAndSendRawTransaction(hex string, privKey string) (string, string, error) {
	signHex, err := client.SignRawTransactionWithKey(hex, []string{privKey}, nil, "ALL")
	if err != nil {
		return "", "", err
	}
	hex = gjson.Get(signHex, "hex").String()
	decodeHex, err := client.DecodeRawTransaction(hex)
	txId := gjson.Get(decodeHex, "txid").String()
	log.Println("SignRawTransactionWithKey DecodeRawTransaction", decodeHex)
	//txId, err = client.SendRawTransaction(hex)
	if err != nil {
		return "", hex, err
	}
	log.Println("SendRawTransaction", txId)
	return txId, hex, nil
}

// create a transaction and signature and send to the network
func (client *Client) BtcCreateAndSignAndSendRawTransaction(fromBitCoinAddress string, privkeys []string, outputItems []TransactionOutputItem, minerFee float64, sequence int) (txId string, err error) {
	_, hex, err := client.BtcCreateAndSignRawTransaction(fromBitCoinAddress, privkeys, outputItems, minerFee, sequence, nil)
	if err != nil {
		return "", err
	}
	txId, err = client.SendRawTransaction(hex)
	if err != nil {
		return "", err
	}
	log.Println("SendRawTransaction", txId)
	return txId, nil
}
