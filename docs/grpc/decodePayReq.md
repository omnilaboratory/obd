## DecodePayReq

DecodePayReq takes an encoded payment request string and attempts to decode it, returning a full description of the conditions encoded within the payment request.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |
| pay_req   |	string	    |The payment request string to be decoded.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| asset_id   |	uint64	    |The ID of an asset.|
| amount   |	int64	    | |
| amt_msat   |	int64	    | |
| destination     |	string	    | |
| payment_hash     |	string	    | |
| num_satoshis     |	int64	    | |
| timestamp     |	int64	    | |
| expiry     |	int64	    | |
| description     |	string	    | |
| description_hash     |	string	    | |
| fallback_addr     |	string	    | |
| cltv_expiry     |	int64	    | |
| route_hints     |	RouteHint[]	    | |
| payment_addr     |	bytes	    | |
| num_msat     |	int64	    | |
| features     |	FeaturesEntry[]	    | |

**RouteHint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| hop_hints   |	HopHint[]	    |A list of hop hints that when chained together can assist in reaching a specific destination.|

**FeaturesEntry**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	uint32		    | |
| value   |	Feature	    | |

**HopHint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| node_id   |	string	    |The public key of the node at the start of the channel.|
| chan_id   |	uint64	    |The unique identifier of the channel.|
| fee_base_msat   |	uint32	    |The base fee of the channel denominated in millisatoshis.|
| fee_proportional_millionths   |	uint32	    |The fee rate of the channel for sending one satoshi across it denominated in millionths of a satoshi.|
| cltv_expiry_delta   |	uint32	    |The time-lock delta of the channel.|

**Feature**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| name   |	string		    | |
| is_required   |	bool	    | |
| is_known   |	bool	    | |

## Example:

<!--
java code example
-->

```java
LightningOuterClass.PayReqString decodePaymentRequest = LightningOuterClass.PayReqString.newBuilder()
                .setPayReq("obto2147485160:11pjp6a0wpp5jzzalddsjjmfvkgp7l5rhvw2qqgqg8awddypfcafrcksrlz3z48qdq2dpsksctgvycqzpgxqyz5vq3q8zqqqp0gsp5lkvghy5rqguw9zj0djytpwh5wthhgfskas8dpcht4vnkxgn4axss9qyyssqd7c0xwc5q7tswf8x9q7h27jmmghr4hxk56vgu3t8guyktsp7rnksu4g6n8y4c4rvww8z59fa0p7l2m4ypszrwg8us93lwfl2f22hx5gpq0cugx")
                .build();
Obdmobile.decodePayReq(decodePaymentRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.PayReq resp = LightningOuterClass.PayReq.parseFrom(bytes);
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
    amount: 100000000
    amt_msat: 0
    asset_id: -2147482136
    cltv_expiry: 40
    destination: "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
    expiry: 86400
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
      key: 17
      value {
        is_known: true
        name: "multi-path-payments"
      }
    }
    payment_addr: "!_\300\276\353\326\347B\'\0169\257\2238\025\254\241\211\2413\177>\271\177\275\2453\261\267\306\257\375"
    payment_hash: "41908fbdeccc2682718cde46bfb627f544bc9a3843457243cf6d4e8194b6928f"
    timestamp: 1679725314
}
```
