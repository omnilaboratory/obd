## RouterOB_SendPaymentV2

RouterOB_SendPaymentV2 attempts to route a payment described by the passed PaymentRequest to the final destination. The call returns a stream of payment updates.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |
| asset_id   |	uint64	    |The ID of an asset.|
| dest   |	bytes	    |The identity pubkey of the payment recipient.|
| asset_amt   |	int64	    |Number of omniAmount to send. this field is only for asset transfer.|
| amt_msat   |	int64	    |Number of millisatoshis to send. this field is only for btc transfer.|
| payment_hash   |	bytes	    |The hash to use within the payment's HTLC|
| final_cltv_delta   |	int32	    |The CLTV delta from the current height that should be used to set the timelock for the final hop.|
| payment_addr   |	bytes	    |An optional payment addr to be included within the last hop of the route.|
| payment_request   |	string	    |A bare-bones invoice for a payment within the Lightning Network. With the details of the invoice, the sender has all the data necessary to send a payment to the recipient. The amount in the payment request may be zero. In that case it is required to set the amt field as well. If no payment request is specified, the following fields are required: dest, amt and payment_hash.|
| timeout_seconds   |	int32	    |An upper limit on the amount of time we should spend when attempting to fulfill the payment. This is expressed in seconds. If we cannot make a successful payment within this time frame, an error will be returned. This field must be non-zero.|
| fee_limit_sat   |	int64	    |The maximum number of satoshis that will be paid as a fee of the payment. If this field is left to the default value of 0, only zero-fee routes will be considered. This usually means single hop routes connecting directly to the destination. To send the payment without a fee limit, use max int here. The fields fee_limit_sat and fee_limit_msat are mutually exclusive.|
| fee_limit_msat   |	int64	    |The maximum number of millisatoshis that will be paid as a fee of the payment. If this field is left to the default value of 0, only zero-fee routes will be considered. This usually means single hop routes connecting directly to the destination. To send the payment without a fee limit, use max int here. The fields fee_limit_sat and fee_limit_msat are mutually exclusive.|
| outgoing_chan_id   |	uint64	    |Deprecated, use outgoing_chan_ids. The channel id of the channel that must be taken to the first hop. If zero, any channel may be used (unless outgoing_chan_ids are set).|
| outgoing_chan_ids   |	uint64[]	    |The channel ids of the channels are allowed for the first hop. If empty, any channel may be used.|
| last_hop_pubkey   |	bytes	    |The pubkey of the last hop of the route. If empty, any hop may be used.|
| cltv_limit   |	int32	    |An optional maximum total time lock for the route. This should not exceed lnd's `--max-cltv-expiry` setting. If zero, then the value of `--max-cltv-expiry` is enforced.|
| route_hints   |	RouteHint[]	    |Optional route hints to reach the destination through private channels.|
| dest_custom_records   |	DestCustomRecordsEntry[]	    |An optional field that can be used to pass an arbitrary set of TLV records to a peer which understands the new records. This can be used to pass application specific data during the payment attempt. Record types are required to be in the custom range >= 65536. When using REST, the values must be encoded as base64.|
| allow_self_payment   |	bool	    |If set, circular payments to self are permitted.|
| dest_features   |	FeatureBit[]	    |Features assumed to be supported by the final node. All transitive feature dependencies must also be set properly. For a given feature bit pair, either optional or remote may be set, but not both. If this field is nil or empty, the router will try to load destination features from the graph as a fallback.|
| max_parts   |	uint32	    |The maximum number of partial payments that may be use to complete the full amount.|
| no_inflight_updates   |	bool	    |If set, only the final payment update is streamed back. Intermediate updates that show which htlcs are still in flight are suppressed.|
| max_shard_size_msat   |	uint64	    |The largest payment split that should be attempted when making a payment if splitting is necessary. Setting this value will effectively cause lnd to split more aggressively, vs only when it thinks it needs to. Note that this value is in milli-satoshis.|
| amp   |	bool	    |If set, an AMP-payment will be attempted.|
| time_pref   |	double	    |The time preference for this payment. Set to -1 to optimize for fees only, to 1 to optimize for reliability only or a value inbetween for a mix.

**RouteHint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| hop_hints   |	HopHint[]	    |A list of hop hints that when chained together can assist in reaching a specific destination.|

**DestCustomRecordsEntry**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	uint64		    | |
| value   |	bytes	    | |

