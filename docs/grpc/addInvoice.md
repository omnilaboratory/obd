## OB_AddInvoice

OB_AddInvoice attempts to add a new invoice to the invoice database. Any duplicated invoices are rejected, therefore all invoices must have a unique payment preimage.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |  
| memo	     |	string		  |	                |  
| asset_id   |	uint32		  |                 |  
| amount     |	omniAmount  |                 | 
| r_preimage |	bytes       |                 | 

## Response:
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| r_hash	         |	bytes		    |	                |  
| payment_request  |	string		  |                 |  
| add_index        |	uint64      |                 | 
| payment_addr     |	bytes       |                 | 

## Example:

<!--
java code example
-->

```java
obdmobile.addInvoice(...)
```

<!--
下面放例子的返回结果 
-->
response:
```
xxxxxx
```


