package rpc

import (
	"LightningOnOmni/config"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
)

var connConfig *ConnConfig

func init() {
	connConfig = &ConnConfig{
		Host: config.Chainnode_Host,
		User: config.Chainnode_User,
		Pass: config.Chainnode_Pass,
	}
	log.Println(connConfig)
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

var client *Client

func NewClient() *Client {
	if client == nil {
		httpClient := http.Client{
			Transport: &http.Transport{
				Proxy:           nil,
				TLSClientConfig: nil,
			},
		}
		connConfig.Host = "http://" + connConfig.Host
		client = &Client{
			config:     connConfig,
			httpClient: httpClient,
		}
	}
	return client
}

func (client *Client) NextID() uint64 {
	return atomic.AddUint64(&client.id, 1)
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
	log.Println(req)

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
		return "", err
	}
	log.Println(gjson.Parse(string(res)).String())
	return gjson.Parse(string(res)).String(), nil
}
