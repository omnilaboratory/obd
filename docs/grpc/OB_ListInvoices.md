## OB_ListInvoices

OB_ListInvoices returns a list of all the invoices currently stored within the database. Any active debug invoices are ignored. It has full support for paginated responses, allowing users to query for specific invoices through their add_index. This can be done by using either the first_index_offset or last_index_offset fields included in the response as the index_offset of the next request. By default, the first 100 invoices created will be returned. Backwards pagination is also supported through the Reversed flag.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| asset_id   |	uint64	    |The ID of an asset.|
| pending_only   |	bool	    |If set, only invoices that are not settled and not canceled will be returned in the response.|
| index_offset   |	uint64	    |The index of an invoice that will be used as either the start or end of a query to determine which invoices should be returned in the response.|
| num_max_invoices   |	uint64	    |The max number of invoices to return in the response to this query.|
| reversed   |	bool	    |If set, the invoices returned will result from seeking backwards from the specified index offset. This can be used to paginate backwards.|
| creation_date_start   |	uint64	    |If set, returns all invoices with a creation date greater than or equal to it. Measured in seconds since the unix epoch.|
| creation_date_end   |	uint64	    |If set, returns all invoices with a creation date less than or equal to it. Measured in seconds since the unix epoch.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| invoices     |	Invoice[]	    |A list of invoices from the time slice of the time series specified in the request.|
| last_index_offset     |	uint64	    |The index of the last item in the set of returned invoices. This can be used to seek further, pagination style.|
| first_index_offset     |	uint64	    |The index of the last item in the set of returned invoices. This can be used to seek backwards, pagination style.|

