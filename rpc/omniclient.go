package rpc

import (
	"errors"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"strings"
	"time"
)

//https://github.com/OmniLayer/omnicore/blob/master/src/omnicore/doc/rpc-api.md
//Returns various state information of the client and protocol.
func (client *Client) OmniGetInfo() (result string, err error) {
	return client.send("omni_getinfo", nil)
}

//Create and broadcast a simple send transaction.
func (client *Client) OmniSend(fromAddress string, toAddress string, propertyId int, amount float64) (result string, err error) {
	_, err = client.ValidateAddress(fromAddress)
	if err != nil {
		return "", err
	}
	_, err = client.ValidateAddress(toAddress)
	if err != nil {
		return "", err
	}
	return client.send("omni_send", []interface{}{fromAddress, toAddress, propertyId, amount})
}

//Create new tokens with manageable supply. https://github.com/OmniLayer/omnicore/blob/master/src/omnicore/doc/rpc-api.md#omni_sendissuancemanaged
func (client *Client) OmniSendIssuanceFixed(fromAddress string, ecosystem int, divisibleType int, name string, data string, amount float64) (result string, err error) {
	if tool.CheckIsString(&fromAddress) == false {
		return "", errors.New("error fromAddress")
	}
	if tool.CheckIsString(&name) == false {
		return "", errors.New("error name")
	}
	if ecosystem != 1 && ecosystem != 2 {
		return "", errors.New("error ecosystem")
	}
	if divisibleType != 1 && divisibleType != 2 {
		return "", errors.New("error divisibleType")
	}
	if amount < 0 {
		return "", errors.New("error amount")
	}
	amountStr := strconv.FormatFloat(amount, 'f', 8, 64)
	if divisibleType == 1 {
		amountStr = strconv.FormatFloat(amount, 'f', 0, 64)
	}
	if tool.CheckIsString(&data) == false {
		data = ""
	}

	_, _ = client.ValidateAddress(fromAddress)

	return client.send("omni_sendissuancefixed", []interface{}{fromAddress, ecosystem, divisibleType, 0, "", "", name, "", data, amountStr})
}

//Create new tokens with manageable supply. https://github.com/OmniLayer/omnicore/blob/master/src/omnicore/doc/rpc-api.md#omni_sendissuancemanaged
func (client *Client) OmniSendIssuanceManaged(fromAddress string, ecosystem int, divisibleType int, name string, data string) (result string, err error) {
	if tool.CheckIsString(&fromAddress) == false {
		return "", errors.New("error fromAddress")
	}
	if tool.CheckIsString(&name) == false {
		return "", errors.New("error name")
	}
	if ecosystem != 1 && ecosystem != 2 {
		return "", errors.New("error ecosystem")
	}
	if divisibleType != 1 && divisibleType != 2 {
		return "", errors.New("error divisibleType")
	}
	if tool.CheckIsString(&data) == false {
		data = ""
	}
	_, _ = client.ValidateAddress(fromAddress)
	return client.send("omni_sendissuancemanaged", []interface{}{fromAddress, ecosystem, divisibleType, 0, "", "", name, "", data})
}

// Issue or grant new units of managed tokens. https://github.com/OmniLayer/omnicore/blob/master/src/omnicore/doc/rpc-api.md#omni_sendgrant
func (client *Client) OmniSendGrant(fromAddress string, propertyId int64, amount float64, memo string) (result string, err error) {
	if tool.CheckIsString(&fromAddress) == false {
		return "", errors.New("error fromAddress")
	}
	if propertyId < 0 {
		return "", errors.New("error propertyId")
	}

	s, err := client.OmniGetProperty(propertyId)
	if err != nil {
		return "", err
	}
	if amount < 0 {
		return "", errors.New("error amout")
	}
	amountStr := strconv.FormatFloat(amount, 'f', 0, 64)
	divisible := gjson.Get(s, "divisible").Bool()
	if divisible {
		amountStr = strconv.FormatFloat(amount, 'f', 8, 64)
	}
	if tool.CheckIsString(&memo) == false {
		memo = ""
	}
	return client.send("omni_sendgrant", []interface{}{fromAddress, "", propertyId, amountStr, memo})
}

