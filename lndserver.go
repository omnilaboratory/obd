package main

import (
	"LightningOnOmni/routers"
	"LightningOnOmni/rpc"
	"log"
	"net/http"
	"time"
)

func main() {

	testRpcToChainNode()
	return

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

	config := &rpc.ConnConfig{
		Host: "62.234.216.108:18332",
		User: "omniwallet",
		Pass: "cB3]iL2@eZ1?cB2?",
	}

	client := rpc.NewClient(config)

	//tx0 := "4031d11e9bd5bbf27419cba674fdeaf063aa0391fb897129f30b2fce638a89be"
	//result, e := client.GetTransacionById(tx0)
	//if e != nil {
	//	log.Println(e)
	//	return
	//}
	//details := gjson.Get(result, "details")
	//for _, item := range details.Array() {
	//	log.Println(item)
	//	log.Println(item.Get("address"))
	//}

	//
	//s, e := client.GetTxOut("4031d11e9bd5bbf27419cba674fdeaf063aa0391fb897129f30b2fce638a89be",0)
	//if e!=nil {
	//	log.Println(e)
	//}
	//log.Println(gjson.Get(s,"confirmations"))

	var keys []string
	keys = append(keys, "n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn")
	keys = append(keys, "n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA")
	result, e := client.CreateMultiSig(1, keys)
	if e != nil {
		log.Println(e)
	}
	log.Println(result)

	//result, e := client.GetTransactionById("4031d11e9bd5bbf27419cba674fdeaf063aa0391fb897129f30b2fce638a89be")
	//if e!=nil {
	//	log.Println(e)
	//	return
	//}
	//log.Println(result)

}
