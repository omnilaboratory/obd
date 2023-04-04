## GetNodeInfo

GetNodeInfo returns the latest advertised, aggregated, and authenticated channel information for the specified node identified by its public key.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| pub_key     |	string	    |The 33-byte hex-encoded compressed public of the target node.|
| include_channels     |	bool	    |If true, will include all known channels associated with the node.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| node     |	LightningNode	    |An individual vertex/node within the channel graph. A node is connected to other nodes by one or more channel edges emanating from it. As the graph is directed, a node will also have an incoming edge attached to it for each outgoing edge.|
| num_channels     |	uint32	    |The total number of channels for the node.|
| total_capacity     |	int64	    |The sum of all channels capacity for the node, denominated in satoshis.|
| channels     |	ChannelEdge[]	    |A list of all public channels for the node.|

**LightningNode**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| last_update     |	uint32	    | |
| pub_key     |	string	    | |
| alias     |	string	    | |
| addresses     |	NodeAddress[]	    | |
| color     |	string	    | |
| features     |	FeaturesEntry[]	    | |
| custom_records     |	CustomRecordsEntry[]	    |Custom node announcement tlv records.|

**ChannelEdge**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| channel_id     |	uint64	    |The unique channel ID for the channel. The first 3 bytes are the block height, the next 3 the index within the block, and the last 2 bytes are the output index for the channel.|
| chan_point     |	string	    | |
| last_update     |	uint32	    | |
| node1_pub     |	string	    | |
| node2_pub     |	string	    | |
| capacity     |	int64	    | |
| node1_policy     |	RoutingPolicy	    | |
| node2_policy     |	RoutingPolicy	    | |
| custom_records     |	CustomRecordsEntry[]	    |Custom channel announcement tlv records.|

**NodeAddress**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| network     |	string	    | |
| addr     |	string	    | |

**FeaturesEntry**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key     |	uint32	    | |
| value     |	Feature	    | |

**Feature**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| name     |	string	    | |
| is_required     |	bool	    | |
| is_known     |	bool	    | |

**RoutingPolicy**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| time_lock_delta     |	uint32	    | |
| min_htlc     |	int64	    | |
| fee_base_msat     |	int64	    | |
| fee_rate_milli_msat     |	int64	    | |
| disabled     |	bool	    | |
| max_htlc_msat     |	uint64	    | |
| last_update     |	uint32	    | |
| custom_records     |	CustomRecordsEntry[]	    |Custom channel update tlv records.|

**CustomRecordsEntry**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key     |	uint64	    | |
| value     |	bytes	    | |


## Example:

<!--
java code example
-->

```java
LightningOuterClass.NodeInfoRequest nodeInfoRequest = LightningOuterClass.NodeInfoRequest.newBuilder()
                .setPubKey("03ae7822b1fb00b0b465bd647adae597e89b69e38d0190bf2df992377c19745426")
                .build();
Obdmobile.getNodeInfo(nodeInfoRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {
        e.printStackTrace();
    }

    @Override
    public void onResponse(byte[] bytes) {
        if (bytes == null) {
            return;
        }
        new Handler(Looper.getMainLooper()).post(new Runnable() {
            @Override
            public void run() {
                try {
                    LightningOuterClass.NodeInfo nodeInfo = LightningOuterClass.NodeInfo.parseFrom(bytes);
                } catch (InvalidProtocolBufferException e) {
                    e.printStackTrace();
                }
            }
        });
    }
});
```

<!--
The response for the example
-->
response:
```
{
  node{
        alias: "hahaha"
        color: "#3399ff"
        features {
          key: 31
          value {
            is_known: true
            name: "amp"
          }
        }
        features {
          key: 7
          value {
            is_known: true
            name: "gossip-queries"
          }
        }
        features {
          key: 0
          value {
            is_known: true
            is_required: true
            name: "data-loss-protect"
          }
        }
        features {
          key: 14
          value {
            is_known: true
            is_required: true
            name: "payment-addr"
          }
        }
        features {
          key: 17
          value {
            is_known: true
            name: "multi-path-payments"
          }
        }
        features {
          key: 2023
          value {
            is_known: true
            name: "script-enforced-lease"
          }
        }
        features {
          key: 9
          value {
            is_known: true
            name: "tlv-onion"
          }
        }
        features {
          key: 5
          value {
            is_known: true
            name: "upfront-shutdown-script"
          }
        }
        features {
          key: 45
          value {
            is_known: true
            name: "explicit-commitment-type"
          }
        }
        features {
          key: 12
          value {
            is_known: true
            is_required: true
            name: "static-remote-key"
          }
        }
        features {
          key: 23
          value {
            is_known: true
            name: "anchors-zero-fee-htlc-tx"
          }
        }
        last_update: 1679388098
        pub_key: "03ae7822b1fb00b0b465bd647adae597e89b69e38d0190bf2df992377c19745426"
      }
  num_channels: 6
  total_capacity: 7200120000
}
```