# Lightning On Omnilayer | OmniBOLT Daemon
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/omnilaboratory/obd/blob/master/LICENSE) [![](https://img.shields.io/badge/standard%20readme-OK-brightgreen)](https://github.com/omnilaboratory/obd/blob/master/README.md) [![](https://img.shields.io/badge/golang-%3E%3D1.9.0-orange)](https://golang.org/dl/) [![](https://img.shields.io/badge/protocol-OmniBOLT-brightgreen)](https://github.com/omnilaboratory/OmniBOLT-spec) 
[![](https://img.shields.io/badge/API%20V0.3-Document-blue)](https://api.omnilab.online) 

obd implements the [OmniBOLT](https://github.com/omnilaboratory/OmniBOLT-spec) specification, which enables Lightning network to be Omnilayer assets aware. Compile the source code and run the binary executable file, we will have an OmniBOLT deamon(OBD) providing all services for lightning network.   


# Table of Contents

 * [Dependency](https://github.com/omnilaboratory/obd#dependency)
 * [Installation](https://github.com/omnilaboratory/obd#installation)
	* [Step 1: fetch the source code](https://github.com/omnilaboratory/obd#step-1-fetch-the-source-code)
	* [Step 2: set up OmniCore node](https://github.com/omnilaboratory/obd#step-2)
	* [Step 3: compile and run OmniBOLT daemon](https://github.com/omnilaboratory/obd#step-3-compile-and-run-omnibolt-daemon)
	* [Step 4: test channel operations using Websocket testing tool](https://github.com/omnilaboratory/obd#step-4-test-channel-operations-using-websocket-testing-tool)
	* [Step 5: channel operations on test site](https://github.com/omnilaboratory/obd#step-5-channel-operations-on-test-site)
		* [login](https://github.com/omnilaboratory/obd#login)
	<!-- Removed by Neo Carmack 2020-06-09 -->		
	<!-- 	* [create channel](https://github.com/omnilaboratory/obd#create-channel)
		* [deposit](https://github.com/omnilaboratory/obd#deposit)
		* [payments in channel](https://github.com/omnilaboratory/obd#payments-in-a-channel)
		* close channel (TBD) -->

	<!-- Added by Kevin Zhang 2019-11-19 -->
	<!-- Removed by Neo Carmack 2020-06-09 -->
	<!-- 	* [Step 6: transfer assets through HTLC](https://github.com/omnilaboratory/obd#step-6-transfer-assets-through-HTLC) -->

 * [API Document](https://github.com/omnilaboratory/obd#api-document)
 * [How to Contribute](https://github.com/omnilaboratory/obd#how-to-contribute)
 * [Current Features](https://github.com/omnilaboratory/obd#current-features)
 * [Comming Features](https://github.com/omnilaboratory/obd#comming-features)
 * [Related Projects](https://github.com/omnilaboratory/obd#related-projects)

# Dependency

[Omnicore 0.18](https://github.com/OmniLayer/omnicore/tree/develop), which is currently in develop branch, and will be to finalize the release soon. 

Omnicore 0.18 integrates the latest BTC core 0.18, which enables relative time locker used in RSM contracts and HTL contracts.

# Installation
The following instruction works for Ubuntu 14.04.4 LTS, golang 1.10 or later.

## step 1: fetch the source code

on your terminal:

```
$ git clone https://github.com/omnilaboratory/obd.git
```

or if you already set up your local git repo, you just need to fetch the latest version: 

```
$ git pull origin master
```

check if all updated:

```
$ git remote -v
origin	https://github.com/omnilaboratory/obd.git (fetch)
origin	https://github.com/omnilaboratory/obd.git (push)
```

## Step 2: 
#### option 1: Remote Omnicore node 
Use our remote OmniCore node. Go to `\config\conf.ini`, you will see:
```
[chainNode]
host=62.234.216.108:18332
user=omniwallet
pass=cB3]iL2@eZ1?cB2?
```
This is a tesing full node for our community to run/call/test omni commands remotely. The OmniBOLT daemon invokes Omni RPC commands from this node, if you use this configuration. It is the most conveniente way to get started.

The other option uses local omnicore full node:  

#### option 2: Local Omnicore node 
[Install OmniCore](https://github.com/OmniLayer/omnicore#installation) on your local machine. Omnicore requires a full BTC core node, which may take days to synchronize the whole BTC database to your local device. After finishing synchronization, you can run omni/BTC commands for experiments, such as constructing raw transactions or generating new addresses.

Edit the configure file: `\config\conf.ini`
```
[chainNode]
host=127.0.0.1:port
user=your user name
pass=your password
```

## Step 3: Compile and Run OmniBOLT Daemon

Wait till all data downloaded.

```
$ go build obdserver.go
```
which generates the executable binary file `obdserver` under the source code directory. 

if you want to generate exe file for windows platform, use this:
```
$ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build obdserver.go
```
you will see an obdserver.exe file generated under the same directory.

Run:
```
$ ./obdserver
```
The terminal displays:
```
2019/08/23 23:05:15 rpcclient.go:23: &{62.234.216.108:18332 omniwallet cB3]iL2@eZ1?cB2?}
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ws                       --> LightningOnOmni/routers.wsClientConnect (3 handlers)
```
Which tells us the daemon is running. We are going to use WebSocket online testing tools to test our lightning commands.


## Step 4: Test channel operations using Websocket testing tool.

**NOTE: You should replace all of relevant data by the exact value that your own OBD generates for you**

OmniBOLT deamon(OBD) exposes WebSocket services for client interaction. For ease of use, we released GUI tool to help users to get started. Go to the [GUI tool repository](https://github.com/omnilaboratory/DebuggingTool) to download and try it.

Another option is to use web socket test client for Chrome to do experiments. Install it from:
```
https://chrome.google.com/webstore/detail/websocket-test-client/fgponpodhbmadfljofbimhhlengambbn?hl=en
```
Make sure your browser supports WebSocket, as displayed in this screenshot.

<p align="center">
  <img width="500" alt="Screenshot of Websocket online testing site" src="https://github.com/omnilaboratory/OmniBOLT-spec/blob/master/imgs/WebSocketTestSite.png">
</p>

Input `ws://127.0.0.1:60020/ws`, press `Open`. If on the right text pannel, displays `OPENED`, then we are ready to send messeages to OBD.

The first message is to sign up as `Alice`. input the following request into the Request box, and press `SEND`:

```json
{
	"type":101
}
```

In the `Message Log` pannel, displays the response message from OBD:

*Return mnemonic words by hirarchecal deterministic wallet system.*

```json
{
    "type":101,
    "status":true,
    "from":"c2215a60-8b81-439f-8cb3-11ba51691076",
    "to":"c2215a60-8b81-439f-8cb3-11ba51691076",
    "result":"two ribbon knee leaf easy pottery hobby pony mule test bridge liar sand mirror decline gasp focus this park undo rough cricket portion ignore"
}
```

Then go to login as `Alice`. input the following message and press `SEND`:

*The mnemonic words is as a login name.*

```json
{
    "type":1,
    "data":{
        "mnemonic":"two ribbon knee leaf easy pottery hobby pony mule test bridge liar sand mirror decline gasp focus this park undo rough cricket portion ignore"
    }
}
```

In the `Message Log` pannel, displays the response message from OBD:

*A SHA256 string of mnemonic words as a user id.*

```json
{
    "type":1,
    "status":true,
    "from":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
    "to":"all",
    "result":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2 login"
}
```

It works.

## Step 5: Channel Operations on test site

For the convenience of brand new users, who are willing to learn the specification and source code, we built a test site for testing OBD commands and it is for testing only. The URL is:

```
ws://62.234.216.108:60020/ws
```
Open two chrom browsers, left is Alice and the right is Bob. Input URL and click `OPEN`, then both status will show `OPENED`.


### Sign up

1、Alice sign up

Websocket request:

```
{
	"type":101
}
```

OBD responses:

```json
{
    "type":101,
    "status":true,
    "from":"c2215a60-8b81-439f-8cb3-11ba51691076",
    "to":"c2215a60-8b81-439f-8cb3-11ba51691076",
    "result":"two ribbon knee leaf easy pottery hobby pony mule test bridge liar sand mirror decline gasp focus this park undo rough cricket portion ignore"
}
```

2、Bob sign up

Websocket request:

```
{
	"type":101
}
```

OBD responses:

```json
{
    "type":101,
    "status":true,
    "from":"cec4e1db-ef38-4508-a9bf-8c5976df1916",
    "to":"cec4e1db-ef38-4508-a9bf-8c5976df1916",
    "result":"outer exhibit burger screen onion dog ensure net depth scan steel field pizza group veteran doctor rhythm inch dawn rotate gravity index modify utility"
}
```

### Login

1、Alice login

Websocket request:

```json
{
	"type":1,
    "data":{
        "mnemonic":"two ribbon knee leaf easy pottery hobby pony mule test bridge liar sand mirror decline gasp focus this park undo rough cricket portion ignore"
    }
}
```

OBD responses:

```json
{
    "type":1,
    "status":true,
    "from":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
    "to":"all",
    "result":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2 login"
}
```

2、Bob login

Websocket request:

```json
{
    "type":1,
    "data":{
        "mnemonic":"outer exhibit burger screen onion dog ensure net depth scan steel field pizza group veteran doctor rhythm inch dawn rotate gravity index modify utility"
    }
}
```

OBD responses:

```json
{
    "type":1,
    "status":true,
    "from":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be",
    "to":"all",
    "result":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be login"
}
```

*A SHA256 string of mnemonic words as a user id.*

Alice's id is: 7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2

Bob's   id is: f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be


Following operations can be found in the instruction of GUI tool or the online API document:


# API Document
Please visit OBD [online API documentation](https://api.omnilab.online) for the lastest update.

# How to Contribute
OmniBOLT Daemon is MIT licensed open source software. Hopefully you can get started by going through the above steps, but Lightning network is not that easy to develop. Anyone is welcome to join us in this journey, and please be nice to each other, don't bring any illegal/private stuff, abuse or racial into our community.

Please submit issues to this repo or help us with those open ones.

Guidelines:

  * read the [OmniBOLT](https://github.com/omnilaboratory/OmniBOLT-spec) spec. If you have any question over there, raise issues in that repo.
  * ask questions or talk about things in Issues.
  * make branches and raise pull-request, even if working on the main repository.
  * don't copy/past any code from anywhere else in contribution, because we have limited resource to compare source codes to avoid legal issues. What we can do is to read your code, run tests of your newly developed modules and read your comments in your branch to see if it is solving a real problem. 
  * better running `go fmt` before commiting any code.
  * add test to any package you commit.
  * write/contribute light client testing tools, such as a HTML page supporting WebSocket, so that new programmers can have an intuitive experience to get started. That helps. We will release our tools for testing.


Join us in [OmniBOLT slack channel](https://join.slack.com/t/omnibolt/shared_invite/enQtNzY2MDIzNzY0MzU5LTFlZTNlZjJhMzQxZTU2M2NhYmFjYjc1ZGZmODYwMWE3YmM0YjNhZWQyMDU2Y2VlMWIxYWFjN2YwMjlmYjUxNzA)

# Current Features (Jun 10, 2020)

* Generate OBD(OmniBOLT Daemon) addresss for users.  
* Open Poon-Dryja Channel.  
* BTC and Omni assets in funding and transaction.  
* fund and close channel.  
* Commitment Transaction within a channel.  
* List latest commitment transaction in a channel.   
* List all commmitment trsactions in a channel.  
* List all the breach remedy transactions in a channel.  
* Surveil the broadcasting commitment transactions and revockable delivery transactions.  
* Execute penalty.   
* HTL contracts, supported by RSMC, HED, BR, RD, HT, HTRD transactions.  
* Multiple channel management for a single OBD instance, scaling out in performance.  
* Multi hop payment using HTLC.  
* Light client implementation.  
* Peer to peer communication module.  
* GUI tool.  
* Atomic swap among channels.  

# Comming Features

* Invoice system.   
* Balance and transaction history.  
* Plugin for current lnd and c-lightning implementation.  
* Interoperability of obd and lnd.  
* to be updated, pursuant to the development of OmniBOLT specification.

# Experimental Features

* Smart Contract System on top of OBD channel. 
* "Payment is Settlement" theory.


# Related projects: 

[https://github.com/OmniLayer/omniwallet](https://github.com/OmniLayer/omniwallet)

[https://github.com/OmniLayer/omnicore](https://github.com/OmniLayer/omnicore)

[https://github.com/OmniLayer/OmniJ](https://github.com/OmniLayer/OmniJ)

[https://github.com/OmniLayer/spec](https://github.com/OmniLayer/spec)

[https://github.com/omnilaboratory/OmniBOLT-spec](https://github.com/omnilaboratory/OmniBOLT-spec)

[https://github.com/lightningnetwork/lightning-rfc](https://github.com/lightningnetwork/lightning-rfc)

[https://github.com/lightningnetwork/lnd](https://github.com/lightningnetwork/lnd)

[https://github.com/omnilaboratory/OmniWalletMobile](https://github.com/omnilaboratory/OmniWalletMobile)





 


