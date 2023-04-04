## OB_AddInvoice

OB_AddInvoice attempts to add a new invoice to the invoice database. Any duplicated invoices are rejected, therefore all invoices must have a unique payment preimage.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |
| asset_id   |	uint64	    |The ID of an asset.|
| memo	     |	string		  |An optional memo to attach along with the invoice. Used for record keeping purposes for the invoice's creator, and will also be set in the description field of the encoded payment request if the description_hash field is not being used.|  
| r_preimage     |	bytes  |The hex-encoded preimage (32 byte) which will allow settling an incoming HTLC payable to this preimage. When using REST, this field must be encoded as base64.| 
| r_hash |	bytes       |The hash of the preimage. When using REST, this field must be encoded as base64. Note: Output only, don't specify for creating an invoice.| 
| value |	int64       |The value of this invoice in satoshis The fields value and value_msat are mutually exclusive.| 
| value_msat |	int64       |The value of this invoice in millisatoshis The fields value and value_msat are mutually exclusive.| 
| amount   |	int64	    |The value of this invoice in millisatoshis The fields value and amount are mutually exclusive.|
| settled   |	bool	    |Whether this invoice has been fulfilled. The field is deprecated. Use the state field instead (compare to SETTLED).|
| creation_date   |	int64	    |When this invoice was created. Measured in seconds since the unix epoch. Note: Output only, don't specify for creating an invoice.|
| settle_date   |	int64	    |When this invoice was settled. Measured in seconds since the unix epoch. Note: Output only, don't specify for creating an invoice.|
| payment_request   |	string	    |A bare-bones invoice for a payment within the Lightning Network. With the details of the invoice, the sender has all the data necessary to send a payment to the recipient. Note: Output only, don't specify for creating an invoice.|
| description_hash   |	bytes	    |Hash (SHA-256) of a description of the payment. Used if the description of payment (memo) is too long to naturally fit within the description field of an encoded payment request. When using REST, this field must be encoded as base64.|
| expiry   |	int64	    |Payment request expiry time in seconds. Default is 86400 (24 hours).|
| fallback_addr   |	string	    |Fallback on-chain address.|
| cltv_expiry   |	uint64	    |Delta to use for the time-lock of the CLTV extended to the final hop.|
| route_hints   |	RouteHint[]	    |Route hints that can each be individually used to assist in reaching the invoice's destination.|
| private   |	bool	    |Whether this invoice should include routing hints for private channels. Note: When enabled, if value and value_msat are zero, a large number of hints with these channels can be included, which might not be desirable.|
| add_index   |	uint64	    |The "add" index of this invoice. Each newly created invoice will increment this index making it monotonically increasing. Callers to the SubscribeInvoices call can use this to instantly get notified of all added invoices with an add_index greater than this one. Note: Output only, don't specify for creating an invoice.|
| settle_index   |	uint64	    |The "settle" index of this invoice. Each newly settled invoice will increment this index making it monotonically increasing. Callers to the SubscribeInvoices call can use this to instantly get notified of all settled invoices with an settle_index greater than this one. Note: Output only, don't specify for creating an invoice.|
| amt_paid   |	int64	    |Deprecated, use amt_paid_sat or amt_paid_msat.|
| amt_paid_sat   |	int64	    |The amount that was accepted for this invoice, in satoshis. This will ONLY be set if this invoice has been settled. We provide this field as if the invoice was created with a zero value, then we need to record what amount was ultimately accepted. Additionally, it's possible that the sender paid MORE that was specified in the original invoice. So we'll record that here as well. Note: Output only, don't specify for creating an invoice.|
| amt_paid_msat   |	int64	    |The amount that was accepted for this invoice, in millisatoshis. This will ONLY be set if this invoice has been settled. We provide this field as if the invoice was created with a zero value, then we need to record what amount was ultimately accepted. Additionally, it's possible that the sender paid MORE that was specified in the original invoice. So we'll record that here as well. Note: Output only, don't specify for creating an invoice.|
| state   |	InvoiceState	    |The state the invoice is in. Note: Output only, don't specify for creating an invoice.|
| htlcs   |	InvoiceHTLC[]	    |List of HTLCs paying to this invoice [EXPERIMENTAL]. Note: Output only, don't specify for creating an invoice.|
| features   |	FeaturesEntry[]	    |List of features advertised on the invoice. Note: Output only, don't specify for creating an invoice.|
| is_keysend   |	bool	    |Indicates if this invoice was a spontaneous payment that arrived via keysend [EXPERIMENTAL]. Note: Output only, don't specify for creating an invoice.|
| payment_addr   |	bytes	    |The payment address of this invoice. This value will be used in MPP payments, and also for newer invoices that always require the MPP payload for added end-to-end security. Note: Output only, don't specify for creating an invoice.|
| is_amp   |	bool		    |Signals whether or not this is an AMP invoice.|
| amp_invoice_state   |	AmpInvoiceStateEntry[]			    |[EXPERIMENTAL]: Maps a 32-byte hex-encoded set ID to the sub-invoice AMP state for the given set ID. This field is always populated for AMP invoices, and can be used along side LookupInvoice to obtain the HTLC information related to a given sub-invoice. Note: Output only, don't specify for creating an invoice.|

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

