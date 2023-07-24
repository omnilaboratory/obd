## pendingChannels

PendingChannels returns a list of all the channels that are currently considered "pending". A channel is pending if it has finished the funding workflow and is waiting for confirmations for the funding txn, or is in the process of closure, either initiated cooperatively or non-cooperatively.

## Arguments:
This request has no parameters.

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| total_limbo_balance     |	int64	    |The balance in satoshis encumbered in pending channels.|
| pending_open_channels     |	PendingOpenChannel[]	    |Channels pending opening.|
| pending_closing_channels     |	ClosedChannel[]	    |Deprecated: Channels pending closing previously contained cooperatively closed channels with a single confirmation. These channels are now considered closed from the time we see them on chain.|
| pending_force_closing_channels     |	ForceClosedChannel[]	    |Channels pending force closing.|
| waiting_close_channels     |	WaitingCloseChannel[]	    |Channels waiting for closing tx to confirm.|

## Example:

<!--
java code example
-->

```java
Obdmobile.pendingChannels(LightningOuterClass.PendingChannelsRequest.newBuilder().build().toByteArray(), new Callback() {
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
            LightningOuterClass.PendingChannelsResponse resp = LightningOuterClass.PendingChannelsResponse.parseFrom(bytes);
            mPendingOpenChannelsList = resp.getPendingOpenChannelsList();
            mPendingClosedChannelsList = resp.getPendingClosingChannelsList();
            mPendingForceClosedChannelsList = resp.getPendingForceClosingChannelsList();
            mPendingWaitingCloseChannelsList = resp.getWaitingCloseChannelsList();
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
 PendingOpenChannels[
       {
        channel {
        asset_capacity: 0
        asset_id: -2147482136
        create_time: 1687795462
        btc_capacity: 20000
        channel_point: "3f6a515cfd28ebc48067ea341018c892cc67f725df2e9f42704a4782493341c0:1"
        commitment_type: ANCHORS
        commitment_type_value: 3
        initiator: INITIATOR_LOCAL
        initiator_value: 1
        local_balance: 2500000000
        local_chan_reserve_sat: 50000000
        num_forwarding_packages: 0
        remote_balance: 2500000000
        remote_chan_reserve_sat: 50000000
        remote_node_pub: "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
        }
        commit_fee: 562
        commit_weight: 860
        fee_per_kw: 500
       }
    ]
    WaitingCloseChannels[
       {
        channel {
        asset_capacity: 100000
        asset_id: -2147482136
        create_time: 1687795462
        btc_capacity: 20000
        chan_status_flags: "ChanStatusCoopBroadcasted|ChanStatusLocalCloseInitiator"
        channel_point: "088408eb1fb17e8f9b65f5ee56c0820dd7c412f0a5f276195b84c096d6c846ff:1"
        commitment_type: ANCHORS
        commitment_type_value: 3
        initiator: INITIATOR_LOCAL
        initiator_value: 1
        local_balance: 50000
        local_chan_reserve_sat: 1000
        num_forwarding_packages: 0
        remote_balance: 50000
        remote_chan_reserve_sat: 1000
        remote_node_pub: "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
        }
        commitments {
        local_commit_fee_sat: 304
        local_txid: "53c95e80f986c455cceeca842233b15659894497b7cff9a92a899fc7488c68cc"
        remote_commit_fee_sat: 304
        remote_pending_commit_fee_sat: 0
        remote_txid: "5d0636a05e8d2b4e6fc83c303df5661e5fd8376a46851c560b1a39cd72b5e31c"
        }
        limbo_balance: 50000
       } 
    ]   
}
```