// Revoke units of managed tokens. https://github.com/OmniLayer/omnicore/blob/master/src/omnicore/doc/rpc-api.md#omni_sendrevoke
func (client *Client) OmniSendRevoke(fromAddress string, propertyId int64, amount float64, memo string) (result string, err error) {
	if tool.CheckIsString(&fromAddress) == false {
		return "", errors.New("error fromAddress")
	}
	if propertyId < 0 {
		return "", errors.New("error propertyId")
	}
	if amount < 0 {
		return "", errors.New("error amout")
	}

	s, err := client.OmniGetProperty(propertyId)
	if err != nil {
		return "", err
	}
	amountStr := strconv.FormatFloat(amount, 'f', 0, 64)
	divisible := gjson.Get(s, "divisible").Bool()
	if divisible {
		amountStr = strconv.FormatFloat(amount, 'f', 8, 64)
	}

	if tool.CheckIsString(&memo) == false {
		memo = ""
	}

	_, _ = client.ValidateAddress(fromAddress)

	return client.send("omni_sendrevoke", []interface{}{fromAddress, propertyId, amountStr, memo})
}

func (client *Client) OmniGetbalance(address string, propertyId int) (result string, err error) {
	_, err = client.ValidateAddress(address)
	if err != nil {
		return "", err
	}
	return client.send("omni_getbalance", []interface{}{address, propertyId})
}

//Get detailed information about an Omni transaction.
func (client *Client) OmniGettransaction(txid string) (result string, err error) {
	return client.send("omni_gettransaction", []interface{}{txid})
}

//Get detailed information about an Omni transaction.
func (client *Client) OmniGetProperty(propertyId int64) (result string, err error) {
	return client.send("omni_getproperty", []interface{}{propertyId})
}

//Get detailed information about an Omni transaction.
func (client *Client) OmniListProperties() (result string, err error) {
	return client.send("omni_listproperties", []interface{}{})
}

func (client *Client) OmniGetAllBalancesForAddress(address string) (result string, err error) {
	_, err = client.ValidateAddress(address)
	if err != nil {
		return "", err
	}

	return client.send("omni_getallbalancesforaddress ", []interface{}{address})
}

//List wallet transactions, optionally filtered by an address and block boundaries.
//https://github.com/OmniLayer/omnicore/blob/master/src/omnicore/doc/rpc-api.md#omni_listtransactions
func (client *Client) OmniListTransactions(address string, count int, skip int) (result string, err error) {
	_, _ = client.ValidateAddress(address)
	if tool.CheckIsString(&address) == false {
		address = "*"
	}
	if count < 0 {
		count = 10
	}
	if skip < 0 {
		skip = 0
	}
	return client.send("omni_listtransactions", []interface{}{address, count, skip})
}

func (client *Client) omniCreatePayloadSimpleSend(propertyId int64, amount float64) (result string, err error) {
	return client.send("omni_createpayload_simplesend", []interface{}{propertyId, decimal.NewFromFloat(amount).String()})
}
func (client *Client) omniCreateRawtxOpreturn(rawtx string, payload string) (result string, err error) {
	return client.send("omni_createrawtx_opreturn", []interface{}{rawtx, payload})
}
func (client *Client) omniCreateRawtxChange(rawtx string, prevtxs []map[string]interface{}, destination string, fee float64) (result string, err error) {
	return client.send("omni_createrawtx_change", []interface{}{rawtx, prevtxs, destination, fee})
}
func (client *Client) omniCreateRawtxReference(rawtx string, destination string) (result string, err error) {
	return client.send("omni_createrawtx_reference", []interface{}{rawtx, destination})
}

func (client *Client) OmniRawTransaction(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, propertyId int64, amount float64, minerFee float64, sequence int) (txid string, err error) {
	_, hex, err := client.OmniCreateAndSignRawTransaction(fromBitCoinAddress, privkeys, toBitCoinAddress, propertyId, amount, minerFee, sequence)
	if err != nil {
		return "", err
	}

	result, err := client.OmniDecodeTransaction(hex)
	if err == nil {
		log.Println(result)
	} else {
		log.Println(err)
	}

	//8 send
	txid, err = client.SendRawTransaction(hex)
	if err != nil {
		return "", err
	}
	log.Println("8 send", txid)

	return txid, nil
}

