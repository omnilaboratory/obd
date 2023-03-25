## RouterSendToRouteV2

RouterSendToRouteV2 attempts to make a payment via the specified route. This method differs from SendPayment in that it allows users to specify a full route manually. This can be used for things like rebalancing, and atomic swaps.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |
| payment_hash   |	bytes	    |The payment hash to use for the HTLC.|
| route   |	Route	    |Route that should be used to attempt to complete the payment.|
| skip_temp_err   |	bool	    |Whether the payment should be marked as failed when a temporary error is returned from the given route. Set it to true so the payment won't be failed unless a terminal error is occurred, such as payment timeout, no routes, incorrect payment details, or insufficient funds.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| attempt_id     |	uint64	    |The unique ID that is used for this attempt.|
| status     |		HTLCStatus		    |The status of the HTLC.|
| route     |		Route		    |The route taken by this HTLC.|
| attempt_time_ns     |		int64		    |The time in UNIX nanoseconds at which this HTLC was sent.|
| resolve_time_ns     |		int64		    |The time in UNIX nanoseconds at which this HTLC was settled or failed. This value will not be set if the HTLC is still IN_FLIGHT.|
| failure     |		Failure		    |Detailed htlc failure info.|
| preimage     |		bytes		    |The preimage that was used to settle the HTLC.|

**HTLCStatus**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| IN_FLIGHT   |	0	    | |
| SUCCEEDED   |	1	    | |
| FAILED   |	2	    | |

**Route**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| asset_id   |	uint64	    |The ID of an asset.|
| total_time_lock   |	uint32	    |The cumulative (final) time lock across the entire route. This is the CLTV value that should be extended to the first hop in the route. All other hops will decrement the time-lock as advertised, leaving enough time for all hops to wait for or present the payment preimage to complete the payment.|
| total_fees   |	int64	    |The sum of the fees paid at each hop within the final route. In the case of a one-hop payment, this value will be zero as we don't need to pay a fee to ourselves.|
| total_amt   |	int64	    |The total amount of funds required to complete a payment over this route. This value includes the cumulative fees at each hop. As a result, the HTLC extended to the first-hop in the route will need to have at least this many satoshis, otherwise the route will fail at an intermediate node due to an insufficient amount of fees.|
| hops   |	Hops[]	    |Contains details concerning the specific forwarding details at each hop.|
| total_fees_msat   |	int64	    |The total fees in millisatoshis.|
| total_amt_msat   |	int64	    |The total amount in millisatoshis.|

**Failure**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| code   |	FailureCode	    |Failure code as defined in the Lightning spec.|
| channel_update   |	ChannelUpdate	    |An optional channel update message.|
| htlc_msat   |	uint64	    |A failure type-dependent htlc value.|
| onion_sha_256   |	bytes	    |The sha256 sum of the onion payload.|
| cltv_expiry   |	uint32	    |A failure type-dependent cltv expiry value.|
| flags   |	uint32	    |A failure type-dependent flags value.|
| failure_source_index   |	uint32	    |The position in the path of the intermediate or final node that generated the failure message. Position zero is the sender node.|
| height   |	uint32	    |A failure type-dependent block height.|

**Hop**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| chan_id   |	uint64	    |The unique channel ID for the channel. The first 3 bytes are the block height, the next 3 the index within the block, and the last 2 bytes are the output index for the channel.|
| chan_capacity   |	int64	    | |
| amt_to_forward   |	int64	    | |
| fee   |	int64	    | |
| expiry   |	uint32	    | |
| amt_to_forward_msat   |	int64	    | |
| fee_msat   |	int64	    | |
| pub_key   |	string	    |An optional public key of the hop. If the public key is given, the payment can be executed without relying on a copy of the channel graph.|
| tlv_payload   |	bool	    |If set to true, then this hop will be encoded using the new variable length TLV format. Note that if any custom tlv_records below are specified, then this field MUST be set to true for them to be encoded properly.|
| mpp_record   |	MPPRecord	    |An optional TLV record that signals the use of an MPP payment. If present, the receiver will enforce that the same mpp_record is included in the final hop payload of all non-zero payments in the HTLC set. If empty, a regular single-shot payment is or was attempted.|
| amp_record   |	AMPRecord	    |An optional TLV record that signals the use of an AMP payment. If present, the receiver will treat all received payments including the same (payment_addr, set_id) pair as being part of one logical payment. The payment will be settled by XORing the root_share's together and deriving the child hashes and preimages according to BOLT XX. Must be used in conjunction with mpp_record.|
| custom_records   |	CustomRecordsEntry[]	    |An optional set of key-value TLV records. This is useful within the context of the SendToRoute call as it allows callers to specify arbitrary K-V pairs to drop off at each hop within the onion.|
| metadata   |	bytes	    |The payment metadata to send along with the payment to the payee.|

