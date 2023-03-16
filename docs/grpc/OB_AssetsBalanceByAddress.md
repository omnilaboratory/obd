## OB_AssetsBalanceByAddress

OB_AssetsBalanceByAddress returns a list of all token balances for a given address.

## Arguments:
| Field		   |	gRPC Type		|	 Description  |
| -------- 	   |	---------       |    ---------    |  
| address	   |	string		    |	 The address of wallet balances.|

## Response:
| Field		         |	gRPC Type		|	   Description    |
| -------- 	         |	---------       |      ---------      |  
| propertyid            |	int64	        |The property identifier.|
| name            |	string	        |The name of the property.|
| balance            |	int64	        |The available balance of the address.|
| reserved            |	string	        |The amount reserved by sell offers and accepts.|
| frozen            |	string	        |The amount frozen by the issuer (applies to managed properties only).|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.AssetsBalanceByAddressRequest asyncAssetsBalanceRequest = LightningOuterClass.AssetsBalanceByAddressRequest.newBuilder()
        .setAddress(User.getInstance().getWalletAddress(mContext))
        .build();
Obdmobile.oB_AssetsBalanceByAddress(asyncAssetsBalanceRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.AssetsBalanceByAddressResponse resp = LightningOuterClass.AssetsBalanceByAddressResponse.parseFrom(bytes);
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
[ 
    {
        balance: 9999900000
        frozen: "0.00000000"
        name: "Usd"
        propertyid: 2147485160
        reserved: "0.00000000"
    }   
]
```