func (client *Client) OmniCreateAndSignRawTransaction(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, propertyId int64, amount float64, minerFee float64, sequence int) (txid, hex string, err error) {
	beginTime := time.Now()
	log.Println("OmniCreateAndSignRawTransaction beginTime", beginTime.String())
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.GetOmniDustBtc() {
		return "", "", errors.New("wrong amount")
	}

	pMoney := config.GetOmniDustBtc()
	if minerFee < config.GetOmniDustBtc() {
		minerFee = 0.00003
	}

	_, _ = client.ValidateAddress(fromBitCoinAddress)
	resultListUnspent, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", "", err
	}

	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	if len(arrayListUnspent) == 0 {
		return "", "", errors.New("empty balance")
	}
	log.Println("listunspent", arrayListUnspent)

	out, _ := decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(pMoney)).Round(8).Float64()
	balance := 0.0
	for _, item := range arrayListUnspent {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Round(8).Float64()
		if balance >= out {
			break
		}
	}

	log.Println("1 balance", balance)
	if balance < out {
		return "", "", errors.New("not enough balance")
	}

	//2.Omni_createpayload_simplesend
	payload, err := client.omniCreatePayloadSimpleSend(propertyId, amount)
	if err != nil {
		return "", "", err
	}
	log.Println("2 payload " + payload)

	inputs := make([]map[string]interface{}, 0, len(arrayListUnspent))
	for _, item := range arrayListUnspent {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		node["confirmations"] = item.Get("confirmations").Int()
		node["spendable"] = item.Get("spendable").Bool()
		node["solvable"] = item.Get("solvable").Bool()
		node["amount"] = item.Get("amount").Float()
		node["address"] = item.Get("address").String()
		node["account"] = item.Get("account").String()
		node["scriptPubKey"] = item.Get("scriptPubKey").String()
		inputs = append(inputs, node)
	}

	outputs := make(map[string]interface{})

	//3.CreateRawTransaction
	createrawtransactionStr, err := client.CreateRawTransaction(inputs, outputs)
	if err != nil {
		return "", "", err
	}
	log.Println("3 createrawtransactionStr", createrawtransactionStr)

	//4.Omni_createrawtx_opreturn
	opreturn, err := client.omniCreateRawtxOpreturn(createrawtransactionStr, payload)
	if err != nil {
		return "", "", err
	}
	log.Println("4 opreturn", opreturn)

	//5. Omni_createrawtx_reference
	reference, err := client.omniCreateRawtxReference(opreturn, toBitCoinAddress)
	if err != nil {
		return "", "", err
	}
	log.Println("5 reference", reference)

	//6.Omni_createrawtx_change
	prevtxs := make([]map[string]interface{}, 0, len(arrayListUnspent))
	for _, item := range arrayListUnspent {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		node["scriptPubKey"] = item.Get("scriptPubKey").String()
		node["value"] = item.Get("amount").Float()
		prevtxs = append(prevtxs, node)
	}
	change, err := client.omniCreateRawtxChange(reference, prevtxs, fromBitCoinAddress, minerFee)
	if err != nil {
		return "", "", err
	}
	log.Println("6 change", change)

	if privkeys == nil || len(privkeys) == 0 {
		privkeys = nil
	}
	//7 sign
	signHex, err := client.SignRawTransactionWithKey(change, privkeys, inputs, "ALL")
	if err != nil {
		return "", "", err
	}

	hex = gjson.Get(signHex, "hex").String()
	log.Println("7 SignRawTransactionWithKey", hex)
	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("7 DecodeSignRawTransactionWithKey", decodeHex)
	txid = gjson.Get(decodeHex, "txid").String()

	result, err := client.OmniDecodeTransaction(hex)
	if err == nil {
		log.Println(result)
	} else {
		log.Println(err)
	}
	log.Println("OmniCreateAndSignRawTransaction endTime.Sub(beginTime)", time.Now().Sub(beginTime).String())
	return txid, hex, nil
}