**FeatureBit**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| DATALOSS_PROTECT_REQ   |	0	    | |  
| DATALOSS_PROTECT_OPT     |	1	    | |
| INITIAL_ROUING_SYNC     |	3	    | |
| UPFRONT_SHUTDOWN_SCRIPT_REQ     |	4	    | |
| UPFRONT_SHUTDOWN_SCRIPT_OPT     |	5	    | |
| GOSSIP_QUERIES_REQ     |	6	    | |
| GOSSIP_QUERIES_OPT     |	7	    | |
| TLV_ONION_REQ     |	8	    | |
| TLV_ONION_OPT     |	9	    | |
| EXT_GOSSIP_QUERIES_REQ     |	10	    | |
| EXT_GOSSIP_QUERIES_OPT     |	11	    | |
| STATIC_REMOTE_KEY_REQ     |	12	    | |
| STATIC_REMOTE_KEY_OPT     |	13	    | |
| PAYMENT_ADDR_REQ     |	14	    | |
| PAYMENT_ADDR_OPT     |	15	    | |
| MPP_REQ     |	16	    | |
| MPP_OPT     |	17	    | |
| WUMBO_CHANNELS_REQ     |	18	    | |
| WUMBO_CHANNELS_OPT     |	19	    | |
| ANCHORS_REQ     |	20	    | |
| ANCHORS_OPT     |	21	    | |
| ANCHORS_ZERO_FEE_HTLC_REQ     |	22	    | |
| ANCHORS_ZERO_FEE_HTLC_OPT     |	23	    | |
| AMP_REQ     |	30	    | |
| AMP_OPT     |	31	    | |

**HopHint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| node_id   |	string	    |The public key of the node at the start of the channel.|
| chan_id   |	uint64	    |The unique identifier of the channel.|
| fee_base_msat   |	uint32	    |The base fee of the channel denominated in millisatoshis.|
| fee_proportional_millionths   |	uint32	    |The fee rate of the channel for sending one satoshi across it denominated in millionths of a satoshi.|
| cltv_expiry_delta   |	uint32	    |The time-lock delta of the channel.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| asset_id   |	uint64	    |The ID of an asset.|
| payment_hash   |	string	    |The payment hash.|
| value   |	int64	    |Deprecated, use value_sat or value_msat.|
| creation_date   |	int64	    |Deprecated, use creation_time_ns.|
| fee   |	int64	    |Deprecated, use fee_sat or fee_msat.|
| payment_preimage   |	string	    |The payment preimage.|
| value_sat   |	int64	    |The value of the payment in satoshis.|
| value_msat   |	int64	    |The value of the payment in milli-satoshis.|
| payment_request   |	string	    |The optional payment request being fulfilled.|
| status   |	PaymentStatus	    |The status of the payment.|
| fee_sat   |	int64	    |The fee paid for this payment in satoshis.|
| fee_msat   |	int64	    |The fee paid for this payment in milli-satoshis.|
| creation_time_ns   |	int64	    |The time in UNIX nanoseconds at which the payment was created.|
| htlcs   |		HTLCAttempt[]	    |The HTLCs made in attempt to settle the payment.|
| payment_index   |		uint64	    |The creation index of this payment. Each payment can be uniquely identified by this index, which may not strictly increment by 1 for payments made in older versions of obd.|
| failure_reason   |		PaymentFailureReason	    | |

**PaymentStatus**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| UNKNOWN   |	0	    | |  
| IN_FLIGHT     |	1	    | |
| SUCCEEDED     |	2	    | |
| FAILED     |	3	    | |

**HTLCAttempt**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| attempt_id   |	uint64	    |The unique ID that is used for this attempt.|
| status   |	HTLCStatus	    |The status of the HTLC.|
| route   |	Route	    |The route taken by this HTLC.|
| attempt_time_ns   |	int64	    |The time in UNIX nanoseconds at which this HTLC was sent.|
| resolve_time_ns   |	int64	    |The time in UNIX nanoseconds at which this HTLC was settled or failed. This value will not be set if the HTLC is still IN_FLIGHT.|
| failure   |	Failure	    |Detailed htlc failure info.|
| preimage   |	bytes	    |The preimage that was used to settle the HTLC.|

