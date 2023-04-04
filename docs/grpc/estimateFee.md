## EstimateFee

EstimateFee asks the chain backend to estimate the fee rate and total fees for a transaction that pays to multiple specified outputs.

When using REST, the AddrToAmount map type can be set by appending &AddrToAmount[<address>]=<amount_to_send> to the URL. Unfortunately this map type doesn't appear in the REST API documentation because of a bug in the grpc-gateway library.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| AddrToAmount   |	AddrToAmountEntry[]	    |The map from addresses to amounts for the transaction.|
| target_conf   |	int32	    |The target number of blocks that this transaction should be confirmed by.|
| min_confs   |	int32	    |The minimum number of confirmations each one of your outputs used for the transaction must satisfy.|
| spend_unconfirmed   |	bool	    |Whether unconfirmed outputs should be used as inputs for the transaction.|

**AddrToAmountEntry**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	string	    | |
| value   |	int64	    | |

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| fee_sat     |	int64	    |The total fee in satoshis.|
| feerate_sat_per_byte     |	int64	    |Deprecated, use sat_per_vbyte. The fee rate in satoshi/vbyte.|
| sat_per_vbyte     |	uint64	    |The fee rate in satoshi/vbyte.|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.EstimateFeeRequest asyncEstimateFeeRequest = LightningOuterClass.EstimateFeeRequest.newBuilder()
                .putAddrToAmount("moR475qgPtKpb3znbuevyGK5zNbsEfCBmD", 100000000)
                .setTargetConf(1)
                .build();
Obdmobile.estimateFee(asyncEstimateFeeRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.EstimateFeeResponse resp = LightningOuterClass.EstimateFeeResponse.parseFrom(bytes);
            long feeStr = resp.getFeeSat();
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
    fee_sat: 510
    feerate_sat_per_byte: 2
    sat_per_vbyte: 2
}
```