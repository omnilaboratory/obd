package rpc

import (
	"errors"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
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
	return client.send("omni_send", []interface{}{fromAddress, toAddress, propertyId, tool.FloatToString(amount, 8)})
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
		//log.Println(decodeHex)
	} else {
		log.Println(err)
	}
	txId := gjson.Get(decodeHex, "txid").String()
	if err != nil {
		return "", hex, err
	}

	return txId, hex, nil
}

func GetBtcMinerAmount(total float64) float64 {
	out, _ := decimal.NewFromFloat(total).Div(decimal.NewFromFloat(4.0)).Sub(decimal.NewFromFloat(config.GetOmniDustBtc())).Round(8).Float64()
	return out
}

// ins*150 + outs*34 + 10 + 80 = transaction size
// https://shimo.im/docs/5w9Fi1c9vm8yp1ly
//https://bitcoinfees.earn.com/api/v1/fees/recommended
func (client *Client) GetMinerFee() float64 {
	price := client.EstimateSmartFee()
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
