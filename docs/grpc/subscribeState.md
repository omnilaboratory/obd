## subscribeState  

SubscribeState subscribes to the state of the wallet. The current wallet state will always be delivered immediately.

#### Arguments:
This request has no parameters.         | 


#### Response:
| Field		         |	gRPC Type		|	   Description  |
| -------- 	         |	---------       |      ---------    |  
| state              |	WalletState	    |A set of single-chan static channel backups.|
**WalletState**
| Name		         |	Number		    |	   Description  |
| -------- 	         |	---------       |      ---------    |  
| NON_EXISTING       |	0	            |NON_EXISTING means that the wallet has not yet been initialized.|
| LOCKED             |	1	            |LOCKED means that the wallet is locked and requires a password to unlock.|
| UNLOCKED           |	2	            |UNLOCKED means that the wallet was unlocked successfully, but RPC server isn't ready.|
| RPC_ACTIVE         |	3	            |RPC_ACTIVE means that the obd(lnd,oblnd) server is active but not fully ready for calls.|
| SERVER_ACTIVE      |	4	            |SERVER_ACTIVE means that the obd(lnd,oblnd) server is ready to accept calls.|
| WAITING_TO_START   |	255	            |WAITING_TO_START means that node is waiting to become the leader in a cluster and is not started yet.|
#### Example:

<!--
java code example
-->

```java
Stateservice.SubscribeStateRequest subscribeStateRequest = Stateservice.SubscribeStateRequest.newBuilder().build();
Obdmobile.subscribeState(subscribeStateRequest.toByteArray(),new RecvStream(){
    @Override
    public void onError(Exception e) {
        e.printStackTrace();
    }

    @RequiresApi(api = Build.VERSION_CODES.N)
    @Override
    public void onResponse(byte[] bytes) {
        if (bytes == null){
            return;
        }
        try {
            Stateservice.SubscribeStateResponse subscribeStateResponse = Stateservice.SubscribeStateResponse.parseFrom(bytes);
            int walletState = subscribeStateResponse.getStateValue();
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

    }
});
```

<!--
下面放例子的返回结果 
-->
response:
```
{
    state: WAITING_TO_START
    state_value: 255
}
```


