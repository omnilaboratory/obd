# Features

## Current Features

(latest updated Sep 1, 2022)

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
	* Surveil broadcasting commitment transactions and revockable delivery transactions.  
	* Execute penalty.   

* Query:  
	* List latest commitment transaction in a channel.   
	* List all commmitment trsactions in a channel.  
	* List all the breach remedy transactions in a channel.  
	* Balances. 
	* Channels.
  
* Network:  
	* Peer to peer communication module, using libP2P.   
 
* Application contracts:
	* Atomic swap among channels.  

* Tracker:
	* Tracker for network statistics.  
	* Balance and transaction history.   


**Tools:**  
* [Polar](https://github.com/omnilaboratory/polar/releases/tag/v1.3.0) for one-click obd network building.  
* [gRPC services](https://omnilaboratory.github.io/obd/#/grpc-api) in exclusive mode.  
* JS SDK(experimental wallet toolkit for light client) version 0.1  
	* Generate seeds and pub/priv key pairs.  
	* backup and restore keys.  
	* Websocket API to operate remote/local obd node.  



## Coming Features   
(latest updated Oct 23, 2022)
 
* October 2022, a benchmark for obd.
* November 2022, Android OmniBOLT wallet for Bitcoin and Omnilayer assets.  
* December 2022, DEX based on atomic swap protocol.
* Earlier Dec 2022, Update core version 0.1 to new omnicore with SegWit and `sendToMany` supported.
* March 2023, NFT support in OmniBOLT android wallet.   
* (to be estimated) Service quality statistics by the tracker network.
* to be updated, pursuant to the development of OmniBOLT specification.


## Experimental Features

* Smart Contract System on top of OBD channel. 
* "Payment is Settlement" theory. 

