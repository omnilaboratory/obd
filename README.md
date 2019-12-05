# LightningOnOmni | OmniBOLT Daemon
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/LightningOnOmnilayer/LightningOnOmni/blob/master/LICENSE) [![](https://img.shields.io/badge/standard%20readme-OK-brightgreen)](https://github.com/LightningOnOmnilayer/LightningOnOmni/blob/master/README.md) [![](https://img.shields.io/badge/golang-%3E%3D1.9.0-orange)](https://golang.org/dl/) [![](https://img.shields.io/badge/protocol-OmniBOLT-brightgreen)](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec) 
[![](https://img.shields.io/badge/API%20V0.3-Document-blue)](https://api.omnilab.online) 

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

	<!-- Added by Kevin Zhang 2019-11-19 -->
	* [Step 6: transfer assets through HTLC](https://github.com/LightningOnOmnilayer/LightningOnOmni#step-6-transfer-assets-through-HTLC)

 * [API Document](https://github.com/LightningOnOmnilayer/LightningOnOmni#api-document)
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

Before compile `obdserver.go` file, you should be get all related packages by run `go get` command in terminal.
Example: `go get google.golang.org/grpc`

This is a full list for need to get packages.

```
google.golang.org/grpc
github.com/gin-gonic/gin
github.com/gorilla/websocket
github.com/satori/go.uuid
github.com/tidwall/gjson
google.golang.org/grpc
golang.org/x/net/context
github.com/shopspring/decimal
github.com/asdine/storm
github.com/asdine/storm/q
github.com/btcsuite/btcd/chaincfg
github.com/btcsuite/btcutil
golang.org/x/crypto/ripemd160
github.com/btcsuite/btcutil/base58
golang.org/x/crypto/salsa20
github.com/go-ini/ini
github.com/tyler-smith/go-bip39
github.com/tyler-smith/go-bip32
```

<br/>

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

**NOTE: You should replace all of relevant data by the exact value that your own OBD generates for you**

Since OmniBOLT deamon(OBD) exposes WebSocket services, we use web socket test client for Chrome to do experiments. Install it from:
```
https://chrome.google.com/webstore/detail/websocket-test-client/fgponpodhbmadfljofbimhhlengambbn?hl=en
```
Make sure your browser supports WebSocket, as displayed in this screenshot.

<p align="center">
  <img width="500" alt="Screenshot of Websocket online testing site" src="https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/imgs/WebSocketTestSite.png">
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


### Generate a wallet address

A wallet address is used for:

* Create channels
* Deposit to channels
* Withdraw from channels
* As a temp address for create commitment transactions


1、Generate Alice's address

Websocket request:

```
{
    "type":-200
}
```

OBD responses:

```json
{
    "type":-200,
    "status":true,
    "from":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
    "to":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
    "result":{
        "address":"mkb9Muc1Qjt2XT5onkqwf3NWAP6BerDwsG",
        "index":0,
        "private_key":"d33390f109115d130f8d1da429c9ac8bda85870c2bb4501959402c007a9a4f6f",
        "pub_key":"02cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba",
        "wif":"cUfFRyzqokaHuhEgwT85tbGpf7WpG8NN7N55nWjFQKFtwVbSNmA2"
    }
}
```

2、Generate Bob's address

Websocket request:

```
{
    "type":-200
}
```

OBD responses:

```json
{
    "type":-200,
    "status":true,
    "from":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be",
    "to":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be",
    "result":{
        "address":"n1ZEtzhL1HJk5zQbgbez2WWNnP1jiHbW9x",
        "index":0,
        "private_key":"13dcf0808819c529ce18c6139f1098ed78045d6cf486a2217f4b71561c2c589f",
        "pub_key":"025b4b752ed16203597a064189ecce30a1055c324cd5644fcc531e8f8f20737185",
        "wif":"cNFK6XSftD18er8a1VWZKx3BqLfxAMQwy3veoJXtdxptpbb7vrig"
    }
}
```

### Data of Alice, Bob, Channel

Alice's:

```
"user_id":      7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2
"address":      mkb9Muc1Qjt2XT5onkqwf3NWAP6BerDwsG
"private_key":  d33390f109115d130f8d1da429c9ac8bda85870c2bb4501959402c007a9a4f6f
"pub_key":"     02cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba
```

Bob's:

```
"user_id":      f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be
"address":      n1ZEtzhL1HJk5zQbgbez2WWNnP1jiHbW9x
"private_key":  13dcf0808819c529ce18c6139f1098ed78045d6cf486a2217f4b71561c2c589f
"pub_key":      025b4b752ed16203597a064189ecce30a1055c324cd5644fcc531e8f8f20737185
```

Channel's

```
"channel_address":   2NFSDbQx3xCunJXixqp9KLSvuK6k7VTmVNu
"channel_address_redeem_script":"522102cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba21025b4b752ed16203597a064189ecce30a1055c324cd5644fcc531e8f8f2073718552ae"
"channel_address_script_pub_key":"a914f367037f9c5774c06ecd61fb7ea75bf82739af6387"
```


### Create channel

Alice sends request to Bob for creating a channel between them:

*recipient_peer_id is Bob's user_id.*

Websocket request:

```json
{
    "type":-32,
    "data":{
        "funding_pubkey":"02cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba"
    },
    "recipient_peer_id":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be"
}
```

OBD creats the complete message for Alice and routes it to Bob:

```json
{
    "type":-32,
    "status":true,
    "from":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
    "to":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be",
    "result":{
        "chain_hash":"1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P",
        "channel_reserve_satoshis":0,
        "delayed_payment_base_point":"",
        "dust_limit_satoshis":0,
        "fee_rate_per_kw":0,
        "funding_address":"mkb9Muc1Qjt2XT5onkqwf3NWAP6BerDwsG",
        "funding_pubkey":"02cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba",
        "funding_satoshis":0,
        "htlc_base_point":"",
        "htlc_minimum_msat":0,
        "max_accepted_htlcs":0,
        "max_htlc_value_in_flight_msat":0,
        "payment_base_point":"",
        "push_msat":0,
        "revocation_base_point":"",
        "temporary_channel_id":[244,48,63,50,134,204,212,189,12,187,107,11,70,62,207,155,171,60,160,197,150,207,10,69,98,178,240,202,231,132,220,110],
        "to_self_delay":0
    }
}
```

In Bob's browser, he will see the message, and he accepts the request, by sending the following message back:

Websocket request:

*funding_pubkey is the pubkey Bob's wallet address*

```json
{
	"type":-33,
	"data":{
		"temporary_channel_id":[244,48,63,50,134,204,212,189,12,187,107,11,70,62,207,155,171,60,160,197,150,207,10,69,98,178,240,202,231,132,220,110],		"funding_pubkey":"025b4b752ed16203597a064189ecce30a1055c324cd5644fcc531e8f8f20737185",
		"approval":true
	}
}
```

OBD responses:

```json
{
    "type":-33,
    "status":true,
    "from":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be",
    "to":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
    "result":{
        "accept_at":"2019-11-29T16:09:26.983149+08:00",
        "address_a":"mkb9Muc1Qjt2XT5onkqwf3NWAP6BerDwsG",
        "address_b":"n1ZEtzhL1HJk5zQbgbez2WWNnP1jiHbW9x",
        "chain_hash":"1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P",
        "channel_address":"2NFSDbQx3xCunJXixqp9KLSvuK6k7VTmVNu",
        "channel_address_redeem_script":"522102cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba21025b4b752ed16203597a064189ecce30a1055c324cd5644fcc531e8f8f2073718552ae",
        "channel_address_script_pub_key":"a914f367037f9c5774c06ecd61fb7ea75bf82739af6387",
        "channel_id":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
        "channel_reserve_satoshis":0,
        "close_at":"0001-01-01T00:00:00Z",
        "create_at":"2019-11-29T16:04:46.970979+08:00",
        "create_by":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
        "curr_state":20,
        "delayed_payment_base_point":"",
        "dust_limit_satoshis":0,
        "fee_rate_per_kw":0,
        "funding_address":"mkb9Muc1Qjt2XT5onkqwf3NWAP6BerDwsG",
        "funding_pubkey":"02cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba",
        "funding_satoshis":0,
        "htlc_base_point":"",
        "htlc_minimum_msat":0,
        "id":3,
        "max_accepted_htlcs":0,
        "max_htlc_value_in_flight_msat":0,
        "payment_base_point":"",
        "peer_id_a":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
        "peer_id_b":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be",
        "pub_key_a":"02cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba",
        "pub_key_b":"025b4b752ed16203597a064189ecce30a1055c324cd5644fcc531e8f8f20737185",
        "push_msat":0,
        "revocation_base_point":"",
        "temporary_channel_id":[244,48,63,50,134,204,212,189,12,187,107,11,70,62,207,155,171,60,160,197,150,207,10,69,98,178,240,202,231,132,220,110],
        "to_self_delay":0
    }
}
```

When you test, you should replace the `temporary_channel_id` by the exact value that OBD generates for you.


### Deposit

**Deposit BTC for miner fee**

Alice tells her OBD to create a funding transaction. Since on Omnilayer/BTC, the miner fee is bitcoin, so that first we need to deposit small amount of bitcoin used in withdraw money, creating temporary multi-sig addresses for transactions in RSMC and HTLC.  

After that, we need to create the second part of transactions that deposit tokens which both sides agreed to fund the channel.  

This is the place that OBD differs from LND or other lightning implementation, since they are for BTC only, so there is no need for them to deposit extra bitcoin for fee.

*Another difference is that we need two transactions for the miner fees, one for the channel, and the other for a internal temporary multi-sig address. So Alice may raise request twice, and Bob needs to answer twice.*

*It is not finalized yet.*

Alice(`mkb9Muc1Qjt2XT5onkqwf3NWAP6BerDwsG`) deposits 0.001 btc to the channel (`2NFSDbQx3xCunJXixqp9KLSvuK6k7VTmVNu`), where fee is 0.00001, and then abtains the transaction ID:

**request**
```json
{
    "type":1009,
    "data":{
    "fromBitCoinAddress":"mkb9Muc1Qjt2XT5onkqwf3NWAP6BerDwsG",
    "fromBitCoinAddressPrivKey":"d33390f109115d130f8d1da429c9ac8bda85870c2bb4501959402c007a9a4f6f",
    "toBitCoinAddress":"2NFSDbQx3xCunJXixqp9KLSvuK6k7VTmVNu",
    "amount":0.001,
    "minerFee":0.00001,
    }
}
```

OBD responses with the transaction script and other payloads:

```json
{
    "type": 1009, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "hex": "0200000002634ad0a2468850f4bed537ffc2a28aa6395cb2c34efe54b321135bae298d5d79020000006a4730440220421bc0e7cbeaebb5ad7e0559d5be01f50143816243904bed8c4fe2972717ad0b02206bf8b4f3dc911b6e8c6748b248e6e59dd4b43e4b5a197379863040f374a979a10121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffffe2fa86ac404f71b97e6b34010549a82c9f3ee1a150c54b7daec99c0da83f8482010000006a47304402207d52f8a36791ce26361809b9a7b9da4175bf3b5ad6004d6a46ecb4a329dfdb370220217c481b71bcafbf3e93fe72837c28e48f38474ea1341bae030867e37be5c08c0121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff02102700000000000017a914f64403be27af8af0a8abc21aed584b06f80adf30876a190f00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac00000000", 
        "txid": "b18aab6599f1661963763281c83ddd7f6de51813881b2ee563008c021d31fcd4"
    }
}
```

**Deposit Omni Assets**

Alice starts to deposit omni assets to the channel. This is quite similar to the the btc funding procedure.

**request**

```json
{
	"type":2001,
	"data":{
        "from_address":"muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit",			
        "from_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8",
        "to_address":"2NFhMhDJT9TsnBCG6L2amH3eDXxgwW6EJh7",
        "amount":50,
        "property_id": 121
	}
}
```

**OBD responses:**

```json
{
    "type":2001,
    "status":true,
    "from":"<user_id>",
    "to":"<user_id>",
    "result":{
        "hex":"0200000001e47f333e4a377c9877ee8aedccca476dd5a6bf7aae2116923c937ebfbd173df1010000006a47304402205d721941d28ec7a6a427d0da51cf89e70772548d0829b627919a3ebf8722e96a02207658e909db233dfa6d4ca4f1a3a08fc0325c045c519d2e0e8f6af56ea6e334f00121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff03a6b50e00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac0000000000000000166a146f6d6e690000000000000079000000012a05f2001c0200000000000017a914f64403be27af8af0a8abc21aed584b06f80adf308700000000"
    }
}
```

Then Alice tells Bob she created a deposit transaction:

```json
{
	"type":-34,
	"data":{
		"temporary_channel_id":[68,9,34,176,221,163,195,216,120,239,152,94,138,101,252,83,99,125,195,221,146,3,0,128,166,224,203,99,101,48,20,164],		"funding_tx_hex":"0200000001e47f333e4a377c9877ee8aedccca476dd5a6bf7aae2116923c937ebfbd173df1010000006a47304402205d721941d28ec7a6a427d0da51cf89e70772548d0829b627919a3ebf8722e96a02207658e909db233dfa6d4ca4f1a3a08fc0325c045c519d2e0e8f6af56ea6e334f00121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff03a6b50e00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac0000000000000000166a146f6d6e690000000000000079000000012a05f2001c0200000000000017a914f64403be27af8af0a8abc21aed584b06f80adf308700000000",	
        "temp_address_pub_key":"0380874d124f259b31ee8cf3256d784f0269ae9cf3b577e5c271c452572f8b28e5","temp_address_private_key":"cVBFoaRumDJYntRRV244KUj7kyrGauGhT6bZcf15xfhGCh9mAbVp",		"channel_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8"
	}
}
```

Bob replies he knows this deposit, and then OBD creates commitment transaction, revockable delivery transaction and breach remedy transaction:

```json
{
	"type":-35,
	"data":{
		"channel_id":[174,36,154,103,145,76,58,237,32,61,201,81,17,156,135,216,66,28,83,203,251,152,138,102,158,113,131,32,241,229,43,75],		"fundee_channel_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",
		"approval":true
	}
}
```

OBD sends a message to both Alice and Bob, reporting the status(true or false) of all the internal transactions:

```json
{
    "type":-35,
    "status":true,
    "from":"<user_id>",
    "to":"<user_id>",
    "result":{
        "amount_a":50,
        "amount_b":0,
        "channel_id":[174,36,154,103,145,76,58,237,32,61,201,81,17,156,135,216,66,28,83,203,251,152,138,102,158,113,131,32,241,229,43,75],
        "channel_info_id":1,
        "create_at":"2019-11-04T13:17:15.1972409+08:00",
        "create_by":"<user_id>",
        "curr_state":20,
        "fundee_sign_at":"2019-11-04T13:18:19.6544966+08:00",
        "funder_address":"muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit",
        "funder_pub_key_2_for_commitment":"0380874d124f259b31ee8cf3256d784f0269ae9cf3b577e5c271c452572f8b28e5",
        "funding_output_index":2,
        "funding_tx_hex":"0200000001e47f333e4a377c9877ee8aedccca476dd5a6bf7aae2116923c937ebfbd173df1010000006a47304402205d721941d28ec7a6a427d0da51cf89e70772548d0829b627919a3ebf8722e96a02207658e909db233dfa6d4ca4f1a3a08fc0325c045c519d2e0e8f6af56ea6e334f00121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff03a6b50e00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac0000000000000000166a146f6d6e690000000000000079000000012a05f2001c0200000000000017a914f64403be27af8af0a8abc21aed584b06f80adf308700000000",
        "funding_txid":"492be5f12083719e668a98fbcb531c42d8879c1151c93d20ed3a4c91679a24ae",
        "id":1,
        "peer_id_a":"<user_id>",
        "peer_id_b":"<user_id>",
        "property_id":121
    }
}
```

### Payments in a channel

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

```json
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

```json
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

<!-- Added by Kevin Zhang 2019-11-19 -->

## Step 6: Transfer assets through HTLC

HTLC(Hashed Timelock Contract) is the second foundamental module in lightning network. In general, any two clients need no direct channels between them for transferring tokens or exchanging information, they can use their direct channels to build a bridge for this purpose. HTLC is designed for chaining the channels together in delivering messages from one client to another.

`
[Alice --(10 USDT)--> Bob] ==(Bob has two channels)== [Bob --(10 USDT)--> Carol] ==(Carol has two channels)== [Carol --(10 USDT)--> David]
`  

**[A B]** stands for the channel built by A and B

A formal HTL contract describes the following procedure:

If Bob can tell Alice the secret R, which is the pre-image of <code>Hash(R)</code> that some one else (Carol) in the chain shared with Bob 3 days ago in exchange of 10 USDT in the channel <code>[Bob Carol]</code>, then Bob will get the 10 USDT fund inside the channel <code>[Alice Bob]</code>, otherwise Alice gets her 10 USDT back.

Readers shall find the latest specification in [OmniBOLT 04: HTLC and Payment Routing](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/OmniBOLT-04-HTLC-and-Payment-Routing.md)

<br/>

### Prepare Data for Client

Payments via HTLC involve many clients and channels among them, in order for demonstration the complete process, we simply setup three clients: Alice, Bob and Carol, and assume only Bob established channel with other clients. OBD has its backend routing algorithm to find the right path for a payment, but it is not exposed to developers.

**Alice's data:**

```
adderss: muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit
pubkey:  029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3c
privkey: cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8
```

**Bob's data:**

```
adderss: mtSJixJ8eCguXDAdkGGoQu3nG1n77a6td8
pubkey:  03da1b78a5ab4a5e4e13515e5104dbfc1d2d349d89039c15eda9c0118b7edaa91f
privkey: cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt
```

**Carol's data:**

```
adderss: mgoiBkppoJMc8cC8XRYNvFEjath5DrKqj8
pubkey:  034094927aa69a96d82d7e67146cf9b8dcd775919d1373d5319454e6004c0cdf7a
privkey: cMxR8h9z5oKrdyuXVR9uzBbyyaJz1karxH1FW5xezhKzxQc7sCJV
```

<br/>

### Prepare Data for Channel Address

For testing, we generate two multisig addresses.

*A 2-2 Multisig address is used to save the assets of both participant,* 
*similar to a safe box, it takes 2 people's keys to open it.*

*A 2-2 Multisig address is created by 2 pubkeys of bitcoin addresses. Example:*

```json
./omnicore-cli createmultisig 2 '["026337bd8737618b61816c94fb2786d5a386d56cdb7ab68ceb4eafe6fb28452525","0286877b505ddda65fb60b8f4f5d584a5ad41158c2eaedb93fca881efadee315be"]'

Response：
{
  "address": "2MvL76hcC7zf41Z5NLRNxAFxtk28RMYstQk",
  "redeemScript": "5221026337bd8737618b61816c94fb2786d5a386d56cdb7ab68ceb4eafe6fb28452525210286877b505ddda65fb60b8f4f5d584a5ad41158c2eaedb93fca881efadee315be52ae"
}
```

**Channel [Alice Bob]:**

```
channel_address：2NFhMhDJT9TsnBCG6L2amH3eDXxgwW6EJh7
redeem_script：5221029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3c2103da1b78a5ab4a5e4e13515e5104dbfc1d2d349d89039c15eda9c0118b7edaa91f52ae
channelId: [174,36,154,103,145,76,58,237,32,61,201,81,17,156,135,216,66,28,83,203,251,152,138,102,158,113,131,32,241,229,43,75]
```

**Channel [Bob Carol]:**

```
channel_address：2MzQW254vB6mHsUvLHxCnKZ73Gcw7kSrvsd
redeem_script：522103da1b78a5ab4a5e4e13515e5104dbfc1d2d349d89039c15eda9c0118b7edaa91f21034094927aa69a96d82d7e67146cf9b8dcd775919d1373d5319454e6004c0cdf7a52ae
channel:
[223,177,75,185,186,22,47,155,145,238,242,1,158,247,192,1,48,183,197,192,190,72,49,233,62,65,156,103,111,172,109,51]
```

<br/>

### Clients sign up and login for HTLC testing

Three clients sign up.

**Alice sign up:**

```json
{
	"type":101
}
```

**Bob sign up:**

```json
{
	"type":101
}
```

**Carol sign up:**

```json
{
	"type":101
}
```

Three clients login.

**Alice login:**

```json
{
    "type":1,
    "data":{
        "mnemonic":"<Your own OBD generates for you>"
    }
}
```

**Bob login:**

```json
{
    "type":1,
    "data":{
        "mnemonic":"<Your own OBD generates for you>"
    }
}
```

**Carol login:**

```json
{
    "type":1,
    "data":{
        "mnemonic":"<Your own OBD generates for you>"
    }
}
```

<br/>

### Open Channel between Alice and Bob（A2B）

Alice sends request to her OBD instance, her OBD helps her complete the message, and routes her request to Bob's OBD for creating a channel between them. 

**Alice send the request:**

```json
{
	"type":-32,
    "data":{
        "funding_pubkey":"029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3c"
    },
    "recipient_peer_id":"<Bob's user_id>"
}
```

**OBD Responses:**

```json
{
    "type":-32,
    "status":true,
    "from":"7da8d2441e0ad67040a274902f1965ee1a5c3fdd86f1ddc3280eda5230e006f2",
    "to":"f38e72f6bf69c69ad1cdc0040550bafb86d5c4d35bd04542fcf5fc5ecb2135be",
    "result":{
        "chain_hash":"1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P",
        "channel_reserve_satoshis":0,
        "delayed_payment_base_point":"",
        "dust_limit_satoshis":0,
        "fee_rate_per_kw":0,
        "funding_address":"mkb9Muc1Qjt2XT5onkqwf3NWAP6BerDwsG",
        "funding_pubkey":"02cba1d7a57cdfe83fd8ee4d868fb4b926ab893fd2d8df7706e6e0829c60e923ba",
        "funding_satoshis":0,
        "htlc_base_point":"",
        "htlc_minimum_msat":0,
        "max_accepted_htlcs":0,
        "max_htlc_value_in_flight_msat":0,
        "payment_base_point":"",
        "push_msat":0,
        "revocation_base_point":"",
        "temporary_channel_id":[244,48,63,50,134,204,212,189,12,187,107,11,70,62,207,155,171,60,160,197,150,207,10,69,98,178,240,202,231,132,220,110],
        "to_self_delay":0
    }
}
```

**Bob replies:**

```json
{
	"type":-33,
	"data":{
		"temporary_channel_id":[68,9,34,176,221,163,195,216,120,239,152,94,138,101,252,83,99,125,195,221,146,3,0,128,166,224,203,99,101,48,20,164],		"funding_pubkey":"03da1b78a5ab4a5e4e13515e5104dbfc1d2d349d89039c15eda9c0118b7edaa91f",
		"approval":true
	}
}
```

**OBD Responses:**

```json
{
    "type": -33, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "accept_at": "2019-11-04T10:59:51.1997943+08:00", 
        "address_a": "muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit", 
        "address_b": "mtSJixJ8eCguXDAdkGGoQu3nG1n77a6td8", 
        "chain_hash": "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P", 
        "channel_address": "2NFhMhDJT9TsnBCG6L2amH3eDXxgwW6EJh7", 
        "channel_address_redeem_script": "5221029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3c2103da1b78a5ab4a5e4e13515e5104dbfc1d2d349d89039c15eda9c0118b7edaa91f52ae", 
        "channel_address_script_pub_key": "a914f64403be27af8af0a8abc21aed584b06f80adf3087", 
        "channel_id": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
        "channel_reserve_satoshis": 0, 
        "close_at": "0001-01-01T00:00:00Z", 
        "create_at": "2019-11-04T10:58:08.2357582+08:00", 
        "create_by": "alice", 
        "curr_state": 20, 
        "delayed_payment_base_point": "", 
        "dust_limit_satoshis": 0, 
        "fee_rate_per_kw": 0, 
        "funding_address": "muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit", 
        "funding_pubkey": "029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3c", 
        "funding_satoshis": 0, 
        "htlc_base_point": "", 
        "htlc_minimum_msat": 0, 
        "id": 1, 
        "max_accepted_htlcs": 0, 
        "max_htlc_value_in_flight_msat": 0, 
        "payment_base_point": "", 
        "peer_id_a": "alice", 
        "peer_id_b": "bob", 
        "pub_key_a": "029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3c", 
        "pub_key_b": "03da1b78a5ab4a5e4e13515e5104dbfc1d2d349d89039c15eda9c0118b7edaa91f", 
        "push_msat": 0, 
        "revocation_base_point": "", 
        "temporary_channel_id": [68,9,34,176,221,163,195,216,120,239,152,94,138,101,252,83,99,125,195,221,146,3,0,128,166,224,203,99,101,48,20,164],
        "to_self_delay": 0
    }
}
```

<br/>

### Open Channel between Bob and Carol（B2C）

Bob sends request to his OBD instance, his OBD helps he complete the message, 
and routes his request to Carol's OBD for creating a channel between them. 

**Bob send the request:**

```json
{
	"type":-32,
    "data":{
        "funding_pubkey":"03da1b78a5ab4a5e4e13515e5104dbfc1d2d349d89039c15eda9c0118b7edaa91f"
    },
    "recipient_peer_id":"<Carol's user_id>"
}
```

**OBD Responses:**

Similar json data structure above.

**Carol replies:**

```json
{
	"type":-33,
	"data":{
		"temporary_channel_id":[26,200,121,127,242,0,84,191,162,35,118,90,99,71,229,123,238,190,22,226,54,211,38,113,229,165,241,132,153,48,99,86],		"funding_pubkey":"034094927aa69a96d82d7e67146cf9b8dcd775919d1373d5319454e6004c0cdf7a",
		"approval":true
	}
}
```

**OBD Responses:**

Similar json data structure above.

<br/>

### Alice deposit BTC to channel A2B for miner fee

Need three times deposit, every time can be 0.0001 btc for miner fee.

*Because currently Omni Core does not support one-to-many transfer, when creating*
*Cxa or Cxb commitment transactions, there are three outputs: 1) directly address*
*2) RSMC address  3) HTLC address. So, we need to construct three raw omni transactions.*
*This is why we need three times deposit BTC to channel for miner fee.*

**Alice send BTC to channel by 1009:**

```json
{
    "type":1009,
    "data":{
        "from_address":"muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit",
        "from_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8",
        "to_address":"2NFhMhDJT9TsnBCG6L2amH3eDXxgwW6EJh7",
        "amount":0.0001,
        "miner_fee":0.00001
    }
}
```

**OBD Responses:**

```json
{
    "type": 1009, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "hex": "0200000002634ad0a2468850f4bed537ffc2a28aa6395cb2c34efe54b321135bae298d5d79020000006a4730440220421bc0e7cbeaebb5ad7e0559d5be01f50143816243904bed8c4fe2972717ad0b02206bf8b4f3dc911b6e8c6748b248e6e59dd4b43e4b5a197379863040f374a979a10121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffffe2fa86ac404f71b97e6b34010549a82c9f3ee1a150c54b7daec99c0da83f8482010000006a47304402207d52f8a36791ce26361809b9a7b9da4175bf3b5ad6004d6a46ecb4a329dfdb370220217c481b71bcafbf3e93fe72837c28e48f38474ea1341bae030867e37be5c08c0121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff02102700000000000017a914f64403be27af8af0a8abc21aed584b06f80adf30876a190f00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac00000000", 
        "txid": "b18aab6599f1661963763281c83ddd7f6de51813881b2ee563008c021d31fcd4"
    }
}
```

**Alice send a notification to bob:**

```json
{
	"type":-3400,
	"data":{
		"temporary_channel_id":[68,9,34,176,221,163,195,216,120,239,152,94,138,101,252,83,99,125,195,221,146,3,0,128,166,224,203,99,101,48,20,164],		"channel_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8",		"funding_tx_hex":"0200000002634ad0a2468850f4bed537ffc2a28aa6395cb2c34efe54b321135bae298d5d79020000006a4730440220421bc0e7cbeaebb5ad7e0559d5be01f50143816243904bed8c4fe2972717ad0b02206bf8b4f3dc911b6e8c6748b248e6e59dd4b43e4b5a197379863040f374a979a10121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffffe2fa86ac404f71b97e6b34010549a82c9f3ee1a150c54b7daec99c0da83f8482010000006a47304402207d52f8a36791ce26361809b9a7b9da4175bf3b5ad6004d6a46ecb4a329dfdb370220217c481b71bcafbf3e93fe72837c28e48f38474ea1341bae030867e37be5c08c0121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff02102700000000000017a914f64403be27af8af0a8abc21aed584b06f80adf30876a190f00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac00000000"
	}
}
```

**OBD Responses:**

```json
{
    "type": -3400, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "amount": 0.0001, 
        "funding_txid": "b18aab6599f1661963763281c83ddd7f6de51813881b2ee563008c021d31fcd4", 
        "temporary_channel_id": [68,9,34,176,221,163,195,216,120,239,152,94,138,101,252,83,99,125,195,221,146,3,0,128,166,224,203,99,101,48,20,164]
    }
}
```

**Bob replies:**

```json
{
	"type":-3500,
	"data":{
		"temporary_channel_id":[68,9,34,176,221,163,195,216,120,239,152,94,138,101,252,83,99,125,195,221,146,3,0,128,166,224,203,99,101,48,20,164],		"channel_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",		"funding_txid":"b18aab6599f1661963763281c83ddd7f6de51813881b2ee563008c021d31fcd4",
		"approval":true
	}
}
```

**OBD Responses:**

```json
{
    "type": -3500, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "channel_id": [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
        "create_at": "2019-11-04T13:05:52.1148727+08:00", 
        "id": 1, 
        "owner": "alice", 
        "temporary_channel_id": [68,9,34,176,221,163,195,216,120,239,152,94,138,101,252,83,99,125,195,221,146,3,0,128,166,224,203,99,101,48,20,164],
        "tx_hash": "0200000001d4fc311d028c0063e52e1b881318e56d7fdd3dc8813276631966f19965ab8ab100000000d9004730440220167874a0697aeebb170adfd418cbab33a39a837099be4a829d8c71d4a1933e0c0220145afac21e206ed8b37e38aaa13788b91dc7819bc2caeb760c5b13fb62a34d820147304402201fcc09eff15e178704bcda267bcc221fa677ff3f9c0f04139a458696b121ef780220502e7ff9f919d4fe165f0b4abf39000779c7c49b47064b3c752db2668a81750601475221029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3c2103da1b78a5ab4a5e4e13515e5104dbfc1d2d349d89039c15eda9c0118b7edaa91f52aeffffffff01581b0000000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac00000000", 
        "txid": "17405cc66b345247ded162a21ecfbd98a6e9b85a6d2dfba32b2421b9670efe4f"
    }
}
```

<br/>

### Alice deposit USDT to channel A2B for transferring

Deposit USDT for transfer to Bob. This USDT is asset Alice would like to transfer to Carol.

*For testing purpose, we issued a testing asset on Omni Layer.*
*The asset name is OST-P1-Test, property id is 121.*

**Alice send:**

```json
{
	"type":2001,
	"data":{
        "from_address":"muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit",			
        "from_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8",
        "to_address":"2NFhMhDJT9TsnBCG6L2amH3eDXxgwW6EJh7",
        "amount":50,
        "property_id": 121
	}
}
```

**OBD Responses:**

```json
{
    "type": 2001, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "hex": "0200000001e47f333e4a377c9877ee8aedccca476dd5a6bf7aae2116923c937ebfbd173df1010000006a47304402205d721941d28ec7a6a427d0da51cf89e70772548d0829b627919a3ebf8722e96a02207658e909db233dfa6d4ca4f1a3a08fc0325c045c519d2e0e8f6af56ea6e334f00121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff03a6b50e00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac0000000000000000166a146f6d6e690000000000000079000000012a05f2001c0200000000000017a914f64403be27af8af0a8abc21aed584b06f80adf308700000000"
    }
}
```

<br/>

*At this step, in order to create C1a commitment transaction, we need a Alice1's data*
*for creating a multisig address by Alice1's pubkey and Bob's pubkey.*

**Alice1's temp address data:**

```shell
address:  mq8t9iEHoYzw4EzByo9p1uCVBWDnFU4JwW
privkey:  cVBFoaRumDJYntRRV244KUj7kyrGauGhT6bZcf15xfhGCh9mAbVp
pubkey:   0380874d124f259b31ee8cf3256d784f0269ae9cf3b577e5c271c452572f8b28e5
```

**Alice send a notification to bob:**

```json
{
    "type":-34,
	"data":{
		"temporary_channel_id":[68,9,34,176,221,163,195,216,120,239,152,94,138,101,252,83,99,125,195,221,146,3,0,128,166,224,203,99,101,48,20,164],		"funding_tx_hex":"0200000001e47f333e4a377c9877ee8aedccca476dd5a6bf7aae2116923c937ebfbd173df1010000006a47304402205d721941d28ec7a6a427d0da51cf89e70772548d0829b627919a3ebf8722e96a02207658e909db233dfa6d4ca4f1a3a08fc0325c045c519d2e0e8f6af56ea6e334f00121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff03a6b50e00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac0000000000000000166a146f6d6e690000000000000079000000012a05f2001c0200000000000017a914f64403be27af8af0a8abc21aed584b06f80adf308700000000",	
		"temp_address_pub_key":"0380874d124f259b31ee8cf3256d784f0269ae9cf3b577e5c271c452572f8b28e5",		"temp_address_private_key":"cVBFoaRumDJYntRRV244KUj7kyrGauGhT6bZcf15xfhGCh9mAbVp",		"channel_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8"
	}
}
```

**OBD Responses:**

```json
{
    "type": -34, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "amount_a": 50, 
        "amount_b": 0, 
        "channel_id": [174,36,154,103,145,76,58,237,32,61,201,81,17,156,135,216,66,28,83,203,251,152,138,102,158,113,131,32,241,229,43,75],
        "channel_info_id": 1, 
        "create_at": "2019-11-04T13:17:15.1972409+08:00", 
        "create_by": "alice", 
        "curr_state": 10, 
        "fundee_sign_at": "0001-01-01T00:00:00Z", 
        "funder_address": "muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit", 
        "funder_pub_key_2_for_commitment": "0380874d124f259b31ee8cf3256d784f0269ae9cf3b577e5c271c452572f8b28e5", 
        "funding_output_index": 2, 
        "funding_tx_hex": "0200000001e47f333e4a377c9877ee8aedccca476dd5a6bf7aae2116923c937ebfbd173df1010000006a47304402205d721941d28ec7a6a427d0da51cf89e70772548d0829b627919a3ebf8722e96a02207658e909db233dfa6d4ca4f1a3a08fc0325c045c519d2e0e8f6af56ea6e334f00121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff03a6b50e00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac0000000000000000166a146f6d6e690000000000000079000000012a05f2001c0200000000000017a914f64403be27af8af0a8abc21aed584b06f80adf308700000000", 
        "funding_txid": "492be5f12083719e668a98fbcb531c42d8879c1151c93d20ed3a4c91679a24ae", 
        "id": 1, 
        "peer_id_a": "alice", 
        "peer_id_b": "bob", 
        "property_id": 121
    }
}
```

**Bob replies:**

```json
{
	"type":-35,
	"data":{
		"channel_id":[174,36,154,103,145,76,58,237,32,61,201,81,17,156,135,216,66,28,83,203,251,152,138,102,158,113,131,32,241,229,43,75],		"fundee_channel_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",
		"approval":true
	}
}
```

**OBD Responses:**

```json
{
    "type": -35, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "amount_a": 50, 
        "amount_b": 0, 
        "channel_id": [174,36,154,103,145,76,58,237,32,61,201,81,17,156,135,216,66,28,83,203,251,152,138,102,158,113,131,32,241,229,43,75],
        "channel_info_id": 1, 
        "create_at": "2019-11-04T13:17:15.1972409+08:00", 
        "create_by": "alice", 
        "curr_state": 20, 
        "fundee_sign_at": "2019-11-04T13:18:19.6544966+08:00", 
        "funder_address": "muYrqVWTKnkaVAMuqn59Ta6GL912ixpxit", 
        "funder_pub_key_2_for_commitment": "0380874d124f259b31ee8cf3256d784f0269ae9cf3b577e5c271c452572f8b28e5", 
        "funding_output_index": 2, 
        "funding_tx_hex": "0200000001e47f333e4a377c9877ee8aedccca476dd5a6bf7aae2116923c937ebfbd173df1010000006a47304402205d721941d28ec7a6a427d0da51cf89e70772548d0829b627919a3ebf8722e96a02207658e909db233dfa6d4ca4f1a3a08fc0325c045c519d2e0e8f6af56ea6e334f00121029cf4b150da0065d5c08bf088e8a5367d35ff72e4e79b39efb401530d19fa3f3cffffffff03a6b50e00000000001976a91499ee096d15bc8ae4ac8856a918ff2c4572877fa488ac0000000000000000166a146f6d6e690000000000000079000000012a05f2001c0200000000000017a914f64403be27af8af0a8abc21aed584b06f80adf308700000000", 
        "funding_txid": "492be5f12083719e668a98fbcb531c42d8879c1151c93d20ed3a4c91679a24ae", 
        "id": 1, 
        "peer_id_a": "alice", 
        "peer_id_b": "bob", 
        "property_id": 121
    }
}
```

<br/>

### Bob deposit BTC to channel B2C for miner fee

Same workflow above.
Need three times deposit, every time can be 0.0001 btc for miner fee.

### Bob deposit USDT to channel B2C for transferring

Same workflow above.
Deposit USDT for transfer to Carol. This USDT is asset Alice would like to transfer to Carol.

<br/>


### Launch a HTLC

We will launch a HTLC (Hashed Timelock Contract) transfer for testing purpose. 
It tests Alice transfer assets to Carol through Bob (middleman).

a) There IS NOT a direct channel between Alice and Carol.

b) There is a direct channel between Alice and Bob.

c) There is a direct channel between Bob and Carol.

