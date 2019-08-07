package main

import (
	"LightningOnOmni/routers"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {

	testRpcToChainNode()

	routersInit := routers.InitRouter()
	server := &http.Server{
		Addr:           ":60020",
		Handler:        routersInit,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())
}

func testRpcToChainNode() {
	httpClient := http.Client{
		Transport: &http.Transport{
			Proxy:           nil,
			TLSClientConfig: nil,
		},
	}
	url := "http://62.234.216.108:18332"
	rawParams := make([]json.RawMessage, 0, 1)
	tx0 := "4031d11e9bd5bbf27419cba674fdeaf063aa0391fb897129f30b2fce638a89be"
	marshalledParam, err := json.Marshal(tx0)
	if err == nil {
		rawMessage := json.RawMessage(marshalledParam)
		rawParams = append(rawParams, rawMessage)
	}
	req := &Request{
		Jsonrpc: "1.0",
		ID:      1,
		Method:  "gettransaction",
		Params:  rawParams,
	}
	marshaledJSON, _ := json.Marshal(req)
	bodyReader := bytes.NewReader(marshaledJSON)
	httpReq, err := http.NewRequest("POST", url, bodyReader)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth("omniwallet", "cB3]iL2@eZ1?cB2?")
	httpResponse, err := httpClient.Do(httpReq)
	if err != nil {
		return
	}
	// Read the raw bytes and close the response.
	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	httpResponse.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading json reply: %v", err)
		return
	}
	var resp rawResponse
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		err = fmt.Errorf("status code: %d, response: %q",
			httpResponse.StatusCode, string(respBytes))
		return
	}
	res, err := resp.result()
	var f interface{}
	err = json.Unmarshal(res, &f)
	if err != nil {
		fmt.Println(err)
	}
	strJson := string(res)
	details := gjson.Get(strJson, "details")
	for _, item := range details.Array() {
		log.Println(item)
		log.Println(item.Get("address"))
	}
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
	panic("implement me")
}

type Request struct {
	Jsonrpc string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      interface{}       `json:"id"`
}

type RPCErrorCode int
type RPCError struct {
	Code    RPCErrorCode `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
}

//func rpcClient() *rpcclient.Client {
//	connCfg := &rpcclient.ConnConfig{
//		Host:         "62.234.216.108:18332",
//		User:         "omniwallet",
//		Pass:         "cB3]iL2@eZ1?cB2?",
//		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
//		DisableTLS:   true, // Bitcoin core does not provide TLS by default
//	}
//	client, err := rpcclient.New(connCfg, nil)
//	if err != nil {
//		fmt.Println(err)
//	}
//	defer client.Shutdown()
//
//	//var account string ="ab"
//	//btcjson.NewGetNewAddressCmd(&account);
//
//	str :=""
//	hashes, err := chainhash.NewHashFromStr(str)
//	tx, err := client.GetTransaction(hashes)
//	if err==nil {
//		log.Println(tx)
//	}
//
//	address, err := client.GetNewAddress("")
//	if err != nil {
//		fmt.Println(err)
//	} else {
//		fmt.Println(address)
//	}
//
//	result, err := client.GetMyInfo()
//	fmt.Println(result)
//
//	// Get the current block count.
//	blockCount, err := client.GetBlockCount()
//	if err != nil {
//		fmt.Println(err)
//	}
//	log.Printf("Block count: %d", blockCount)
//
//	return client
//}
