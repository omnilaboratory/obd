## verifyChanBackup

VerifyChanBackup allows a caller to verify the integrity of a channel backup snapshot. This method will accept either a packed Single or a packed Multi. Specifying both will result in an error.

## Arguments:
This request has no parameters.


## Response:
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| chan_backups|	ChannelBackups	    |The set of new channels that have been added since the last channel backup snapshot was requested.|  
| multi_chan_backup  |	bytes	|A multi-channel backup that covers all open channels currently known to obd(lnd,oblnd).|
**ChannelBackup**
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| chan_point|	ChannelPoint	    |Identifies the channel that this backup belongs to.|  
| chan_backup  |	bytes	|Is an encrypted single-chan backup. this can be passed to RestoreChannelBackups, or the WalletUnlocker Init and Unlock methods in order to trigger the recovery protocol. When using REST, this field must be encoded as base64.|
**ChannelBackups**
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| chan_backups|	ChannelBackup[]	    |A set of single-chan static channel backups.|
**ChannelPoint**
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| funding_txid_bytes|	string    |Txid of the funding transaction. When using REST, this field must be encoded as base64.|
| funding_txid_str|	bytes    |Hex-encoded string representing the byte-reversed hash of the funding transaction.|
| output_index|	int    |The index of the output of the funding transaction|
**MultiChanBackup**
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| chan_point|	ChannelPoint    |Identifies the channel that this backup belongs to.|
| multi_chan_backup|	bytes    |A single encrypted blob containing all the static channel backups of the channel listed above. This can be stored as a single file or blob, and safely be replaced with any prior/future versions. When using REST, this field must be encoded as base64.|

## Example:

<!--
java code example
-->

```java
Obdmobile.verifyChanBackup(chanBackupSnapshot.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {
        if(e.getMessage().equals("rpc error: code = Unknown desc = invalid single channel backup: chacha20poly1305: message authentication failed")){

        }else if(e.getMessage().trim().equals("rpc error: code = Unknown desc = only one Single is accepted at a time")){
        }else if(e.getMessage().equals("rpc error: code = Unknown desc = invalid multi channel backup: chacha20poly1305: message authentication failed")){}else{
        }
        e.printStackTrace();
    }
    @Override
    public void onResponse(byte[] bytes) {

    }
});
```

<!--
下面放例子的返回结果 
-->
response:
```
This response has no parameters.
```