**Alice launch a request to transfer to Carol:**

```json
{
	"type": -40,
	"data": {
		"property_id": 121,
		"amount": 5,
		"recipient_peer_id": "<Carol's user_id>"
	}
}
```

**OBD Responses:**

```json
{
    "type": -40, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "amount": 5, 
        "approval": false, 
        "property_id": 121, 
        "request_hash": "742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636"
    }
}
```

Carol receieved request from Alice, and Carol generate a preimage R as a secret.

Carol send H (Hash of the preimage R) to Alice. Alice will looking for a middleman
that can transfer assets directly to Carol. Alice will tell the middleman if can 
produce to Alice an unknown 20-byte random input data R from a known hash H, 
within three days, then Alice will settle the contract by paying the middleman assets.

**Carol send H (Hash_Preimage_R) to Alice**

```json
{
	"type": -41,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
		"property_id": 121,
		"amount": 5,
		"approval":true
	}
}
```

**OBD Responses:**

```json
{
    "type": -41, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "approval": true, 
        "h": "83519233492eb05ddd547757f2c3d151ad9392b2ebf48fc1a88e07e61dd82a45", 
        "id": 1, 
        "request_hash": "742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636"
    }
}
```

<br/>

### Alice looking for a path to transfer to Carol and Send H to Bob

