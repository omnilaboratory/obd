package omnicore

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
)

//创建btc的raw交易：输入为未广播的预交易,输出为交易hex，支持单签和多签，如果是单签，就需要后续步骤再签名
func BtcCreateRawTransactionForUnsendInputTx(fromBitCoinAddress string, inputItems []bean.TransactionInputItem, outputItems []bean.TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (retMap map[string]interface{}, err error) {
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return nil, errors.New("toBitCoinAddress is empty")
	}

	if minerFee <= config.GetOmniDustBtc() {
		minerFee = GetMinerFee()
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

	dataToTracker := make(map[string]interface{})
	dataToTracker["inputs"] = inputs
	dataToTracker["outputs"] = output
	bytes, err := json.Marshal(dataToTracker)
	hex := conn2tracker.CreateRawTransaction(string(bytes))
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

// create a raw transaction and no sign , not send to the network,get the hash of signature
func BtcCreateRawTransaction(fromBitCoinAddress string, outputItems []bean.TransactionOutputItem, minerFee float64, sequence int, redeemScript *string) (retMap map[string]interface{}, err error) {
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return nil, errors.New("toBitCoinAddress is empty")
	}

	if minerFee <= 0 {
		minerFee = GetMinerFee()
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

	result := conn2tracker.ListUnspent(fromBitCoinAddress)
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
	hex := conn2tracker.CreateRawTransaction(string(bytes))
	if hex == "" {
		return nil, errors.New("error createRawTransaction")
	}

	retMap = make(map[string]interface{})
	retMap["hex"] = hex
	retMap["inputs"] = inputs
	retMap["total_in_amount"] = balance
	return retMap, nil
}

// From channelAddress to temp multi address, to Create CommitmentTx
func OmniCreateRawTransactionUseSingleInput(txType int, resultListUnspent string, fromBitCoinAddress string, toBitCoinAddress string, propertyId int64, amount float64, minerFee float64, sequence int, redeemScript *string, usedTxid string) (retMap map[string]interface{}, currUseTxid string, err error) {
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsAddress(toBitCoinAddress) == false {
		return nil, "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.GetOmniDustBtc() {
		return nil, "", errors.New("wrong amount")
	}

	pMoney := config.GetOmniDustBtc()

	arrayListUnspent := gjson.Parse(resultListUnspent).Array()

	balance := 0.0
	inputs := make([]map[string]interface{}, 0, 0)
	currUseTxid = ""
	for _, item := range arrayListUnspent {
		currUseTxid = item.Get("txid").String()
		if usedTxid != "" && strings.Contains(usedTxid, currUseTxid) {
			continue
		}
		inputAmount := item.Get("amount").Float()
		if inputAmount > pMoney {
			node := make(map[string]interface{})
			node["txid"] = item.Get("txid").String()
			node["vout"] = item.Get("vout").Int()
			node["amount"] = inputAmount
			node["scriptPubKey"] = item.Get("scriptPubKey").String()
			if redeemScript != nil {
				node["redeemScript"] = *redeemScript
			}
			balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(node["amount"].(float64))).Round(8).Float64()
			minerFee = tool.GetBtcMinerAmount(balance)
			inputs = append(inputs, node)
			break
		}
	}

	if currUseTxid == "" {
		return nil, "", errors.New("not found the miner fee input")
	}

	minMinerFee := config.GetMinMinerFee(len(inputs))
	if minerFee < minMinerFee {
		minerFee = minMinerFee
	}

	out, _ := decimal.NewFromFloat(pMoney).
		Add(decimal.NewFromFloat(minerFee)).
		Round(8).
		Float64()

	retMap, err = createOmniRawTransaction(balance, out, amount, minerFee, propertyId, inputs, toBitCoinAddress, toBitCoinAddress, redeemScript)
	if err != nil {
		return nil, "", err
	}
	return retMap, currUseTxid, nil
}

// 通道地址的剩余input全部花掉
func OmniCreateRawTransactionUseRestInput(txType int, resultListUnspent string, fromBitCoinAddress string, usedTxid string, toBitCoinAddress, changeToAddress string, propertyId int64, amount float64, minerFee float64, redeemScript *string) (retMap map[string]interface{}, err error) {
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsAddress(toBitCoinAddress) == false {
		return nil, errors.New("toBitCoinAddress is empty")
	}
	if amount < config.GetOmniDustBtc() {
		return nil, errors.New("wrong amount")
	}

	pMoney := config.GetOmniDustBtc()

	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	inputs := make([]map[string]interface{}, 0, 0)
	for _, item := range arrayListUnspent {
		txid := item.Get("txid").String()
		if (usedTxid != "" && strings.Contains(usedTxid, txid) == false) || len(usedTxid) == 0 {
			node := make(map[string]interface{})
			node["txid"] = txid
			node["vout"] = item.Get("vout").Int()
			node["amount"] = item.Get("amount").Float()
			node["scriptPubKey"] = item.Get("scriptPubKey").String()
			if redeemScript != nil {
				node["redeemScript"] = *redeemScript
			}
			inputs = append(inputs, node)
		}
	}

	minMinerFee := config.GetMinMinerFee(len(inputs))
	if minerFee < minMinerFee {
		minerFee = minMinerFee
	}
	out, _ := decimal.NewFromFloat(minerFee).
		Add(decimal.NewFromFloat(pMoney)).
		Round(8).
		Float64()

	balance := 0.0
	for _, item := range inputs {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item["amount"].(float64))).Round(8).Float64()
		if balance >= out {
			break
		}
	}
	return createOmniRawTransaction(balance, out, amount, minerFee, propertyId, inputs, toBitCoinAddress, changeToAddress, redeemScript)
}

