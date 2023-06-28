## subscribeChannelChange
subscribeChannelChange allows a client to sub-subscribe to the most up to date information concerning the state of all channel backups. Each time a new channel is added, we return the new set of channels, along with a multi-chan backup containing the backup info for all channels. Each time a channel is closed, we send a new update, which contains new new chan back ups, but the updated set of encrypted multi-chan backups with the closed channel(s) removed.

## Arguments:
This request has no parameters.


## Response:
| Field		              |	gRPC Type		      |	  Description   |
| -------- 	            |	---------         |    ---------    |  
| Res   |	string	  |The res.|  

## Example:

<!--
java code example
-->

```java
Obdmobile.subscribeChannelChange(LightningOuterClass.ChannelBackupSubscription.newBuilder().build().toByteArray(), new RecvStream() {
    @Override
    public void onError(Exception e) {
        LogUtils.e(TAG, e.getMessage());
    }

    @Override
    public void onResponse(byte[] bytes) {
        if (bytes == null) {
            return;
        }
        try {
            LightningOuterClass.SubscribeChannelChangeRes resp = LightningOuterClass.SubscribeChannelChangeRes.parseFrom(bytes);
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
    res: 1
}
```