Alice will looking for a path to reach Carol by pathfinding algorithm.
For testing purpose now, there is three client Alice, Bob, Carol, and 
they are all connected to an OBD server. 

The real situation should be that a network of many OBD servers is 
spread all over the world, a large number of clients are connected 
to these OBDs, and then we need to find the shortest path to transfer 
assets from one client to another.

<br/>

**Alice found a path and launch a request to middleman Bob**

**Alice send H to Bob**

```json
{
	"type": -42,
	"data": {
		"h":"83519233492eb05ddd547757f2c3d151ad9392b2ebf48fc1a88e07e61dd82a45"
	}
}
```

**OBD Responses:**

```json
{
    "type": -42, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "h": "83519233492eb05ddd547757f2c3d151ad9392b2ebf48fc1a88e07e61dd82a45", 
        "request_hash": "742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636"
    }
}
```

<br/>

Bob (middleman) receieved request from Alice and agree.

**Bob's temp address data for create HTLC commitment transactions:**

```shell
Bob RSMC
address:  mqtffoYNN7J3pZnkeVgoXfpQ1pz11ZuyX2
privkey:  cN3o7Se2qcSq7Z3wYL9DLxn1cr5Du5rmmxsQeoeF6LC7yXcZZotN
pubkey:   03c5d2756dea6d6259080d7e1ab3f8597e7e9a83b5667eff70ea49ca3addb6f293

Bob HTLC
address:  mnznzruCDWz3QQUCw4wvC3NoNup2kdTdkU
privkey:  cPRy5pB8Ek2DabfQ74x8giqDdPtwTvptgRnq8qEP7KduzdmMFmJM
pubkey:   03d2edfe1f0a527f70473dbacb386e4e6a9cc0ea0cabf71f6c0a3dd516a8e6099f
```

