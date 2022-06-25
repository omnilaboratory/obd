## introduction

The gRPC service is a service offered in exclusive mode, which means the user must run an obd node locally, and he alone can access the node. The obd node supports both assets and BTC, and the service of BTC is provided by lnd (based on lnd's codebase).

For exclusive and non-custodial modes, please refer to the [Architecture section](https://omnilaboratory.github.io/obd/#/Architecture).


## mapping of asset related interfaces

(updated Jun 25, 2022, keep updating)

The new asset-related gRPC interfaces of an obd node are listed here, and other common/bitcoin interfaces are provided by the original code of lnd. 

obd Interface: the asset-related gRPC interface.  
sub-service: the service it belongs to.  
Argument added: the newly added arguments.  
bitcoin-only lnd interface: the doc link of the original lnd interface. 

assetID: the asset id defined by Omnilayer is an unsigned 32-bit integer.  
omniAmount: is defined by Omnilayer, see [OmniBOLT spec 3](https://github.com/omnilaboratory/OmniBOLT-spec/blob/master/OmniBOLT-03-RSMC-and-OmniLayer-Transactions.md#string-to-int64)


| obd interface	    |	sub service		        		|	Argument added	    |   Request/Response    |  bitcoin-only lnd interface   |  
| -------- 	        |	-----------------------		|	-------------------	|  -------------------	|  -------------------	        |   
| AddHoldInvoiceRequest	      |	Invoices		    |	assetID: uint32, amount: omniAmount    | Request, Response | https://api.lightning.community/#addholdinvoice  |
| AddInvoice                	|	Lightning		    | assetID: uint32, amount: omniAmount    | Request, Response | https://api.lightning.community/#addinvoice      |
| ChannelBalance              |	Lightning       |	assetID: uint32, amount: omniAmount    | Response          | https://api.lightning.community/#channelbalance  |
| GetChanInfo                 |	Lightning       |	assetID: uint32, asset_capacity: omniAmount          | Response          | https://api.lightning.community/#getchaninfo     |
| GetNodeInfo                 |	Lightning       |	assetID: uint32, total_asset_capacity: omniAmount    | Response          | https://api.lightning.community/#getnodeinfo  |
| OpenChannel 	              |	Lightning		    |	assetID: uint32, asset_capacity: omniAmount          | Request           | https://api.lightning.community/#channelbalance  |
| Sendpaymentv2 	            |	Router		      |	assetID: uint32, asset_amount: omniAmount            | Request,Response          | https://api.lightning.community/#sendpaymentv2 |
