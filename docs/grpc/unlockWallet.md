## unlockWallet

UnlockWallet is used at startup of obd(lnd,oblnd) to provide a password to unlock the wallet database.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |  
| wallet_password	     |	bytes		  |wallet_password should be the current valid passphrase for the daemon. This will be required to decrypt on-disk material that the daemon requires to function properly. When using REST, this field must be encoded as base64.|
| recovery_window	     |	int		  |recovery_window is an optional argument specifying the address lookahead when restoring a wallet seed. The recovery window applies to each individual branch of the BIP44 derivation paths. Supplying a recovery window of zero indicates that no addresses should be recovered, such after the first initialization of the wallet.|
| channel_backups	     |	ChanBackupSnapshot		  |channel_backups is an optional argument that allows clients to recover the settled funds within a set of channels. This should be populated if the user was unable to close out all channels and sweep funds before partial or total data loss occurred. If specified, then after on-chain recovery of funds, obd(lnd,oblnd) begin to carry out the data loss recovery protocol in order to recover the funds in each channel from a remote force closed transaction.|
| stateless_init	     |	bool		  |stateless_init is an optional argument instructing the daemon NOT to create any *.macaroon files in its file system.|
**ChanBackupSnapshot**
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| single_chan_backups|	ChannelBackups	    |The set of new channels that have been added since the last channel backup snapshot was requested.|  
| multi_chan_backup  |	MultiChanBackup	|A multi-channel backup that covers all open channels currently known to obd(lnd,oblnd).|
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
This response has no parameters.

## Example:

<!--
java code example
-->

```java
String password = "12345678"
Walletunlocker.UnlockWalletRequest unlockWalletRequest = Walletunlocker.UnlockWalletRequest.newBuilder().setWalletPassword(ByteString.copyFromUtf8(password )).build();
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
下面放例子的返回结果 
-->
response:
```
This response has no parameters.
```


