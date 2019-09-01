# LightningOnOmni | OmniBOLT Daemon

LightningOnOmni implements the [OmniBOLT](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec) specification, which enables Lightning network to be Omnilayer assets aware. Compile the source code and run the binary executable file, we will have an OmniBOLT deamon(OBD) providing all services for lightning network.   


# Table of Contents

 * [Dependency](https://github.com/LightningOnOmnilayer/LightningOnOmni#dependency)
 * [Installation](https://github.com/LightningOnOmnilayer/LightningOnOmni#installation)
 * [How to Contribute](https://github.com/LightningOnOmnilayer/LightningOnOmni#how-to-contribute)
 * [Current Features](https://github.com/LightningOnOmnilayer/LightningOnOmni#current-features)
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
Since OmniBOLT deamon(OBD) exposes WebSocket services, we use online WS testing tool to do experiments. Go to:
```
https://www.websocket.org/echo.html
```
Make sure your browser supports WebSocket, as displayed in this screenshot.

<p align="center">
  <img width="500" alt="Screenshot of Websocket online testing site" src="https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/imgs/WebSocketTestSite.png">
</p>

Input `ws://127.0.0.1:60020/ws`, press `Connect`. If on the right text pannel, displays `CONNECTED`, then we are ready to send messeages to OBD.

The first message is to get new Omni address for a channel. input the following request into the Message box, and press `SEND`:
```
{"type":1001", data":"email"}
```

In the right side text pannel, displays the response message from OBD:
```
RECEIVED: {"type":1001,"status":true,"sender":"59dfb5e2-f1dc-46c6-8ff3-dfc9f2f1ea82","result":"mzCihFnTFyZUo76QMKovoWWJAPkBqDi63J"}
```

It works

## Step 5: Channel Operations

For example:

[type: -32 openchannel](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/OmniBOLT-02-peer-protocol.md#the-open_channel-message)

Alice:
```
 {"type":0,
    "sender":"44ff0d17-13d0-4741-9f8d-e59b17011965",
    "recipient":"",
    "data":"{\"id\":1,
              \"chain_hash\":\"1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P\",   
              \"temporary_channel_id\":[229,183,118,180,41,204,14,173,33,18,101,64,250,6,244,29,115,151,105,108,147,205,77,16,175,249,148,105,117,192,181,34],
              \"funding_satoshis\":0,
              \"push_msat\":0,
              \"dust_limit_satoshis\":0,
              \"max_htlc_value_in_flight_msat\":0,
              \"channel_reserve_satoshis\":0,
              \"htlc_minimum_msat\":0,
              \"feerate_per_kw\":0,
              \"to_self_delay\":0,
              \"max_accepted_htlcs\":0,
              \"funding_pubkey\":\"n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn\",
              \"revocation_basepoint\":\"\",
              \"payment_basepoint\":\"\",
              \"delayed_payment_basepoint\":\"\",
              \"htlc_basepoint\":\"\",
              \"create_at\":\"2019-08-23T09:22:16.0104522+08:00\"
    }"
} 
```

[type: -33 ChannelAccept](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/OmniBOLT-02-peer-protocol.md#the-accept_channel-message)
Bob:
```
{
  "type":-33,
  "data":[229,183,118,180,41,204,14,173,33,18,101,64,250,6,244,29,115,151,105,108,147,205,77,16,175,249,148,105,117,192,181,34]
}
```

To test "open_channel" in light client mode, open two browsers, one for Alice, and one for Bob. Alice sends [open_channel](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/OmniBOLT-02-peer-protocol.md#the-open_channel-message) to Bob, and Bob replies with APPROAVAL:

<p align="center">
  <img width="600" alt="create channel" src="https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/imgs/CreateChannel.png">
</p>


# How to Contribute
OmniBOLT Daemon is MIT licensed open source software. Hopefully you can get started by doing the above steps, but Lightning network is not that easy to develop. Anyone is welcome to join this journey, and please be nice to each other, don't bring any illegal/private stuff, abuse or racial into our community.   

Please submit issues to this repo or help us with those open ones.

Guidelines:

  * read the [OmniBOLT](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec) spec. If you have any question over there, raise issues in that repo.
  * ask questions or talk about things in Issues.
  * make branches and raise pull-request, even if working on the main repository.
  * dont copy/past any code from somewhere else in contribution, because we have limited resource to compare source codes to avoid legal issues. What we can do is to read your code, run tests of your newly developed modules and read your comments in your branch to see if it is soving a real problem. 
  * better running go fmt before pushing any code.
  * add test to any package.
  * write/contribute light client testing tools, such as a HTML page supporting WebSocket, so that new programmers can have an intuitive experience to get started. That helps. We will release our tools for testing,  
 


# Current Features

* Generate user OBD(OmniBOLT Daemon) address.  
* Open Poon-Dryja Channel.
* Deposit, close.
* Commitment Transaction within a channel.


# Related projects: 

[https://github.com/OmniLayer/omniwallet](https://github.com/OmniLayer/omniwallet)

[https://github.com/OmniLayer/omnicore](https://github.com/OmniLayer/omnicore)

[https://github.com/OmniLayer/OmniJ](https://github.com/OmniLayer/OmniJ)

[https://github.com/OmniLayer/spec](https://github.com/OmniLayer/spec)

[https://github.com/LightningOnOmnilayer/Omni-BOLT-spec](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec)

[https://github.com/lightningnetwork/lightning-rfc](https://github.com/lightningnetwork/lightning-rfc)

[https://github.com/lightningnetwork/lnd](https://github.com/lightningnetwork/lnd)

[https://github.com/LightningOnOmnilayer/OmniWalletMobile](https://github.com/LightningOnOmnilayer/OmniWalletMobile)





 


