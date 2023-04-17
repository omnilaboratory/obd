## OB_SafeBox_CloseChannel

OB_SafeBox_CloseChannel attempts to close an active channel identified by its channel outpoint (ChannelPoint). The actions of this method can additionally be augmented to attempt a force close after a timeout period in the case of an inactive peer. If a non-force close (cooperative closure) is requested, then the user can specify either a target number of blocks until the closure transaction is confirmed, or a manual fee rate. If neither are specified, then a default lax, block confirmation target is used.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| channel_point   |	ChannelPoint	    |The outpoint (txid:index) of the funding transaction. With this value, Bob will be able to generate a signature for Alice's version of the commitment transaction.|
| force   |	bool	    |If true, then the channel will be closed forcibly. This means the current commitment transaction will be signed and broadcast.|
| target_conf   |	int32	    |The target number of blocks that the closure transaction should be confirmed by.|
| sat_per_byte   |	int64	    |Deprecated, use sat_per_vbyte. A manual fee rate set in sat/vbyte that should be used when crafting the closure transaction.|
| delivery_address   |	string	    |An optional address to send funds to in the case of a cooperative close. If the channel was opened with an upfront shutdown script and this field is set, the request to close will fail because the channel must pay out to the upfront shutdown addresss.|
| sat_per_vbyte   |	uint64	    |A manual fee rate set in sat/vbyte that should be used when crafting the closure transaction.|
| max_fee_per_vbyte   |	uint64	    |The maximum fee rate the closer is willing to pay. NOTE: This field is only respected if we're the initiator of the channel.|

**ChannelPoint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| funding_txid_bytes   |	bytes	    |Txid of the funding transaction. When using REST, this field must be encoded as base64.|
| funding_txid_str   |	string	    |Hex-encoded string representing the byte-reversed hash of the funding transaction.|
| output_index   |	uint32	    |The index of the output of the funding transaction.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| close_pending     |	PendingUpdate	    | |
| chan_close     |	ChannelCloseUpdate	    | |

**PendingUpdate**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| txid   |	bytes	    | |
| output_index   |	uint32	    | |

**ChannelCloseUpdate**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| closing_txid   |	bytes	    | |
| success   |	bool	    | |

## Example:

<!--
java code example
-->

```java
String mChannelPoint = "095d9402b8cf68ba3937c582eb6fdb51b31a7af8d9ae53044e303c860644a694:0"
LightningOuterClass.ChannelPoint point = LightningOuterClass.ChannelPoint.newBuilder()
                .setFundingTxidStr(mChannelPoint.substring(0, mChannelPoint.indexOf(':')))
                .setOutputIndex(Character.getNumericValue(mChannelPoint.charAt(mChannelPoint.length() - 1)))
                .build();
LightningOuterClass.CloseChannelRequest closeChannelRequest = LightningOuterClass.CloseChannelRequest.newBuilder()
        .setChannelPoint(point)
        .build();
Obdmobile.ob_SafeBox_CloseChannel(closeChannelRequest.toByteArray(), new RecvStream() {
    @Override
    public void onError(Exception e) {
        if (e.getMessage().equals("EOF")) {
            return;
        }
        e.printStackTrace()       
    }

    @Override
    public void onResponse(byte[] bytes) {
        try {
            LightningOuterClass.CloseStatusUpdate resp = LightningOuterClass.CloseStatusUpdate.parseFrom(bytes);
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
   close_pending{
    txid: "] 350\024 350 227g 340 241 236_(341251 334f<a 344 246 255%L 315 346 303tL 301 277\003(375\366"
   }
   chan_close {
    closing_txid: "] 350\024 350 227g 340 241 236_(341251 334f<a 344 246 255%L 315 346 303tL 301 277\003(375\366"
   }
}
```