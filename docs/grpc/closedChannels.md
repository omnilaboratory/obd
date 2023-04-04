## ClosedChannels

ClosedChannels returns a description of all the closed channels that this node was a participant in.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| cooperative     |	bool	    | |
| local_force     |	bool	    | |
| remote_force     |	bool	    | |
| breach     |	bool	    | |
| funding_canceled     |	bool	    | |
| abandoned     |	bool	    | |

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| channels     |	ChannelCloseSummary[]	    | |

**ChannelCloseSummary**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| channel_point     |	string	    |The outpoint (txid:index) of the funding transaction.|
| chan_id     |	uint64	    |The unique channel ID for the channel.|
| chain_hash     |	string	    |The hash of the genesis block that this channel resides within.|
| closing_tx_hash     |	string	    |The txid of the transaction which ultimately closed this channel.|
| remote_pubkey     |	string	    |Public key of the remote peer that we formerly had a channel with.|
| capacity     |	int64	    |Total capacity of the channel.|
| close_height     |	uint32	    |Height at which the funding transaction was spent.|
| settled_balance     |	int64	    |Settled balance at the time of channel closure.|
| time_locked_balance     |	int64	    |The sum of all the time-locked outputs at the time of channel closure.|
| close_type     |	ClosureType	    |Details on how the channel was closed.|
| open_initiator     |	Initiator	    |Open initiator is the party that initiated opening the channel. Note that this value may be unknown if the channel was closed before we migrated to store open channel information after close.|
| close_initiator     |	Initiator	    |Close initiator indicates which party initiated the close. This value will be unknown for channels that were cooperatively closed before we started tracking cooperative close initiators. Note that this indicates which party initiated a close, and it is possible for both to initiate cooperative or force closes, although only one party's close will be confirmed on chain.|
| resolutions     |	Resolution[]	    | |
| alias_scids     |	uint64[]	    |This lists out the set of alias short channel ids that existed for the closed channel. This may be empty.|
| zero_conf_confirmed_scid     |	uint64	    |The confirmed SCID for a zero-conf channel.|

**ClosureType**
| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| COOPERATIVE_CLOSE     |	0	    | |
| LOCAL_FORCE_CLOSE     |	1	    | |
| REMOTE_FORCE_CLOSE     |	2	    | |
| BREACH_CLOSE     |	3	    | |
| FUNDING_CANCELED     |	4	    | |
| ABANDONED     |	5	    | |

**Initiator**
| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| INITIATOR_UNKNOWN     |	0	    | |
| INITIATOR_LOCAL     |	1	    | |
| INITIATOR_REMOTE     |	2	    | |
| INITIATOR_BOTH     |	3	    | |

**Resolution**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| resolution_type     |	ResolutionType	    |The type of output we are resolving.|
| outcome     |	ResolutionOutcome	    |The outcome of our on chain action that resolved the outpoint.|
| outpoint     |	OutPoint	    |The outpoint that was spent by the resolution.|
| amount_sat     |	uint64	    |The amount that was claimed by the resolution.|
| sweep_txid     |	string	    |The hex-encoded transaction ID of the sweep transaction that spent the output.|

**ResolutionType**
| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| TYPE_UNKNOWN     |	0	    | |
| ANCHOR     |	1	    |We resolved an anchor output.|
| INCOMING_HTLC     |	2	    |We are resolving an incoming htlc on chain. This if this htlc is claimed, we swept the incoming htlc with the preimage. If it is timed out, our peer swept the timeout path.|
| OUTGOING_HTLC     |	3	    |We are resolving an outgoing htlc on chain. If this htlc is claimed, the remote party swept the htlc with the preimage. If it is timed out, we swept it with the timeout path.|
| COMMIT     |	4	    |We force closed and need to sweep our time locked commitment output.|

**ResolutionOutcome**
| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| OUTCOME_UNKNOWN     |	0	    |Outcome unknown.|
| CLAIMED     |	1	    |An output was claimed on chain.|
| UNCLAIMED     |	2	    |An output was left unclaimed on chain.|
| ABANDONED     |	3	    |ResolverOutcomeAbandoned indicates that an output that we did not claim on chain, for example an anchor that we did not sweep and a third party claimed on chain, or a htlc that we could not decode so left unclaimed.|
| FIRST_STAGE     |	4	    |If we force closed our channel, our htlcs need to be claimed in two stages. This outcome represents the broadcast of a timeout or success transaction for this two stage htlc claim.|
| TIMEOUT     |	5	    |A htlc was timed out on chain.|

**OutPoint**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| txid_bytes     |	bytes	    |Raw bytes representing the transaction id.|
| txid_str     |	string	    |Reversed, hex-encoded string representing the transaction id.|
| output_index     |	uint32	    |The index of the output on the transaction.|

## Example:

<!--
java code example
-->

```java
Obdmobile.closedChannels(LightningOuterClass.ClosedChannelsRequest.newBuilder().build().toByteArray(), new Callback() {
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
            LightningOuterClass.ClosedChannelsResponse resp = LightningOuterClass.ClosedChannelsResponse.parseFrom(bytes);
            mClosedChannelsList = resp.getChannelsList();
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
  channels:[
              asset_capacity: 1000000000
              asset_id: -2147482136
              btc_capacity: 20000
              chain_hash: "000000000933ea01ad0ee984209779baaec3ced90fa3f408719526f8d77f4943"
              chan_id: 2666073804802228225
              channel_point: "60eba778e3e0e154385e9665b2c2f6808a5a5818bf1448cd415372430c6d5dc3:1"
              close_height: 2424791
              close_initiator: INITIATOR_LOCAL
              close_initiator_value: 1
              closing_tx_hash: "f9282bb7665cca998b55f5a3efc6286964c572ba76b4873cbbe5263cc837b04d"
              open_initiator: INITIATOR_LOCAL
              open_initiator_value: 1
              remote_pubkey: "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
              settled_balance: 18892
              time_locked_balance: 0
            ]
}
```