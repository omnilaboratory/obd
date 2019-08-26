package rpc

import (
	"LightningOnOmni/config"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
)

func (client *Client) CreateMultiSig(minSignNum int, keys []string) (result string, err error) {
	return client.send("createmultisig", []interface{}{minSignNum, keys})
}

func (client *Client) dumpPrivKey(address string) (result string, err error) {
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
	if len(address) < 1 {
		return "", errors.New("address not exist")
	}
	keys := []string{
		address,
	}
	return client.send("listunspent", []interface{}{0, 999999, keys})
}

func (client *Client) CreateRawTransaction(inputs []map[string]interface{}, outputs map[string]float64) (result string, err error) {
	return client.send("createrawtransaction", []interface{}{inputs, outputs})
}

func (client *Client) SignRawTransactionWithKey(hex string, privkeys []string, prevtxs []map[string]interface{}, sighashtype string) (result string, err error) {
	return client.send("signrawtransaction", []interface{}{hex, prevtxs, privkeys, sighashtype})
}

func (client *Client) SendRawTransaction(hex string) (result string, err error) {
	return client.send("sendrawtransaction", []interface{}{hex})
}

func (client *Client) DecodeRawTransaction(hex string) (result string, err error) {
	return client.send("decoderawtransaction", []interface{}{hex})
}

func (client *Client) GetBlockCount() (result string, err error) {
	return client.send("getblockcount", nil)
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
	return client.send("verifymessage", []interface{}{address, signature, message})
}
func (client *Client) DecodeScript(hexString string) (result string, err error) {
	return client.send("decodescript", []interface{}{hexString})
}

func (client *Client) Validateaddress(address string) (ismine bool, err error) {
	result, err := client.send("validateaddress", []interface{}{address})
	if err != nil {
		return false, err
	}
	log.Println(result)
	return gjson.Get(result, "ismine").Bool(), nil
}
func (client *Client) Importaddress(address string) (err error) {
	result, err := client.send("importaddress", []interface{}{address, nil, false})
	if err != nil {
		return err
	}
	log.Println(result)
	return nil
}

// create a transaction and just signnature , not send to the network,get the hash of signature
func (client *Client) BtcCreateAndSignRawTransaction(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, amount float64, minerFee float64, sequence *int) (hex string, err error) {
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
	out, _ := decimal.NewFromFloat(fee).Add(decimal.NewFromFloat(amount)).Float64()

	balance := 0.0
	scriptPubKey := ""
	var inputs []map[string]interface{}
	for _, item := range array {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		if item.Get("redeemScript").Exists() {
			node["redeemScript"] = item.Get("redeemScript")
		}
		if sequence != nil {
			node["sequence"] = sequence
		}
		if scriptPubKey == "" {
			scriptPubKey = item.Get("scriptPubKey").String()
		}
		inputs = append(inputs, node)

		balance, _ = decimal.NewFromFloat(balance).Add(decimal.NewFromFloat(item.Get("amount").Float())).Float64()
		if balance >= out {
			break
		}
	}
	log.Println("input list ", inputs)

	if len(inputs) == 0 || balance < out {
		return "", errors.New("not enough balance")
	}

	drawback, _ := decimal.NewFromFloat(balance).Sub(decimal.NewFromFloat(out)).Float64()
	output := make(map[string]float64)
	output[toBitCoinAddress] = amount
	output[fromBitCoinAddress] = drawback

	hex, err = client.CreateRawTransaction(inputs, output)

	if err != nil {
		return "", err
	}

	log.Println("CreateRawTransaction", hex)

	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("DecodeRawTransaction", decodeHex)

	for _, item := range inputs {
		item["scriptPubKey"] = scriptPubKey
	}

	if privkeys == nil || len(privkeys) == 0 {
		privkeys = nil
	}
	signHex, _ := client.SignRawTransactionWithKey(hex, privkeys, inputs, "ALL")

	hex = gjson.Get(signHex, "hex").String()
	log.Println("SignRawTransactionWithKey", hex)
	decodeHex, _ = client.DecodeRawTransaction(hex)
	log.Println("DecodeRawTransaction", decodeHex)

	return hex, err
}

// create a transaction and signature and send to the network
func (client *Client) BtcCreateAndSignAndSendRawTransaction(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, amount float64, minerFee float64, sequence *int) (txId string, err error) {
	hex, err := client.BtcCreateAndSignRawTransaction(fromBitCoinAddress, privkeys, toBitCoinAddress, amount, minerFee, sequence)
	if err != nil {
		return "", err
	}

	txId, err = client.SendRawTransaction(hex)
	if err != nil {
		return "", err
	}

	return txId, nil
}
