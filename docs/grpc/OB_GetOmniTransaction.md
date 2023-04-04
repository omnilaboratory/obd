## OB_GetOmniTransaction

OB_GetOmniTransaction is used for getting detailed information about an Omni transaction.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| txid   |	string	    |The hash of the transaction to lookup.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| txid     |	string	    |The hex-encoded hash of the transaction.|
| sendingaddress     |	string	    |The Bitcoin address of the sender.|
| referenceaddress     |	string	    |A Bitcoin address used as reference (if any).|
| ismine     |	bool	    |Whether the transaction involes an address in the wallet.|
| confirmations     |	int32	    |The number of transaction confirmations.|
| fee     |	string	    |The transaction fee in bitcoins.|
| blocktime     |	int64	    |The timestamp of the block that contains the transaction.|
| valid     |	bool	    |Whether the transaction is valid.|
| positioninblock     |	int32	    |The position (index) of the transaction within the block.|
| version     |	int32	    |The transaction version.|
| type_int     |	int32	    |The transaction type as number.|
| type     |	string	    |Other transaction type specific properties.|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.GetOmniTransactionRequest getOmniTransactionRequest = LightningOuterClass.GetOmniTransactionRequest.newBuilder()
                .setTxid("088408eb1fb17e8f9b65f5ee56c0820dd7c412f0a5f276195b84c096d6c846ff")
                .build();
Obdmobile.oB_GetOmniTransaction(getOmniTransactionRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.AssetTx resp = LightningOuterClass.AssetTx.parseFrom(bytes);
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
{
    amount: "10.00000000"
    block: 2424771
    blockhash: "0000000000000026b1b196a337fe089ddbb611c535bcd98297cb6fe02dafb78b"
    blocktime: 1679032184
    confirmations: 507
    divisible: true
    fee: "0.00000808"
    positioninblock: 25
    propertyid: 2147485160
    referenceaddress: "2MvNTdR6JH4crnqZ3eTYQnWgk3aJU7gCsnz"
    sendingaddress: "moR475qgPtKpb3znbuevyGK5zNbsEfCBmD"
    txid: "964c374b60d90fda598b8128f4e4f14d9772582d66d6d097bdcaf79489550c2f"
    type: "Simple Send"
    type_int: 0
    valid: true
    version: 0
}
```
