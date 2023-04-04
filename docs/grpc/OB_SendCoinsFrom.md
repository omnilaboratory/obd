## OB_SendCoinsFrom

OB_SendCoinsFrom executes a request to send coins to a particular address. Unlike SendMany, this RPC call only allows creating a single output at a time. If neither target_conf, or sat_per_vbyte are set, then the internal wallet will consult its fee model to determine a fee for the default confirmation target.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| asset_id   |	uint64	    |The ID of an asset.|
| addr   |	string	    |The address to send coins to.|
| from   |	String	    |The address to send coins from.|
| amount   |	int64	    |The amount in satoshis to send.|
| asset_amount   |	int64	    |The asset amount in satoshis to send.|
| target_conf   |	int32	    |The target number of blocks that this transaction should be confirmed by.|
| sat_per_vbyte   |	uint64	    |A manual fee rate set in sat/vbyte that should be used when crafting the transaction.|
| sat_per_byte   |	int64	    |Deprecated, use sat_per_vbyte. A manual fee rate set in sat/vbyte that should be used when crafting the transaction.|
| send_all   |	bool	    |If set, then the amount field will be ignored, and lnd will attempt to send all the coins under control of the internal wallet to the specified address.|
| label   |	string	    |An optional label for the transaction, limited to 500 characters.|
| min_confs   |	int32	    |The minimum number of confirmations each one of your outputs used for the transaction must satisfy.|
| spend_unconfirmed   |	bool	    |Whether unconfirmed outputs should be used as inputs for the transaction.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| txid     |	string	    |The transaction ID of the transaction.|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.SendCoinsFromRequest sendRequest;
if (assetId == 0) {
    sendRequest = LightningOuterClass.SendCoinsFromRequest.newBuilder()
            .setAddr("mrJRHs3LkvtSnpAWHHr8fB8N5k3HxNEwKQ")
            .setFrom("moR475qgPtKpb3znbuevyGK5zNbsEfCBmD")
            .setAmount(100000000)
            .setTargetConf(1)
            .build();
} else {
    sendRequest = LightningOuterClass.SendCoinsFromRequest.newBuilder()
            .setAssetId((int) 2147485160)
            .setAddr("mrJRHs3LkvtSnpAWHHr8fB8N5k3HxNEwKQ")
            .setFrom("moR475qgPtKpb3znbuevyGK5zNbsEfCBmD")
            .setAssetAmount(100000000)
            .setTargetConf(1)
            .build();
        }
Obdmobile.oB_SendCoinsFrom(sendRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {

    }

    @Override
    public void onResponse(byte[] bytes) {
        if (bytes == null) {
            return;
        }
        try {
            LightningOuterClass.SendCoinsResponse resp = LightningOuterClass.SendCoinsResponse.parseFrom(bytes);
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
    txid: "7280ed4975df80612d7a91349202bf58a683b5e9eacfdoec6686bfed579bf40c"
}
```
