package rpc

import (
	"LightningOnOmni/config"
	"LightningOnOmni/tool"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"strings"
)

//https://github.com/OmniLayer/omnicore/blob/master/src/omnicore/doc/rpc-api.md
//Returns various state information of the client and protocol.
func (client *Client) OmniGetinfo() (result string, err error) {
	return client.send("omni_getinfo", nil)
}

//Create and broadcast a simple send transaction.
func (client *Client) OmniSend(fromAddress string, toAddress string, propertyId int, amount float64) (result string, err error) {
	return client.send("omni_send", []interface{}{fromAddress, toAddress, propertyId, amount})
}

func (client *Client) OmniGetbalance(address string, propertyId int) (result string, err error) {
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
	return client.send("omni_getallbalancesforaddress ", []interface{}{address})
}

//List wallet transactions, optionally filtered by an address and block boundaries.
func (client *Client) OmniListTransactions(count int, skip int) (result string, err error) {
	if count < 0 {
		count = 10
	}
	if skip < 0 {
		skip = 0
	}
	return client.send("omni_listtransactions", []interface{}{"*", count, skip})
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
	//8 send
	txid, err = client.SendRawTransaction(hex)
	if err != nil {
		return "", err
	}
	log.Println("8 send", txid)

	return txid, nil
}

func (client *Client) OmniCreateAndSignRawTransaction(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, propertyId int64, amount float64, minerFee float64, sequence int) (txid, hex string, err error) {
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.Dust {
		return "", "", errors.New("wrong amount")
	}

	pMoney := 0.0000054
	if minerFee < config.Dust {
		minerFee = 0.00003
	}

	if ismine, _ := client.ValidateAddress(fromBitCoinAddress); ismine == false {
		err = client.ImportAddress(fromBitCoinAddress)
	}

	if ismine, _ := client.ValidateAddress(toBitCoinAddress); ismine == false {
		err = client.ImportAddress(toBitCoinAddress)
	}

	resultListUnspent, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", "", err
	}

	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	if len(arrayListUnspent) == 0 {
		return "", "", errors.New("empty balance")
	}
	log.Println("listunspent", arrayListUnspent)

	out, _ := decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(pMoney)).Float64()
	balance := 0.0
	for _, item := range arrayListUnspent {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Float64()
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
	//outputs["data"]="e4bda0e5a5bdefbc8ce4b896e7958ce38082"
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

	return txid, hex, nil
}

// From channelAddress to temp multi address, to Create CommitmentTx
func (client *Client) OmniCreateAndSignRawTransactionForCommitmentTx(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, propertyId int64, amount float64, minerFee float64, sequence int, redeemScript *string, usedTxid string) (txid, hex string, currUseTxid string, err error) {
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", "", "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", "", "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.Dust {
		return "", "", "", errors.New("wrong amount")
	}

	pMoney := GetOmniDustBtc()
	if minerFee < config.Dust {
		minerFee = GetMinerFee()
	}

	out, _ := decimal.NewFromFloat(minerFee).
		Add(decimal.NewFromFloat(minerFee)).
		Float64()

	client.ValidateAddress(fromBitCoinAddress)
	client.ValidateAddress(toBitCoinAddress)

	resultListUnspent, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", "", "", err
	}
	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	log.Println("listunspent", arrayListUnspent)
	if len(arrayListUnspent) < 2 {
		return "", "", "", errors.New("wrong input num, need 2  btc  inputs for miner fee ")
	}

	inputs := make([]map[string]interface{}, 0, 0)
	currUseTxid = ""
	for _, item := range arrayListUnspent {
		currUseTxid = item.Get("txid").String()
		if usedTxid != "" && strings.Contains(usedTxid, currUseTxid) {
			continue
		}
		inputAmount := item.Get("amount").Float()
		if inputAmount >= out {
			node := make(map[string]interface{})
			node["txid"] = item.Get("txid").String()
			node["vout"] = item.Get("vout").Int()
			node["amount"] = inputAmount
			node["scriptPubKey"] = item.Get("scriptPubKey").String()
			if redeemScript != nil {
				node["redeemScript"] = *redeemScript
			}
			inputs = append(inputs, node)
			break
		}
	}

	if currUseTxid == "" {
		return "", "", "", errors.New("not found the miner fee input")
	}

	out, _ = decimal.NewFromFloat(out).
		Add(decimal.NewFromFloat(pMoney)).
		Float64()
	balance := 0.0
	for _, item := range inputs {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item["amount"].(float64))).Float64()
		if balance >= out {
			break
		}
	}

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

	return txid, hex, currUseTxid, nil
}

func (client *Client) OmniCreateAndSignRawTransactionForCommitmentTxToBob(fromBitCoinAddress string, usedTxid string, privkeys []string, toBitCoinAddress, changeToAddress string, propertyId int64, amount float64, minerFee float64, sequence int, redeemScript *string) (txid, hex string, err error) {
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.Dust {
		return "", "", errors.New("wrong amount")
	}

	pMoney := GetOmniDustBtc()
	if minerFee < config.Dust {
		minerFee = GetMinerFee()
	}

	out, _ := decimal.NewFromFloat(minerFee).
		Add(decimal.NewFromFloat(pMoney)).
		Float64()

	_, _ = client.ValidateAddress(fromBitCoinAddress)
	_, _ = client.ValidateAddress(toBitCoinAddress)

	resultListUnspent, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", "", err
	}
	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	log.Println("listunspent", arrayListUnspent)
	if len(arrayListUnspent) < 3 {
		return "", "", errors.New("wrong input num, need 3 input:one for omni token, 2  btc  inputs for miner fee ")
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

	balance := 0.0
	for _, item := range inputs {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item["amount"].(float64))).Float64()
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

	return txid, hex, nil
}

func (client *Client) OmniCreateAndSignRawTransactionForUnsendInputTx(fromBitCoinAddress string, privkeys []string, inputItems []TransactionInputItem, toBitCoinAddress, changeToAddress string, propertyId int64, amount float64, minerFee float64, sequence int, redeemScript *string) (txid string, hex string, err error) {
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", "", errors.New("toBitCoinAddress is empty")
	}

	if len(inputItems) == 0 {
		return "", "", errors.New("inputItems is empty")
	}

	if amount < config.Dust {
		return "", "", errors.New("wrong amount")
	}

	pMoney := 0.00000540
	if minerFee < config.Dust {
		minerFee = GetMinerFee()
	}

	if ismine, _ := client.ValidateAddress(fromBitCoinAddress); ismine == false {
		err = client.ImportAddress(fromBitCoinAddress)
	}

	if ismine, _ := client.ValidateAddress(toBitCoinAddress); ismine == false {
		err = client.ImportAddress(toBitCoinAddress)
	}

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
	out, _ := decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(pMoney)).Float64()
	balance := 0.0
	for _, item := range inputs {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item["amount"].(float64))).Float64()
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

	//3.CreateRawTransaction
	outputs := make(map[string]interface{})
	//outputs["data"] = "e4bda0e5a5bdefbc8ce4b896e7958ce38082"

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
	change, err := client.omniCreateRawtxChange(reference, prevtxs, changeToAddress, out)
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

	return txid, hex, nil
}
