# OmniGo | in Golang
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/omnilaboratory/obd/blob/master/LICENSE) [![](https://img.shields.io/badge/golang-%3E%3D1.9.0-orange)](https://golang.org/dl/) [![](https://img.shields.io/badge/Spec-OmniLayer-brightgreen)](https://github.com/OmniLayer/spec) 
  

OmniGo implements part of Omnilayer spec in Go (golang), for constructing omni transaction offline, which is frequently used by obd lightning channels in payments. With OmniGo, OBD will have such advantages:  

* Quick deployment. Less dependency to a full node, which may take weeks to sync blocks database.  
* High efficiency in constructing transactions, hence HTLCs. Benchmark below reports 900+ TPS.  
* Agile architecture. Both high end server and small size device can deploy a lightning node.  

Full node will be used in funding/closing channels, monioring/punishing counterparties, broadcasting txs. These are not high frequent operatiosn and can be delegated to trackers.  

## Construct simple send transaction

[Example using omnicore rpc](https://github.com/OmniLayer/omnicore/wiki/Use-the-raw-transaction-API-to-create-a-Simple-Send-transaction) has the same output hex as the following code:   

## 1) List unspent outputs

List unspent outputs for 1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc.
```
omnicore-cli "listunspent" 0 999999 '["1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc"]'
```

```
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
```

## 2) Construct payload

```
const jsonStream = `
		{ "txid" : "c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de" , "vout" : 0} 
		{ "txid" : "ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3" , "vout" : 2 } 
	   ` 
	 
	payload, payload_hex := Omni_createpayload_simplesend("2", "0.1", true)
	fmt.Println("expect: 00000000000000020000000000989680")
	fmt.Println("export:")
	fmt.Println(payload_hex)
```

## 3) Construct transaction base
```
tx, tx_hex, _ := CreateRawTransaction(jsonStream, wire.TxVersion)
	fmt.Printf("The final base transaction hex")
	fmt.Printf("%s \n ", tx_hex)
```

## 4) Attach payload output
```
tx, _ = Omni_createrawtx_opreturn(tx, payload, payload_hex)

	expected := "0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff010000000000000000166a146f6d6e690000000000000002000000000098968000000000"
	fmt.Println("step 4 expects:")
	fmt.Println(expected)

	fmt.Println("step 4 exports: ")
	exported := TxToHex(tx)
	fmt.Println(exported)

	if strings.Compare(expected, exported) == 0 {
		fmt.Println("succeed in 4) Attach reference/receiver output")
	}

```

## 5) Attach reference/receiver output
```
//5) Attach reference/receiver output
	receiver := "1Njbpr7EkLA1R8ag8bjRN7oks7nv5wUn3o"
	tx, _ = Omni_createrawtx_reference(tx, receiver, &chaincfg.MainNetParams)

	expected = "0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff020000000000000000166a146f6d6e690000000000000002000000000098968022020000000000001976a914ee692ea81da1b12d3dd8f53fd504865c9d843f5288ac00000000"
	fmt.Println("step 5 expects: ")
	fmt.Println(expected)

	fmt.Println("step 5 exports: ")
	exported = TxToHex(tx)
	fmt.Println(exported)

	if strings.Compare(expected, exported) == 0 {
		fmt.Println("succeed in 5) Attach reference/receiver output")
	}
```

## 6) Specify miner fee and attach change output (as needed)

```
fmt.Println("step 6: attach change output to the sender")

	const prev_txs_json = `
		{"txid":"c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de","vout":0,"scriptPubKey":"76a914c6734676a08e3c6438bd95fa62c57939c988a17b88ac","value":"0.001"} 
		{"txid":"ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3","vout":2,"scriptPubKey":"76a914c6734676a08e3c6438bd95fa62c57939c988a17b88ac","value":"0.0083566"}
		`
	tx, _ = Omni_createrawtx_change(tx, prev_txs_json,
		"1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc",
		"0.0006", &chaincfg.MainNetParams)

	expected = "0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff036a5a0d00000000001976a914c6734676a08e3c6438bd95fa62c57939c988a17b88ac0000000000000000166a146f6d6e690000000000000002000000000098968022020000000000001976a914ee692ea81da1b12d3dd8f53fd504865c9d843f5288ac00000000"
	fmt.Println("step 6 expects: ")
	fmt.Println(expected)

	exported = TxToHex(tx)
	fmt.Println("step 6 exports: ")
	fmt.Println(exported)
```

## 7) Decode Raw Transaction Hex
```
fmt.Println("Hex from step 6 may be different than what expected, because omnicore c++ reorder the outputs. ")
	fmt.Println("Use DecodeRawtrnasaction to verify the two hexs. ")

	transaction_json_expected := DecodeRawTransaction(expected, &chaincfg.MainNetParams)
	transaction_json_exported := DecodeRawTransaction(exported, &chaincfg.MainNetParams)

	fmt.Println("expect")
	fmt.Println(transaction_json_expected)
	fmt.Println("export")
	fmt.Println(transaction_json_exported)
```






