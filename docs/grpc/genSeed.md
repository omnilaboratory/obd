## addInvoice
<!-- 
中文用注释符号注释掉。创建一个收款Invoice 
-->  

GenSeed is the first method that should be used to instantiate a new lnd instance. This method allows a caller to generate a new aezeed cipher seed given an optional passphrase. If provided, the passphrase will be necessary to decrypt the cipherseed to expose the internal wallet seed.

Once the cipherseed is obtained and verified by the user, the InitWallet method should be used to commit the newly generated seed, and create the wallet 

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