**Bob replies:**

```json
{
	"type": -44,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
		"approval":true,
		"channel_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",
		"last_temp_address_private_key":"cURFDvYsF49hazcGK3i4344H1r3pSjHwdqL2yQ85qxdzpub3rozx",
		"curr_rsmc_temp_address_pub_key":"03c5d2756dea6d6259080d7e1ab3f8597e7e9a83b5667eff70ea49ca3addb6f293",
		"curr_rsmc_temp_address_private_key":"cN3o7Se2qcSq7Z3wYL9DLxn1cr5Du5rmmxsQeoeF6LC7yXcZZotN",
		"curr_htlc_temp_address_pub_key":"03d2edfe1f0a527f70473dbacb386e4e6a9cc0ea0cabf71f6c0a3dd516a8e6099f",
		"curr_htlc_temp_address_private_key":"cPRy5pB8Ek2DabfQ74x8giqDdPtwTvptgRnq8qEP7KduzdmMFmJM"
	}
}

```

**OBD Responses:**

```json
{
    "type": -44, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "approval": true, 
        "request_hash": "742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636"
    }
}
```

<br/>

Bob agree request, Alice create related commitment transactions between Alice and Bob.

**Alice's temp address data for create HTLC commitment transactions:**

```shell
Alice RSMC
address: mmvTaVWx9EtwRHMkmhLbZF2JBGZc12ym2o
privkey: cSYJ3vwcgMPDqegXFJ2YgCYNNgKS9tNxCaRkZnn3ourQSdGNkJCk
pubkey:  03dab6d7b005e8b15a2dc8d7005b45111876813c24a54ff15316a76ba376cf020f

Alice HTLC
address: mtj8ChNcwkJi3ktB4apPTakknpDdiErTDX
privkey: cR7wXNwPjMrDCpnJoinTiMK384YyKNTfyctLQ2CCdQobdanEqgAs
pubkey:  03d16de84b72460055b18e6d572b49c4ab0e1d889c0bf0705becb22e16b65ca916

Alice HTna
address: mvYRwC7zTVhxNWeEgnUrdazMERvbP2yZpP
privkey: cNzNyejXtgC4ySXCVXqa6egVinYLDtGhkRkGd271ZW6AJrVKYZ2w
pubkey:  030cb3034f7374d5bb614e27169df99d346748a1a7a365a27b1138f5db7ad2b0f3
```

