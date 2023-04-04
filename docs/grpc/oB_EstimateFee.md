## OB_EstimateFee

OB_EstimateFee asks the chain backend to estimate the fee rate and total fees for a transaction that pays to multiple specified outputs.

When using REST, the AddrToAmount map type can be set by appending &AddrToAmount[<address>]=<amount_to_send> to the URL. Unfortunately this map type doesn't appear in the REST API documentation because of a bug in the grpc-gateway library.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| asset_id   |	uint64	    |The ID of an asset.|
| addr   |	String	    |The address to send coins to.|
| from   |	String	    |The address to send coins from.|
| target_conf   |	int32	    |The target number of blocks that this transaction should be confirmed by.|
| amount   |	int64	    |The amount in satoshis to send.|
| asset_amount   |	int64	    |The asset amount in satoshis to send.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| fee_sat     |	int64	    |The total fee in satoshis.|
| feerate_sat_per_byte     |	int64	    |TDeprecated, use sat_per_vbyte. The fee rate in satoshi/vbyte.|
| sat_per_vbyte     |	uint64	    |The fee rate in satoshi/vbyte.|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.ObEstimateFeeRequest asyncEstimateFeeRequest;
if (assetId == 0) {
    asyncEstimateFeeRequest = LightningOuterClass.ObEstimateFeeRequest.newBuilder()
            .setAddr("mrJRHs3LkvtSnpAWHHr8fB8N5k3HxNEwKQ")
            .setFrom("moR475qgPtKpb3znbuevyGK5zNbsEfCBmD")
            .setAmount(100000000)
            .setTargetConf(1)
            .build();
} else {
    asyncEstimateFeeRequest = LightningOuterClass.ObEstimateFeeRequest.newBuilder()
            .setAssetId((int) 2147485160)
            .setAddr("mrJRHs3LkvtSnpAWHHr8fB8N5k3HxNEwKQ")
            .setFrom("moR475qgPtKpb3znbuevyGK5zNbsEfCBmD")
            .setAssetAmount(100000000)
            .setTargetConf(1)
            .build();
        }
Obdmobile.oB_EstimateFee(asyncEstimateFeeRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {

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