## oB_SetDefaultAddress

oB_SetDefaultAddress is set as the default address.

## Arguments:
| Field		   |	gRPC Type		|	 Description  |
| -------- 	   |	---------       |    ---------    |  
| address	   |	string		    |	 The address of wallet balances.|

## Response:
This response has no parameters.

## Example:

<!--
java code example
-->

```java
LightningOuterClass.SetDefaultAddressRequest setDefaultAddressRequest = LightningOuterClass.SetDefaultAddressRequest.newBuilder()
        .setAddress("moR475qgPtKpb3znbuevyGK5zNbsEfCBmD")
        .build();
Obdmobile.oB_SetDefaultAddress(setDefaultAddressRequest.toByteArray(), new Callback() {
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