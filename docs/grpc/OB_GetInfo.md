## OB_GetInfo

OB_GetInfo returns general information concerning the lightning node including it's identity pubkey, alias, the chains it is connected to, and information concerning the number of open+pending channels.

## Arguments:
This request has no parameters.

## Response:
| Field		         |	gRPC Type		|	   Description    |
| -------- 	         |	---------       |      ---------      |  
| version            |	string	        |The version of the LND software that the node is running.|
| commit_hash            |	string	        |The SHA1 commit hash that the daemon is compiled with.|
| identity_pubkey            |	string	        |The identity pubkey of the current node.|
| alias            |	string	        |If applicable, the alias of the current node, e.g. "bob"|
| color            |	string	        |The color of the current node in hex code format.|
| num_pending_channels            |	uint32	        |Number of pending channels.|
| num_active_channels            |	uint32	        |Number of active channels.|
| num_inactive_channels            |	uint32	        |Number of inactive channels.|
| num_peers            |	uint32	        |Number of peers.|
| block_height            |	uint32	        |The node's current view of the height of the best block.|
| block_hash            |	string	        |The node's current view of the hash of the best block.|
| best_header_timestamp            |	int64	        |Timestamp of the block best known to the wallet.|
| synced_to_chain            |	bool	        |Whether the wallet's view is synced to the main chain.|
| synced_to_graph            |	bool	        |Whether we consider ourselves synced with the public channel graph.|
| testnet            |	bool	        |Whether the current node is connected to testnet. This field is deprecated and the network field should be used instead.|
| chains            |	Chain[]	        |A list of active chains the node is connected to.|
| uris            |	string[]	        |The URIs of the current node.|
| features            |	FeaturesEntry[]	        |Features that our node has advertised in our init message, node announcements and invoices.|
| require_htlc_interceptor            |	bool	        |Indicates whether the HTLC interceptor API is in always-on mode.|
| store_final_htlc_resolutions            |	bool	        |Indicates whether final htlc resolutions are stored on disk.|

**Chain**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| chain   |	string	    |    The blockchain the node is on (eg bitcoin, litecoin).|  
| network     |	string	    |    The network the node is on (eg regtest, testnet, mainnet).|

**FeaturesEntry**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	uint32	    |    |  
| value     |	Feature	    |    |

**Feature**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| name   |	string	    |    |  
| is_required     |	bool	    |    |
| is_known     |	bool	    |    |

## Example:

<!--
java code example
-->

```java
Obdmobile.oB_GetInfo(LightningOuterClass.GetInfoRequest.newBuilder().build().toByteArray(), new Callback() {
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
            LightningOuterClass.GetInfoResponse resp = LightningOuterClass.GetInfoResponse.parseFrom(bytes);
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
    alias: "alice"
    best_header_timestamp: 1678778514
    block_hash: "0000000003ebfac2c84263680d3441a58118b30a7f5523c088d1440f94faca91"
    block_height: 2424335
    chains {
      chain: "bitcoin"
      network: "testnet"
    }
    color: "#3399ff"
    commit_hash: "e7fde326e9aab360c6d92955f6877931c2b25d05"
    features {
      key: 17
      value {
        is_known: true
        name: "multi-path-payments"
      }
    }
    features {
      key: 23
      value {
        is_known: true
        name: "anchors-zero-fee-htlc-tx"
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
      key: 7
      value {
        is_known: true
        name: "gossip-queries"
      }
    }
    features {
      key: 31
      value {
        is_known: true
        name: "amp"
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
      key: 30
      value {
        is_known: true
        is_required: true
        name: "amp"
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
      key: 9
      value {
        is_known: true
        name: "tlv-onion"
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
      key: 2023
      value {
        is_known: true
        name: "script-enforced-lease"
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
    identity_pubkey: "03ae7822b1fb00b0b465bd647adae597e89b69e38d0190bf2df992377c19745426"
    num_active_channels: 2
    num_peers: 1
    pubkey_bech32: "ln1qwh8sg43lvqtpdr9h4j84kh9jl5fk60r35qep0edlxfrwlqew32zv0yv3sq"
    synced_to_chain: true
    synced_to_graph: true
    testnet: true
    version: "0.14.2-beta commit=ticker/v1.1.0-492-ge7fde326e-dirty"
}
```
