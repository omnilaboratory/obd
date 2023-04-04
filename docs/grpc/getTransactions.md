## GetTransactions

GetTransactions returns a list describing all the known transactions relevant to the wallet.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| start_height   |	int32	    |The height from which to list transactions, inclusive. If this value is greater than end_height, transactions will be read in reverse.|  
| end_height     |	int32	    |The height until which to list transactions, inclusive. To include unconfirmed transactions, this value should be set to -1, which will return transactions from start_height until the current chain tip and unconfirmed transactions. If no end_height is provided, the call will default to this option.|
| account     |	string	    |An optional filter to only include transactions relevant to an account.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| transactions     |	Transaction[]	    |The list of transactions relevant to the wallet.|

**Transaction**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| tx_hash   |	string	    |The transaction hash.|  
| amount     |	int64	    |The transaction amount, denominated in satoshis.|
| num_confirmations     |	int32	    |The number of confirmations.|
| block_hash     |	string	    |The hash of the block this transaction was included in.|
| block_height     |	int32	    |The height of the block this transaction was included in.|
| time_stamp     |	int64	    |Timestamp of this transaction.|
| total_fees     |	int64	    |Fees paid for this transaction.|
| dest_addresses     |	string[]	    |Addresses that received funds for this transaction. Deprecated as it is now incorporated in the output_details field.|
| output_details     |	OutputDetail[]	    |Outputs that received funds for this transaction.|
| raw_tx_hex     |	string	    |The raw transaction hex.|
| label     |	string	    |A label that was optionally set on transaction broadcast.|
| previous_outpoints     |	PreviousOutPoint[]	    |PreviousOutpoints/Inputs of this transaction.|

**Transaction**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| output_type   |	OutputScriptType	    |The type of the output.|  
| address     |	string	    |The address.|
| pk_script     |	string	    |The pkscript in hex.|
| output_index     |	int64	    |The output index used in the raw transaction.|
| amount     |	int64	    |The value of the output coin in satoshis.|
| is_our_address     |	bool	    |Denotes if the output is controlled by the internal wallet.|

**PreviousOutPoint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| outpoint   |	string	    |The outpoint in format txid:n.|  
| is_our_output     |	bool	    |Denotes if the outpoint is controlled by the internal wallet. The flag will only detect p2wkh, np2wkh and p2tr inputs as its own.|

**OutputScriptType**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| SCRIPT_TYPE_PUBKEY_HASH   |	0	    | |  
| SCRIPT_TYPE_SCRIPT_HASH     |	1	    | |
| SCRIPT_TYPE_WITNESS_V0_PUBKEY_HASH     |	2	    | |
| SCRIPT_TYPE_WITNESS_V0_SCRIPT_HASH     |	3	    | |
| SCRIPT_TYPE_PUBKEY     |	4	    | |
| SCRIPT_TYPE_MULTISIG     |	5	    | |
| SCRIPT_TYPE_NULLDATA     |	6	    | |
| SCRIPT_TYPE_NON_STANDARD     |	7	    | |
| SCRIPT_TYPE_WITNESS_UNKNOWN     |	 8	    | |
| SCRIPT_TYPE_WITNESS_V1_TAPROOT     |	 9	    | |

## Example:

<!--
java code example
-->

```java
Obdmobile.getTransactions(LightningOuterClass.GetTransactionsRequest.newBuilder().build().toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {
        e.printStackTrace();
    }

    @Override
    public void onResponse(byte[] bytes) {
        if (bytes == null) {
            return;
        }
        try {
            LightningOuterClass.TransactionDetails resp = LightningOuterClass.TransactionDetails.parseFrom(bytes);           
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }
    }
});
```

