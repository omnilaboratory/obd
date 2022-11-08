# OmniBOLT Daemon | Smart Asset Lightning Network
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/omnilaboratory/obd/blob/master/LICENSE) [![](https://img.shields.io/badge/standard%20readme-OK-brightgreen)](https://github.com/omnilaboratory/obd/blob/master/README.md) [![](https://img.shields.io/badge/golang-%3E%3D1.9.0-orange)](https://golang.org/dl/) [![](https://img.shields.io/badge/protocol-OmniBOLT-brightgreen)](https://github.com/omnilaboratory/OmniBOLT-spec) 
[![](https://img.shields.io/badge/API%20V0.3-Document-blue)](https://api.omnilab.online) 

<p align="center">
  <img width="500" alt="OmniBOLT-banner" src="docs/assets/omni-lightning-3.jpg">
</p>


OBD implements the [OmniBOLT](https://github.com/omnilaboratory/OmniBOLT-spec) specification, and it is an open source, off-chain decentralized platform, build upon BTC/OmniLayer network, implements basic multi hop HTLC payment, multi-currency atomic swap, and more off-chain contracts on the network of [smart assets lightning channels](https://github.com/omnilaboratory/OmniBOLT-spec/blob/master/OmniBOLT-02-peer-protocol.md#omni-address).  

<p align="center">
  <img width="500" alt="None Custodial OmniBOLT Daemon" src="docs/assets/None-Custodial-OmniBOLT-Daemon-2.png">
</p>
    

The latest document/tutorial has been moved to [OmniBOLT - Developers](https://omnilaboratory.github.io/obd/#/OBD-README).  

To know how obd works, jump to the [OmniBOLT - Architecture](https://omnilaboratory.github.io/obd/#/Architecture).  

The latest features and ETA is here: [OmniBOLT - Features and Roadmap](https://omnilaboratory.github.io/obd/#/features).  

## Quick Start:

Users can get quickly started with graphic or command line tools:  

#### Polar
The graphic terminal is provided by Polar, and is customized for obd: [Polar - GUI Terminal Customized](https://github.com/omnilaboratory/polar/releases). 

Polar helps Lightning Network application developers quickly build networks locally on their computers. Here is a short video demo: [https://twitter.com/omni_bolt/status/1549709303921410048?s=20&t=-M9Y4L0Bw_VialiSVPgqmA](https://twitter.com/omni_bolt/status/1549709303921410048?s=20&t=-M9Y4L0Bw_VialiSVPgqmA)  

#### Docker

Docker helps people to quickly interact with obd and omnicore via command line tools. OBD uses `docker-compose` to package `obd`, `omnicored`, and `btcd` together to make deploying these daemons easily. Please check the `docker-compose.yml` config file for all the configurations under:

```
https://github.com/omnilaboratory/lnd/tree/obd/docker/obtest
```

We compiled and deployed images for your testing:
```
omnicored: ccr.ccs.tencentyun.com/omnicore/omnicored:0.0.1
obd: ccr.ccs.tencentyun.com/omnicore/ob-lnd:${lnd-version:-0.0.7
```
Now we can:  

#### Issue tokens, Build obd network, and Lightning pay tokens

[test-shell-template.md](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md) instructs how to start with command line tool, including: 
* [start nodes](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md#alice-and-bob-cli)
* [build a network with three nodes(Alice, Bob, Carol)](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md#connect) 
* [generate address](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md#newaddress)
* [issue assets using omnicore](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md#send_coin)
* [fund each node by assets](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md#send_coin)  
* [open channels](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md#openchannel) 
* [create invoices](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md#addinvoice) 
* [pay invoices](https://github.com/omnilaboratory/lnd/blob/obd/docker/obtest/test-shell-template.md#sendpayment)  

The testing asset id is `--asset_id = 2147483651`.  

To issue assets on the Bitcoin/Omnilayer mainnet, you should deploy an omnicore full node and execute the cli to issue. For non-developers, we recommend you to visit the official [https://www.omniwallet.org/](https://www.omniwallet.org/) (https://github.com/OmniLayer/omniwallet) for easier and quicker manage your assets.  

#### Backend and Faucet(on Regtest)

The [Omnicore Proxy](https://github.com/omnilaboratory/omnicore-proxy) offers the backend public anonymous omni/bitcoin services for obd nodes.  
It is specified in `lnd/start-a.sh` when starts an OBD node:  
```
exec ./lnd-debug \
    --autopilot.active \
    --maxpendingchannels=100 \
    --noseedbackup \
    "--lnddir=~/apps/oblnd" \
    "--$CHAIN.active" \
    "--$CHAIN.$NETWORK" \
    "--$CHAIN.node"="$BACKEND" \
    "--$BACKEND.rpchost"="regnet.oblnd.top:18332" \
    "--rpclisten=$HOSTNAME:10010" \
    "--rpclisten=localhost:10010" \
    --listen=0.0.0.0:9736 \
    "--restlisten=0.0.0.0:18080" \
    --debuglevel="$DEBUG" \
    --$BACKEND.zmqpubrawblock=tcp://regnet.oblnd.top:28332 \
    --$BACKEND.zmqpubrawtx=tcp://regnet.oblnd.top:28333 \
    "$@"
```

The proxy decouples the lightning node and the full Bitcoin/Omnilayer node, to lower the barriers of OBD deployment, especially for mobile nodes. 

The complete white-listed services are: [https://github.com/omnilaboratory/omnicore-proxy/blob/master/whitelist_proxy/whitelist_proxy.go](https://github.com/omnilaboratory/omnicore-proxy/blob/master/whitelist_proxy/whitelist_proxy.go)
 

## Community

Discord: [http://discord.gg/2QYqzSMZuy](http://discord.gg/2QYqzSMZuy)  
Slack: [https://join.slack.com/t/omnibolt/shared_invite/zt-ad732myf-1G7lXpHPkFH_yRcilwT4Ig](https://join.slack.com/t/omnibolt/shared_invite/zt-ad732myf-1G7lXpHPkFH_yRcilwT4Ig)

## Video Tutorials
Video tutorials can be found here:   
  
* [OmniBOLT Technology guide part I](https://youtu.be/G-T_uwqzDAI)  

## Mainnet warning
OBD is still in an early stage, and `sendToMany` in omnicore is still under development. We do not recommend running it on mainnet with real money just yet, which may lead to loss of funds unless you want to take a reckless adventure.  

## Security
If you discover a vulnerability, weakness, or threat that can potentially compromise the security of obd, we ask you to keep it confidential and submit your concern directly to [the team](mailto:neo.carmack@gmail.com?subject=%5BGitHub%5D%20OmniBOLT%20Security).

 
