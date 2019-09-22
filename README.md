# LightningOnOmni | OmniBOLT Daemon
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/LightningOnOmnilayer/LightningOnOmni/blob/master/LICENSE) [![](https://img.shields.io/badge/standard%20readme-OK-brightgreen)](https://github.com/LightningOnOmnilayer/LightningOnOmni/blob/master/README.md) [![](https://img.shields.io/badge/golang-%3E%3D1.9.0-orange)](https://golang.org/dl/) [![](https://img.shields.io/badge/protocol-OmniBOLT-brightgreen)](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec) 

LightningOnOmni implements the [OmniBOLT](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec) specification, which enables Lightning network to be Omnilayer assets aware. Compile the source code and run the binary executable file, we will have an OmniBOLT deamon(OBD) providing all services for lightning network.   


# Table of Contents

 * [Dependency](https://github.com/LightningOnOmnilayer/LightningOnOmni#dependency)
 * [Installation](https://github.com/LightningOnOmnilayer/LightningOnOmni#installation)
	* [Step 1: fetch the source code](https://github.com/LightningOnOmnilayer/LightningOnOmni#step-1-fetch-the-source-code)
	* [Step 2: set up OmniCore node](https://github.com/LightningOnOmnilayer/LightningOnOmni#step-2)
	* [Step 3: compile and run OmniBOLT daemon](https://github.com/LightningOnOmnilayer/LightningOnOmni#step-3-compile-and-run-omnibolt-daemon)
	* [Step 4: test channel operations using Websocket testing tool](https://github.com/LightningOnOmnilayer/LightningOnOmni#step-4-test-channel-operations-using-websocket-testing-tool)
	* [Step 5: channel operations on test site](https://github.com/LightningOnOmnilayer/LightningOnOmni#step-5-channel-operations-on-test-site)
		* [login](https://github.com/LightningOnOmnilayer/LightningOnOmni#login)
		* [create channel](https://github.com/LightningOnOmnilayer/LightningOnOmni#create-channel)
		* [deposit](https://github.com/LightningOnOmnilayer/LightningOnOmni#deposit)
		* [payments in channel](https://github.com/LightningOnOmnilayer/LightningOnOmni#payments-in-a-channel)
		* close channel (TBD)
 * [How to Contribute](https://github.com/LightningOnOmnilayer/LightningOnOmni#how-to-contribute)
 * [Current Features](https://github.com/LightningOnOmnilayer/LightningOnOmni#current-features)
 * [Comming Features](https://github.com/LightningOnOmnilayer/LightningOnOmni#comming-features)
 * [Related Projects](https://github.com/LightningOnOmnilayer/LightningOnOmni#related-projects)

# Dependency

[Omnicore 0.18](https://github.com/OmniLayer/omnicore/tree/develop), which is currently in develop branch, and will be to finalize the release soon. 

Omnicore 0.18 integrates the latest BTC core 0.18, which enables relative time locker used in RSM contracts and HTL contracts.

# Installation
The following instruction works for Ubuntu 14.04.4 LTS, golang 1.10 or later.

## step 1: fetch the source code

on your terminal:

```
$ git clone https://github.com/LightningOnOmnilayer/LightningOnOmni.git
```

or if you already set up your local git repo, you just need to fetch the latest version: 

```
$ git pull origin master
```

check if all updated:

```
$ git remote -v
origin	https://github.com/LightningOnOmnilayer/LightningOnOmni.git (fetch)
origin	https://github.com/LightningOnOmnilayer/LightningOnOmni.git (push)
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
This is a tesing full node for our community to run/call/test omni commands remotely. The OmniBOLT daemon invocates Omni RPC commands from this node, if you use this configuration. It is the most conveniente way to get started.

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
If you fail to install gRPC by `go get google.golang.org/grpc` and other gRPC related packages used in this project,try this:
```
$ mkdir -p $GOPATH/src/google.golang.org/
$ cd $GOPATH/src/google.golang.org
$ git clone https://github.com/grpc/grpc-go grpc
$ git clone https://github.com/googleapis/go-genproto genproto
```

During compilation, if you come across:
```
cannot find package "golang.org/x/net/context" in any of:
	/usr/local/go/src/golang.org/x/net/context (from $GOROOT)
...
```

That's because these packages had been moved to github. Use the following commands to fix:
```
go to your GOPATH and:
$ cd src
$ mkdir golang.org
$ cd golang.org
$ mkdir x
$ cd x
$ git clone https://github.com/golang/net.git
$ git clone https://github.com/golang/sys.git
$ git clone https://github.com/golang/text.git
```

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
Since OmniBOLT deamon(OBD) exposes WebSocket services, we use web socket test client for Chrome to do experiments. Install it from:
```
https://chrome.google.com/webstore/detail/websocket-test-client/fgponpodhbmadfljofbimhhlengambbn?hl=en
```
Make sure your browser supports WebSocket, as displayed in this screenshot.

<p align="center">
  <img width="500" alt="Screenshot of Websocket online testing site" src="https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/imgs/WebSocketTestSite.png">
</p>

Input `ws://127.0.0.1:60020/ws`, press `Open`. If on the right text pannel, displays `OPENED`, then we are ready to send messeages to OBD.

The first message is to login as `Alice`. input the following request into the Request box, and press `SEND`:
```
{
	"type":1,
	"data":{
        "peer_id":"alice"
        }
}
```

In the `Message Log` pannel, displays the response message from OBD:
```
{"type":1,"status":true,"sender":"alice","result":"alice login"}
```

It works.

## Step 5: Channel Operations on test site

We built a test site for testing OBD commands for new users who is willing to learn the specification and source code. Remember this site is for testing only. The URL is:

```
ws://62.234.216.108:60020/ws
```
Open two chrom browsers, left is Alice and the right is Bob. Input URL and click `OPEN`, then both status will show `OPENED`.

### login
```
1、alice login
{
	"type":1,
	"data":{
        "peer_id":"alice"
        }
}
2、bob login
{
	"type":1,
	"data":{
        "peer_id":"bob"
        }
}
```

### create channel

Alice sends request to Bob for creating a channel between them:

**reqest:**
```
    {
    	"type":-32,
    		"data":{"funding_pubkey":"0389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0",
			"funding_address":"mtu1CPCHK1yfTCwiTquSKRHcBrW2mHmfJH"
    		},
	"recipient_peer_id":"bob"
    }
```
OBD creats the complete message for Alice and route it to Bob:
```
{
	"type":-32,
	"status":true,
	"sender":"bob",
	"result":{
		"type":-32,
		"status":true,
		"sender":"bob",
		"result":		
			{"chain_hash":"1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P",
			"channel_reserve_satoshis":0,
			"delayed_payment_basepoint":"",
			"dust_limit_satoshis":0,
			"feerate_per_kw":0,
			"funding_address":"mtu1CPCHK1yfTCwiTquSKRHcBrW2mHmfJH",
			"funding_pubkey":"0389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0",
			"funding_satoshis":0,
			"htlc_basepoint":"",
			"htlc_minimum_msat":0,
			"max_accepted_htlcs":0,
			"max_htlc_value_in_flight_msat":0,
			"payment_basepoint":"",
			"push_msat":0,
			"revocation_basepoint":"",
			"temporary_channel_id":[115,110,9,131,137,123,219,126,153,157,22,1,117,48,237,221,100,2,148,125,222,216,233,4,201,195,248,13,230,112,81,178],
			"to_self_delay":0
	}
}
```
In Bob's browser, he will see the message, and he accept the request, by sending the following message back:
```
{
	"type":-33,
	"data":{
		"temporary_channel_id":[115,110,9,131,137,123,219,126,153,157,22,1,117,48,237,221,100,2,148,125,222,216,233,4,201,195,248,13,230,112,81,178],
		"funding_pubkey":"0303391b3681f48f5f181bbfbdea741b9a2fdac0e8d99def43b6faed78bb8a4e28",
		"funding_address":"n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA",
		"attitude":true
	}
}
```

When you test, you should replace the `temporary_channel_id` by the exact value that OBD generates for you.

### deposit
Alice(mtu1CPCHK1yfTCwiTquSKRHcBrW2mHmfJH) deposits 0.01 to the channel (2N3vGUfBSNALGGxUo8gTYpVQAmLwjXomLhF), where fee is 0.00001, and then abtains the transaction ID:

**request**
```
{
    	"type":1009,
    	"data":{
		"fromBitCoinAddress":"mtu1CPCHK1yfTCwiTquSKRHcBrW2mHmfJH",
		"fromBitCoinAddressPrivKey":"cTBs2yp9DFeJhsJmg9ChFDuC694oiVjSakmU7s6CFr35dfhcko1V",
		"toBitCoinAddress":"2N3vGUfBSNALGGxUo8gTYpVQAmLwjXomLhF",
		"amount":0.001,
		"minerFee":0.00001,
    	}
}
```

OBD responses with the transaction script and other payloads:

```
{
	"type":1009,
	"status":true,
	"sender":"1323c03c-d60f-4465-af15-174d97938309",
	"result":{
		"hex":"02000000..... too long to display, you shall get your response from OBD......014eba8ac00000000",
		"txid":"e25c3f5763b11cda6ea56ad8eda7e5677a037aa614208e46acfa5d8484e89f39"
	}
}
```

then Alice tells Bob she created a deposit transaction:
```
{
	"type":-34,
	"data":{
		"temporary_channel_id":[138,61,69,205,17,204,61,156,218,215,97,165,250,225,12,179,46,100,75,124,17,22,112,193,15,17,84,236,116,199,108,126],
		"property_id":0,
 		"funding_tx_hex":"0200e28a6..... too long to display, you shall get your transaction hex from OBD .......a88ac00000000",
		"temp_address_pub_key":"03ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d",
    		"temp_address_private_key":"cSgTisoiZLzH5vrwHBMAXLC5nvND2ffcqqDtejMg12rEVrUMeP5R",
		"channel_address_private_key":"cTBs2yp9DFeJhsJmg9ChFDuC694oiVjSakmU7s6CFr35dfhcko1V"
	}
}
```
Bob replies he knows this deposit, and then OBD creates commitment transaction, revockable delivery transaction and breach remedy transaction:
```
{
	"type":-35,
	"data":{
		"channel_id":[57,159,232,132,132,93,250,172,70,142,32,20,166,122,3,122,103,229,167,237,216,106,165,110,218,28,177,99,87,63,92,226],
		"fundee_channel_address_private_key":"cUC9UsuybBiS7ZBFBhEFaeuhBXbPSm6yUBZVaMSD2DqS3aiBouvS",
		"attitude":true
    	}
    }
```

OBD sends a message to both Alice and Bob, reporting the status(true or false) of all the internal transactions:
```
{
	"type":-35,
	"status":true,
	"sender":"bob",
	"result":
	{
		"amount_a":0.001,
		"amount_b":0,
		"channel_id":[57,159,232,132,132,93,250,172,70,142,32,20,166,122,3,122,103,229,167,237,216,106,165,110,218,28,177,99,87,63,92,226],
		"channel_info_id":1,
		"create_at":"2019-09-10T17:39:04.5786763+08:00",
		"create_by":"alice",
		"curr_state":20,
		"fundee_sign_at":"2019-09-10T17:39:19.8482475+08:00",
		"funder_pub_key_2_for_commitment":"03ea01f8b137df5744ec2b0b91bc46139cabf228403264df65f6233bd7f0cbd17d",
		"funding_output_index":0,
		"funding_tx_hex":"02000000014eba4c789910f6a37455eda8da82517992e1e28a69a5b1d82233a8cc364ee0a4010000006a47304402207f1c6a788846bad8d1bc1f4a785dcb8c83aa8c90dc132e0bb5059b26b1a31c640220435a70c37cba9c9464e03f395588385253c2a73320a820da6e87dcfd0450b5fd01210389cc1a24ee6aa7e9b8133df08b60ee2fc41ea2a37e50ebafb4392d313594f1c0ffffffff02a08601000000000017a91475138ee96bf42cec92a6815d4fd47b821fbdeceb8799fe0100000000001976a91492c53581aa6f00960c4a1a50039c00ffdbe9e74a88ac00000000",
		"funding_txid":"e25c3f5763b11cda6ea56ad8eda7e5677a037aa614208e46acfa5d8484e89f39",
		"id":1,
		"peer_id_a":"alice",
		"peer_id_b":"bob",
		"property_id":0
	}
}
```

### payments in a channel

Now Alice can use this channel to pay to Bob. Here are some data generated by OBD used in contructing temporary multi-sig addresses and temporary transactions of i-th commitment transaction, which will be invalidated after new(i+1) commitment transaction is created.

You shall find your corresponding temporary data within the responses of OBD:
```
	last_temp_pub_key： n2gj8MDzUU7JZ6eVF5VpXcL4wUfaDXzTfJ
	last_temp_private_key： cSgTisoiZLzH5vrwHBMAXLC5nvND2ffcqqDtejMg12rEVrUMeP5R
	curr_temp_address： mpnHbpARXjUBcf6vib7E3jjD6Zv4CrvYuW
	curr_temp_private_key： cP8vR19XbtytyHgyBh1SV5dVAMLrLR2rzSU9EAcQTCUHij61u5C2
	curr_temp_pub_key: 02b8302d22a50fd84f34d528ff98998a6959bc7fb8f45b5f3fb44e23101aa5d8f2
```

Alice sends to Bob:

```
{
	"type":-351,
	"data":{
		"channel_id":[57,159,232,132,132,93,250,172,70,142,32,20,166,122,3,122,103,229,167,237,216,106,165,110,218,28,177,99,87,63,92,226],
		"amount":0.0001,
		"curr_temp_address_pub_key":"02b8302d22a50fd84f34d528ff98998a6959bc7fb8f45b5f3fb44e23101aa5d8f2",
		"curr_temp_address_private_key":"cP8vR19XbtytyHgyBh1SV5dVAMLrLR2rzSU9EAcQTCUHij61u5C2",
		"property_id":0,
		"channel_address_private_key":"cTBs2yp9DFeJhsJmg9ChFDuC694oiVjSakmU7s6CFr35dfhcko1V",
		"last_temp_address_private_key":"cSgTisoiZLzH5vrwHBMAXLC5nvND2ffcqqDtejMg12rEVrUMeP5R"
	}
}
```

then Bob replies the money is well receieved:
```
{
	"type":-352,
	"data":{
		"channel_id":[57,159,232,132,132,93,250,172,70,142,32,20,166,122,3,122,103,229,167,237,216,106,165,110,218,28,177,99,87,63,92,226],
		"curr_temp_address_pub_key":"0277bf9e0df3ffe2d8f22356fb198e9b0a2237b8c51bde77e3c7da5df09a4dba05",
		"curr_temp_address_private_key":"cQm58g783uvXbFjTjLN5STR3TrPiKueC5f9SCkAc9kSew4dj2Y2i",
		"last_temp_private_key":"",
		"request_commitment_hash":"822b8c9980e75fbb9c73d475da9c1e432ae5b664bfbb6ae0c4689333f67ea07e",
		"channel_address_private_key":"cUC9UsuybBiS7ZBFBhEFaeuhBXbPSm6yUBZVaMSD2DqS3aiBouvS",
		"attitude":true
	}
}
```

Remember to watch the display of `Message log` window, to see what OBD replies in each communication between Alice and Bob. These are usefull information when we try to debug OBD and to write real code for it.

This document is not completed yet, and will be updated during our programming.


# How to Contribute
OmniBOLT Daemon is MIT licensed open source software. Hopefully you can get started by doing the above steps, but Lightning network is not that easy to develop. Anyone is welcome to join this journey, and please be nice to each other, don't bring any illegal/private stuff, abuse or racial into our community.

Please submit issues to this repo or help us with those open ones.

Guidelines:

  * read the [OmniBOLT](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec) spec. If you have any question over there, raise issues in that repo.
  * ask questions or talk about things in Issues.
  * make branches and raise pull-request, even if working on the main repository.
  * don't copy/past any code from anywhere else in contribution, because we have limited resource to compare source codes to avoid legal issues. What we can do is to read your code, run tests of your newly developed modules and read your comments in your branch to see if it is solving a real problem. 
  * better running `go fmt` before commiting any code.
  * add test to any package you commit.
  * write/contribute light client testing tools, such as a HTML page supporting WebSocket, so that new programmers can have an intuitive experience to get started. That helps. We will release our tools for testing.



# Current Features

* Generate user OBD(OmniBOLT Daemon) address.
* Open Poon-Dryja Channel.
* BTC and Omni assets in funding and transaction.
* fund and close channel.
* Commitment Transaction within a channel.

# Comming Features

* HTL contracts, supported by HED, BR, RD, HT, HTRD transactions.
* Multiple channel management for one OBD, scaling out in performance.
* Payment across multiple channels.
* OBD communication.
* Light client implementation.
* Balance and transaction history.
* to be updated, pursuant to the development of OmniBOLT specification.


# Related projects: 

[https://github.com/OmniLayer/omniwallet](https://github.com/OmniLayer/omniwallet)

[https://github.com/OmniLayer/omnicore](https://github.com/OmniLayer/omnicore)

[https://github.com/OmniLayer/OmniJ](https://github.com/OmniLayer/OmniJ)

[https://github.com/OmniLayer/spec](https://github.com/OmniLayer/spec)

[https://github.com/LightningOnOmnilayer/Omni-BOLT-spec](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec)

[https://github.com/lightningnetwork/lightning-rfc](https://github.com/lightningnetwork/lightning-rfc)

[https://github.com/lightningnetwork/lnd](https://github.com/lightningnetwork/lnd)

[https://github.com/LightningOnOmnilayer/OmniWalletMobile](https://github.com/LightningOnOmnilayer/OmniWalletMobile)





 


