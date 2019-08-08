package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
)

var config *ConnConfig

func init() {
	config = &ConnConfig{
		Host: "62.234.216.108:18332",
		User: "omniwallet",
		Pass: "cB3]iL2@eZ1?cB2?",
	}
}

type ConnConfig struct {
	// Host is the IP address and port of the RPC server you want to connect to.
	Host string
	// User is the username to use to authenticate to the RPC server.
	User string
	// Pass is the passphrase to use to authenticate to the RPC server.
	Pass string
}

type Client struct {
	id uint64 // atomic, so must stay 64-bit aligned
	// config holds the connection configuration assoiated with this client.
	config *ConnConfig
	// httpClient is the underlying HTTP client to use when running in HTTP POST mode.
	httpClient http.Client
}

type Request struct {
	Jsonrpc string            `json:"jsonrpc"`
	ID      interface{}       `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

type rawResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
}

func (r rawResponse) result() (result []byte, err error) {
	if r.Error != nil {
		return nil, r.Error
	}
	return r.Result, nil
}

func (r *RPCError) Error() string {
	return r.Error()
}

type RPCError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewClient() *Client {
	httpClient := http.Client{
		Transport: &http.Transport{
			Proxy:           nil,
			TLSClientConfig: nil,
		},
	}
	config.Host = "http://" + config.Host
	client := &Client{
		config:     config,
		httpClient: httpClient,
	}
	return client
}

func (client *Client) NextID() uint64 {
	return atomic.AddUint64(&client.id, 1)
}

func (client *Client) GetTransactionById(txid string) (result string, err error) {
	params := make([]interface{}, 0, 1)
	params = append(params, txid)
	return client.send("gettransaction", params)
}
func (client *Client) GetTxOut(txid string, num int) (result string, err error) {
	params := make([]interface{}, 0, 2)
	params = append(params, txid)
	params = append(params, num)
	return client.send("gettxout", params)
}
func (client *Client) CreateMultiSig(minSignNum int, keys []string) (result string, err error) {
	params := make([]interface{}, 0, 2)
	params = append(params, minSignNum)
	params = append(params, keys)
	return client.send("createmultisig", params)
}

func (client *Client) GetNewAddress(label string) (result string, err error) {
	params := make([]interface{}, 0, 1)
	params = append(params, label)
	return client.send("getnewaddress", params)
}

func (client *Client) ListUnspent(address string) (result string, err error) {
	if len(address) < 1 {
		return "", errors.New("address not exist")
	}

	params := make([]interface{}, 0, 3)
	params = append(params, 0)
	params = append(params, 9999999)
	keys := []string{
		address,
	}
	params = append(params, keys)
	return client.send("listunspent", params)
}

type RawTxInput struct {
	Txid     string `json:"txid"`
	Vout     int    `json:"vout"`
	Sequence int    `json:"sequence"`
}

func (client *Client) CreateRawTransaction(inputs []map[string]interface{}, outputs map[string]float64) (result string, err error) {
	params := make([]interface{}, 0, 2)
	params = append(params, inputs)
	params = append(params, outputs)
	return client.send("createrawtransaction", params)
}

func (client *Client) SignRawTransactionWithKey(hex string, privkeys []string, prevtxs []map[string]interface{}, sighashtype string) (result string, err error) {
	params := make([]interface{}, 0, 4)
	params = append(params, hex)
	params = append(params, prevtxs)
	params = append(params, privkeys)
	params = append(params, sighashtype)
	return client.send("signrawtransaction", params)
}

func (client *Client) SendRawTransaction(hex string) (result string, err error) {
	params := make([]interface{}, 0, 1)
	params = append(params, hex)
	return client.send("sendrawtransaction", params)
}

func (client *Client) DecodeRawTransaction(hex string) (result string, err error) {
	params := make([]interface{}, 0, 1)
	params = append(params, hex)
	return client.send("decoderawtransaction", params)
}

func (client *Client) BtcRawTransactionMultiSign(fromBitCoinAddress string, privkeys []string, toBitCoinAddress string, amount float64, mineFee float64) (txId string, err error) {
	result, err := client.ListUnspent(fromBitCoinAddress)
	if err != nil {
		return "", err
	}

	array := gjson.Parse(result).Array()
	if len(array) == 0 {
		return "", errors.New("empty balance")
	}
	log.Println("listunspent", array)

	fee := mineFee
	out, _ := decimal.NewFromFloat(fee).Add(decimal.NewFromFloat(amount)).Float64()
	var inputs []map[string]interface{}

	balance := 0.0
	scriptPubKey := ""

	for _, item := range array {
		node := make(map[string]interface{})
		node["txid"] = item.Get("txid").String()
		node["vout"] = item.Get("vout").Int()
		if len(privkeys) > 0 {
			node["redeemScript"] = item.Get("redeemScript")
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

	hex, err := client.CreateRawTransaction(inputs, output)

	if err != nil {
		return "", err
	}

	log.Println("CreateRawTransaction", hex)

	decodeHex, _ := client.DecodeRawTransaction(hex)
	log.Println("DecodeRawTransaction", decodeHex)

	for _, item := range inputs {
		item["scriptPubKey"] = scriptPubKey
	}

	var signHex string
	if len(privkeys) > 0 {
		signHex, _ = client.SignRawTransactionWithKey(hex, privkeys, inputs, "ALL")
	} else {
		signHex, _ = client.SignRawTransactionWithKey(hex, nil, inputs, "ALL")
	}

	hex = gjson.Get(signHex, "hex").String()
	log.Println(hex)
	decodeHex, _ = client.DecodeRawTransaction(hex)
	log.Println("DecodeRawTransaction", decodeHex)
	txId, err = client.SendRawTransaction(string(hex))
	if err != nil {
		return "", err
	}

	return txId, nil
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

//Returns various state information of the client and protocol.
func (client *Client) Omni_getinfo() (result string, err error) {
	return client.send("omni_getinfo", nil)
}

//Returns the token balance for a given address and property.
func (client *Client) Omni_getbalance(address string, propertyId int) (result string, err error) {
	params := make([]interface{}, 0, 2)
	params = append(params, address)
	params = append(params, propertyId)
	return client.send("omni_getbalance", params)
}

//Get detailed information about an Omni transaction.
func (client *Client) Omni_gettransaction(txid string) (result string, err error) {
	params := make([]interface{}, 0, 1)
	params = append(params, txid)
	return client.send("omni_gettransaction", params)
}

//List wallet transactions, optionally filtered by an address and block boundaries.
func (client *Client) Omni_listtransactions(count int, skip int) (result string, err error) {
	if count < 0 {
		count = 10
	}
	if skip < 0 {
		skip = 0
	}
	params := make([]interface{}, 0, 3)
	params = append(params, "*")
	params = append(params, count)
	params = append(params, skip)
	return client.send("omni_listtransactions", params)
}

//Create and broadcast a simple send transaction.
func (client *Client) Omni_send(fromAddress string, toAddress string, propertyId int, amount float64) (result string, err error) {
	params := make([]interface{}, 0, 4)
	params = append(params, fromAddress)
	params = append(params, toAddress)
	params = append(params, propertyId)
	params = append(params, amount)
	return client.send("omni_send", params)
}

func (client *Client) send(method string, params []interface{}) (result string, err error) {
	rawParams := make([]json.RawMessage, 0, len(params))
	for _, item := range params {
		marshaledParam, err := json.Marshal(item)
		if err == nil {
			rawParams = append(rawParams, marshaledParam)
		}
	}

	log.Println()
	req := &Request{
		Jsonrpc: "1.0",
		ID:      client.NextID(),
		Method:  method,
		Params:  rawParams,
	}
	marshaledJSON, e := json.Marshal(req)
	if e != nil {
		return "", e
	}

	bodyReader := bytes.NewReader(marshaledJSON)

	httpReq, err := http.NewRequest("POST", client.config.Host, bodyReader)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(client.config.User, client.config.Pass)
	httpResponse, err := client.httpClient.Do(httpReq)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if httpResponse.StatusCode == 500 {
		return "", errors.New("can not get data from server")
	}
	// Read the raw bytes and close the response.
	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	httpResponse.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading json reply: %v", err)
		return "", err
	}

	var resp rawResponse
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		err = fmt.Errorf("status code: %d, response: %q", httpResponse.StatusCode, string(respBytes))
		return "", err
	}
	res, err := resp.result()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return gjson.Parse(string(res)).String(), nil
}
