## DeletePayment

DeletePayment deletes an outgoing payment from DB. Note that it will not attempt to delete an In-Flight payment, since that would be unsafe.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |
| payment_hash   |	bytes	    |Payment hash to delete.|
| failed_htlcs_only   |	bool	    |Only delete failed HTLCs from the payment, not the payment itself.|

## Response:
This response has no parameters.

## Example:

<!--
java code example
-->

```java
LightningOuterClass.DeletePaymentRequest deletePaymentRequest = LightningOuterClass.DeletePaymentRequest.newBuilder()
                .setPaymentHash(byteStringFromHex("fd4c4943bf84694847960d8b7c50f37a23d6e77196fec0b5aab69cba969f453f"))
                .setFailedHtlcsOnly(false)
                .build();
Obdmobile.deletePayment(deletePaymentRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {

    }

    @Override
    public void onResponse(byte[] bytes) {

    }
});
private static ByteString byteStringFromHex(String hexString) {
    byte[] hexBytes = BaseEncoding.base16().decode(hexString.toUpperCase());
    return ByteString.copyFrom(hexBytes);
}
```

<!--
The response for the example
-->
response:
```
This response has no parameters.
```