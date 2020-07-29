# OBD GUI Tool
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/omnilaboratory/obd/blob/master/LICENSE) [![](https://img.shields.io/badge/standard%20readme-OK-brightgreen)](https://github.com/omnilaboratory/obd/blob/master/README.md) [![](https://img.shields.io/badge/protocol-OmniBOLT-brightgreen)](https://github.com/omnilaboratory/OmniBOLT-spec) 
[![](https://img.shields.io/badge/API%20V0.3-Document-blue)](https://api.omnilab.online) 


OmniBOLT Daemon Debugging Tool. This is a graphic user interface for developers. Every API on the left navigation tree has been defined in [api.omnilab.online](https://api.omnilab.online). This tool is in fact a lighning wallet.  

If you come to this tool in developing modules for OmniBOLT, you must have successfully installed and ran OBD on your local or remote machine. If you do not familiar with the whole process yet, we suggest you [install](https://github.com/omnilaboratory/obd#table-of-contents) OBD first and try to run it for a quick experience.  

In this tutorial, you can connect either your own OBD node, or the [nodes in the testnet](https://github.com/omnilaboratory/DebuggingTool#Nodes-in-testnet) we configured for our community. 


<p align="center">
  <img width="500" alt="Debugging Tool Screenshot" src="https://github.com/omnilaboratory/DebuggingTool/blob/master/doc/img/image_screen.png">
</p>

* [Installation](https://github.com/omnilaboratory/DebuggingTool#installation-clone-this-project-and-run-it-by-chrome)
* [Nodes in the testnet](https://github.com/omnilaboratory/DebuggingTool#Nodes-in-testnet)
* [Operations](https://github.com/omnilaboratory/DebuggingTool#operations) 
	* [Step 1: connect to an OBD node](https://github.com/omnilaboratory/DebuggingTool#step-1-connect-to-an-obd-node)
	* [Step 2: signup a new user](https://github.com/omnilaboratory/DebuggingTool#step-2-signup-a-new-user)
	* [Step 3: login using mnemonic words](https://github.com/omnilaboratory/DebuggingTool#step-3-login-using-mnemonic-words)
	* [Step 4: connect another user](https://github.com/omnilaboratory/DebuggingTool#step-4-connect-another-user)
	* [Step 5: open channel](https://github.com/omnilaboratory/DebuggingTool#step-5-open-channel)
	* [Step 6: create an invoice](https://github.com/omnilaboratory/DebuggingTool#step-6-create-an-invoice)
	* [Step 7: channel operations](https://github.com/omnilaboratory/DebuggingTool#step-7-channel-operations)
* [Auto pilot mode](https://github.com/omnilaboratory/DebuggingTool#auto-pilot-mode)
* [JS-SDK](https://github.com/omnilaboratory/DebuggingTool#js-sdk)
* [how to use APIs](https://github.com/omnilaboratory/DebuggingTool#how-to-use-apis)
 	

## Installation: clone this project and run it by Chrome

```
git clone https://github.com/omnilaboratory/DebuggingTool
```

Chrome must run with argument `disable-web-security` to disable web security stratigy. For example, if you are in Windows terminal:
```
chrome.exe --args --disable-web-security --user-data-dir=/any_temp_directory_for_chrome_data
```

Then open index.html under the DebuggingTool directory.

> Remark: We use some local json files to store the content of each interface, such as interface name, description, parameter list, etc.
In the js code, we use jQuery’s `$.getJSON(jsonFile, function(result) {}` method to parse these json files. Chrome by default prohibits the access of this form of cross-domain resource.  

There are currently three solutions:  
1) Start Chrome by disabling web security policy: --disable-web-security.  
2) Run a web server, put the debugging tool in the corresponding directory, so that Chrome can visit the url normally.  
3) Use Safari or other browsers to run the debugging tool.  

## Nodes in testnet

Here is the list of nodes in the testnet. will be updating accordingly:   

``` 
ws://62.234.216.108:60020/wstest
ws://62.234.188.160:60020/wstest
```



## Operations 

### Step 1: connect to an OBD node

<p align="center">
  <img width="750" alt="Connect Screenshot" src="https://github.com/omnilaboratory/DebuggingTool/blob/master/doc/img/connect.png">
</p>

1. click "Connect to OBD node";  
2. on the right panel, appears the input box. connect you local obd node or one of the [nodes in the testnet](https://github.com/omnilaboratory/DebuggingTool#Nodes-in-testnet), for example: `ws://62.234.216.108:60020/wstest`. Just replace the default `127.0.0.1:60020` by `62.234.216.108:60020`  
3. click "connect". wait for several seconds, the status bar on top of the window will show the status "connected".

If you want to build and deploy your own OBD, you shall go through the installation intruction at [obd repository](https://github.com/omnilaboratory/obd#installation).

### Step 2: signup a new user

<p align="center">
  <img width="750" alt="SignUp" src="https://github.com/omnilaboratory/DebuggingTool/blob/master/doc/img/signup.png">
</p>

1. click "signUp";  
2. click "invoke API";  
3. OBD responses the new set of mnemonic words for the currently connected new user. The mnemonic words is the identity to carry out all the following operations in obd network; Record this mnemonic words in somewhere safe. For example, write it down on a paper.  


### Step 3: login using mnemonic words

<p align="center">
  <img width="750" alt="login" src="https://github.com/omnilaboratory/DebuggingTool/blob/master/doc/img/login.png">
</p>

1. click "login";  
2. in the input box of "Input Parameters", past the mnemonic words generated in step 2, and  
3. click "invoke API";  
4. OBD responses 3 arguments, telling user where and who he is:  

```
nodeAddress : /ip4/62.234.216.108/tcp/3001/p2p/QmP1mQMzDRV2bKWdhwvPWFubSAz1gqJY44RjdYm3G5DFeF
```
`nodeAddress` is the node location of the OBD server you connected. In this tutorial, it is the node "ws://62.234.216.108:60030/ws" you connected in step 1.  

```
nodePeerId : QmP1mQMzDRV2bKWdhwvPWFubSAz1gqJY44RjdYm3G5DFeF
``` 
`nodePeerId` is part of the complete `nodeAddress`. Because on one server, there can be thousans of OBD running on it, every OBD has a unique `nodePeerId` to be identified.  

```
userPeerId : 30dfbc0e1b42c4cb50410b7a08186ce405a92fff235480608425bf4b0207e5ad
```
This is the user id, which is used together with `nodeAddress` to tell someone else that "i'm here, please connect me by connectP2PPeer".  

We assume this first window belongs to user Alice.

### Step 4: connect another user

Open another browser window, open index.html, and signup another user to get his `nodeAddress` and `userPeerId`. In this tutorial, this second window belongs to user Bob.

Switch back to Alice's window, we shall input Bob's `nodeAddress` and `userPeerId` to build connection with Bob.  

<p align="center">
  <img width="750" alt="connectNode" src="https://github.com/omnilaboratory/DebuggingTool/blob/master/doc/img/connectNode.png">
</p>

1. switch back to Alice's window;  
2. click "connectP2PPeer";  
3. input the `nodeAddress` into the "NodeAddress" input box;  
4. click "invoke API";  

### Step 5: open channel

Click openChannel, input the arguments required by this function and click "invoke API", wait Bob's response. 

In the other window you just opened in step 4 for Bob, you will see an incoming message asking for opening a channel with Bob. Click "acceptChannel", leave the default values that the js SDK filled for you, response Alice to accept the "openChannel" request.  

### Step 6: create an invoice

<p align="center">
  <img width="500" alt="connectNode" src="https://github.com/omnilaboratory/DebuggingTool/blob/master/doc/img/createInvoice.png">
</p>


1. switch back to Alice's window;  
2. click "createInvoice";  
3. input the `property_id`, `amount`, `h` `expiry_time` and short memo, where `h` is the locker (`hash(r)`) that payer uses it to lock a payment, only you can unlock it by your secrete `r`;  
4. click "invoke API", you will see the beth32 encoded invoice string and QR code are created;  

Share ths invoice string or QR code to anyone (not only Bob) who is going to pay you. 


### Step 7: channel operations

Then you are able to keep going with other operations to dive deeper into OmniBOLT.  

Online API documents lists all the channel operations step by step, and testing demo data as well. Please visit OBD [online API documentation](https://api.omnilab.online) to learn how to fill in arguments to work with OBD.  

## Auto pilot mode

During forwarding HTLCs, a channel shall construct tens of temporary unbroadcast transactions, so lots of private keys needed to be manually managed. Enable this mode, the obd will be your surrogate, will help you fill in all the private keys, and automatically answer the remote requests.

<p align="center">
  <img width="150" alt="auto_pilot_mode" src="https://github.com/omnilaboratory/DebuggingTool/blob/master/doc/img/auto_pilot_mode.png">
</p>

## JS SDK

The JS SDK is located under [js/obdapi.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/obdapi.js). It implments a complete set of APIs for HD wallets and client applications. It manages pub/priv keys genertion and helps developers automatically fill in the OBD arguments which is hardly to be manually finished.

## How to use APIs

JS/TypeScript examples can be found under [js/common.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js).

Or you will seek help via [online API documentation](https://api.omnilab.online). 