**Invoice**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| memo   |	string	    |An optional memo to attach along with the invoice. Used for record keeping purposes for the invoice's creator, and will also be set in the description field of the encoded payment request if the description_hash field is not being used.|
| r_preimage   |	bytes	    |The hex-encoded preimage (32 byte) which will allow settling an incoming HTLC payable to this preimage. When using REST, this field must be encoded as base64.|
| r_hash   |	bytes	    |The hash of the preimage. When using REST, this field must be encoded as base64. Note: Output only, don't specify for creating an invoice.|
| value   |	int64	    |The value of this invoice in satoshis The fields value and value_msat are mutually exclusive.|
| value_msat   |	int64	    |The value of this invoice in millisatoshis The fields value and value_msat are mutually exclusive.|
| settled   |	bool		    |Whether this invoice has been fulfilled. The field is deprecated. Use the state field instead (compare to SETTLED).|
| creation_date   |	int64		    |When this invoice was created. Measured in seconds since the unix epoch. Note: Output only, don't specify for creating an invoice.|
| settle_date   |	int64		    |When this invoice was settled. Measured in seconds since the unix epoch. Note: Output only, don't specify for creating an invoice.|
| payment_request   |	string		    |A bare-bones invoice for a payment within the Lightning Network. With the details of the invoice, the sender has all the data necessary to send a payment to the recipient. Note: Output only, don't specify for creating an invoice.|
| description_hash   |	bytes		    |Hash (SHA-256) of a description of the payment. Used if the description of payment (memo) is too long to naturally fit within the description field of an encoded payment request. When using REST, this field must be encoded as base64.|
| expiry   |	int64		    |Payment request expiry time in seconds. Default is 86400 (24 hours).|
| fallback_addr   |	string		    |Fallback on-chain address.|
| cltv_expiry   |	uint64		    |Delta to use for the time-lock of the CLTV extended to the final hop.|
| route_hints   |	RouteHint[]		    |Route hints that can each be individually used to assist in reaching the invoice's destination.|
| private   |	bool		    |Whether this invoice should include routing hints for private channels. Note: When enabled, if value and value_msat are zero, a large number of hints with these channels can be included, which might not be desirable.|
| add_index   |	uint64		    |The "add" index of this invoice. Each newly created invoice will increment this index making it monotonically increasing. Callers to the SubscribeInvoices call can use this to instantly get notified of all added invoices with an add_index greater than this one. Note: Output only, don't specify for creating an invoice.|
| settle_index   |	uint64		    |The "settle" index of this invoice. Each newly settled invoice will increment this index making it monotonically increasing. Callers to the SubscribeInvoices call can use this to instantly get notified of all settled invoices with an settle_index greater than this one. Note: Output only, don't specify for creating an invoice.|
| amt_paid   |	int64		    |Deprecated, use amt_paid_sat or amt_paid_msat.|
| amt_paid_sat   |	int64		    |The amount that was accepted for this invoice, in satoshis. This will ONLY be set if this invoice has been settled. We provide this field as if the invoice was created with a zero value, then we need to record what amount was ultimately accepted. Additionally, it's possible that the sender paid MORE that was specified in the original invoice. So we'll record that here as well. Note: Output only, don't specify for creating an invoice.|
| amt_paid_msat   |	int64		    |The amount that was accepted for this invoice, in millisatoshis. This will ONLY be set if this invoice has been settled. We provide this field as if the invoice was created with a zero value, then we need to record what amount was ultimately accepted. Additionally, it's possible that the sender paid MORE that was specified in the original invoice. So we'll record that here as well. Note: Output only, don't specify for creating an invoice.|
| state   |	InvoiceState		    |The state the invoice is in. Note: Output only, don't specify for creating an invoice.|
| htlcs   |	InvoiceHTLC[]		    |List of HTLCs paying to this invoice [EXPERIMENTAL]. Note: Output only, don't specify for creating an invoice.|
| features   |	FeaturesEntry[]		    |List of features advertised on the invoice. Note: Output only, don't specify for creating an invoice.|
| is_keysend   |	bool		    |Indicates if this invoice was a spontaneous payment that arrived via keysend [EXPERIMENTAL]. Note: Output only, don't specify for creating an invoice.|
| payment_addr   |	bytes		    |The payment address of this invoice. This value will be used in MPP payments, and also for newer invoices that always require the MPP payload for added end-to-end security. Note: Output only, don't specify for creating an invoice.|
| is_amp   |	bool		    |Signals whether or not this is an AMP invoice.|
| amp_invoice_state   |	AmpInvoiceStateEntry[]		    |[EXPERIMENTAL]: Maps a 32-byte hex-encoded set ID to the sub-invoice AMP state for the given set ID. This field is always populated for AMP invoices, and can be used along side LookupInvoice to obtain the HTLC information related to a given sub-invoice. Note: Output only, don't specify for creating an invoice.|

**RouteHint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| hop_hints   |	HopHint[]	    |A list of hop hints that when chained together can assist in reaching a specific destination.|

**InvoiceState**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| OPEN   |	0	    | |  
| SETTLED     |	1	    | |
| CANCELED     |	2	    | |
| ACCEPTED     |	3	    | |

**InvoiceHTLC**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| chan_id   |	uint64	    |Short channel id over which the htlc was received.|
| htlc_index   |	uint64	    |Index identifying the htlc on the channel.|
| amt_msat   |	uint64	    |The amount of the htlc in msat.|
| accept_height   |	int32	    |Block height at which this htlc was accepted.|
| accept_time   |	int64	    |Time at which this htlc was accepted.|
| resolve_time   |	int64	    |Time at which this htlc was settled or canceled.|
| expiry_height   |	int32	    |Block height at which this htlc expires.|
| state   |	InvoiceHTLCState	    |Current state the htlc is in.|
| custom_records   |	CustomRecordsEntry[]		    |The total amount of the mpp payment in msat.|
| mpp_total_amt_msat   |	uint64		    |The total amount of the mpp payment in msat.|
| amp   |	AMP		    |Details relevant to AMP HTLCs, only populated if this is an AMP HTLC.|

**FeaturesEntry**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	uint32		    | |
| value   |	Feature	    | |

**AmpInvoiceStateEntry**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	string		    | |
| value   |	AMPInvoiceState	    | |