<!--
The response for the example
-->
response:
```
[
    transactions {
      amount: 18892
      block_hash: "00000000000054c7b6dd51f64ccdb21e75ea3205ea32ad158381197c7f89fdd5"
      block_height: 2425273
      dest_addresses: "msajYt7BaFpa8G2EiqDYaj4uLqo3o1Ksmh"
      dest_addresses: "moR475qgPtKpb3znbuevyGK5zNbsEfCBmD"
      label: "0:closechannel:shortchanid-2666612565498134529"
      num_confirmations: 4
      raw_tx_hex: "02000000017843bdc84346e86fb708cd940e8d481f4a7cdeabdb7272b4a5b9818cb5f3e7c801000000da0047304402205bae82b741d0ba16b05c2e5c7be0b8b4f9b6b4c424253ed3b40cf4108e6ebf1502206572dd0c47f928a9e66b1d5f82a31de21b6be21802af352f6a242e5c72e8e6b701483045022100f827ecf2ae466b051f4e815c6f1adc1769e393ac0ec7113466bfb7d594dcf9d602205575877b6a53c54c4e9748e70c3486eaa031a34d8b5c03dfc3f79f0cf58bba7401475221038296a4309c9e6b102e4ee871adf1f23b5a7bdb4ab08701714e6377da5cff8502103f52464de2b8ffba354fc44439a74db8b204670f93e345742851b38dad505e85652aeffffffff030000000000000000216a1f6f6d6e6900000007800005e80201000000005f5e100020000000005f5e10022020000000000001976a914845888e933764d28f5591f4027a4cc934598b64c88accc490000000000001976a91456a35c756353fbb8a9dbbeba58ff6e894c2c753a88ac00000000
      time_stamp: 167939157
      total_fees: 0
      tx_hash: "84b195a40a969b791b509c7a722cf0289a3979feda86d65d0891e9dd547f6ea0"
    }
    transactions {
      amount: -20510
      block_hash: "0000000000000465d280fd0f6f9833bbae02992c7d6dc7daf23a81ef7ddab14b"
      block_height: 2425270
      dest_addresses: "2N76MnDzN16Cc9415gZHxq4vTjJtQNpWZ6S"
      dest_addresses: "moR475qgPtKpb3znbuevyGK5zNbsEfCBmD"
      label: "0:openchannel:shortchanid-2666612565498134529"
      num_confirmations: 7
      raw_tx_hex: "0200000001bc8e48e019c79ab663d4e8be0f5745ceb3067a3c92e98a5db81dc4430ac563c4010000006a47304402204dc63e41a03c2ff881c7c0f4767b23d66193ecf224ee0f5e9ad3e4e3b9945dd802204d95ce5b00108dad6e0b1cd88193864dc2551d8bd45a3345ce63d4923142d6b5012103e95394729ca398a4ca0c886ae63d7e818344da76bd1d6cb42cf6c9e036d7908900000000030000000000000000166a146f6d6e6900000000800005e8000000000bebc200204e00000000000017a91497e485960053edded935b4dd5b2893ec5154c9a98762971100000000001976a91456a35c756353fbb8a9dbbeba58ff6e894c2c753a88ac00000000"
      time_stamp: 1679389104
      total_fees: 510
      tx_hash: "c8e7f3b58c81b9a5b47272dbabde7c4a1f488d0e94cd08b76fe84643c8bd4378"
    }
    transactions {
      amount: 1173376
      block_hash: "0000000000008ff937f4a42b2df754f54862954ea82606e34e0f15074af39bb0"
      block_height: 2424793
      dest_addresses: "mmcksPxvD9eQc1vdJcTuq3xHFKUhQe1Y7g"
      dest_addresses: "moR475qgPtKpb3znbuevyGK5zNbsEfCBmD"
      num_confirmations: 484
      raw_tx_hex: "02000000000101c4dc89ae1551bacb7e09442e877fb1c0b6aa8eca0fdf4fdcd362c2f9a2c4db240100000000feffffff0249055d74030000001976a91442e9b69824cca155c938b2625c4055936f694b3588ac80e71100000000001976a91456a35c756353fbb8a9dbbeba58ff6e894c2c753a88ac024730440220679b425fb929dca1bba219ede5886486d0c0e305cf00d8fe064b9c3b10034044022022feabd65e9c0e2fad40d60795df33fcafa63c71eae7c6443e53b8f0c00eff7701210368411b53f8f64150b7675c2595a98c8bef49efa9b715b647ce0b5d531430add4d8ff2400"
      time_stamp: 1679045972
      total_fees: 0
      tx_hash: "c463c50a43c41db85d8ae9923c7a06b3ce45570fbee8d463b69ac719e0488ebc"
    }           
]
```