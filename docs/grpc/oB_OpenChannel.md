## OB_OpenChannel

OB_OpenChannel attempts to open a singly funded channel specified in the request to a remote peer. Users are able to specify a target number of blocks that the funding transaction should be confirmed in, or a manual fee rate to us for the funding transaction. If neither are specified, then a lax block confirmation target is used. Each OpenStatusUpdate will return the pending channel ID of the in-progress channel. Depending on the arguments specified in the OpenChannelRequest, this pending channel ID can then be used to manually progress the channel funding flow.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| sat_per_vbyte   |	uint64	    |A manual fee rate set in sat/vbyte that should be used when crafting the funding transaction.|
| node_pubkey   |	bytes	    |The pubkey of the node to open a channel with. When using REST, this field must be encoded as base64.|
| node_pubkey_string   |	string	    |The hex encoded pubkey of the node to open a channel with. Deprecated now that the REST gateway supports base64 encoding of bytes fields.|
| local_funding_amount   |	int64	    |The number of satoshis the wallet should commit to the channel.|
| push_sat   |	int64	    |The number of satoshis to push to the remote side as part of the initial commitment state.|
| target_conf   |	int32	    |The target number of blocks that the funding transaction should be confirmed by.|
| sat_per_byte   |	int64	    |Deprecated, use sat_per_vbyte. A manual fee rate set in sat/vbyte that should be used when crafting the funding transaction.|
| private   |	bool	    |Whether this channel should be private, not announced to the greater network.|
| min_htlc_msat   |	int64	    |The minimum value in millisatoshi we will require for incoming HTLCs on the channel.|
| remote_csv_delay   |	uint32	    |The delay we require on the remote's commitment transaction. If this is not set, it will be scaled automatically with the channel size.|
| min_confs   |	int32	    |The minimum number of confirmations each one of your outputs used for the funding transaction must satisfy.|
| spend_unconfirmed   |	bool	    |Whether unconfirmed outputs should be used as inputs for the funding transaction.|
| close_address   |	string	    |Close address is an optional address which specifies the address to which funds should be paid out to upon cooperative close. This field may only be set if the peer supports the option upfront feature bit (call listpeers to check). The remote peer will only accept cooperative closes to this address if it is set. Note: If this value is set on channel creation, you will not be able to cooperatively close out to a different address.|
| funding_shim   |	FundingShim	    |Funding shims are an optional argument that allow the caller to intercept certain funding functionality. For example, a shim can be provided to use a particular key for the commitment key (ideally cold) rather than use one that is generated by the wallet as normal, or signal that signing will be carried out in an interactive manner (PSBT based).|
| remote_max_value_in_flight_msat   |	uint64	    |The maximum amount of coins in millisatoshi that can be pending within the channel. It only applies to the remote party.|
| remote_max_htlcs   |	uint32	    |The maximum number of concurrent HTLCs we will allow the remote party to add to the commitment transaction.|
| max_local_csv   |	uint32	    |Max local csv is the maximum csv delay we will allow for our own commitment transaction.|
| commitment_type   |	CommitmentType	    |The explicit commitment type to use. Note this field will only be used if the remote peer supports explicit channel negotiation.|
| zero_conf   |	bool	    |If this is true, then a zero-conf channel open will be attempted.|
| scid_alias   |	bool	    |If this is true, then an option-scid-alias channel-type open will be attempted.|
| base_fee   |	uint64	    |The base fee charged regardless of the number of milli-satoshis sent.|
| fee_rate   |	uint64	    |The fee rate in ppm (parts per million) that will be charged in proportion of the value of each forwarded HTLC.|
| use_base_fee   |	bool	    |If use_base_fee is true the open channel announcement will update the channel base fee with the value specified in base_fee. In the case of a base_fee of 0 use_base_fee is needed downstream to distinguish whether to use the default base fee value specified in the config or 0.|
| use_fee_rate   |	bool	    |If use_fee_rate is true the open channel announcement will update the channel fee rate with the value specified in fee_rate. In the case of a fee_rate of 0 use_fee_rate is needed downstream to distinguish whether to use the default fee rate value specified in the config or 0.|
| remote_chan_reserve_sat   |	uint64	    |The number of satoshis we require the remote peer to reserve. This value, if specified, must be above the dust limit and below 20% of the channel capacity.|