// From channelAddress to temp multi address, to Create CommitmentTx
func (client *Client) OmniCreateAndSignRawTransactionUseSingleInput(txType int, fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, propertyId int64, amount float64, minerFee float64, sequence int, redeemScript *string, usedTxid string) (txid, hex string, currUseTxid string, err error) {
	beginTime := time.Now()
	log.Println("OmniCreateAndSignRawTransactionUseSingleInput beginTime", beginTime.String())
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", "", "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", "", "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.GetOmniDustBtc() {
		return "", "", "", errors.New("wrong amount")
	}

	pMoney := config.GetOmniDustBtc()
	if minerFee < pMoney {
		minerFee = config.GetMinerFee()
	}

	_, _ = client.ValidateAddress(fromBitCoinAddress)
	_, _ = client.ValidateAddress(toBitCoinAddress)

	resultListUnspent, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", "", "", err
	}
	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	log.Println("listunspent", arrayListUnspent)
	inputCount := 3 + txType
	if len(arrayListUnspent) < inputCount {
		return "", "", "", errors.New("wrong input num, need " + strconv.Itoa(inputCount) + " input:one for omni token, " + strconv.Itoa(inputCount-1) + "  btc  inputs for miner fee ")
	}

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
			inputs = append(inputs, node)
			break
		}
	}

	if currUseTxid == "" {
		return "", "", "", errors.New("not found the miner fee input")
	}

	minMinerFee := config.GetMinMinerFee(len(inputs))
	if minerFee < minMinerFee {
		minerFee = minMinerFee
	}

	out, _ := decimal.NewFromFloat(pMoney).
		Add(decimal.NewFromFloat(minerFee)).
		Round(8).
		Float64()
	log.Println("1 balance", balance)
	if balance < out {
		return "", "", "", errors.New("not enough balance")
	}

	//2.Omni_createpayload_simplesend
	payload, err := client.omniCreatePayloadSimpleSend(propertyId, amount)
	if err != nil {
		return "", "", "", err
	}
	log.Println("2 payload " + payload)

	outputs := make(map[string]interface{})
	//3.CreateRawTransaction
	createrawtransactionStr, err := client.CreateRawTransaction(inputs, outputs)
	if err != nil {
		return "", "", "", err
	}
	log.Println("3 createrawtransactionStr", createrawtransactionStr)

	//4.Omni_createrawtx_opreturn
	opreturn, err := client.omniCreateRawtxOpreturn(createrawtransactionStr, payload)
	if err != nil {
		return "", "", "", err
	}
	log.Println("4 opreturn", opreturn)

	//5. Omni_createrawtx_reference
	reference, err := client.omniCreateRawtxReference(opreturn, toBitCoinAddress)
	if err != nil {
		return "", "", "", err
	}
	log.Println("5 reference", reference)

	//6.Omni_createrawtx_change
	prevtxs := make([]map[string]interface{}, 0, len(inputs))
	for _, item := range inputs {
		node := make(map[string]interface{})
		node["txid"] = item["txid"]
		node["vout"] = item["vout"]
		node["scriptPubKey"] = item["scriptPubKey"]
		node["value"] = item["amount"]
		if redeemScript != nil {
			node["redeemScript"] = *redeemScript
		}
		prevtxs = append(prevtxs, node)
	}
	change, err := client.omniCreateRawtxChange(reference, prevtxs, toBitCoinAddress, minerFee)
	if err != nil {
		return "", "", "", err
	}
	log.Println("6 change", change)

	if privkeys == nil || len(privkeys) == 0 {
		privkeys = nil
	}
	//7 sign
	signHex, err := client.SignRawTransactionWithKey(change, privkeys, inputs, "ALL")
	if err != nil {
		return "", "", "", err
	}

	hex = gjson.Get(signHex, "hex").String()
	log.Println("7 SignRawTransactionWithKey", hex)
	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("7 DecodeSignRawTransactionWithKey", decodeHex)
	txid = gjson.Get(decodeHex, "txid").String()

	result, err := client.OmniDecodeTransaction(hex)
	if err == nil {
		log.Println(result)
	} else {
		log.Println(err)
	}
	log.Println("OmniCreateAndSignRawTransaction endTime.Sub(beginTime)", time.Now().Sub(beginTime).String())
	return txid, hex, currUseTxid, nil
}