func OmniCreateRawTransactionUseUnsendInput(fromBitCoinAddress string, inputItems []bean.TransactionInputItem, toBitCoinAddress, changeToAddress string, propertyId int64, amount float64, minerFee float64, sequence int, redeemScript *string) (retMap map[string]interface{}, err error) {
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsAddress(toBitCoinAddress) == false {
		return nil, errors.New("toBitCoinAddress is empty")
	}
	if tool.CheckIsAddress(changeToAddress) == false {
		return nil, errors.New("changeToAddress is empty")
	}

	if len(inputItems) == 0 {
		return nil, errors.New("inputItems is empty")
	}

	if amount < config.GetOmniDustBtc() {
		return nil, errors.New("wrong amount")
	}
	pMoney := config.GetOmniDustBtc()

	inputs := make([]map[string]interface{}, 0, 0)
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		node["amount"] = item.Amount
		node["scriptPubKey"] = item.ScriptPubKey
		if sequence > 0 {
			node["sequence"] = sequence
		}
		if redeemScript != nil {
			node["redeemScript"] = *redeemScript
		}

		inputs = append(inputs, node)
	}
	out, _ := decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(pMoney)).Round(8).Float64()
	balance := 0.0
	for _, item := range inputs {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item["amount"].(float64))).Round(8).Float64()
		if balance >= out {
			break
		}
	}

	return createOmniRawTransaction(balance, out, amount, minerFee, propertyId, inputs, toBitCoinAddress, changeToAddress, redeemScript)
}

// From channelAddress to temp multi address, to Create CommitmentTx
func GetInputInfo(fromBitCoinAddress string, txid, redeemScript string) (info map[string]interface{}, err error) {
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, errors.New("fromBitCoinAddress is empty")
	}

	resultListUnspent := conn2tracker.ListUnspent(fromBitCoinAddress)
	if resultListUnspent == "" {
		return nil, err
	}

	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	for _, item := range arrayListUnspent {
		currUseTxid := item.Get("txid").String()
		if currUseTxid == txid {
			node := make(map[string]interface{})
			node["txid"] = currUseTxid
			node["vout"] = item.Get("vout").Int()
			node["amount"] = item.Get("amount").Float()
			node["scriptPubKey"] = item.Get("scriptPubKey").String()
			if &redeemScript != nil {
				node["redeemScript"] = redeemScript
			}
			return node, nil
		}
	}

	return nil, errors.New("not found input info")
}

