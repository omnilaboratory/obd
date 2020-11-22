# OmniGo | in Golang
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/omnilaboratory/obd/blob/master/LICENSE) [![](https://img.shields.io/badge/golang-%3E%3D1.9.0-orange)](https://golang.org/dl/) [![](https://img.shields.io/badge/Spec-OmniLayer-brightgreen)](https://github.com/OmniLayer/spec) 
  

OmniGo implements part of Omnilayer spec in Go (golang), for constructing omni transaction offline, which is frequently used by obd lightning channels in payments. With OmniGo, OBD will have such advantages:  

* Less dependency to a full node, which may takes weeks to sync blocks database, so that any small size device can deploy a lightning node.  
* High efficiency in constructing transactions, hence HTLCs. Benchmark below reports 900+ TPS.  
* 

Full node will be used in opening and closing channels. 

## Construct simple send transaction

[Example using omnicore rpc](https://github.com/OmniLayer/omnicore/wiki/Use-the-raw-transaction-API-to-create-a-Simple-Send-transaction) has the same output hex as the following code:   

## 1) List unspent outputs

List unspent outputs for 1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc.
`
omnicore-cli "listunspent" 0 999999 '["1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc"]'
`
`
[
    ...,
    {
        "txid" : "c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de",
        "vout" : 0,
        "address" : "1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc",
        "account" : "",
        "scriptPubKey" : "76a914c6734676a08e3c6438bd95fa62c57939c988a17b88ac",
        "amount" : 0.00100000,
        "confirmations" : 0,
        "spendable" : true
    },
    {
        "txid" : "ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3",
        "vout" : 2,
        "address" : "1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc",
        "account" : "",
        "scriptPubKey" : "76a914c6734676a08e3c6438bd95fa62c57939c988a17b88ac",
        "amount" : 0.00835660,
        "confirmations" : 1416,
        "spendable" : true
    }
]
`

## 2) Construct payload

`
const jsonStream = `
		{ "txid" : "c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de" , "vout" : 0} 
		{ "txid" : "ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3" , "vout" : 2 } 
	   ` 
	 
payload, payload_hex := Omni_createpayload_simplesend("2", "0.1", true)
fmt.Println("expect: 00000000000000020000000000989680")
fmt.Println("export:")
fmt.Println(payload_hex)
`


