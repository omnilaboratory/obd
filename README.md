# LightningOnOmni 

LightningOnOmni implements the [OmniBOLT](https://github.com/LightningOnOmnilayer/Omni-BOLT-spec) specification, which enables Lightning network to be Omnilayer assets aware. 

# Dependency

[Omnicore 0.18](https://github.com/OmniLayer/omnicore/tree/develop), which is currently in develop branch, and will be to finalize the release soon. 

Omnicore 0.18 integrates the latest BTC core 0.18, which enables relative time locker used in RSM contracts and HTL contracts.

# Installation
The following installation works for Ubuntu 14.04.4 LTS.

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
### option 1: Remote Omnicore node 
Use our remote OmniCore node. Go to `\config\conf.ini`, you will see:
```
[chainNode]
host=62.234.216.108:18332
user=omniwallet
pass=cB3]iL2@eZ1?cB2?
```
This is our tesing full node for community to run/call omni commands remotely. The omni-lightning node invocates Omni RPC commands from this node.

### option 2: Local Omnicore node 
[Install OmniCore](https://github.com/OmniLayer/omnicore#installation) on your local machine. Omnicore requires a full BTC core node, which may take days to synchronize the whole BTC database to your local device. After finish synchronization, you can run omni/BTC commands for experiments, such as constructing raw transactions, generating new addresses.

Then edit the configure file: `\config\conf.ini`
```
[chainNode]
host=127.0.0.1:port
user=your user name
pass=your password
```

## Step 3: Run omni-lightning node
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
go build olndserver.go
```
which generates the executable binary file. 

## Step 4: Test channel operations using Websocket



TBD

# Current Features

* Generate user OLND address.  
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





 


