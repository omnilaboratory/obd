## listChannels

ListChannels returns a description of all the open channels that this node is a participant in.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| active_only   |	bool	    |    |  
| inactive_only     |	bool	    |    |
| public_only     |	bool	    |    |
| private_only     |	bool	    |    |
| peer     |	bytes	    |Filters the response for channels with a target peer's pubkey. If peer is empty, all channels will be returned.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| channels     |	Channel[]	    |channels.|

**Channel**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| active   |	bool	    |Whether this channel is active or not.|  
| remote_pubkey     |	string	    |The identity pubkey of the remote node.|
| channel_point     |	string	    |The outpoint (txid:index) of the funding transaction. With this value, Bob will be able to generate a signature for Alice's version of the commitment transaction.|
| chan_id     |	uint64	    |The unique channel ID for the channel. The first 3 bytes are the block height, the next 3 the index within the block, and the last 2 bytes are the output index for the channel.|
| capacity     |	int64	    |The total amount of funds held in this channel.|
| local_balance     |	int64	    |This node's current balance in this channel.|
| remote_balance     |	int64	    |The counterparty's current balance in this channel.|
| commit_fee     |	int64	    |The amount calculated to be paid in fees for the current set of commitment transactions. The fee amount is persisted with the channel in order to allow the fee amount to be removed and recalculated with each channel state update, including updates that happen after a system restart.|
| commit_weight     |	int64	    |The weight of the commitment transaction.|
| fee_per_kw     |	int64	    |The required number of satoshis per kilo-weight that the requester will pay at all times, for both the funding transaction and commitment transaction. This value can later be updated once the channel is open.|
| unsettled_balance     |	int64	    |The unsettled balance in this channel.|
| total_satoshis_sent     |	int64	    |The total number of satoshis we've sent within this channel.|
| total_satoshis_received     |	int64	    |The total number of satoshis we've received within this channel.|
| num_updates     |	uint64	    |The total number of updates conducted within this channel.|
| pending_htlcs     |	HTLC[]	    |The list of active, uncleared HTLCs currently pending within the channel.|
| csv_delay     |	uint32	    |Deprecated. The CSV delay expressed in relative blocks. If the channel is force closed, we will need to wait for this many blocks before we can regain our funds.|
| private     |	bool	    |Whether this channel is advertised to the network or not.|
| initiator     |	bool	    |True if we were the ones that created the channel.|
| chan_status_flags     |	string	    |A set of flags showing the current state of the channel.|
| local_chan_reserve_sat     |	int64	    |Deprecated. The minimum satoshis this node is required to reserve in its balance.|
| remote_chan_reserve_sat     |	int64	    |Deprecated. The minimum satoshis the other node is required to reserve in its balance.|
| static_remote_key     |	bool	    |Deprecated. Use commitment_type.|
| commitment_type     |	CommitmentType	    |The commitment type used by this channel.|
| lifetime     |	int64	    |The number of seconds that the channel has been monitored by the channel scoring system. Scores are currently not persisted, so this value may be less than the lifetime of the channel [EXPERIMENTAL].|
| uptime     |	int64	    |The number of seconds that the remote peer has been observed as being online by the channel scoring system over the lifetime of the channel [EXPERIMENTAL].|
| close_address     |	string	    |Close address is the address that we will enforce payout to on cooperative close if the channel was opened utilizing option upfront shutdown. This value can be set on channel open by setting close_address in an open channel request. If this value is not set, you can still choose a payout address by cooperatively closing with the delivery_address field set.|
| push_amount_sat     |	uint64	    |The amount that the initiator of the channel optionally pushed to the remote party on channel open. This amount will be zero if the channel initiator did not push any funds to the remote peer. If the initiator field is true, we pushed this amount to our peer, if it is false, the remote peer pushed this amount to us.|
| thaw_height     |	uint32	    |This uint32 indicates if this channel is to be considered 'frozen'. A frozen channel doest not allow a cooperative channel close by the initiator. The thaw_height is the height that this restriction stops applying to the channel. This field is optional, not setting it or using a value of zero will mean the channel has no additional restrictions. The height can be interpreted in two ways: as a relative height if the value is less than 500,000, or as an absolute height otherwise.|
| local_constraints     |	ChannelConstraints	    |List constraints for the local node.|
| remote_constraints     |	ChannelConstraints	    |List constraints for the remote node.|
| alias_scids     |	uint64[]	    |This lists out the set of alias short channel ids that exist for a channel. This may be empty.|
| zero_conf     |	bool	    |Whether or not this is a zero-conf channel.|
| zero_conf_confirmed_scid     |	uint64	    |This is the confirmed / on-chain zero-conf SCID.|

