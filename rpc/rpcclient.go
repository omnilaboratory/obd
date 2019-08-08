package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func (client *Client) CreateRawTransaction(inputs []RawTxInput, outputs []map[string]float32, locktime int) (result string, err error) {
	params := make([]interface{}, 0, 3)
	params = append(params, inputs)
	params = append(params, outputs)
	params = append(params, locktime)
	return client.send("createrawtransaction", params)
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

func (client *Client) send(method string, params []interface{}) (result string, err error) {
	rawParams := make([]json.RawMessage, 0, len(params))
	for _, item := range params {
		marshaledParam, err := json.Marshal(item)
		if err == nil {
			rawParams = append(rawParams, marshaledParam)
		}
	}

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
	return string(res), nil
}