**FailureCode**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| RESERVED   |	0	    |The numbers assigned in this enumeration match the failure codes as defined in BOLT #4. Because protobuf 3 requires enums to start with 0, a RESERVED value is added.|
| INCORRECT_OR_UNKNOWN_PAYMENT_DETAILS   |	1	    | |
| INCORRECT_PAYMENT_AMOUNT   |	2	    | |
| FINAL_INCORRECT_CLTV_EXPIRY   |	3	    | |
| FINAL_INCORRECT_HTLC_AMOUNT   |	4	    | |
| FINAL_EXPIRY_TOO_SOON   |	5	    | |
| INVALID_REALM   |	6	    | |
| EXPIRY_TOO_SOON   |	  7	    | |
| INVALID_ONION_VERSION   |	  8	    | |
| INVALID_ONION_HMAC   |	  9	    | |
| INVALID_ONION_KEY   |	  10	    | |
| AMOUNT_BELOW_MINIMUM   |	  11	    | |
| FEE_INSUFFICIENT   |	  12	    | |
| INCORRECT_CLTV_EXPIRY   |	  13	    | |
| CHANNEL_DISABLED   |	  14	    | |
| TEMPORARY_CHANNEL_FAILURE   |	  15	    | |
| REQUIRED_NODE_FEATURE_MISSING   |	  16	    | |
| REQUIRED_CHANNEL_FEATURE_MISSING   |	  17	    | |
| UNKNOWN_NEXT_PEER   |	  18	    | |
| TEMPORARY_NODE_FAILURE   |	  19	    | |
| PERMANENT_NODE_FAILURE   |	  20	    | |
| PERMANENT_CHANNEL_FAILURE   |	  21	    | |
| EXPIRY_TOO_FAR   |	  22    | |
| MPP_TIMEOUT   |	  23	    | |
| INVALID_ONION_PAYLOAD   |	  24	    | |
| INTERNAL_FAILURE   |	  997	    |An internal error occurred.|
| UNKNOWN_FAILURE   |	  998	    |The error source is known, but the failure itself couldn't be decoded.|
| UNREADABLE_FAILURE   |	  999	    |An unreadable failure result is returned if the received failure message cannot be decrypted. In that case the error source is unknown.|

**ChannelUpdate**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| signature   |	bytes	    |The signature that validates the announced data and proves the ownership of node id.|
| chain_hash   |	bytes	    |The target chain that this channel was opened within. This value should be the genesis hash of the target chain. Along with the short channel ID, this uniquely identifies the channel globally in a blockchain.|
| chan_id   |	uint64	    |The unique description of the funding transaction.|
| timestamp   |	uint32	    |A timestamp that allows ordering in the case of multiple announcements. We should ignore the message if timestamp is not greater than the last-received.|
| message_flags   |	uint32	    |The bitfield that describes whether optional fields are present in this update. Currently, the least-significant bit must be set to 1 if the optional field MaxHtlc is present.|
| channel_flags   |	uint32	    |The bitfield that describes additional meta-data concerning how the update is to be interpreted. Currently, the least-significant bit must be set to 0 if the creating node corresponds to the first node in the previously sent channel announcement and 1 otherwise. If the second bit is set, then the channel is set to be disabled.|
| time_lock_delta   |	uint32	    |The minimum number of blocks this node requires to be added to the expiry of HTLCs. This is a security parameter determined by the node operator. This value represents the required gap between the time locks of the incoming and outgoing HTLC's set to this node.|
| htlc_minimum_msat   |	uint64	    |The minimum HTLC value which will be accepted.|
| base_fee   |	uint32	    |The base fee that must be used for incoming HTLC's to this particular channel. This value will be tacked onto the required for a payment independent of the size of the payment.|
| fee_rate   |	uint32	    |The fee rate that will be charged per millionth of a satoshi.|
| htlc_maximum_msat   |	uint64	    |The maximum HTLC value which will be accepted.|
| extra_opaque_data   |	bytes	    |The set of data that was appended to this message, some of which we may not actually know how to iterate or parse. By holding onto this data, we ensure that we're able to properly validate the set of signatures that cover these new fields, and ensure we're able to make upgrades to the network in a forwards compatible manner.|