**HopHint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| node_id   |	string	    |The public key of the node at the start of the channel.|
| chan_id   |	uint64	    |The unique identifier of the channel.|
| fee_base_msat   |	uint32	    |The base fee of the channel denominated in millisatoshis.|
| fee_proportional_millionths   |	uint32	    |The fee rate of the channel for sending one satoshi across it denominated in millionths of a satoshi.|
| cltv_expiry_delta   |	uint32	    |The time-lock delta of the channel.|

**InvoiceHTLCState**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| ACCEPTED     |	0	    | |
| SETTLED     |	1	    | |
| CANCELED     |	2	    | |

**CustomRecordsEntry**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	uint64		    | |
| value   |	bytes	    | |

**AMP**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| root_share   |	bytes		    |An n-of-n secret share of the root seed from which child payment hashes and preimages are derived.|
| set_id   |	bytes	    |An identifier for the HTLC set that this HTLC belongs to.|
| child_index   |	uint32	    |A nonce used to randomize the child preimage and child hash from a given root_share.|
| hash   |	bytes	    |The payment hash of the AMP HTLC.|
| preimage   |	bytes	    |The preimage used to settle this AMP htlc. This field will only be populated if the invoice is in InvoiceState_ACCEPTED or InvoiceState_SETTLED.|

**Feature**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| name   |	string		    | |
| is_required   |	bool	    | |
| is_known   |	bool	    | |

**AMPInvoiceState**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| state   |	InvoiceHTLCState		    |The state the HTLCs associated with this setID are in.|
| settle_index   |	uint64	    |The settle index of this HTLC set, if the invoice state is settled.|
| settle_time   |	int64	    |The time this HTLC set was settled expressed in unix epoch.|
| amt_paid_msat   |	int64	    |The total amount paid for the sub-invoice expressed in milli satoshis.|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.ListInvoiceRequest invoiceRequest = LightningOuterClass.ListInvoiceRequest.newBuilder()
                    .setAssetId((int) 2147485160)
                    .setIsQueryAsset(true)
                    .setNumMaxInvoices(100)
                    .setStartTime(Long.parseLong(1677600000)
                    .build();
Obdmobile.oB_ListInvoices(invoiceRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.ListInvoiceResponse resp = LightningOuterClass.ListInvoiceResponse.parseFrom(bytes);
            List invoicesList =  resp.getInvoicesList();
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
    first_index_offset: 1
    invoices {
      add_index: 1
      amount: 0
      amt_paid_msat: 159101031
      asset_id: -2147482136
      cltv_expiry: 40
      creation_date: 1678789110
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
      htlcs {
        accept_height: 2424351
        accept_time: 1678789111
        amt_msat: 159101031
        chan_id: 2665597716268253185
        expiry_height: 2424394
        htlc_index: 0
        mpp_total_amt_msat: 159101031
        resolve_time: 1678789111
        state: SETTLED
        state_value: 1
      }
      memo: "lucky"
      payment_addr: "\301=`\266p\023\034al\206\232\030\344\222\360\357\035\230\250\272\371\352zl\335\356\355D\310\232\005\221"
      payment_request: "obto2147485160:1591010310n1pjpqj0kpp5vtqs4c0lhrtcwxpmmghzs67665eanzmnahzay0fphe3cpzkjd6xsdqgd36kx6mecqzpgxqyz5vq3q8zqqqp0gsp5cy7kpdnszvwxzmyxngvwfyhsauwe3296l8485mxaamk5fjy6qkgs9qyyssqf7a4dtzlpxwkelh67qh4kacc85geq7gw5ha4lt28rwer8vlfur4sfj0lgthdf4ygwwshxkzc893q38d56c94ed0cemwcj5s50u980qgqjfxerx"
      r_hash: "b\301\n\341\377\270\327\207\030;\332.(k\332\3253\331\213s\355\305\322=!\276c\200\212\322n\215"
      r_preimage: "\324H>e~I\344\320\343\023\341\224\027f\316\207\277<\303\317\233mT\216^\330\310\324\024L\215\235"
      settle_date: 1678789111
      settle_index: 1
      settled: true
      state: SETTLED
      state_value: 1
      value_msat: 159101031
    }
}
```