**Alice create HTLC commitment transactions:**

```json
{
	"type": -45,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
		"channel_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8",
		"last_temp_address_private_key":"cVsKRbL4ijWULmbU78nghKYL79GLYo7q9ccgmSR5c6zJWKfEEdJN",
		"curr_rsmc_temp_address_pub_key":"03dab6d7b005e8b15a2dc8d7005b45111876813c24a54ff15316a76ba376cf020f",
		"curr_rsmc_temp_address_private_key":"cSYJ3vwcgMPDqegXFJ2YgCYNNgKS9tNxCaRkZnn3ourQSdGNkJCk",
		"curr_htlc_temp_address_pub_key":"03d16de84b72460055b18e6d572b49c4ab0e1d889c0bf0705becb22e16b65ca916",
		"curr_htlc_temp_address_private_key":"cR7wXNwPjMrDCpnJoinTiMK384YyKNTfyctLQ2CCdQobdanEqgAs",
        "curr_htlc_temp_address_for_ht1a_pub_key":"030cb3034f7374d5bb614e27169df99d346748a1a7a365a27b1138f5db7ad2b0f3",
		"curr_htlc_temp_address_for_ht1a_private_key":"cNzNyejXtgC4ySXCVXqa6egVinYLDtGhkRkGd271ZW6AJrVKYZ2w"
	}
}
```

