## oB_WalletBalanceByAddress

oB_WalletBalanceByAddress returns total unspent outputs(confirmed and unconfirmed), all confirmed unspent outputs and all unconfirmed unspent outputs under control of the wallet.

## Arguments:
| Field		   |	gRPC Type		|	 Description  |
| -------- 	   |	---------       |    ---------    |  
| address	   |	string		    |	 The address of wallet balances.|

## Response:
| Field		         |	gRPC Type		|	   Description    |
| -------- 	         |	---------       |      ---------      |  
| total_balance            |	int64	        |The balance of the wallet.|
| confirmed_balance            |	int64	        |The confirmed balance of a wallet(with >= 1 confirmations).|
| unconfirmed_balance            |	int64	        |The unconfirmed balance of a wallet(with 0 confirmations).|
| address            |	string	        |The address of wallet.|
| account_balance            |	AccountBalanceEntryp[]	        |A mapping of each wallet account's name to its balance.|

**AccountBalanceEntry**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| key   |	string	    |    |  
| value     |	WalletAccountBalance	    |    |

**WalletAccountBalance**
| Field		            |	gRPC Type		    |	 Description  |
| -------- 	            |	---------           |    ---------    |  
| confirmed_balance   |	int64	    |  The confirmed balance of the account (with >= 1 confirmations).|  
| unconfirmed_balance     |	int64	    |  The unconfirmed balance of the account (with 0 confirmations).|

## Example:

<!--
java code example
-->

```java
LightningOuterClass.WalletBalanceByAddressRequest walletBalanceByAddressRequest = LightningOuterClass.WalletBalanceByAddressRequest.newBuilder()
        .setAddress("moR475qgPtKpb3znbuevyGK5zNbsEfCBmD")
        .build();
Obdmobile.oB_WalletBalanceByAddress(walletBalanceByAddressRequest.toByteArray(), new Callback() {
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
            LightningOuterClass.WalletBalanceByAddressResponse resp = LightningOuterClass.WalletBalanceByAddressResponse.parseFrom(bytes);
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
    account_balance {
      key: "default"
      value {
        confirmed_balance: 1559747
        unconfirmed_balance: 0
      }
    }
    address: "moR475qgPtKpb3znbuevyGK5zNbsEfCBmD"
    confirmed_balance: 1559747
    total_balance: 1559747
    unconfirmed_balance: 0
}
```