func (client *Client) OmniCreateAndSignRawTransactionUseRestInput(txType int, fromBitCoinAddress string, usedTxid string, privkeys []string, toBitCoinAddress, changeToAddress string, propertyId int64, amount float64, minerFee float64, sequence int, redeemScript *string) (txid, hex string, err error) {
	beginTime := time.Now()
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.GetOmniDustBtc() {
		return "", "", errors.New("wrong amount")
	}

	pMoney := config.GetOmniDustBtc()
	if minerFee < config.GetOmniDustBtc() {
		minerFee = config.GetMinerFee()
	}

	_, _ = client.ValidateAddress(fromBitCoinAddress)
	_, _ = client.ValidateAddress(toBitCoinAddress)

	resultListUnspent, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", "", err
	}
	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	log.Println("listunspent", arrayListUnspent)
	inputCount := 3 + txType
	if len(arrayListUnspent) < inputCount {
		return "", "", errors.New("wrong input num, need " + strconv.Itoa(inputCount) + " input:one for omni token, " + strconv.Itoa(inputCount-1) + "  btc  inputs for miner fee ")
	}

	inputs := make([]map[string]interface{}, 0, 0)
	for _, item := range arrayListUnspent {
		txid := item.Get("txid").String()
		if usedTxid != "" && strings.Contains(usedTxid, txid) == false {
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

	log.Println("1 balance")
	if balance < out {
		return "", "", errors.New("not enough balance")
	}

	//2.Omni_createpayload_simplesend
	payload, err := client.omniCreatePayloadSimpleSend(propertyId, amount)
	if err != nil {
		return "", "", err
	}
	log.Println("2 payload ")

	outputs := make(map[string]interface{})
	//3.CreateRawTransaction
	createrawtransactionStr, err := client.CreateRawTransaction(inputs, outputs)
	if err != nil {
		return "", "", err
	}
	log.Println("3 createrawtransactionStr", createrawtransactionStr)

	//4.Omni_createrawtx_opreturn
	opreturn, err := client.omniCreateRawtxOpreturn(createrawtransactionStr, payload)
	if err != nil {
		return "", "", err
	}
	log.Println("4 opreturn")

	//5. Omni_createrawtx_reference
	reference, err := client.omniCreateRawtxReference(opreturn, toBitCoinAddress)
	if err != nil {
		return "", "", err
	}
	log.Println("5 reference")

	//6.Omni_createrawtx_change
	prevtxs := make([]map[string]interface{}, 0, len(inputs))
	for _, item := range inputs {
		node := make(map[string]interface{})
		node["txid"] = item["txid"]
		node["vout"] = item["vout"]
		node["scriptPubKey"] = item["scriptPubKey"]
		node["value"] = item["amount"]
		if redeemScript != nil {
			node["redeemScript"] = *redeemScript
		}
		prevtxs = append(prevtxs, node)
	}
	change, err := client.omniCreateRawtxChange(reference, prevtxs, changeToAddress, minerFee)
	if err != nil {
		return "", "", err
	}
	log.Println("6 change")

	if privkeys == nil || len(privkeys) == 0 {
		privkeys = nil
	}
	//7 sign
	signHex, err := client.SignRawTransactionWithKey(change, privkeys, inputs, "ALL")
	if err != nil {
		return "", "", err
	}

	hex = gjson.Get(signHex, "hex").String()
	log.Println("7 SignRawTransactionWithKey")
	decodeHex, _ := client.DecodeRawTransaction(hex)
	//log.Println("7 DecodeSignRawTransactionWithKey", decodeHex)
	txid = gjson.Get(decodeHex, "txid").String()

	result, err := client.OmniDecodeTransaction(hex)
	if err == nil {
		log.Println(result)
	} else {
		log.Println(err)
	}
	log.Println("OmniCreateAndSignRawTransactionUseRestInput endTime.Sub(beginTime)", time.Now().Sub(beginTime).String())
	return txid, hex, nil
}

func (client *Client) OmniCreateAndSignRawTransactionUseUnsendInput(fromBitCoinAddress string, privkeys []string, inputItems []TransactionInputItem, toBitCoinAddress, changeToAddress string, propertyId int64, amount float64, minerFee float64, sequence int, redeemScript *string) (txid string, hex string, err error) {
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", "", errors.New("toBitCoinAddress is empty")
	}

	if len(inputItems) == 0 {
		return "", "", errors.New("inputItems is empty")
	}

	if amount < config.GetOmniDustBtc() {
		return "", "", errors.New("wrong amount")
	}

	pMoney := config.GetOmniDustBtc()
	if minerFee < config.GetOmniDustBtc() {
		minerFee = config.GetMinerFee()
	}

	_, _ = client.ValidateAddress(fromBitCoinAddress)
	_, _ = client.ValidateAddress(toBitCoinAddress)
	_, _ = client.ValidateAddress(changeToAddress)

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
	log.Println("1 balance", balance)
	if balance < out {
		return "", "", errors.New("not enough balance")
	}

	//2.Omni_createpayload_simplesend
	payload, err := client.omniCreatePayloadSimpleSend(propertyId, amount)
	if err != nil {
		return "", "", err
	}
	//log.Println("2 payload " + payload)

	//3.CreateRawTransaction
	outputs := make(map[string]interface{})
	//outputs["data"] = "e4bda0e5a5bdefbc8ce4b896e7958ce38082"

	createrawtransactionStr, err := client.CreateRawTransaction(inputs, outputs)
	if err != nil {
		return "", "", err
	}

	log.Println("3 createrawtransactionStr", createrawtransactionStr)
	result, err := client.DecodeRawTransaction(createrawtransactionStr)
	if err != nil {
		return "", "", err
	}
	log.Println(result)

	//4.Omni_createrawtx_opreturn
	opreturn, err := client.omniCreateRawtxOpreturn(createrawtransactionStr, payload)
	if err != nil {
		return "", "", err
	}
	//log.Println("4 opreturn", opreturn)

	//5. Omni_createrawtx_reference
	reference, err := client.omniCreateRawtxReference(opreturn, toBitCoinAddress)
	if err != nil {
		return "", "", err
	}
	//log.Println("5 reference", reference)

	result, err = client.DecodeRawTransaction(reference)
	if err != nil {
		return "", "", err
	}
	//log.Println(result)
	//6.Omni_createrawtx_change
	prevtxs := make([]map[string]interface{}, 0, 0)
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		node["scriptPubKey"] = item.ScriptPubKey
		node["value"] = item.Amount
		if redeemScript != nil {
			node["redeemScript"] = *redeemScript
		}
		prevtxs = append(prevtxs, node)
	}
	change, err := client.omniCreateRawtxChange(reference, prevtxs, changeToAddress, minerFee)
	if err != nil {
		return "", "", err
	}
	//log.Println("6 change", change)

	result, err = client.DecodeRawTransaction(change)
	if err != nil {
		return "", "", err
	}
	log.Println(result)

	if privkeys == nil || len(privkeys) == 0 {
		privkeys = nil
	}

	//7 sign
	signHex, err := client.SignRawTransactionWithKey(change, privkeys, inputs, "ALL")
	if err != nil {
		return "", "", err
	}

	hex = gjson.Get(signHex, "hex").String()
	//log.Println("7 SignRawTransactionWithKey", hex)
	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("7 DecodeSignRawTransactionWithKey", decodeHex)
	txid = gjson.Get(decodeHex, "txid").String()

	result, err = client.OmniDecodeTransaction(hex)
	if err != nil {
		log.Println(err)
	}
	//log.Println(result)

	return txid, hex, nil
}

func (client *Client) OmniSignRawTransactionForUnsend(hex string, inputItems []TransactionInputItem, privKey string) (string, string, error) {

	var inputs []map[string]interface{}
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		node["amount"] = item.Amount
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
	if err == nil {
		log.Println(decodeHex)
	} else {
		log.Println(err)
	}
	txId := gjson.Get(decodeHex, "txid").String()
	if err != nil {
		return "", hex, err
	}

	result, err := client.OmniDecodeTransaction(hex)
	if err == nil {
		log.Println(result)
	} else {
		log.Println(err)
	}

	return txId, hex, nil
}
