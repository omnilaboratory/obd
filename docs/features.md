# Features

## Current Features

(latest updated Sep 23, 2020)

**Core 0.1:**  

* Channel:  
	* Generate OBD(OmniBOLT Daemon) address.  
	* Open Poon-Dryja Channel.  
	* Omni assets and Bitcoin funding.  
	* fund and close channel.  

* Payment:  
	* Commitment Transaction within a channel.  
	* HTL contracts, supported by RSMC, HED, BR, RD, HT, HTRD transactions.  
	* Multiple channel management for a single OBD instance, scaling out in performance.  
	* Multi hop payment using HTLC.  
	* Invoice system. 
	* Surveil the broadcasting commitment transactions and revockable delivery transactions.  
	* Execute penalty.   

* Query:  
	* List latest commitment transaction in a channel.   
	* List all commmitment trsactions in a channel.  
	* List all the breach remedy transactions in a channel.  
	* Balance. 
	* Channels.
  
* Network:  
	* Peer to peer communication module, using libP2P.   
 
* Application contracts:
	* Atomic swap among channels.  

* Tracker:
	* Tracker for network statistics.  
	* Balance and transaction history.   


**Tools:**  
* GUI tool version 0.1 for developers.  
* JS SDK(wallet toolkit for light client) version 0.1  
	* Generate seeds and pub/priv key pairs.  
	* backup and restore keys.  
	* Websocket API to operate remote/local obd node.  



## Coming Features   
(latest updated Sep 23, 2020)
 
* OBD as a plugin for current lnd implementation. (~5 weeks, middle of Oct, 2020)  
* Interoperability of obd and lnd. (~5 weeks, middle of october, 2020)  
* Add obd grpc interfaces to lnd interface package. lnd clients will be able to operate obd channel by the newly added interfaces. (6~8 weeks, earlier Nov, 2020)  
* Outsource channel monitoring and penalizing malicious activity.(to be estimated)  
* Update core version 0.1 to new omnicore with SegWit and sendmany supported.  (~5 weeks, earlier Nov, 2020)  
* Service quality statistics by tracker. (to be estimated)
* to be updated, pursuant to the development of OmniBOLT specification.  
 


## Experimental Features

* Smart Contract System on top of OBD channel. 
* "Payment is Settlement" theory.