**FundingShim**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| chan_point_shim   |	ChanPointShim	    |A channel shim where the channel point was fully constructed outside of lnd's wallet and the transaction might already be published.|
| psbt_shim   |	PsbtShim	    |A channel shim that uses a PSBT to fund and sign the channel funding transaction.|

**CommitmentType**

| Name		            |	Number		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| UNKNOWN_COMMITMENT_TYPE     |	0	    |Returned when the commitment type isn't known or unavailable.|
| LEGACY     |	1	    |A channel using the legacy commitment format having tweaked to_remote keys.|
| STATIC_REMOTE_KEY     |	2	    |A channel that uses the modern commitment format where the key in the output of the remote party does not change each state. This makes back up and recovery easier as when the channel is closed, the funds go directly to that key.|
| ANCHORS     |	3	    |A channel that uses a commitment format that has anchor outputs on the commitments, allowing fee bumping after a force close transaction has been broadcast.|
| SCRIPT_ENFORCED_LEASE     |	4	    |A channel that uses a commitment type that builds upon the anchors commitment format, but in addition requires a CLTV clause to spend outputs paying to the channel initiator. This is intended for use on leased channels to guarantee that the channel initiator has no incentives to close a leased channel before its maturity date.|

**ChanPointShim**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| amt   |	int64	    |The size of the pre-crafted output to be used as the channel point for this channel funding.|
| chan_point   |	ChannelPoint		    |The target channel point to refrence in created commitment transactions.|
| local_key   |	KeyDescriptor		    |Our local key to use when creating the multi-sig output.|
| remote_key   |	bytes		    |The key of the remote party to use when creating the multi-sig output.|
| pending_chan_id   |	bytes		    |If non-zero, then this will be used as the pending channel ID on the wire protocol to initate the funding request. This is an optional field, and should only be set if the responder is already expecting a specific pending channel ID.|
| thaw_height   |	uint32		    |This uint32 indicates if this channel is to be considered 'frozen'. A frozen channel does not allow a cooperative channel close by the initiator. The thaw_height is the height that this restriction stops applying to the channel. The height can be interpreted in two ways: as a relative height if the value is less than 500,000, or as an absolute height otherwise.|

**PsbtShim**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| pending_chan_id   |	bytes	    |A unique identifier of 32 random bytes that will be used as the pending channel ID to identify the PSBT state machine when interacting with it and on the wire protocol to initiate the funding request.|
| base_psbt   |	bytes	    |An optional base PSBT the new channel output will be added to. If this is non-empty, it must be a binary serialized PSBT.|
| no_publish   |	bool	    |If a channel should be part of a batch (multiple channel openings in one transaction), it can be dangerous if the whole batch transaction is published too early before all channel opening negotiations are completed. This flag prevents this particular channel from broadcasting the transaction after the negotiation with the remote peer. In a batch of channel openings this flag should be set to true for every channel but the very last.|

**ChannelPoint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| funding_txid_bytes   |	bytes	    |Txid of the funding transaction. When using REST, this field must be encoded as base64.|
| funding_txid_str   |	string	    |Hex-encoded string representing the byte-reversed hash of the funding transaction.|
| output_index   |	uint32	    |The index of the output of the funding transaction.|

**KeyDescriptor**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| raw_key_bytes   |	bytes	    |The raw bytes of the key being identified.|
| key_loc   |	KeyLocator	    |The key locator that identifies which key to use for signing.|

**KeyLocator**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key_family   |	int32	    |The family of key being identified.|
| key_index   |	int32	    |The precise index of the key being identified.|

