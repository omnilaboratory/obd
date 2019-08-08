package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
)

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

func NewClient(config *ConnConfig) *Client {
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

func (c *Client) NextID() uint64 {
	return atomic.AddUint64(&c.id, 1)
}

func (client *Client) Send(method string, txId string) (result string, err error) {

	rawParams := make([]json.RawMessage, 0, 1)
	marshalledParam, err := json.Marshal(txId)
	if err == nil {
		rawMessage := json.RawMessage(marshalledParam)
		rawParams = append(rawParams, rawMessage)
	}
	req := &Request{
		Jsonrpc: "1.0",
		ID:      client.NextID(),
		Method:  method,
		Params:  rawParams,
	}

	marshaledJSON, _ := json.Marshal(req)
	bodyReader := bytes.NewReader(marshaledJSON)

	httpReq, err := http.NewRequest("POST", client.config.Host, bodyReader)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(client.config.User, client.config.Pass)
	httpResponse, err := client.httpClient.Do(httpReq)
	if err != nil {
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
		err = fmt.Errorf("status code: %d, response: %q",
			httpResponse.StatusCode, string(respBytes))
		return "", err
	}
	res, err := resp.result()
	var f interface{}
	err = json.Unmarshal(res, &f)
	if err != nil {
		fmt.Println(err)
	}
	return string(res), nil
}
