## unlockWallet

UnlockWallet is used at startup of obd(lnd,oblnd) to provide a password to unlock the wallet database.

## Arguments:
**ChannelBackup**

| Field		            |	gRPC Type		|	Description  |
| -------- 	            |	---------       |   ---------    |  
| chan_point            |	ChannelPoint	|   Identifies the channel that this backup belongs to.|  
| chan_backup           |	bytes	        |   Is an encrypted single-chan backup. this can be passed to RestoreChannelBackups, or the WalletUnlocker Init and Unlock methods in order to trigger the recovery protocol. When using REST, this field must be encoded as base64.|

**ChannelBackups**

| Field		        |	gRPC Type		    |	 Description  |
| -------- 	        |	---------           |    ---------    |  
| chan_backups      |	ChannelBackup[]	    |    A set of single-chan static channel backups.|

**ChannelPoint**

| Field		            |	gRPC Type		|	 Description  |
| -------- 	            |	---------       |    ---------    |  
| funding_txid_bytes    |	string          |    Txid of the funding transaction. When using REST, this field must be encoded as base64.|
| funding_txid_str      |	bytes           |    Hex-encoded string representing the byte-reversed hash of the funding transaction.|
| output_index          |	int             |    The index of the output of the funding transaction|

**MultiChanBackup**

| Field		         |	gRPC Type		|	 Description  |
| -------- 	         |	---------       |    ---------    |  
| chan_point         |	ChannelPoint    |    Identifies the channel that this backup belongs to.|
| multi_chan_backup  |	bytes           |    A single encrypted blob containing all the static channel backups of the channel listed above. This can be stored as a single file or blob, and safely be replaced with any prior/future versions. When using REST, this field must be encoded as base64.|


## Response:
This response has no parameters.

## Example:

<!--
java code example
-->

```java
String password = "12345678"
Walletunlocker.UnlockWalletRequest unlockWalletRequest 
    = Walletunlocker.UnlockWalletRequest.newBuilder()
    .setWalletPassword(ByteString.copyFromUtf8(password ))
    .build();
    Obdmobile.unlockWallet(unlockWalletRequest.toByteArray(), new Callback() {
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
The response for the example
-->
response:
```
This response has no parameters.
```


