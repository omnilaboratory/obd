## restoreChannelBackups 

RestoreChannelBackups accepts a set of singular channel backups, or a single encrypted multi-chan backup and attempts to recover any funds remaining within the channel. If we are able to unpack the backup, then the new channel will be shown under listchannels, as well as pending channels.

## Arguments:
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


## Response:
This request has no parameters.

## Example:

<!--
java code example
-->

```java
LightningOuterClass.MultiChanBackup multiChanBackup =  chanBackupSnapshot.getMultiChanBackup();
LightningOuterClass.RestoreChanBackupRequest restoreChanBackupRequest = LightningOuterClass.RestoreChanBackupRequest.newBuilder()
        .setMultiChanBackup(multiChanBackup.getMultiChanBackup())
        .build();
Log.e(TAG, "multi Channel restoreChanBackupRequest Str" + String.valueOf(restoreChanBackupRequest));
Obdmobile.restoreChannelBackups(restoreChanBackupRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {
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


