# OmniBOLT Daemon | Smart Asset Lightning Network
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/omnilaboratory/obd/blob/master/LICENSE) [![](https://img.shields.io/badge/standard%20readme-OK-brightgreen)](https://github.com/omnilaboratory/obd/blob/master/README.md) [![](https://img.shields.io/badge/golang-%3E%3D1.9.0-orange)](https://golang.org/dl/) [![](https://img.shields.io/badge/protocol-OmniBOLT-brightgreen)](https://github.com/omnilaboratory/OmniBOLT-spec) 
[![](https://img.shields.io/badge/API%20V0.3-Document-blue)](https://api.omnilab.online) 

<p align="center">
  <img width="500" alt="OmniBOLT-banner" src="docs/assets/omni-lightning.png">
</p>


OBD implements the [OmniBOLT](https://github.com/omnilaboratory/OmniBOLT-spec) specification, and it is an open source, off-chain decentralized platform, build upon BTC/OmniLayer network, implements basic multi hop HTLC payment, multi-currency atomic swap, and more off-chain contracts on the network of [smart assets lightning channels](https://github.com/omnilaboratory/OmniBOLT-spec/blob/master/OmniBOLT-02-peer-protocol.md#omni-address).  

<p align="center">
  <img width="750" alt="None Custodial OmniBOLT Daemon" src="docs/assets/None-Custodial-OmniBOLT-Daemon-2.png">
</p>

In addition, OBD is designed special for inbound liquidity providers. A daemon allows thousands of remote light clients connections, including connections from [LND wallets](https://omnilaboratory.github.io/obd/#/Architecture?id=lnd-integrated). 

Clone, compile the source code and run the binary executable file, you will have a featured OmniBOLT deamon(OBD) to start the journey of lightning network.     

The latest document/tutorial has been moved to [OmniBOLT - Developers](https://omnilaboratory.github.io/obd/#/OBD-README).  

To know how OBD works, jump to the [OmniBOLT - Architecture](https://omnilaboratory.github.io/obd/#/Architecture).  

The latest features and ETA is here: [OmniBOLT - Features and Roadmap](https://omnilaboratory.github.io/obd/#/features).  

## Quick Start:

#### Graphic Terminal
Graphic terminal is provided by Polar, and is custumized for OBD: [Polar - GUI Terminal](https://github.com/omnilaboratory/polar/releases). 

Polar helps Lightning Network application developers quickly build networks locally on their computers. Here is a short video demo: [https://twitter.com/omni_bolt/status/1549709303921410048?s=20&t=-M9Y4L0Bw_VialiSVPgqmA](https://twitter.com/omni_bolt/status/1549709303921410048?s=20&t=-M9Y4L0Bw_VialiSVPgqmA)  

#### Docker

Docker helps people to quickly interact with OBD and omnicore via their command line tools. OBD uses `docker-compose` to package `obd`, `omnicored`, and `btcd` together to make deploying these daemons easily. Please check the `docker-compose.yml` config file for all the configurations under:

```
https://github.com/omnilaboratory/lnd/tree/obd/docker/obtest
```

We compiled and deployed images for your testing:
```
omnicored: ccr.ccs.tencentyun.com/omnicore/omnicored:0.0.1
obd: ccr.ccs.tencentyun.com/omnicore/ob-lnd:${lnd-version:-0.0.7
```

[test-shell-template.md](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md) instructs how to build a network with three nodes(Alice, Bob, Carol), fund each node, build channels, create invoices and pay invoices.  

The testing asset id is `--asset_id = 2147483651`.  

#### Faucet

The faucet we deployed for developers is in this repository: [https://github.com/omnilaboratory/omnicore-faucet-api](https://github.com/omnilaboratory/omnicore-faucet-api) 

** main api**： 
* mine  
* send_coin  
* get asset balance  
* list assets  
* query asset  
* create asset  

#### Omnicore Proxy
Omnicore Proxy offers public anonymous omni services for obd nodes: [https://github.com/omnilaboratory/btcwallet/tree/publicrpc](https://github.com/omnilaboratory/btcwallet/tree/publicrpc) 


## Community

Discord: [http://discord.gg/2QYqzSMZuy](http://discord.gg/2QYqzSMZuy)
Slack: [https://join.slack.com/t/omnibolt/shared_invite/zt-ad732myf-1G7lXpHPkFH_yRcilwT4Ig](https://join.slack.com/t/omnibolt/shared_invite/zt-ad732myf-1G7lXpHPkFH_yRcilwT4Ig)

## Video Tutorials
Video tutorials can be found here:   
  
* [OmniBOLT Technology guide part I](https://youtu.be/G-T_uwqzDAI)  


## Security
If you discover a vulnerability, weakness, or threat that can potentially compromise the security of obd, we ask you to keep it confidential and submit your concern directly to [the team](mailto:neo.carmack@gmail.com?subject=%5BGitHub%5D%20OmniBOLT%20Security).

 
