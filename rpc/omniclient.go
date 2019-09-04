package rpc

import (
	"LightningOnOmni/config"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
)

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

func (client *Client) omniCreatepayloadSimpleSend(propertyId int, amount float64) (result string, err error) {
	return client.send("omni_createpayload_simplesend", []interface{}{propertyId, decimal.NewFromFloat(amount).String()})
}
func (client *Client) omniCreaterawtxOpreturn(rawtx string, payload string) (result string, err error) {
	return client.send("omni_createrawtx_opreturn", []interface{}{rawtx, payload})
}
func (client *Client) omniCreaterawtxChange(rawtx string, prevtxs []map[string]interface{}, destination string, fee float64) (result string, err error) {
	return client.send("omni_createrawtx_change", []interface{}{rawtx, prevtxs, destination, fee})
}
func (client *Client) omniCreaterawtxReference(rawtx string, destination string) (result string, err error) {
	return client.send("omni_createrawtx_reference", []interface{}{rawtx, destination})
}

func (client *Client) OmniRawTransaction(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, propertyId int, amount float64, minerFee float64, sequence *int) (txid string, err error) {
	if len(fromBitCoinAddress) < 1 {
		return "", errors.New("fromBitCoinAddress is empty")
	}
	if len(toBitCoinAddress) < 1 {
		return "", errors.New("toBitCoinAddress is empty")
	}
	if amount < config.Dust {
		return "", errors.New("wrong amount")
	}
	if minerFee < config.Dust {
		return "", errors.New("minerFee too small")
	}

	if ismine, _ := client.ValidateAddress(fromBitCoinAddress); ismine == false {
		err = client.ImportAddress(fromBitCoinAddress)
	}
	if ismine, _ := client.ValidateAddress(toBitCoinAddress); ismine == false {
		err = client.ImportAddress(toBitCoinAddress)
	}

	result, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", err
	}

	array := gjson.Parse(result).Array()
	if len(array) == 0 {
		return "", errors.New("empty balance")
	}
	log.Println("listunspent", array)

	fee := minerFee
	dustMoney := 0.00000546
	out, _ := decimal.NewFromFloat(fee).Add(decimal.NewFromFloat(dustMoney)).Float64()
	balance := 0.0
	for _, item := range array {
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
	payload, err := client.omniCreatepayloadSimpleSend(propertyId, amount)
	if err != nil {
		return "", err
	}
	log.Println("2 payload " + payload)

	inputs := make([]map[string]interface{}, 0, len(array))
	for _, item := range array {
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
	opreturn, err := client.omniCreaterawtxOpreturn(createrawtransactionStr, payload)
	if err != nil {
		return "", err
	}
	log.Println("4 opreturn", opreturn)

	//5. Omni_createrawtx_reference
	reference, err := client.omniCreaterawtxReference(opreturn, toBitCoinAddress)
	if err != nil {
		return "", err
	}
	log.Println("5 reference", reference)

	//6.Omni_createrawtx_change
	prevtxs := make([]map[string]interface{}, 0, len(array))
	for _, item := range array {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		node["scriptPubKey"] = item.Get("scriptPubKey").String()
		node["value"] = item.Get("amount").Float()
		prevtxs = append(prevtxs, node)
	}
	change, err := client.omniCreaterawtxChange(reference, prevtxs, fromBitCoinAddress, minerFee)
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

	hex := gjson.Get(signHex, "hex").String()
	log.Println("7 SignRawTransactionWithKey", hex)
	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("7 DecodeSignRawTransactionWithKey", decodeHex)

	//8 send
	txid, err = client.SendRawTransaction(hex)
	if err != nil {
		return "", err
	}
	log.Println("8 send", txid)

	return txid, nil
}
