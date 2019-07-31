package main

import (
	"LightningOnOmni/routers"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"log"
	"net/http"
	"time"
)

func main() {

	rpcClient()

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

func rpcClient() *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         "62.234.216.108:18332",
		User:         "omniwallet",
		Pass:         "cB3]iL2@eZ1?cB2?",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Shutdown()

	//var account string ="ab"
	//btcjson.NewGetNewAddressCmd(&account);

	address, err := client.GetNewAddress("")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(address)
	}

	result, err := client.GetMyInfo()
	fmt.Println(result)

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("Block count: %d", blockCount)

	return client
}