**OBD Responses:**

```json
{
    "type": -45, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "h": "e7626f2b7207006d6515399c587c09c3bfb5ed3b12f63c12b0d40e634f9dd9a3", 
        "h_and_r_info_request_hash": "1fe82bc9152741670c4ee2b4853df9346c1cc63fce6d1c896e7eeca8cc62c9d9"
    }
}
```

<br/>

### Bob Send H to Carol through the Path

Setp 2 of the path, Bob (middleman) has got the H and send it to Carol. 

That's mean if Carol can produce to Bob an unknown 20-byte random input 
data R from a known hash H, within two days, then Bob will settle the 
contract by paying Carol assets.

<br/>

**Bob (middleman) Send H to Carol (destination):**

```json
{
	"type": -43,
	"data": {
		"h":"83519233492eb05ddd547757f2c3d151ad9392b2ebf48fc1a88e07e61dd82a45",
        "h_and_r_info_request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636"
	}
}
```

**OBD Responses:**

```json
{
    "type": -43, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "h": "e7626f2b7207006d6515399c587c09c3bfb5ed3b12f63c12b0d40e634f9dd9a3", 
        "request_hash": "1fe82bc9152741670c4ee2b4853df9346c1cc63fce6d1c896e7eeca8cc62c9d9"
    }
}
```

<br/>

Carol (destination) receieved request from Bob and agree.

**Carol's temp address data for create HTLC commitment transactions:**

```shell
Carol RSMC
address: mwFhZchMtq4y9jSvmkAaoyAve1vZ4gfCvB
privkey: cTTFKJ3N8W4qHcpwj19NVVaEYAXBxf8DmBj9g7owLTWQ3mXXfD51
pubkey:  03dd26ec67e15bde83b527be45a1c64f420821ba78ebc5eb9d3fe1a8ae3cd1f6d9

Carol HTLC
address: mmsNgwiBhJLcv7Pup7gvsNKg82ECcLMjdF
privkey: cVW35aEjd56ZFHTSDtq9iUzHAfH6rWGVJcmYBn9QEp9ygSfkgZeG
pubkey:  03e96c6692bef50af7c6c777ff1bd65b1134d18c98be801e00f8e6247db65950b8
```

**Carol replies:**

```json
{
	"type": -44,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
		"approval":true,
		"channel_address_private_key":"cMxR8h9z5oKrdyuXVR9uzBbyyaJz1karxH1FW5xezhKzxQc7sCJV",
		"last_temp_address_private_key":"",
		"curr_rsmc_temp_address_pub_key":"03dd26ec67e15bde83b527be45a1c64f420821ba78ebc5eb9d3fe1a8ae3cd1f6d9",
		"curr_rsmc_temp_address_private_key":"cTTFKJ3N8W4qHcpwj19NVVaEYAXBxf8DmBj9g7owLTWQ3mXXfD51",
		"curr_htlc_temp_address_pub_key":"03e96c6692bef50af7c6c777ff1bd65b1134d18c98be801e00f8e6247db65950b8",
		"curr_htlc_temp_address_private_key":"cVW35aEjd56ZFHTSDtq9iUzHAfH6rWGVJcmYBn9QEp9ygSfkgZeG"
	}
}
```

**OBD Responses:**

```json
{
    "type": -44, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "approval": true, 
        "request_hash": "742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636"
    }
}
```

<br/>

Carol agree request, Bob create related commitment transactions between Bob and Carol.

**Bob's temp address data for create HTLC commitment transactions:**

```shell
Bob RSMC
address: mxHpHW9Sc7JByG3m9nmjC77jexKcMe8mwi
privkey: cTMZ3csaDWvods2jnMkNJZpz98DRWmFQXb4C1vDFTibcST5g3SNb
pubkey:  033c6284ac5c2409cbf2a49103ff05715f5a0497a0490cdc038248ba37c10e8ccb

Bob HTLC
address: mqkWnkNfhUR7niBVehuBfdXDmhHCL71ohG
privkey: cNon6RZ9uLq6EPpGYt8tDZjQLMWVZDAXcrFy3LH1ZmHQJDbKWnye
pubkey:  025c7cab6f5724a507ad7268bfb6820a3b6902b09a99e1b37241a6b8ede33cc2f1

Bob HTnb
address: mzQmxkY35FaXfzDKyTQPWzxG7k3vZJqkeP
privkey: cUoDGr5cdNarcv43YXFdXBY2zf9721y6u6MiDnk56TSJWCGKvTbL
pubkey:  02977ddeffc04ac0c99c74db308a4db39e60b338d99f3d1661f5ae24f3e17ad414
```

**Bob create HTLC commitment transactions:**