func OmniCreateRawTransaction(fromBitCoinAddress string, toBitCoinAddress string, propertyId int64, amount float64, minerFee float64) (retMap map[string]interface{}, err error) {
	if tool.CheckIsAddress(fromBitCoinAddress) == false {
		return nil, errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsAddress(toBitCoinAddress) == false {
		return nil, errors.New("toBitCoinAddress is empty")
	}
	if amount < config.GetOmniDustBtc() {
		return nil, errors.New("wrong amount")
	}

	_, err = ParsePropertyId(strconv.Itoa(int(propertyId)))
	if err != nil {
		return nil, err
	}

	pMoney := config.GetOmniDustBtc()
	if minerFee < config.GetOmniDustBtc() {
		minerFee = 0.00003
	}

	balanceResult := conn2tracker.OmniGetBalancesForAddress(fromBitCoinAddress, int(propertyId))
	if balanceResult == "" {
		return nil, errors.New("empty omni balance")
	}

	omniBalance := gjson.Get(balanceResult, "balance").Float()
	if omniBalance < amount {
		return nil, errors.New("not enough omni balance")
	}

	resultListUnspent := conn2tracker.ListUnspent(fromBitCoinAddress)
	if resultListUnspent == "" {
		return nil, errors.New("enmpty ListUnspent")
	}

	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	if len(arrayListUnspent) == 0 {
		return nil, errors.New("empty balance")
	}

	out, _ := decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(pMoney)).Round(8).Float64()

	balance := 0.0
	inputs := make([]map[string]interface{}, 0, len(arrayListUnspent))
	for _, item := range arrayListUnspent {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		node["scriptPubKey"] = item.Get("scriptPubKey").String()
		node["amount"] = item.Get("amount").Float()
		inputs = append(inputs, node)
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Round(8).Float64()
		if balance >= out {
			break
		}
	}

	retMap, err = createOmniRawTransaction(balance, out, amount, minerFee, propertyId, inputs, toBitCoinAddress, fromBitCoinAddress, nil)
	if err != nil {
		return nil, err
	}
	return retMap, nil
}

func createOmniRawTransaction(balance, out, amount, minerFee float64, propertyId int64, inputs []map[string]interface{}, toBitCoinAddress, changeToAddress string, redeemScript *string) (retMap map[string]interface{}, err error) {
	if balance < out {
		return nil, errors.New("not enough balance")
	}

	payloadBytes, payloadHex := Omni_createpayload_simplesend(strconv.Itoa(int(propertyId)), tool.FloatToString(amount, 8), true)

	inputsBytes, _ := json.Marshal(inputs)
	inputsStr := string(inputsBytes)
	inputsStr = strings.TrimLeft(inputsStr, "[")
	inputsStr = strings.TrimRight(inputsStr, "]")
	inputsStr = strings.ReplaceAll(inputsStr, "},", "}")
	rawTx, _, _ := CreateRawTransaction(inputsStr, 2)

	opreturnTxMsg, err := Omni_createrawtx_opreturn(rawTx, payloadBytes, payloadHex)

	referenceTxMsg, err := Omni_createrawtx_reference(opreturnTxMsg, toBitCoinAddress, tool.GetCoreNet())

	//6.Omni_createrawtx_change
	prevtxs := make([]map[string]interface{}, 0, 0)
	for _, item := range inputs {
		node := make(map[string]interface{})
		node["txid"] = item["txid"]
		node["vout"] = item["vout"]
		node["scriptPubKey"] = item["scriptPubKey"]
		value := decimal.NewFromFloat(item["amount"].(float64))
		node["value"] = value.String()
		if redeemScript != nil {
			node["redeemScript"] = *redeemScript
		}
		prevtxs = append(prevtxs, node)
	}

	prevtxsBytes, _ := json.Marshal(prevtxs)
	prevtxsStr := string(prevtxsBytes)
	prevtxsStr = strings.TrimLeft(prevtxsStr, "[")
	prevtxsStr = strings.TrimRight(prevtxsStr, "]")
	prevtxsStr = strings.ReplaceAll(prevtxsStr, "},", "}")
	change, err := Omni_createrawtx_change(referenceTxMsg, prevtxsStr, changeToAddress, tool.FloatToString(minerFee, 8), tool.GetCoreNet())

	retMap = make(map[string]interface{})
	retMap["hex"] = TxToHex(change)
	retMap["inputs"] = inputs
	return retMap, nil
}
