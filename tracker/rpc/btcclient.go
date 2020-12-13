package rpc

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
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
	return 0
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

func (client *Client) OmniDecodeTransaction(hex string) (result string, err error) {
	result, err = client.send("omni_decodetransaction", []interface{}{hex})
	return result, err
}

func (client *Client) OmniDecodeTransactionWithPrevTxs(hex string, prevtxs []bean.TransactionInputItem) (result string, err error) {
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
