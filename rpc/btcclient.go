package rpc

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"math"
	"strconv"
	"time"
)

//https://developer.bitcoin.org/reference/rpc/index.html
func (client *Client) GetBlockChainInfo() (result string, err error) {
	return client.send("getblockchaininfo", nil)
}

//https://developer.bitcoin.org/reference/rpc/estimatesmartfee.html
func (client *Client) EstimateSmartFee() (feeRate float64) {
	result, err := client.send("estimatesmartfee", []interface{}{10})
	if err == nil {
		return gjson.Get(result, "feerate").Float() * 100000
	}
	return 2
}
func (client *Client) CreateMultiSig(minSignNum int, keys []string) (result string, err error) {
	return client.send("createmultisig", []interface{}{minSignNum, keys})
}

func (client *Client) AddMultiSigAddress(minSignNum int, keys []string) (result string, err error) {
	for _, item := range keys {
		_, _ = client.ValidateAddress(item)
	}
	return client.send("addmultisigaddress", []interface{}{minSignNum, keys})
}

func (client *Client) DumpPrivKey(address string) (result string, err error) {
	_, err = client.ValidateAddress(address)
	if err != nil {
		return "", err
	}
	return client.send("dumpprivkey", []interface{}{address})
}

func (client *Client) GetTransactionById(txid string) (result string, err error) {
	return client.send("gettransaction", []interface{}{txid})
}

func (client *Client) ListReceivedByAddress(address string) (result string, err error) {
	_, _ = client.ValidateAddress(address)
	return client.send("listreceivedbyaddress", []interface{}{0, false, true, address})
}

var tempTestMemPoolAcceptData map[string]string

//https://bitcoin.org/en/developer-reference#testmempoolaccept
func (client *Client) TestMemPoolAccept(signedhex string) (result string, err error) {
	if tempTestMemPoolAcceptData == nil {
		tempTestMemPoolAcceptData = make(map[string]string)
	}
	hexHash := tool.SignMsgWithSha256([]byte(signedhex))
	if len(tempTestMemPoolAcceptData[hexHash]) > 0 {
		return tempTestMemPoolAcceptData[hexHash], nil
	}
	rawtxs := make([]string, 0)
	rawtxs = append(rawtxs, signedhex)
	result, err = client.send("testmempoolaccept", []interface{}{rawtxs})
	if err == nil {
		tempTestMemPoolAcceptData[hexHash] = result
	}
	return result, err
}

func (client *Client) GetTxOut(txid string, num int) (result string, err error) {
	return client.send("gettxout", []interface{}{txid, num})
}

func (client *Client) GetNewAddress(label string) (address string, err error) {
	return client.send("getnewaddress", []interface{}{label})
}

