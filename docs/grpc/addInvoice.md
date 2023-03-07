## addInvoice
<!-- 
中文用注释符号注释掉。创建一个收款Invoice 
-->  

Create an Invoice and add it to the local database. The key of the local K-V store is the preimage of the hash locker which has to be unique:  

1. Any duplicated invoices are rejected.
2. An invoice must identify an asset ID.  

#### Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |  
| memo	     |	string		  |	                |  
| asset_id   |	uint32		  |                 |  
| amount     |	omniAmount  |                 | 
| r_preimage |	bytes       |                 | 


#### Response:
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| r_hash	         |	bytes		    |	                |  
| payment_request  |	string		  |                 |  
| add_index        |	uint64      |                 | 
| payment_addr     |	bytes       |                 | 

#### Example:

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


