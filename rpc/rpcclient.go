package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/omnilaboratory/obd/config"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/tidwall/gjson"
)

var connConfig *ConnConfig

func init() {
	connConfig = &ConnConfig{
		Host: config.ChainNode_Host,
		User: config.ChainNode_User,
		Pass: config.ChainNode_Pass,
	}
	//log.Println(connConfig)
}

type ConnConfig struct {
	// Host is the IP address and port of the remote omnicore server you want to connect to.
	Host string
	// User is the username used in authentification by the remote RPC server.
	User string
	// Pass is the passphrase used in authentification.
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
	return fmt.Sprintf("Code: %d,Msg: %s", r.Code, r.Message)
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
func (client *Client) CheckVersion() error {
	result, err := client.OmniGetinfo()
	if err != nil {
		return err
	}

	log.Println("omniCoreVersion: "+gjson.Get(result, "omnicoreversion").String()+",", "bitcoinCoreVersion: "+gjson.Get(result, "bitcoincoreversion").String())
	bitcoinCoreVersion := gjson.Get(result, "bitcoincoreversion").String()

	infoes := strings.Split(bitcoinCoreVersion, ".")
	tempInt, _ := strconv.Atoi(infoes[0])
	if tempInt >= 0 {
		return nil
	}
	tempInt, _ = strconv.Atoi(infoes[1])
	if tempInt >= 18 {
		return nil
	}
	return errors.New("error bitcoinCore version " + gjson.Get(result, "bitcoincoreversion").String())
}

func (client *Client) send(method string, params []interface{}) (result string, err error) {
	rawParams := make([]json.RawMessage, 0, len(params))
	for _, item := range params {
		marshaledParam, err := json.Marshal(item)
		if err == nil {
			rawParams = append(rawParams, marshaledParam)
		}
	}
	//method = "./omnicore-cli -conf=/root/.bitcoin/omnicore18data/bitcoin.conf "+method
	//log.Println("request to Rpc server:", method, params)
	req := &Request{
		Jsonrpc: "2.0",
		ID:      client.NextID(),
		Method:  strings.Trim(method, " "),
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
		return "", err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == 403 {
		err = fmt.Errorf("status code: %d, response: %q , your ip is not allowed", httpResponse.StatusCode, httpResponse.Status)
		return "", err
	}

	// Read the raw bytes and close the response.
	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}

	var resp rawResponse
	err = json.Unmarshal(respBytes, &resp)

	if err != nil {
		err = fmt.Errorf("error reading json reply: %v", err)
		return "", err
	}

	if err != nil {
		err = fmt.Errorf("status code: %d, response: %q", httpResponse.StatusCode, err.Error())
		return "", err
	}

	res, err := resp.result()
	if httpResponse.StatusCode != 200 || err != nil {
		if err == nil {
			err = fmt.Errorf("status code: %d, response: %q", httpResponse.StatusCode, httpResponse.Status)
		}
		return "", err
	}
	return gjson.Parse(string(res)).String(), nil
}
