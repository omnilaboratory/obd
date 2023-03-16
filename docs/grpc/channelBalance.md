## ChannelBalance

ChannelBalance returns a report on the total funds(Bitcoin/Satoshi ) across all open channels, categorized in local/remote, pending local/remote and unsettled local/remote balances.

## Arguments:
This request has no parameters.

## Response:
| Field		         |	gRPC Type		|	   Description    |
| -------- 	         |	---------       |      ---------      |  
| balance            |	int64	        |Deprecated. Sum of channels balances denominated in satoshis.|
| pending_open_balance            |	int64	        |Deprecated. Sum of channels pending balances denominated in satoshis.|
| local_balance            |	Amount	        |Sum of channels local balances.|
| remote_balance            |	Amount	        |Sum of channels remote balances.|
| unsettled_local_balance            |	Amount	        |Sum of channels local unsettled balances.|
| unsettled_remote_balance            |	Amount	        |Sum of channels remote unsettled balances.|
| pending_open_local_balance            |	Amount	        |Sum of channels pending local balances.|
| pending_open_remote_balance            |	Amount	        |Sum of channels pending remote balances.|

**Amount**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| sat   |	uint64	    |    Value denominated in satoshis.|  
| msat     |	uint64	    |    Value denominated in milli-satoshis.|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.ChannelBalanceRequest channelBalanceRequest = LightningOuterClass.ChannelBalanceRequest.newBuilder()
        .setAssetId((int) propertyid)
        .build();
Obdmobile.channelBalance(channelBalanceRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.ChannelBalanceResponse resp = LightningOuterClass.ChannelBalanceResponse.parseFrom(bytes);
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
    balance: 0
    local_balance {
      msat: 8778000
    }
    pending_open_balance: 0
    pending_open_local_balance {
      msat: 0
    }
    pending_open_remote_balance {
      msat: 0
    }
    remote_balance {
      msat: 10000000
    }
    unsettled_local_balance {
      msat: 0
    }
    unsettled_remote_balance {
      msat: 0
    }
}
```
