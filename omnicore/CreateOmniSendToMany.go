package omnicore

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

// https://github.com/OmniLayer/omnicore/pull/1252
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
// cpp code in omnicore::  src/omnicore/rpctx.cpp
// omni_sendtomany(const JSONRPCRequest& request)
//
// The output of the transaction is as follows:(https://gist.github.com/dexX7/1138fd1ea084a9db56798e9bce50d0ef#example-three-receivers)
//
// Output Index	|	Script type
//		0		|	Payload with commands
//		1		|	Pay-to-pubkey-hash (recipient 1)
//		2		|	Pay-to-pubkey-hash (recipient 2)
//		3		|	Pay-to-pubkey-hash (not relevant)
//		4		|	Pay-to-script-hash (recipient 3)
//		5		|	Pay-to-pubkey-hash (our change)
//
// In this inplementation, add output index to the receiver_list, for example:
// output_list = `
//					{"output": 1, "address": "mnM5bS3qQTxZSHkgpp2LDcKYL5thWtqCD6", "amount": "3"}
//					{"output": 2, "address": "2N9UopnCBgbC6wAjMaiqrBZjmgNFiZccY7C", "amount": "5"}
//				`
// The output 0 is for payload.

func OmniCreateSendToManyTransaction(from_addres string,
	unspent_list string,
	prev_tx_list string,
	property_id uint32,
	receivers_array []PayloadOutput,
	divisible bool,
	btc_version int32,
	miner_fee_in_btc string,
	defaultNet *chaincfg.Params) (*wire.MsgTx, string, error) {

	// step 2: create payload
	payload, payload_hex := OmniCreatePayloadSendToMany(property_id, receivers_array, divisible)

	// step 3: create raw bitcoin transaction by unspent list
	tx, _, err := CreateRawTransaction(unspent_list, btc_version)

	// step 4: attach payload to output 0
	tx, err = Omni_createrawtx_opreturn(tx, payload, payload_hex)

	// step 5: Attach multiple reference/receiver output
	for output_index := 1; output_index <= len(receivers_array); output_index++ {
		tx, err = Omni_createrawtx_reference(tx, receivers_array[output_index-1].Address, defaultNet)
	}

	// step 6: Specify miner fee and attach change output (as needed)
	// generally, the change returns back to the sender.
	tx, _ = Omni_createrawtx_change(tx, prev_tx_list, from_addres, miner_fee_in_btc, defaultNet)

	return tx, "", err
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