func (client *Client) ListUnspent(address string) (result string, err error) {
	_, err = client.ValidateAddress(address)
	if err != nil {
		return "", err
	}
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

//https://developer.bitcoin.org/reference/rpc/signrawtransactionwithkey.html
func (client *Client) SignRawTransactionWithKey(hex string, privkeys []string, prevtxs []map[string]interface{}, sighashtype string) (result string, err error) {
	return client.send("signrawtransactionwithkey", []interface{}{hex, privkeys, prevtxs, sighashtype})
}

func (client *Client) SendRawTransaction(hex string) (result string, err error) {
	return client.send("sendrawtransaction", []interface{}{hex})
}

var tempDecodeRawTransactionData map[string]string

func (client *Client) DecodeRawTransaction(hex string) (result string, err error) {

	if tempDecodeRawTransactionData == nil {
		tempDecodeRawTransactionData = make(map[string]string)
	}
	hexHash := tool.SignMsgWithSha256([]byte(hex))
	if len(tempDecodeRawTransactionData[hexHash]) > 0 {
		return tempDecodeRawTransactionData[hexHash], nil
	}

	result, err = client.send("decoderawtransaction", []interface{}{hex})
	if err == nil {
		tempDecodeRawTransactionData[hexHash] = result
	}
	return result, err
}

var tempOmniDecodeTransactionData map[string]string

func (client *Client) OmniDecodeTransaction(hex string) (result string, err error) {

	if tempOmniDecodeTransactionData == nil {
		tempOmniDecodeTransactionData = make(map[string]string)
	}
	hexHash := tool.SignMsgWithSha256([]byte(hex))
	if len(tempOmniDecodeTransactionData[hexHash]) > 0 {
		return tempOmniDecodeTransactionData[hexHash], nil
	}
	result, err = client.send("omni_decodetransaction", []interface{}{hex})
	if err == nil {
		tempOmniDecodeTransactionData[hexHash] = result
	}

	return result, err
}

var tempOmniDecodeTransactionData2 map[string]string

func (client *Client) OmniDecodeTransactionWithPrevTxs(hex string, prevtxs []TransactionInputItem) (result string, err error) {

	if tempOmniDecodeTransactionData2 == nil {
		tempOmniDecodeTransactionData2 = make(map[string]string)
	}
	hexHash := tool.SignMsgWithSha256([]byte(hex))
	if len(tempOmniDecodeTransactionData2[hexHash]) > 0 {
		return tempOmniDecodeTransactionData2[hexHash], nil
	}
	result, err = client.send("omni_decodetransaction", []interface{}{hex, prevtxs})
	if err == nil {
		tempOmniDecodeTransactionData2[hexHash] = result
	}
	return result, err
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
	_, err = client.ValidateAddress(address)
	if err != nil {
		return "", err
	}
	return client.send("verifymessage", []interface{}{address, signature, message})
}
func (client *Client) DecodeScript(hexString string) (result string, err error) {
	return client.send("decodescript", []interface{}{hexString})
}

func (client *Client) ValidateAddress(address string) (isValid bool, err error) {
	if tool.CheckIsString(&address) == false {
		return false, errors.New("address not exist")
	}

	if validatedAddress[address] {
		return true, nil
	}

	result, err := client.GetAddressInfo(address)
	if err != nil {
		return false, err
	}
	if gjson.Get(result, "iswatchonly").Bool() == false {
		_, _ = client.send("importaddress", []interface{}{address, "", false})
	}
	//log.Println(result)
	validatedAddress[address] = true

	return gjson.Get(result, "iswatchonly").Bool(), nil
}

var tempGetAddressInfoMap map[string]string

func (client *Client) GetAddressInfo(address string) (result string, err error) {

	if tempGetAddressInfoMap == nil {
		tempGetAddressInfoMap = make(map[string]string)
	}
	if len(tempGetAddressInfoMap[address]) > 0 {
		return tempGetAddressInfoMap[address], nil
	}

	result, err = client.send("getaddressinfo", []interface{}{address})
	if err != nil {
		return "", err
	}

	tempGetAddressInfoMap[address] = result

	return result, nil
}

func (client *Client) ImportPrivKey(privkey string) (result string, err error) {
	return client.send("importprivkey", []interface{}{privkey, "", false})
}

type TransactionOutputItem struct {
	ToBitCoinAddress string
	Amount           float64
}
type TransactionInputItem struct {
	Txid         string  `json:"txid"`
	ScriptPubKey string  `json:"scriptPubKey"`
	RedeemScript string  `json:"redeem_script"`
	Vout         uint32  `json:"vout"`
	Amount       float64 `json:"value"`
}

// create a raw transaction and no sign , not send to the network,get the hash of signature
func (client *Client) BtcCreateRawTransaction(fromBitCoinAddress string, outputItems []TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (retMap map[string]interface{}, err error) {
	if len(fromBitCoinAddress) < 1 {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return nil, errors.New("toBitCoinAddress is empty")
	}

	_, err = client.ValidateAddress(fromBitCoinAddress)
	if err != nil {
		return nil, err
	}

	if minerFee <= 0 {
		minerFee = client.GetMinerFee()
	}

	outTotalAmount := decimal.NewFromFloat(0)
	for _, item := range outputItems {
		outTotalAmount = outTotalAmount.Add(decimal.NewFromFloat(item.Amount))
	}

	if outTotalAmount.LessThan(decimal.NewFromFloat(config.GetOmniDustBtc())) {
		return nil, errors.New("wrong outTotalAmount")
	}

	if minerFee < config.GetOmniDustBtc() {
		return nil, errors.New("minerFee too small")
	}

	result, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return nil, err
	}

	array := gjson.Parse(result).Array()
	if len(array) == 0 {
		return nil, errors.New("empty balance")
	}
	log.Println("listunspent", array)

	out, _ := outTotalAmount.Round(8).Float64()

	balance := 0.0
	inputs := make([]map[string]interface{}, 0)
	for _, item := range array {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		node["amount"] = item.Get("amount").Float()
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
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Round(8).Float64()
		if balance > out {
			break
		}
	}

	if len(inputs) == 0 || balance < out {
		return nil, errors.New("not enough balance")
	}

	minerFeeAndOut, _ := decimal.NewFromFloat(minerFee).Add(outTotalAmount).Round(8).Float64()

	subMinerFee := 0.0
	if balance <= minerFeeAndOut {
		needLessFee, _ := decimal.NewFromFloat(minerFeeAndOut).Sub(decimal.NewFromFloat(balance)).Round(8).Float64()
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

	drawback, _ := decimal.NewFromFloat(balance).Sub(decimal.NewFromFloat(minerFeeAndOut)).Round(8).Float64()
	output := make(map[string]interface{})
	for _, item := range outputItems {
		if item.Amount > 0 {
			output[item.ToBitCoinAddress], _ = decimal.NewFromFloat(item.Amount).Sub(decimal.NewFromFloat(subMinerFee)).Round(8).Float64()
		}
	}
	if drawback > 0 {
		output[fromBitCoinAddress] = drawback
	}
	hex, err := client.CreateRawTransaction(inputs, output)
	if err != nil {
		return nil, err
	}
	retMap = make(map[string]interface{})
	retMap["hex"] = hex
	retMap["inputs"] = inputs
	retMap["total_in_amount"] = balance
	return retMap, nil
}

type NeedSignData struct {
	Hex    string                   `json:"hex"`
	Prvkey string                   `json:"prvkey"`
	Inputs []map[string]interface{} `json:"inputs"`
}

func (client *Client) BtcSignRawTransactionFromJson(dataJson string) (signHex string, err error) {
	inputData := &NeedSignData{}
	err = json.Unmarshal([]byte(dataJson), inputData)
	if err != nil {
		return "", err
	}

	signHexObj, err := client.SignRawTransactionWithKey(inputData.Hex, []string{inputData.Prvkey}, inputData.Inputs, "ALL")
	if err != nil {
		return "", err
	}
	signHex = gjson.Get(signHexObj, "hex").String()
	return signHex, nil
}

// create a transaction and just signnature , not send to the network,get the hash of signature
func (client *Client) BtcCreateAndSignRawTransaction(fromBitCoinAddress string, privkeys []string, outputItems []TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (txid string, hex string, err error) {
	if len(fromBitCoinAddress) < 1 {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return "", "", errors.New("toBitCoinAddress is empty")
	}

	_, err = client.ValidateAddress(fromBitCoinAddress)
	if err != nil {
		return "", "", err
	}

	if minerFee <= 0 {
		minerFee = client.GetMinerFee()
	}

	outTotalAmount := decimal.NewFromFloat(0)
	for _, item := range outputItems {
		outTotalAmount = outTotalAmount.Add(decimal.NewFromFloat(item.Amount))
	}

	if outTotalAmount.LessThan(decimal.NewFromFloat(config.GetOmniDustBtc())) {
		return "", "", errors.New("wrong outTotalAmount")
	}

	if minerFee < config.GetOmniDustBtc() {
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
	out, _ := outTotalAmount.Round(8).Float64()

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
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Round(8).Float64()
	}
	log.Println("input list ", inputs)

	if len(inputs) == 0 || balance < out {
		return "", "", errors.New("not enough balance")
	}

	minerFeeAndOut, _ := decimal.NewFromFloat(minerFee).Add(outTotalAmount).Round(8).Float64()

	subMinerFee := 0.0
	if balance <= minerFeeAndOut {
		needLessFee, _ := decimal.NewFromFloat(minerFeeAndOut).Sub(decimal.NewFromFloat(balance)).Round(8).Float64()
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

	drawback, _ := decimal.NewFromFloat(balance).Sub(decimal.NewFromFloat(minerFeeAndOut)).Round(8).Float64()
	output := make(map[string]interface{})
	for _, item := range outputItems {
		if item.Amount > 0 {
			output[item.ToBitCoinAddress], _ = decimal.NewFromFloat(item.Amount).Sub(decimal.NewFromFloat(subMinerFee)).Round(8).Float64()
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

//创建btc的raw交易：输入为未广播的预交易,输出为交易hex，支持单签和多签，如果是单签，就需要后续步骤再签名
func (client *Client) BtcCreateRawTransactionForUnsendInputTx(fromBitCoinAddress string, inputItems []TransactionInputItem, outputItems []TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (retMap map[string]interface{}, err error) {
	if len(fromBitCoinAddress) < 1 {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return nil, errors.New("toBitCoinAddress is empty")
	}

	_, err = client.ValidateAddress(fromBitCoinAddress)
	if err != nil {
		return nil, err
	}

	if minerFee <= config.GetOmniDustBtc() {
		minerFee = client.GetMinerFee()
	}

	outAmount := decimal.NewFromFloat(0)
	for _, item := range outputItems {
		outAmount = outAmount.Add(decimal.NewFromFloat(item.Amount))
	}

	if outAmount.LessThan(decimal.NewFromFloat(config.GetOmniDustBtc())) {
		return nil, errors.New("wrong outAmount")
	}

	if minerFee < config.GetOmniDustBtc() {
		return nil, errors.New("minerFee too small")
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
		minerFee, _ = decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(subMinerFee)).Round(8).Float64()
	}

	balance := 0.0
	var inputs []map[string]interface{}
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		node["amount"] = item.Amount
		if sequence > 0 {
			node["sequence"] = sequence
		}
		if tool.CheckIsString(redeemScript) {
			node["redeemScript"] = *redeemScript
		}
		node["scriptPubKey"] = item.ScriptPubKey
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Amount)).Round(8).Float64()
		inputs = append(inputs, node)
	}
	//not enough money
	if outAmount.GreaterThan(decimal.NewFromFloat(balance)) {
		return nil, errors.New("not enough balance")
	}

	//not enough money for minerFee
	if outAmount.Equal(decimal.NewFromFloat(balance)) {
		outAmount = outAmount.Sub(decimal.NewFromFloat(minerFee))
	}

	out, _ := decimal.NewFromFloat(minerFee).Add(outAmount).Round(8).Float64()
	log.Println("input list ", inputs)

	if len(inputs) == 0 || balance < out {
		return nil, errors.New("not enough balance")
	}
	drawback, _ := decimal.NewFromFloat(balance).Sub(decimal.NewFromFloat(out)).Round(8).Float64()

	output := make(map[string]interface{})
	totalOutAmount := 0.0
	for _, item := range outputItems {
		if item.Amount > 0 {
			tempAmount, _ := decimal.NewFromFloat(item.Amount).Sub(decimal.NewFromFloat(subMinerFee)).Round(8).Float64()
			output[item.ToBitCoinAddress] = tempAmount
			totalOutAmount += tempAmount
		}
	}
	if drawback > 0 {
		output[fromBitCoinAddress] = drawback
	}
	hex, err := client.CreateRawTransaction(inputs, output)
	if err != nil {
		return nil, err
	}
	retMap = make(map[string]interface{})
	retMap["hex"] = hex
	retMap["inputs"] = inputs
	retMap["total_in_amount"] = balance
	retMap["total_out_amount"] = totalOutAmount
	return retMap, nil
}

//创建btc的raw交易：输入为未广播的预交易,输出为交易hex，支持单签和多签，如果是单签，就需要后续步骤再签名
func (client *Client) BtcCreateAndSignRawTransactionForUnsendInputTx(fromBitCoinAddress string, privkeys []string, inputItems []TransactionInputItem, outputItems []TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (txid string, hex string, err error) {
	beginTime := time.Now()
	if len(fromBitCoinAddress) < 1 {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return "", "", errors.New("toBitCoinAddress is empty")
	}

	_, err = client.ValidateAddress(fromBitCoinAddress)
	if err != nil {
		return "", "", err
	}

	if minerFee <= config.GetOmniDustBtc() {
		minerFee = client.GetMinerFee()
	}

	outAmount := decimal.NewFromFloat(0)
	for _, item := range outputItems {
		outAmount = outAmount.Add(decimal.NewFromFloat(item.Amount))
	}

	if outAmount.LessThan(decimal.NewFromFloat(config.GetOmniDustBtc())) {
		return "", "", errors.New("wrong outAmount")
	}

	if minerFee < config.GetOmniDustBtc() {
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
		minerFee, _ = decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(subMinerFee)).Round(8).Float64()
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
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Amount)).Round(8).Float64()
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

	out, _ := decimal.NewFromFloat(minerFee).Add(outAmount).Round(8).Float64()
	log.Println("input list ", inputs)

	if len(inputs) == 0 || balance < out {
		return "", "", errors.New("not enough balance")
	}
	drawback, _ := decimal.NewFromFloat(balance).Sub(decimal.NewFromFloat(out)).Round(8).Float64()

	output := make(map[string]interface{})
	for _, item := range outputItems {
		if item.Amount > 0 {
			output[item.ToBitCoinAddress], _ = decimal.NewFromFloat(item.Amount).Sub(decimal.NewFromFloat(subMinerFee)).Round(8).Float64()
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
	log.Println("endTime.Sub(beginTime)", time.Now().Sub(beginTime).String())

	return txid, hex, err
}

func (client *Client) BtcSignRawTransactionForUnsend(hex string, inputItems []TransactionInputItem, privKey string) (string, string, error) {

	var inputs []map[string]interface{}
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		node["scriptPubKey"] = item.ScriptPubKey
		node["redeemScript"] = item.RedeemScript
		inputs = append(inputs, node)
	}
	signHex, err := client.SignRawTransactionWithKey(hex, []string{privKey}, inputs, "ALL")
	if err != nil {
		return "", "", err
	}
	hex = gjson.Get(signHex, "hex").String()
	decodeHex, err := client.DecodeRawTransaction(hex)
	txId := gjson.Get(decodeHex, "txid").String()
	if err != nil {
		return "", hex, err
	}
	return txId, hex, nil
}

func (client *Client) BtcSignRawTransaction(hex string, privKey string) (string, string, error) {
	signHex, err := client.SignRawTransactionWithKey(hex, []string{privKey}, nil, "ALL")
	if err != nil {
		return "", "", err
	}
	hex = gjson.Get(signHex, "hex").String()
	decodeHex, err := client.DecodeRawTransaction(hex)
	if err != nil {
		log.Println(err)
	}
	txId := gjson.Get(decodeHex, "txid").String()
	if err != nil {
		return "", hex, err
	}

	//result, err := client.OmniDecodeTransaction(hex)
	//if err == nil {
	//	log.Println(result)
	//} else {
	//	log.Println(err)
	//}

	return txId, hex, nil
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
