package rpc

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"math"
	"strconv"
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

type multiSign struct {
	Address      string `json:"address"`
	RedeemScript string `json:"redeemScript"`
	ScriptPubKey string `json:"scriptPubKey"`
}

func (client *Client) CreateMultiSig(minSignNum int, keys []string) (result string, err error) {
	addr, redeemScript, scriptPubKey := omnicore.CreateMultiSigAddr(keys[0], keys[1], tool.GetCoreNet())
	sign := multiSign{}
	sign.Address = addr
	sign.RedeemScript = redeemScript
	sign.ScriptPubKey = scriptPubKey
	marshal, _ := json.Marshal(&sign)
	return string(marshal), nil

}

func (client *Client) GetTransactionById(txid string) (result string, err error) {
	return client.send("gettransaction", []interface{}{txid})
}

func (client *Client) ListReceivedByAddress(address string) (result string, err error) {
	_, _ = client.ValidateAddress(address)
	return client.send("listreceivedbyaddress", []interface{}{0, false, true, address})
}

//https://bitcoin.org/en/developer-reference#testmempoolaccept
func (client *Client) TestMemPoolAccept(signedhex string) (result string, err error) {
	rawtxs := make([]string, 0)
	rawtxs = append(rawtxs, signedhex)
	result, err = client.send("testmempoolaccept", []interface{}{rawtxs})
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

//https://developer.bitcoin.org/reference/rpc/createrawtransaction.html
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

func (client *Client) DecodeRawTransaction(hex string) (result string, err error) {
	result = omnicore.DecodeRawTransaction(hex, tool.GetCoreNet())
	return result, err
}

func (client *Client) OmniDecodeTransaction(hex string) (result string, err error) {
	result, err = client.send("omni_decodetransaction", []interface{}{hex})
	return result, err
}

func (client *Client) OmniDecodeTransactionWithPrevTxs(hex string, prevtxs []TransactionInputItem) (result string, err error) {
	result, err = client.send("omni_decodetransaction", []interface{}{hex, prevtxs})
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
	if tool.CheckIsAddress(address) == false {
		return false, errors.New("address not exist")
	}
	_, _ = client.send("importaddress", []interface{}{address, "", false})
	return true, nil
}

func (client *Client) GetAddressInfo(address string) (result string, err error) {
	result, err = client.send("getaddressinfo", []interface{}{address})
	if err != nil {
		return "", err
	}
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
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return nil, errors.New("toBitCoinAddress is empty")
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

	result := conn.HttpListUnspentFromTracker(fromBitCoinAddress)
	array := gjson.Parse(result).Array()
	if len(array) == 0 {
		return nil, errors.New("empty balance")
	}

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

	dataToTracker := make(map[string]interface{})
	dataToTracker["inputs"] = inputs
	dataToTracker["outputs"] = output
	bytes, err := json.Marshal(dataToTracker)
	hex := conn.HttpCreateRawTransactionFromTracker(string(bytes))
	if hex == "" {
		return nil, errors.New("error createRawTransaction")
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

//创建btc的raw交易：输入为未广播的预交易,输出为交易hex，支持单签和多签，如果是单签，就需要后续步骤再签名
func (client *Client) BtcCreateRawTransactionForUnsendInputTx(fromBitCoinAddress string, inputItems []TransactionInputItem, outputItems []TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (retMap map[string]interface{}, err error) {
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return nil, errors.New("toBitCoinAddress is empty")
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

	dataToTracker := make(map[string]interface{})
	dataToTracker["inputs"] = inputs
	dataToTracker["outputs"] = output
	bytes, err := json.Marshal(dataToTracker)
	hex = conn.HttpCreateRawTransactionFromTracker(string(bytes))
	if hex == "" {
		return nil, errors.New("error createRawTransaction")
	}

	retMap = make(map[string]interface{})
	retMap["hex"] = hex
	retMap["inputs"] = inputs
	retMap["total_in_amount"] = balance
	retMap["total_out_amount"] = totalOutAmount
	return retMap, nil
}
