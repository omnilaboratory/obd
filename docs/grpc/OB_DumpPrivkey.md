## OB_DumpPrivkey
OB_DumpPrivkey is used to export the keywif of the wallet.

## Arguments:
| Field		   |	gRPC Type		|	 Description  |
| -------- 	   |	---------       |    ---------    |  
| address	   |	string		    |The address of wallet balances.|

## Response:
| Field		              |	gRPC Type		      |	  Description   |
| -------- 	            |	---------         |    ---------    |  
| keyWif	   |	string		    |The keyWif of wallet.|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.DumpPrivkeyRequest dumpPrivkeyRequest = LightningOuterClass.DumpPrivkeyRequest.newBuilder()
                .setAddress("moR475qgPtKpb3znbuevyGK5zNbsEfCBmD")
                .build();
Obdmobile.oB_DumpPrivkey(dumpPrivkeyRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.DumpPrivkeyResponse resp = LightningOuterClass.DumpPrivkeyResponse.parseFrom(bytes);
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
    keyWif: jj123dl2lmd56mdngkpaje
}
```


