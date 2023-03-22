## ListPayments

ListPayments returns a list of all outgoing payments.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| include_incomplete   |	bool	    |If true, then return payments that have not yet fully completed. This means that pending payments, as well as failed payments will show up if this field is set to true. This flag doesn't change the meaning of the indices, which are tied to individual payments.|
| index_offset   |	uint64	    |The index of a payment that will be used as either the start or end of a query to determine which payments should be returned in the response. The index_offset is exclusive. In the case of a zero index_offset, the query will start with the oldest payment when paginating forwards, or will end with the most recent payment when paginating backwards.|
| max_payments   |	uint64	    |The maximal number of payments returned in the response to this query.|
| reversed   |	bool	    |If set, the payments returned will result from seeking backwards from the specified index offset. This can be used to paginate backwards. The order of the returned payments is always oldest first (ascending index order).|
| count_total_payments   |	bool	    |If set, all payments (complete and incomplete, independent of the max_payments parameter) will be counted. Note that setting this to true will increase the run time of the call significantly on systems that have a lot of payments, as all of them have to be iterated through to be counted.|
| creation_date_start   |	uint64	    |If set, returns all invoices with a creation date greater than or equal to it. Measured in seconds since the unix epoch.|
| creation_date_end   |	uint64	    |If set, returns all invoices with a creation date less than or equal to it. Measured in seconds since the unix epoch.|
| asset_id   |	uint64	    |The ID of an asset.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| payments     |	Payment[]	    |The list of payments.|
| first_index_offset     |	uint64	    |The index of the first item in the set of returned payments. This can be used as the index_offset to continue seeking backwards in the next request.|
| last_index_offset     |	uint64	    |The index of the last item in the set of returned payments. This can be used as the index_offset to continue seeking forwards in the next request.|
| total_num_payments     |	uint64	    |Will only be set if count_total_payments in the request was set. Represents the total number of payments (complete and incomplete, independent of the number of payments requested in the query) currently present in the payments database.|

**Payment**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| payment_hash   |	string	    |The payment hash.|  
| value     |	int64	    |Deprecated, use value_sat or value_msat.|
| creation_date     |	int64	    |Deprecated, use creation_time_ns.|
| fee     |	int64	    |Deprecated, use fee_sat or fee_msat.|
| payment_preimage     |	string	    |The payment preimage.|
| value_sat     |	int64	    |The value of the payment in satoshis.|
| value_msat     |	int64	    |The value of the payment in milli-satoshis.|
| payment_request     |	string	    |The optional payment request being fulfilled.|
| status     |	PaymentStatus	    |The status of the payment.|
| fee_sat     |	int64	    |The fee paid for this payment in satoshis.|
| fee_msat     |	int64	    |The fee paid for this payment in milli-satoshis.|
| creation_time_ns     |	int64	    |The time in UNIX nanoseconds at which the payment was created.|
| htlcs     |	HTLCAttempt[]	    |The HTLCs made in attempt to settle the payment.|
| payment_index     |	uint64	    |The creation index of this payment. Each payment can be uniquely identified by this index, which may not strictly increment by 1 for payments made in older versions of lnd.|
| failure_reason     |	PaymentFailureReason	    | |

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
LightningOuterClass.ListPaymentsRequest paymentsRequest = LightningOuterClass.ListPaymentsRequest.newBuilder()
                    .setAssetId((int) 2147485160)
                    .setIsQueryAsset(false)
                    .setIncludeIncomplete(false)
                    .setStartTime(Long.parseLong(1677600000))
                    .build();
Obdmobile.oB_ListPayments(paymentsRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.ListPaymentsResponse resp = LightningOuterClass.ListPaymentsResponse.parseFrom(bytes);
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
    total_num_payments: 5
    first_index_offset: 2003
    last_index_offset: 7003
    payments {
      asset_id: 0
      creation_date: 1678789720
      creation_time_ns: 1678789720353394336
      fee_msat: 0
      htlcs {
        attempt_id: 2005
        attempt_time_ns: 1678789720374191680
        preimage: "\215\1771\260Mnu\t\273u\314\v\367\005\322`\364e\273i\302\031\310\2351\300A\201-\031?\301"
        resolve_time_ns: 1678789720559565638
        route {
          hops {
            amt_to_forward: 0
            amt_to_forward_msat: 2000000
            chan_capacity: 0
            chan_id: 2665559233359052800
            expiry: 2424396
            fee: 0
            fee_msat: 0
            mpp_record {
              payment_addr: "<1\211o]C\314\353(\245ml\346\266\244L\353\266\2266=j\205D\355\323Z\316\022\252BN"
              total_amt_msat: 2000000
            }
            pub_key: "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
            tlv_payload: true
          }
          total_amt_msat: 2000000
          total_fees_msat: 0
          total_time_lock: 2424396
        }
        status: SUCCEEDED
        status_value: 1
      }
      payment_hash: "3c9bf1c44929557d4d931d18a64a2aaaff4253a867c0632dd0e711385ed9e0ee"
      payment_index: 2006
      payment_preimage: "8d7f31b04d6e7509bb75cc0bf705d260f465bb69c219c89d31c041812d193fc1"
      payment_request: "obto0:20u1pjpqnrlpp58jdlr3zf992h6nvnr5v2vj324tl5y5agvlqxxtwsuugnshkeurhqdqqcqzpgxqyz5vq3qpqsp58sccjm6ag0xwk299d4kwdd4yfn4md93k844g238d6ddvuy42gf8q9qyyssq350xgr0qj4etdhx4rj22hdx0vpwjq4gczu8c29z3xpvxjps97svhcpmd3tpzty8nrd662270a92p0qqrmtzq38spx9l4rgd70tm2vkgq9jtz37"
      status: SUCCEEDED
      status_value: 2
      value_msat: 2000000
    }
}
```