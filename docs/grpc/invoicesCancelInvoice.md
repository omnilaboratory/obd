## InvoicesCancelInvoice

InvoicesCancelInvoice cancels a currently open invoice. If the invoice is already canceled, this call will succeed. If the invoice is already settled, it will fail.

## Arguments:
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| payment_hash   |	bytes	    |Hash corresponding to the (hold) invoice to cancel. When using REST, this field must be encoded as base64.|

## Response:
This response has no parameters.

## Example:

<!--
java code example
-->

```java
InvoicesOuterClass.CancelInvoiceMsg cancelInvoiceMsg = InvoicesOuterClass.CancelInvoiceMsg.newBuilder()
                            .setPaymentHash("\244s\224\236R\r\005\271\3450\356g\354v\214\347\300b\t\316\315r\232IfZ\336_\272\203\226\f")
                            .build();
Obdmobile.invoicesCancelInvoice(cancelInvoiceMsg.toByteArray(), new Callback() {
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