## Response:
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| r_hash	         |	bytes		    | |  
| payment_request  |	string		  |A bare-bones invoice for a payment within the Lightning Network. With the details of the invoice, the sender has all the data necessary to send a payment to the recipient.|  
| add_index        |	uint64      |The "add" index of this invoice. Each newly created invoice will increment this index making it monotonically increasing. Callers to the SubscribeInvoices call can use this to instantly get notified of all added invoices with an add_index greater than this one.| 
| payment_addr     |	bytes       |The payment address of the generated invoice. This value should be used in all payments for this invoice as we require it for end to end security.| 

## Example:

<!--
java code example
-->

```java
LightningOuterClass.Invoice asyncInvoiceRequest;
if (mAssetId == 0) {
    asyncInvoiceRequest = LightningOuterClass.Invoice.newBuilder()
            .setAssetId((int) 0)
            .setValueMsat((long) 100000000000)
            .setMemo("invoice")
            .setExpiry(Long.parseLong("86400")) // in seconds
            .setPrivate(false)
            .build();
} else {
    asyncInvoiceRequest = LightningOuterClass.Invoice.newBuilder()
            .setAssetId((int) 2147485160)
            .setAmount((long) 100000000)
            .setMemo("invoice")
            .setExpiry(Long.parseLong("86400")) // in seconds
            .setPrivate(false)
            .build();
}
Obdmobile.oB_AddInvoice(asyncInvoiceRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.AddInvoiceResponse resp = LightningOuterClass.AddInvoiceResponse.parseFrom(bytes);
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
    add_index: 19
    payment_addr: "\375\230\213\222\203\0028\342\212Ol\210\260\272\364r\357t&\026\354\016\320\342\353\253\'c\"u\351\241"
    payment_request: "obto2147485160:11pjp6a0wpp5jzzalddsjjmfvkgp7l5rhvw2qqgqg8awddypfcafrcksrlz3z48qdq2dpsksctgvycqzpgxqyz5vq3q8zqqqp0gsp5lkvghy5rqguw9zj0djytpwh5wthhgfskas8dpcht4vnkxgn4axss9qyyssqd7c0xwc5q7tswf8x9q7h27jmmghr4hxk56vgu3t8guyktsp7rnksu4g6n8y4c4rvww8z59fa0p7l2m4ypszrwg8us93lwfl2f22hx5gpq0cugx"
    r_hash: "\220\205\337\265\260\224\266\226Y\001\367\350;\261\312\000\020\004\037\256kH\024\343\251\036-\001\374Q\025N"
}
```