```json
{
	"type": -45,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
		"channel_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",
		"last_temp_address_private_key":"cNzMo4uyZZHBEsxzfK6QRzVcshR5mjFLbwY9p1n921i6PQAewSBD",
		"curr_rsmc_temp_address_pub_key":"033c6284ac5c2409cbf2a49103ff05715f5a0497a0490cdc038248ba37c10e8ccb",
		"curr_rsmc_temp_address_private_key":"cTMZ3csaDWvods2jnMkNJZpz98DRWmFQXb4C1vDFTibcST5g3SNb",
		"curr_htlc_temp_address_pub_key":"025c7cab6f5724a507ad7268bfb6820a3b6902b09a99e1b37241a6b8ede33cc2f1",
		"curr_htlc_temp_address_private_key":"cNon6RZ9uLq6EPpGYt8tDZjQLMWVZDAXcrFy3LH1ZmHQJDbKWnye",
        "curr_htlc_temp_address_for_ht1a_pub_key":"02977ddeffc04ac0c99c74db308a4db39e60b338d99f3d1661f5ae24f3e17ad414",
		"curr_htlc_temp_address_for_ht1a_private_key":"cUoDGr5cdNarcv43YXFdXBY2zf9721y6u6MiDnk56TSJWCGKvTbL"
	}
}
```

**OBD Responses:**

```json
{
    "type": -45, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "h": "e7626f2b7207006d6515399c587c09c3bfb5ed3b12f63c12b0d40e634f9dd9a3", 
        "h_and_r_info_request_hash": "1fe82bc9152741670c4ee2b4853df9346c1cc63fce6d1c896e7eeca8cc62c9d9"
    }
}
```

<br/>

### Carol Send R to Bob through the Path

So, from previous step (Setp 2), Bob ask Carol if Carol can produce to Bob an unknown 20-byte random input data R from a known hash H, within two days, then Bob will settle the contract by paying assets.

Of course Carol has the preimage R, because she generated it. Then now, Carol send R to Bob.

<br/>

**Carol's temp address data for create HTLC commitment transactions:**

```shell
Carol HTLC HEnb commitment transaction:

address: mt3esQqTd8udMNQK8Vm8EDka2N3uCdquCa
privkey: cR14XVjQ4yXunTnpqXZ1FMangq5bZNqsQ4gnsVpJ1KAMkxZVqo3F
pubkey:  020eaa8f0c0f2761215af43dd7fccb11df8cafffcff4e8f186bd1b8a8a11e5f680
```

**Carol (destination) Send R (Preimage_R) to Bob (middleman):**

```json
{
	"type": -46,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
        "r":"d77d7df74ca33a672802387fffc6575a09dc7c45",
		"channel_address_private_key":"cMxR8h9z5oKrdyuXVR9uzBbyyaJz1karxH1FW5xezhKzxQc7sCJV",
		"curr_htlc_temp_address_private_key":"cVW35aEjd56ZFHTSDtq9iUzHAfH6rWGVJcmYBn9QEp9ygSfkgZeG",		
        "curr_htlc_temp_address_for_he1b_pub_key":"020eaa8f0c0f2761215af43dd7fccb11df8cafffcff4e8f186bd1b8a8a11e5f680",
		"curr_htlc_temp_address_for_he1b_private_key":"cR14XVjQ4yXunTnpqXZ1FMangq5bZNqsQ4gnsVpJ1KAMkxZVqo3F"
	}
}
```

**OBD Responses:**

```json
{
    "type": -46, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "id": 4, 
        "r": "2de142c8006a3462241e96a610b59f3d92d8259c", 
        "request_hash": "1fe82bc9152741670c4ee2b4853df9346c1cc63fce6d1c896e7eeca8cc62c9d9"
    }
}
```

<br/>

Bob receieved the R, and check out if R is correct.
If correct, then create rest HTLC commitment transactions.

**Bob replies and create rest HTLC commitment transactions:**

```json
{
	"type": -47,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
		"r":"d77d7df74ca33a672802387fffc6575a09dc7c45",
		"channel_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",
		"curr_htlc_temp_address_private_key":"cNon6RZ9uLq6EPpGYt8tDZjQLMWVZDAXcrFy3LH1ZmHQJDbKWnye"
	}
}
```

**OBD Responses:**

```json
{
    "type": -47, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "r": "2de142c8006a3462241e96a610b59f3d92d8259c", 
        "request_hash": "1fe82bc9152741670c4ee2b4853df9346c1cc63fce6d1c896e7eeca8cc62c9d9"
    }
}
```

<br/>

### Bob Send R to Alice through the Path

Bob send R to Alice, and Alice will create rest HTLC commitment transactions to pay assets to Bob.

**Bob's temp address data for create rest HTLC commitment transactions:**

```shell
Bob HTLC HEnb commitment transaction:

address: mhDr57jhEWeYg2eYQf7LYHoxZ8ZgXEunaT
privkey: cTeQ2e9Hw6y1RHjCCF9x3MR7pn3yPySgxSYy5rtvmVvM7ZNh9jUZ
pubkey:  03ebdfc067f822e9ae0d76759c422bfd3aee342e21ca716dc16b81b335da73d69e
```

**Bob (middleman) Send R (Preimage_R) to Alice (launcher):**

```json
{
	"type": -46,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
        "r":"d77d7df74ca33a672802387fffc6575a09dc7c45",
		"channel_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",
		"curr_htlc_temp_address_private_key":"cPRy5pB8Ek2DabfQ74x8giqDdPtwTvptgRnq8qEP7KduzdmMFmJM",		
        "curr_htlc_temp_address_for_he1b_pub_key":"03ebdfc067f822e9ae0d76759c422bfd3aee342e21ca716dc16b81b335da73d69e",
		"curr_htlc_temp_address_for_he1b_private_key":"cTeQ2e9Hw6y1RHjCCF9x3MR7pn3yPySgxSYy5rtvmVvM7ZNh9jUZ"
	}
}
```

**OBD Responses:**

```json
{
    "type": -46, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "id": 4, 
        "r": "2de142c8006a3462241e96a610b59f3d92d8259c", 
        "request_hash": "1fe82bc9152741670c4ee2b4853df9346c1cc63fce6d1c896e7eeca8cc62c9d9"
    }
}
```

<br/>

Alice receieved the R, and check out if R is correct.
If correct, then create rest HTLC commitment transactions.

**Alice replies and create rest HTLC commitment transactions:**

```json
{
	"type": -47,
	"data": {
		"request_hash":"742db9677d53316b8faef7c9f40766e4f39dd6b82487c103960e9170de8ce636",
		"r":"d77d7df74ca33a672802387fffc6575a09dc7c45",
		"channel_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8",
		"curr_htlc_temp_address_private_key":"cR7wXNwPjMrDCpnJoinTiMK384YyKNTfyctLQ2CCdQobdanEqgAs"
	}
}
```

**OBD Responses:**

```json
{
    "type": -47, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "r": "2de142c8006a3462241e96a610b59f3d92d8259c", 
        "request_hash": "1fe82bc9152741670c4ee2b4853df9346c1cc63fce6d1c896e7eeca8cc62c9d9"
    }
}
```

<br/>

### Close HTLC

For continue using OBD to transfer assets OR some reasons
Alice need to close this HTLC channel between she and Bob.
Then Alice launch the closing request.

**Alice (launcher) request to close HTLC of A2B channel**

**Alice's temp address data for create commitment transactions:**

```shell
Alice RSMC 
address: mkPtXTRyA53ddhknMnVqNCDdeN2FsXmtwe
privkey: cTiDwaM3y5LE2HuWWgvRTC9mgHiovf2zntjSgCPyLLeuUTmKk1BY
pubkey:  02fed65567b2ab00e2cbb28b46a687ce8fd0894486989cba54975b45bbc6a85ed8
```

```json
{
	"type": -48,
	"data": {
        "channel_id":[174,36,154,103,145,76,58,237,32,61,201,81,17,156,135,216,66,28,83,203,251,152,138,102,158,113,131,32,241,229,43,75],
		"channel_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8",
		"last_rsmc_temp_address_private_key":"cSYJ3vwcgMPDqegXFJ2YgCYNNgKS9tNxCaRkZnn3ourQSdGNkJCk",
		"last_htlc_temp_address_private_key":"cR7wXNwPjMrDCpnJoinTiMK384YyKNTfyctLQ2CCdQobdanEqgAs",
		"last_htlc_temp_address_for_htnx_private_key":"cNzNyejXtgC4ySXCVXqa6egVinYLDtGhkRkGd271ZW6AJrVKYZ2w",
		"curr_rsmc_temp_address_pub_key":"02fed65567b2ab00e2cbb28b46a687ce8fd0894486989cba54975b45bbc6a85ed8",
		"curr_rsmc_temp_address_private_key":"cTiDwaM3y5LE2HuWWgvRTC9mgHiovf2zntjSgCPyLLeuUTmKk1BY"
	}
}
```