## Response:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| chan_pending     |	PendingUpdate	    |Signals that the channel is now fully negotiated and the funding transaction published.|
| chan_open     |	ChannelOpenUpdate	    |Signals that the channel's funding transaction has now reached the required number of confirmations on chain and can be used.|
| psbt_fund     |	ReadyForPsbtFunding	    |Signals that the funding process has been suspended and the construction of a PSBT that funds the channel PK script is now required.|
| pending_chan_id     |	bytes		    |The pending channel ID of the created channel. This value may be used to further the funding flow manually via the FundingStateStep method.|

**PendingUpdate**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| txid   |	bytes	    | |
| output_index   |	uint32	    | |

**ChannelOpenUpdate**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| channel_point   |	ChannelPoint	    | |

**ReadyForPsbtFunding**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| funding_address   |	string	    |The P2WSH address of the channel funding multisig address that the below specified amount in satoshis needs to be sent to.|
| funding_amount   |	int64	    |The exact amount in satoshis that needs to be sent to the above address to fund the pending channel.|
| psbt   |	bytes	    |A raw PSBT that contains the pending channel output. If a base PSBT was provided in the PsbtShim, this is the base PSBT with one additional output. If no base PSBT was specified, this is an otherwise empty PSBT with exactly one output.|

**ChannelPoint**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| funding_txid_bytes   |	bytes	    |Txid of the funding transaction. When using REST, this field must be encoded as base64.|
| funding_txid_str   |	string	    |Hex-encoded string representing the byte-reversed hash of the funding transaction.|
| output_index   |	uint32	    |The index of the output of the funding transaction.|

## Example:

<!--
java code example
-->

```java
String nodeUri = "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61"
byte[] nodeKeyBytes = hexStringToByteArray(pubkey);
LightningOuterClass.OpenChannelRequest openChannelRequest;
if (assetId == 0) {
    openChannelRequest = LightningOuterClass.OpenChannelRequest.newBuilder()
            .setNodePubkey(ByteString.copyFrom(nodeKeyBytes))
            .setTargetConf(448)
            .setPrivate(false)
            .setLocalFundingBtcAmount((long) (100000000))
            .setPushBtcSat((long) (100000000 / 2))
            .setAssetId((int) 0)
            .build();
} else {
    openChannelRequest = LightningOuterClass.OpenChannelRequest.newBuilder()
            .setNodePubkey(ByteString.copyFrom(nodeKeyBytes))
            .setTargetConf(448)
            .setPrivate(false)
            .setLocalFundingBtcAmount(20000)
            .setLocalFundingAssetAmount((long) (100000000))
            .setPushAssetSat((long) (100000000 / 2))
            .setAssetId((int) 2147485160)
            .build();
        }
Obdmobile.oB_OpenChannel(openChannelRequest.toByteArray(), new RecvStream() {
    @Override
    public void onError(Exception e) {
        if (e.getMessage().equals("EOF")) {
            return;
        }
        e.printStackTrace();
    }

    @Override
    public void onResponse(byte[] bytes) {
        try {
            LightningOuterClass.OpenStatusUpdate resp = LightningOuterClass.OpenStatusUpdate.parseFrom(bytes);
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }
    }
});
public static byte[] hexStringToByteArray(String hex) {
    int l = hex.length();
    byte[] data = new byte[l / 2];
    for (int i = 0; i < l; i += 2) {
        data[i / 2] = (byte) ((Character.digit(hex.charAt(i), 16) << 4)
                + Character.digit(hex.charAt(i + 1), 16));
    }
    return data;
}
```

<!--
The response for the example
-->
response:
```
{
    chan_pending {
      output_index: 1
      txid: "\3326\236M\312\017j\377\236\356\224\v\034\244\321\027\312\226\346\320,t\266Jp\025\333a\311P\272^"
      txid_str: "5eba50c961db15704ab6742cd0e696ca17d1a41c0b94ee9eff6a0fca4d9e36da"
    }
    pending_chan_id: "e\021\223\000\214\241G\237\205\031\370*\000\343\0054H\252\261b\022\215\375\256B\234\033\0358\266\243d"
}
```