package omnicore

import (
	"fmt"
	"log"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

//https://github.com/OmniLayer/omnicore/pull/1252
// example 1:
// omni_createpayload_sendtomany 31
// 				'[
//					{"output": 2, "amount": "10.5"},
//					{"output": 3, "amount": "0.5"},
//				 	{"output": 5, "amount": "15.0"}
//				 ]'
// expected export: 000000070000001f0302000000003e95ba80030000000002faf080050000000059682f00
//
func TestOmniCreatePayloadSendToMany(t *testing.T) {

	var property_id uint32 = 31
	const output_list = `
						{"output": 2, "address": "", "amount": "10.5"} 
						{"output": 3, "address": "", "amount": "0.5"} 
						{"output": 5, "address": "", "amount": "15.0"}
						`
	divisible := true

	receiver_array := ExtractReceiverList(output_list, divisible)
	_, payload_hex := OmniCreatePayloadSendToMany(property_id, receiver_array, divisible)
	fmt.Println("expect: 000000070000001f0302000000003e95ba80030000000002faf080050000000059682f00")

	fmt.Println(payload_hex)

}

//https://github.com/OmniLayer/omnicore/pull/1252
// unspent list is from example 4:
// "vin": [
//    {
//		"txid": "2a89b07484fe8aa2b3038d32e2888b9b30de4553e23136e324502a0f049ed6b1",
//		"vout": 3,
//		"scriptSig": {
//		  "asm": "00149cae6190b0802e646e93b9006b7239b036002ca0",
//		  "hex": "1600149cae6190b0802e646e93b9006b7239b036002ca0"
//		},
//		"txinwitness": [
//		  "30440220651224e5c625d8163cb6e2971f58b191109b5dc02fb821694ada765ab424b23602206cd8787258b3538f78f157ac6d1d520a21d3e1febda9477d77120e171d0225a301",
//		  "02d0c8e0f374949dae9c0855d56af23b4d64c3c128fe336cd36a5270b2898a9964"
//		],
//	  }
//	]
//
// The total value is 4.99982296, according to the sum of the outputs 1,2,3: 2 dust + 1 change
// dust 0.00000546 + dust 0.00000540 + change 4.99981210 = 4.99982296
// manually craft a unspent list
//  '[
//		{"txid":"2a89b07484fe8aa2b3038d32e2888b9b30de4553e23136e324502a0f049ed6b1", "vout": 3, "value":"5"}
//	]'
//
// example 2:
// omni_sendtomany "2N9UopnCBgbC6wAjMaiqrBZjmgNFiZccY7C"
//					3
//					'[
//						{"address": "mnM5bS3qQTxZSHkgpp2LDcKYL5thWtqCD6", "amount": "3"},
//						{"address": "2N9UopnCBgbC6wAjMaiqrBZjmgNFiZccY7C", "amount": "5"}
//					]'
//
// expected export is the transaction hex: 55d9f3eb56a5de5afef6b1d5dcdc2300d26d0aafcacbbfe387cba0983e0644fa
//
// In this test, add output index to the receiver_list, for example:
// output_list = `
//					{"output": 1, "address": "mnM5bS3qQTxZSHkgpp2LDcKYL5thWtqCD6", "amount": "3"}
//					{"output": 2, "address": "2N9UopnCBgbC6wAjMaiqrBZjmgNFiZccY7C", "amount": "5"}
//				`
// The output 0 is for payload.
func TestOmniCreateSendToManyTransaction(t *testing.T) {

	const prev_tx_list = `
						{"txid":"2a89b07484fe8aa2b3038d32e2888b9b30de4553e23136e324502a0f049ed6b1", 
						"vout": 3, 
						"scriptPubKey": "",
						"value": "4.99982296"}  
						`

	const unspent_list = `
						{"txid":"2a89b07484fe8aa2b3038d32e2888b9b30de4553e23136e324502a0f049ed6b1", 
						"vout": 3}  
						`

	const receiver_list = `
						{"output": 1, "address": "mnM5bS3qQTxZSHkgpp2LDcKYL5thWtqCD6", "amount": "3"}
						{"output": 2, "address": "2N9UopnCBgbC6wAjMaiqrBZjmgNFiZccY7C", "amount": "5"}
					`
	divisible := true
	receivers_array := ExtractReceiverList(receiver_list, divisible)
	var btc_version int32 = 1
	miner_fee_in_btc := "0.0006"
	from_address := "2N9UopnCBgbC6wAjMaiqrBZjmgNFiZccY7C"
	var property_id uint32 = 3

	//default_net := &chaincfg.MainNetParams  &chaincfg.MainNetParams
	//default_net := &chaincfg.RegressionNetParams // the regtest

	tx, _, err := OmniCreateSendToManyTransaction(from_address,
		unspent_list,
		prev_tx_list,
		property_id,
		receivers_array,
		divisible,
		btc_version,
		miner_fee_in_btc,
		&chaincfg.RegressionNetParams)

	if err != nil {
		log.Fatal(err)
	}

	exported := TxToHex(tx)
	tx_hash := tx.TxHash()

	fmt.Println(exported)
	fmt.Println("tx hash is: ")
	fmt.Println(tx_hash)
}

// Example for adding redeem script to each receiver's output.
func TestOmniCreateSendToManyScriptHashTransaction(t *testing.T) {

	const prev_tx_list = `
						{"txid":"2a89b07484fe8aa2b3038d32e2888b9b30de4553e23136e324502a0f049ed6b1", 
						"vout": 3, 
						"scriptPubKey": "",
						"value": "4.99982296"}  
						`

	const unspent_list = `
						{"txid":"2a89b07484fe8aa2b3038d32e2888b9b30de4553e23136e324502a0f049ed6b1", 
						"vout": 3}  
						`

	const receiver_list = `
						{"output": 1, "address": "mnM5bS3qQTxZSHkgpp2LDcKYL5thWtqCD6", "amount": "3"}
						{"output": 2, "address": "2N9UopnCBgbC6wAjMaiqrBZjmgNFiZccY7C", "amount": "5"}
					`
	divisible := true
	receivers_array := ExtractReceiverList(receiver_list, divisible)
	var btc_version int32 = 1
	miner_fee_in_btc := "0.0006"
	from_address := "2N9UopnCBgbC6wAjMaiqrBZjmgNFiZccY7C"
	var property_id uint32 = 3

	// add hash locker for example
	script := txscript.NewScriptBuilder()
	locker := btcutil.Hash160([]byte("12341234123412341234"))
	script.AddOp(txscript.OP_HASH160)
	script.AddData(locker)
	script.AddOp(txscript.OP_EQUALVERIFY)
	var script_byts []byte
	script_byts, _ = script.Script()

	script_array := make([][]byte, 2)
	script_array[0] = []byte{}
	script_array[1] = script_byts

	tx, _, err := OmniCreateSendToManyScriptHashTransaction(from_address,
		unspent_list,
		prev_tx_list,
		property_id,
		receivers_array,
		script_array,
		divisible,
		btc_version,
		miner_fee_in_btc,
		&chaincfg.RegressionNetParams)

	if err != nil {
		log.Fatal(err)
	}

	exported := TxToHex(tx)
	tx_hash := tx.TxHash()

	fmt.Println(exported)
	fmt.Println("tx hash is: ")
	fmt.Println(tx_hash)
}