**OBD Responses:**

```json
{
    "type": -48, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "channel_id": [174,36,154,103,145,76,58,237,32,61,201,81,17,156,135,216,66,28,83,203,251,152,138,102,158,113,131,32,241,229,43,75],
        "create_at": "2019-11-06T08:45:40.4120228+08:00", 
        "create_by": "alice", 
        "curr_rsmc_temp_address_pub_key": "02fed65567b2ab00e2cbb28b46a687ce8fd0894486989cba54975b45bbc6a85ed8", 
        "curr_state": 10, 
        "id": 5, 
        "request_hash": "fa6cdcd0974eeabbeffca3d70d0a66cd7549b002de7cd56eee1c7e60b94dc0be"
    }
}
```

<br/>

Bob agree the closing request, and create BR (Breach Remedy)
& a newer commitment transactions (known as Cxa or Cxb).

**Bob's temp address data for create commitment transactions:**

```shell
Bob RSMC 
address: n2wKQgrfM5fFXmmA6xNWjWPPFPktJHnqEj
privkey: cU78aif2a4YR5xK8HxBTrPKjdjhD8W4SSZNTw4yFEdwi59JMrYQY
pubkey:  0298bdca47bbb76b1022eb7d18534961a12ce6dd80308c839576602b771e324fba
```

**Bob replies and create BR & a newer commitment transactions:**

```json
{
	"type": -49,
	"data": {
		"request_close_htlc_hash":"fa6cdcd0974eeabbeffca3d70d0a66cd7549b002de7cd56eee1c7e60b94dc0be",
		"channel_address_private_key":"cToieuvo3JjkEUKa3tjd6J98RXKDTo1d2hUSVgKpZ1KwBvGhQFL8",
		"last_rsmc_temp_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",
		"last_htlc_temp_address_private_key":"cPRy5pB8Ek2DabfQ74x8giqDdPtwTvptgRnq8qEP7KduzdmMFmJM",
		"last_htlc_temp_address_for_htnx_private_key":"cTeQ2e9Hw6y1RHjCCF9x3MR7pn3yPySgxSYy5rtvmVvM7ZNh9jUZ",
		"curr_rsmc_temp_address_pub_key":"0298bdca47bbb76b1022eb7d18534961a12ce6dd80308c839576602b771e324fba",
		"curr_rsmc_temp_address_private_key":"cU78aif2a4YR5xK8HxBTrPKjdjhD8W4SSZNTw4yFEdwi59JMrYQY"
	}
}
```

**OBD Responses:**

```json
{
    "type": -49, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "msg": "close htlc success"
    }
}
```

<br/>

For continue using OBD to transfer assets OR some reasons
Carol need to close this HTLC channel between she and Bob.
Then Carol launch the closing request.

**Carol (destination) request to close HTLC of B2C channel**

**Carol's temp address data for create commitment transactions:**

```shell
Carol RSMC 
address: mrWGmCyzEQxKWmBQoGmKDSH9Avo7yb6Vz6
privkey: cNDBq3ZKKQEduVyygfcQRzxbhTS3Gt2zz6VEizkp6WyRXn8RdBtH
pubkey:  03080445b531e1df053ce9f1e3d01cdf679f693b23a991ce74145cb0b2e29a2b2d
```

```json
{
	"type": -48,
	"data": {
        "channel_id":[223,177,75,185,186,22,47,155,145,238,242,1,158,247,192,1,48,183,197,192,190,72,49,233,62,65,156,103,111,172,109,51],
		"channel_address_private_key":"cMxR8h9z5oKrdyuXVR9uzBbyyaJz1karxH1FW5xezhKzxQc7sCJV",
		"last_rsmc_temp_address_private_key":"cTTFKJ3N8W4qHcpwj19NVVaEYAXBxf8DmBj9g7owLTWQ3mXXfD51",
		"last_htlc_temp_address_private_key":"cVW35aEjd56ZFHTSDtq9iUzHAfH6rWGVJcmYBn9QEp9ygSfkgZeG",
		"last_htlc_temp_address_for_htnx_private_key":"cR14XVjQ4yXunTnpqXZ1FMangq5bZNqsQ4gnsVpJ1KAMkxZVqo3F",
		"curr_rsmc_temp_address_pub_key":"03080445b531e1df053ce9f1e3d01cdf679f693b23a991ce74145cb0b2e29a2b2d",
		"curr_rsmc_temp_address_private_key":"cNDBq3ZKKQEduVyygfcQRzxbhTS3Gt2zz6VEizkp6WyRXn8RdBtH"
	}
}
```

**OBD Responses:**

```json
{
    "type": -48, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "channel_id": [223,177,75,185,186,22,47,155,145,238,242,1,158,247,192,1,48,183,197,192,190,72,49,233,62,65,156,103,111,172,109,51],
        "create_at": "2019-11-06T08:49:19.3717723+08:00", 
        "create_by": "carol", 
        "curr_rsmc_temp_address_pub_key": "03080445b531e1df053ce9f1e3d01cdf679f693b23a991ce74145cb0b2e29a2b2d", 
        "curr_state": 10, 
        "id": 7, 
        "request_hash": "9491ad7b9ff1003d6404b6f60845dfb0423d6233f42a3dbe118650c6c0e10232"
    }
}
```

<br/>

Bob agree the closing request, and create BR (Breach Remedy)
& a newer commitment transactions (known as Cxa or Cxb).

**Bob's temp address data for create commitment transactions:**

```shell
Bob RSMC 
address: mimQmQxqBVSCbUjVEir5d9Fi9ij9jqPEdP
privkey: cRNxX8S287DA1hkMZHVwnQiMdQVwBBdqpaGYDP1wrRdzT7pSm5kU
pubkey:  02a08635fb1c664aa2bc1a87e76f8dc0b3170c0d45d0f899b3f192093afa1bcd8c
```

**Bob replies and create BR & a newer commitment transactions:**

```json
{
	"type": -49,
	"data": {
		"request_close_htlc_hash":"9491ad7b9ff1003d6404b6f60845dfb0423d6233f42a3dbe118650c6c0e10232",
		"channel_address_private_key":"cTr7SHQ7FR6Ej8ZU8vDpbt3y9GuF3hnBqn81Tv9Tdi5TeakqKavt",
		"last_rsmc_temp_address_private_key":"cTMZ3csaDWvods2jnMkNJZpz98DRWmFQXb4C1vDFTibcST5g3SNb",
		"last_htlc_temp_address_private_key":"cTeQ2e9Hw6y1RHjCCF9x3MR7pn3yPySgxSYy5rtvmVvM7ZNh9jUZ",
		"last_htlc_temp_address_for_htnx_private_key":"cUoDGr5cdNarcv43YXFdXBY2zf9721y6u6MiDnk56TSJWCGKvTbL",
		"curr_rsmc_temp_address_pub_key":"02a08635fb1c664aa2bc1a87e76f8dc0b3170c0d45d0f899b3f192093afa1bcd8c",
		"curr_rsmc_temp_address_private_key":"cRNxX8S287DA1hkMZHVwnQiMdQVwBBdqpaGYDP1wrRdzT7pSm5kU"
	}
}
```

**OBD Responses:**

```json
{
    "type": -49, 
    "status": true, 
    "from": "<user_id>", 
    "to": "<user_id>", 
    "result": {
        "msg": "close htlc success"
    }
}
```

<!-- Added by Kevin Zhang 2019-11-19 END POINT-->

This document is not completed yet, and will be updated during our programming.

# API Document
Please visit OBD [online API documentation](https://api.omnilab.online) for the lastest update.

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


Join us in [OmniBOLT slack channel](https://join.slack.com/t/omnibolt/shared_invite/enQtNzY2MDIzNzY0MzU5LTFlZTNlZjJhMzQxZTU2M2NhYmFjYjc1ZGZmODYwMWE3YmM0YjNhZWQyMDU2Y2VlMWIxYWFjN2YwMjlmYjUxNzA)

# Current Features

* Generate user OBD(OmniBOLT Daemon) address.  
* Open Poon-Dryja Channel.  
* BTC and Omni assets in funding and transaction.  
* fund and close channel.  
* Commitment Transaction within a channel.  
* List latest commitment transaction in a channel.   
* List all commmitment trsactions in a channel.
* List all the breach remedy transactions in a channel.
* Surveil the broadcasting commitment transactions and revockable delivery transactions.
* Execute penelty. 

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





 