**PaymentFailureReason**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| FAILURE_REASON_NONE   |	0	    |Payment isn't failed (yet).|
| FAILURE_REASON_TIMEOUT   |	1	    |There are more routes to try, but the payment timeout was exceeded.|
| FAILURE_REASON_NO_ROUTE   |	2	    |All possible routes were tried and failed permanently. Or were no routes to the destination at all.|
| FAILURE_REASON_ERROR   |	3	    |A non-recoverable error has occured.|
| FAILURE_REASON_INCORRECT_PAYMENT_DETAILS   |	4	    |Payment details incorrect (unknown hash, invalid amt or invalid final cltv delta).|
| FAILURE_REASON_INSUFFICIENT_BALANCE   |	5	    |Insufficient local balance.|

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
RouterOuterClass.SendPaymentRequest sendPaymentRequest = RouterOuterClass.SendPaymentRequest.newBuilder()
            .setAssetId((int) 2147485160)
            .setPaymentRequest("obto2147485160:11pjp6a0wpp5jzzalddsjjmfvkgp7l5rhvw2qqgqg8awddypfcafrcksrlz3z48qdq2dpsksctgvycqzpgxqyz5vq3q8zqqqp0gsp5lkvghy5rqguw9zj0djytpwh5wthhgfskas8dpcht4vnkxgn4axss9qyyssqd7c0xwc5q7tswf8x9q7h27jmmghr4hxk56vgu3t8guyktsp7rnksu4g6n8y4c4rvww8z59fa0p7l2m4ypszrwg8us93lwfl2f22hx5gpq0cugx")
            .setFeeLimitMsat(calculateAbsoluteFeeLimit(100000000))
            .setTimeoutSeconds(30)
            .setMaxParts(1)
            .build();
Obdmobile.routerOB_SendPaymentV2(sendPaymentRequest.toByteArray(), new RecvStream() {
    @Override
    public void onError(Exception e) {
        if (e.getMessage().equals("EOF")) {
            return;
        }
        e.printStackTrace();
    }

    @Override
    public void onResponse(byte[] bytes) {
        if (bytes == null) {
            return;
        }
        try {
            LightningOuterClass.Payment resp = LightningOuterClass.Payment.parseFrom(bytes);                                                 
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }
    }
});

public static long calculateAbsoluteFeeLimit(long amountSatToSend) {
    long absFee;
    if (amountSatToSend <= RefConstants.LN_PAYMENT_FEE_THRESHOLD) {
        absFee = (long) (Math.sqrt(amountSatToSend));
    } else {
        absFee = (long) (getRelativeSettingsFeeLimit() * amountSatToSend);
    }
    return Math.max(absFee, 3L);
}

public static float getRelativeSettingsFeeLimit() {
    String lightning_feeLimit = "3%";
    String feePercent = lightning_feeLimit.replace("%", "");
    float feeMultiplier = 1f;
    if (!feePercent.equals("None")) {
        feeMultiplier = Integer.parseInt(feePercent) / 100f;
    }
    return feeMultiplier;
}
```

<!--
The response for the example
-->
response:
```
{
    asset_id: 2147485160
    creation_date: 1679726415
    creation_time_ns: 1679726415995149194
    fee_msat: 0
    htlcs {
      attempt_id: 17002
      attempt_time_ns: 1679726416015893725
      preimage: "\333\240\322\200\341\361\310dP\336[63\217\205\201.\200\240\025\027e<\000t\021\016\255+\353*\362"
      resolve_time_ns: 1679726416217377267
      route {
        asset_id: -2147482136
        hops {
          amt_to_forward: 0
          amt_to_forward_msat: 100000000
          chan_capacity: 0
          chan_id: 2666867652199055361
          expiry: 2425824
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
        total_time_lock: 2425824
      }
      status: SUCCEEDED
      status_value: 1
    }
    payment_hash: "41908fbdeccc2682718cde46bfb627f544bc9a3843457243cf6d4e8194b6928f"
    payment_index: 17003
    payment_preimage: "dba0d280e1f1c86450de5b36338f85812e80a01517653c0074110ead2beb2af2"
    payment_request: "obto2147485160:11pjpayczpp5gxggl00vesngyuvvmertld3874ztex3cgdzhys70d48gr99kj28sdqqcqzpgxqyz5vq3q8zqqqp0gsp5y90up0ht6mn5yfcw8xhexwq44jscngfn0ultjlaa55emrd7x4l7s9qyyssqxlde4r7fnj7wttc8f2470pn0cmnev6dnjww44pm3709m97nh0tl354ry0ef3kw232ml6q7w06n4sqn8t0m0dudt5c08kg8uj858edzqpvmvrrg"
    status: SUCCEEDED
    status_value: 2
    value_msat: 100000000
}
```