**ChannelConstraints**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| csv_delay   |	uint32	    |The CSV delay expressed in relative blocks. If the channel is force closed, we will need to wait for this many blocks before we can regain our funds.|  
| chan_reserve_sat     |	uint64	    |The minimum satoshis this node is required to reserve in its balance.|
| dust_limit_sat     |	uint64	    |The dust limit (in satoshis) of the initiator's commitment tx.|
| max_pending_amt_msat     |	uint64	    |The maximum amount of coins in millisatoshis that can be pending in this channel.|
| min_htlc_msat     |	uint64	    |The smallest HTLC in millisatoshis that the initiator will accept.|
| max_accepted_htlcs     |	uint32	    |The total number of incoming HTLC's that the initiator will accept.|

**HTLC**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| incoming   |	bool	    | |  
| amount     |	int64	    | |
| hash_lock     |	bytes	    | |
| expiration_height     |	uint32	    | |
| htlc_index     |	uint64	    |Index identifying the htlc on the channel.|
| forwarding_channel     |	uint64	    |If this HTLC is involved in a forwarding operation, this field indicates the forwarding channel. For an outgoing htlc, it is the incoming channel. For an incoming htlc, it is the outgoing channel. When the htlc originates from this node or this node is the final destination, forwarding_channel will be zero. The forwarding channel will also be zero for htlcs that need to be forwarded but don't have a forwarding decision persisted yet.|
| forwarding_htlc_index     |	uint64	    |Index identifying the htlc on the forwarding channel.|

**CommitmentType**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| UNKNOWN_COMMITMENT_TYPE   |	0	    |Returned when the commitment type isn't known or unavailable.|  
| LEGACY     |	1	    |A channel using the legacy commitment format having tweaked to_remote keys.|
| STATIC_REMOTE_KEY     |	2	    |A channel that uses the modern commitment format where the key in the output of the remote party does not change each state. This makes back up and recovery easier as when the channel is closed, the funds go directly to that key.|
| ANCHORS     |	3	    |A channel that uses a commitment format that has anchor outputs on the commitments, allowing fee bumping after a force close transaction has been broadcast.|
| SCRIPT_ENFORCED_LEASE     |	4	    |A channel that uses a commitment type that builds upon the anchors commitment format, but in addition requires a CLTV clause to spend outputs paying to the channel initiator. This is intended for use on leased channels to guarantee that the channel initiator has no incentives to close a leased channel before its maturity date.|

## Example:

<!--
java code example
-->

```java
Obdmobile.listChannels(LightningOuterClass.ListChannelsRequest.newBuilder().build().toByteArray(), new Callback() {
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
            LightningOuterClass.ListChannelsResponse resp = LightningOuterClass.ListChannelsResponse.parseFrom(bytes);
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
  {
    active: true
    asset_capacity: 0
    btc_capacity: 20000
    chan_id: 2665559233359052800
    chan_status_flags: "ChanStatusDefault"
    channel_point: "095d9402b8cf68ba3937c582eb6fdb51b31a7af8d9ae53044e303c860644a694:0"
    close_address: "moR475qgPtKpb3znbuevyGK5zNbsEfCBmD"
    commit_fee: 1222
    commit_weight: 692
    commitment_type: ANCHORS
    commitment_type_value: 3
    csv_delay: 144
    fee_per_kw: 500
    initiator: true
    lifetime: 62
    local_asset_balance: 0
    local_balance: 8778
    local_chan_reserve_sat: 354
    local_constraints {
      chan_reserve_sat: 354
      csv_delay: 144
      dust_limit_sat: 354
      max_accepted_htlcs: 483
      max_pending_amt_msat: 19800000
      min_htlc_msat: 1
    }
    num_updates: 0
    push_asset_amount_sat: 0
    push_btc_amount_sat: 10000
    remote_asset_balance: 0
    remote_balance: 10000
    remote_chan_reserve_sat: 354
    remote_constraints {
      chan_reserve_sat: 354
      csv_delay: 144
      dust_limit_sat: 354
      max_accepted_htlcs: 483
      max_pending_amt_msat: 19800000
      min_htlc_msat: 1
    }
    remote_pubkey: "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
    total_satoshis_received: 0
    total_satoshis_sent: 0
    unsettled_balance: 0
  }    
]
```