**MPPRecord**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| payment_addr   |	bytes	    |A unique, random identifier used to authenticate the sender as the intended payer of a multi-path payment. The payment_addr must be the same for all subpayments, and match the payment_addr provided in the receiver's invoice. The same payment_addr must be used on all subpayments.|
| total_amt_msat   |	int64	    |The total amount in milli-satoshis being sent as part of a larger multi-path payment. The caller is responsible for ensuring subpayments to the same node and payment_hash sum exactly to total_amt_msat. The same total_amt_msat must be used on all subpayments.|

**AMPRecord**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| root_share   |	bytes	    | |
| set_id   |	bytes	    | |
| child_index   |	uint32	    | |

**CustomRecordsEntry**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	uint64		    | |
| value   |	bytes	    | |

## Example:

<!--
java code example
-->

```java
LightningOuterClass.Route route = {
        asset_id: -2147482136
        hops {
          amt_to_forward: 0
          amt_to_forward_msat: 100000000
          chan_capacity: 0
          chan_id: 2666867652199055361
          expiry: 2425823
          fee: 0
          fee_msat: 0
          mpp_record {
            payment_addr: "!_\300\276\353\326\347B\'\0169\257\2238\025\254\241\211\2413\177>\271\177\275\2453\261\267\306\257\375"
            total_amt_msat: 100000000
          }
          pub_key: "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
          tlv_payload: true
        }
        total_amt_msat: 100000000
        total_fees_msat: 0
        total_time_lock: 2425823
      }
RouterOuterClass.SendToRouteRequest sendToRouteRequest = RouterOuterClass.SendToRouteRequest.newBuilder()
            .setPaymentHash(byteStringFromHex("fd4c4943bf84694847960d8b7c50f37a23d6e77196fec0b5aab69cba969f453f"))
            .setRoute(route)
            .build();
Obdmobile.routerSendToRouteV2(sendToRouteRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {
        e.printStackTrace();                
    }

    @Override
    public void onResponse(byte[] bytes) {
        try {
            LightningOuterClass.HTLCAttempt htlcAttempt = LightningOuterClass.HTLCAttempt.parseFrom(bytes);
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
    attempt_id: 18001
    attempt_time_ns: 1679731626595208560
    failure {
      code: INCORRECT_OR_UNKNOWN_PAYMENT_DETAILS
      code_value: 1
      failure_source_index: 1
      height: 2425788
      htlc_msat: 0
    }
    resolve_time_ns: 1679731626829642362
    route {
      asset_id: -2147482136
      hops {
        amt_to_forward: 0
        amt_to_forward_msat: 100000000
        chan_capacity: 0
        chan_id: 2666867652199055361
        expiry: 2425831
        fee: 0
        fee_msat: 0
        mpp_record {
          payment_addr: "r\212|H\\\276\004\306\307u\355?\262?C\006,\356\334B\fK8\210\224\331\214\r\000\246\2347"
          total_amt_msat: 100000000
        }
        pub_key: "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
        tlv_payload: true
      }
      total_amt_msat: 100000000
      total_fees_msat: 0
      total_time_lock: 2425831
    }
    status: FAILED
    status_value: 2
}
```

