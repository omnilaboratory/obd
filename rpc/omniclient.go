package rpc

import (
	"LightningOnOmni/config"
	"LightningOnOmni/tool"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
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
	hex, err := client.OmniCreateAndSignRawTransaction(fromBitCoinAddress, privkeys, toBitCoinAddress, propertyId, amount, minerFee, sequence)
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
func (client *Client) OmniCreateAndSignRawTransaction(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, propertyId int64, amount float64, minerFee float64, sequence int) (hex string, err error) {
	if tool.CheckIsString(&fromBitCoinAddress) == false {
		return "", errors.New("fromBitCoinAddress is empty")
	}
	if tool.CheckIsString(&toBitCoinAddress) == false {
		return "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.Dust {
		return "", errors.New("wrong amount")
	}

	pMoney := 0.00000546
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
		return "", err
	}

	arrayListUnspent := gjson.Parse(resultListUnspent).Array()
	if len(arrayListUnspent) == 0 {
		return "", errors.New("empty balance")
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
		return "", errors.New("not enough balance")
	}

	//2.Omni_createpayload_simplesend
	payload, err := client.omniCreatePayloadSimpleSend(propertyId, amount)
	if err != nil {
		return "", err
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

	outputs := make(map[string]float64)
	//3.CreateRawTransaction
	createrawtransactionStr, err := client.CreateRawTransaction(inputs, outputs)
	if err != nil {
		return "", err
	}
	log.Println("3 createrawtransactionStr", createrawtransactionStr)

	//4.Omni_createrawtx_opreturn
	opreturn, err := client.omniCreateRawtxOpreturn(createrawtransactionStr, payload)
	if err != nil {
		return "", err
	}
	log.Println("4 opreturn", opreturn)

	//5. Omni_createrawtx_reference
	reference, err := client.omniCreateRawtxReference(opreturn, toBitCoinAddress)
	if err != nil {
		return "", err
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
		return "", err
	}
	log.Println("6 change", change)

	if privkeys == nil || len(privkeys) == 0 {
		privkeys = nil
	}
	//7 sign
	signHex, err := client.SignRawTransactionWithKey(change, privkeys, inputs, "ALL")
	if err != nil {
		return "", err
	}

	hex = gjson.Get(signHex, "hex").String()
	log.Println("7 SignRawTransactionWithKey", hex)
	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("7 DecodeSignRawTransactionWithKey", decodeHex)

	return hex, nil
}

//https://blog.csdn.net/ffzhihua/article/details/80733124
func (client *Client) OmniCreateAndSignRawTransactionForUnsendInputTx(fromBitCoinAddress string, inputItems []TransactionInputItem, privkeys []string, outputItems []TransactionOutputItem, propertyId int64, minerFee float64, sequence *int) (txid, hex string, err error) {
	if len(fromBitCoinAddress) < 1 {
		return "", "", errors.New("fromBitCoinAddress is empty")
	}
	if len(outputItems) < 1 {
		return "", "", errors.New("outputItems is empty")
	}

	outAmount := decimal.NewFromFloat(0)
	for _, item := range outputItems {
		outAmount = outAmount.Add(decimal.NewFromFloat(item.Amount))
	}

	if outAmount.LessThan(decimal.NewFromFloat(config.Dust)) {
		return "", "", errors.New("wrong outAmount")
	}

	pMoney := 0.00000546
	if minerFee < config.Dust {
		minerFee = 0.00003
	}

	out, _ := decimal.NewFromFloat(minerFee).Add(decimal.NewFromFloat(pMoney)).Float64()
	balance := 0.0
	for _, item := range inputItems {
		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Amount)).Float64()
		if balance >= out {
			break
		}
	}

	log.Println("1 balance", balance)
	if balance < out {
		return "", "", errors.New("not enough balance")
	}

	//2.Omni_createpayload_simplesend
	amount, _ := outAmount.Float64()
	payload, err := client.omniCreatePayloadSimpleSend(propertyId, amount)
	if err != nil {
		return "", "", err
	}
	log.Println("2 payload " + payload)

	inputs := make([]map[string]interface{}, 0, len(inputItems))
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		if sequence != nil {
			node["sequence"] = *sequence
		}
		node["scriptPubKey"] = item.ScriptPubKey
		inputs = append(inputs, node)
	}

	outputs := make(map[string]float64)
	//3.CreateRawTransaction
	createrawtransactionStr, err := client.CreateRawTransaction(inputs, outputs)
	if err != nil {
		return "", "", err
	}
	log.Println("3 createrawtransactionStr", createrawtransactionStr)

	//4.omni_createrawtx_opreturn
	opreturn, err := client.omniCreateRawtxOpreturn(createrawtransactionStr, payload)
	if err != nil {
		return "", "", err
	}
	log.Println("4 opreturn", opreturn)

	toBitCoinAddress := ""
	//5. omni_createrawtx_reference
	reference, err := client.omniCreateRawtxReference(opreturn, toBitCoinAddress)
	if err != nil {
		return "", "", err
	}
	log.Println("5 reference", reference)

	//6.omni_createrawtx_change
	prevtxs := make([]map[string]interface{}, 0, len(inputItems))
	for _, item := range inputItems {
		node := make(map[string]interface{})
		node["txid"] = item.Txid
		node["vout"] = item.Vout
		if sequence != nil {
			node["sequence"] = *sequence
		}
		node["scriptPubKey"] = item.ScriptPubKey
		node["value"] = item.Amount

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
