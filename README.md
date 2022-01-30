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

Graphic terminal that assists you get started quickly is here: [OmniBOLT - GUI Terminal](https://omnilaboratory.github.io/obd/#/GUI-tool). 

Join our community to get the latest progress: [communities](https://omnilaboratory.github.io/obd/#/communities)

Video tutorials can be found here:   
  
* [OmniBOLT Technology guide part I](https://youtu.be/G-T_uwqzDAI)  
* [Step 1 -- create and fund channel](https://youtu.be/PbbNk2JCopA)
* [Step 2 -- create invoice](https://youtu.be/Z9UmHFclGdc)
* [Step 3 -- pay invoice](https://youtu.be/NEexFe7R9kc)




## How to Build:

* make linux-amd64   
the complied file will in bin dir ,example: lncli-linux-amd64  obd-linux-amd64  tracker-linux-amd64

* or use go build command directly:
1. go build -o bin/
1. go build -o bin/ ./tracker
1. cd lnd go build -o ../bin/ github.com/lightningnetwork/lnd/cmd/lncli







 
