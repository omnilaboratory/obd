## OB_NewAddress

OB_NewAddress is used to generate a new address.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |  
| type_value	     |	int		  |	      The type of address to generate.|


## Response:
| Field		         |	gRPC Type		|	   Description    |
| -------- 	         |	---------       |      ---------      |  
| address            |	string	        |The address generated.|  

## Example:

<!--
java code example
-->

```java
LightningOuterClass.NewAddressRequest newAddressRequest 
    = LightningOuterClass.NewAddressRequest.newBuilder().setTypeValue(2).build();
Obdmobile.oB_NewAddress(newAddressRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {
        e.printStackTrace();
    }
    @Override
    public void onResponse(byte[] bytes) {
        if(bytes == null){
            return;
        }
        try {
            LightningOuterClass.NewAddressResponse newAddressResponse 
                = LightningOuterClass.NewAddressResponse.parseFrom(bytes);
            String address = newAddressResponse.getAddress();
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
    address: "n17BMaC82DMJX8r2VV5gBFA4QtMU8ELW42"
}
```


