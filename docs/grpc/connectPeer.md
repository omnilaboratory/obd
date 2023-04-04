## ConnectPeer

ConnectPeer attempts to establish a connection to a remote peer. This is at the networking level, and is used for communication between nodes. This is distinct from establishing a channel with a peer.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |
| addr   |	LightningAddress	    |Lightning address of the peer to connect to.|
| perm   |	bool	    |If set, the daemon will attempt to persistently connect to the target peer. Otherwise, the call will be synchronous.|
| timeout   |	uint64	    |The connection timeout value (in seconds) for this request. It won't affect other requests.|

**LightningAddress**

| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| pubkey   |	string	    |The identity pubkey of the Lightning node.|
| host   |	string	    |The network location of the lightning node, e.g. `69.69.69.69:1337` or `localhost:10011`.|

## Response:
This response has no parameters.

## Example:

<!--
java code example
-->

```java
String nodeUri = "025af4448f55bf1e6de1ae0486d4d103427c4e559a62ed7f8035bb1ed1af734f61@192.144.199.67:9735"
LightningOuterClass.LightningAddress lightningAddress = LightningOuterClass.LightningAddress.newBuilder()
                .setHostBytes(ByteString.copyFrom(nodeUri.getHost().getBytes(StandardCharsets.UTF_8)))
                .setPubkeyBytes(ByteString.copyFrom(nodeUri.getPubKey().getBytes(StandardCharsets.UTF_8))).build();
LightningOuterClass.ConnectPeerRequest connectPeerRequest = LightningOuterClass.ConnectPeerRequest.newBuilder().setAddr(lightningAddress).build();
Obdmobile.connectPeer(connectPeerRequest.toByteArray(), new Callback() {
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