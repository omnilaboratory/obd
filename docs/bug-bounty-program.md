# Bug Bounty Program

* `Contact`: Neo Carmack(neocarmack@omnilab.online), Ben Fei(benfei@omnilab.online)

<p align="center">
  <img width="750" alt="OmniBOLT-Bug-Bounty" src="assets/omni-bounty.png">
</p>

## Introduction

(latest updated Oct 3, 2020)

The OmniBOLT network leverages lightning technology to provide quick and safety guarantees for anyone who wishes to build scalable, decentralized payment apps. It can run independently or as a plugin of LND, in which case LND applications are empowered to operate smart asset lightning channels. Working with skilled security researchers and hackers across the globe plays a crucial role in improving the security of our network. Therefore we launched a bug bounty program to find vulnerabilities and pay rewards. 

This bug bounty program encourages participants to identify and submit bugs/flaws/vulnerabilities that could negatively impact OmniBOLT network users. Successful submissions have a chance of being eligible for a bounty reward. The scope of our program and the bounty levels are provided in more detail below.


## Scope

Do code review for logical and security mistakes in our testnet: OBD is a new lightning network that was written from scratch by the community. With the launch of the bug bounty program we put the following components in scope:   

 * `OBD Core`: [the OBD code base](https://github.com/omnilaboratory/obd).
	* channle operations
	* payment
	* query, balance, state
	* network
	* application contracts

 * `JS SDK`: [JS SDK](https://omnilaboratory.github.io/obd/#/js-sdk) is for light client that interacts with local/remote OBD node.
 * `LND Bridge`: [LND bridge](https://github.com/omnilaboratory/lnd) runs OBD as a plugin for LND, and adds interfaces to LND grpc package that empowers LND apps to operate OBD channel. 



Participants can use the GUI tool to access all services without the need to spend any time on the installation, setup, and configuration of OBD node. To get started please visit [this tutorial](https://omnilaboratory.github.io/obd/#/GUI-tool).   

 
## What to look for

The following list is collected by the community, and it should give you some ideas for issues that we regard as high-value submissions. The list is not meant to limit or discourage other types of submissions, but it increases the chances of a successful submission (and bounty award).

 * Current reported attacks to lightning network. 
 * Compromise funds from users who have deposited or received funds on the OBD network.
 * Prevent users from depositing, withdrawing, or transacting funds on the OBD network.
 * Double spend a UTXO of a commitment transaction on a channel and exit it to the main chain (Omnilayer/BTC).
 * Transaction/messages malleability, which causes broadcasting elder commitment transactions.
 * Cheat/Attack activities that a user can not discover and punish.
 * Security weaknesses/attacks on the P2P communication protocol among nodes. 
 * Buffer overflow, private info leak in (remote) message calls.
 * Data type overflow/wrap around, e.g. integer overflow.
 * Concurrency. 
 * Try to include invalid transactions in a block. 
 * Brick the exit priority queue of channels so that no funds can be exited anymore. 
 * Gain access to a system and run OS commands aka getting a shell.
 * Server configuration issues (open ports, firewall).
 * Issues related to third-party libraries used (outdated software).
 * Incorrect usage of the official BIP(33,39,44,..) implementations.
 * Cryptographic primitives security. Incorrect implementation/usage/configuration of:
	 * Elliptic curve (secp256k1, ECDSA,ECDH,ECIES).
	 * Hash algorithms (Keccak-256, Blake2b).
	 * Seeds, pub/priv key generation, storage, and usage.
	 * To be added.
 

 

## How to submit bug reports

Before the OmniBOLT mainnet is online, issues can be sent to OBD repository or neocarmack@omnilab.online. If you assess the impact will be significant, or will probably compromise users' funds, please send it via private mail.  

An issue report shall includes following information:

|              Information            |                         Explantion                   | 
|              -----------            |                       --------------                 | 
|       Type of vulnerability         |  A classification of the type of vulnerability being reported, such as security feature bypass, buffer overflow, and so on. We recommend user refer to https://nvd.nist.gov/vuln/categories.| 
|         Affected Component          |  The component that is affected by the vulnerability. This should include the component’s package name, branch and version information.	  | 
|             Platform                |  OS: Linux, BSD, Windows. Browser: Chrome, Firefox, Edge, Safari.        |
|           Proof-of-concept          |  All steps required to trigger the vulnerability. A description of the vulnerability in the form of text, code, or other form depending on the nature of the vulnerability. The environment dependency should be minimized.   |
|  Vulnerability reproduction output  |  The output from a successful reproduction of the vulnerability. This could consist of debugger output, a screenshot, or some other format that demonstrates a reproduction of the issue. More detailed information like debugger output is preferred.        |
|         Analysis and proposal       |  Optional. Your proposal of how to fix the vulnerability.    |
|               Reference             |  Optional. The references that support your analysis and proprosal.     |




## Bounty Rewards

 * The bounty amount will be determined in USDT and will only be paid out via online OmniWallet. We will need your omni address.   
 * Successful submissions are rewarded based on the severity of the issue.  
 * We generally take real impact into account and use CVSSv3 scoring system as a reference to understand the risk of an issue.  
 * This software is under heavy development, the issue you open is possibly what we already know and are working to fix. Therefore it is a mutual understanding that you may not be the first one who identifies the vulnerability. But you are still eligible for bounty rewards if you propose a novel solution to it.  


The following table gives an overview of the reward structure:   
				
|  Component Category  |        Low       |      Medium       |        High       |      Critical      | 
|  ------------------  |  --------------  |  ---------------  |  ---------------  |   ---------------  |
|     OBD Core         |  up to 250 USD	  |  up to 1,500 USD  |  up to 7,500 USD  |  up to 15,000 USD  | 
|      JS SDK  	       |  up to 250 USD	  |  up to 1,500 USD  |  up to 7,500 USD  |  up to 15,000 USD  | 
|    LND Bridge        |  up to 250 USD	  |  up to 1,500 USD  |  up to 7,500 USD  |  up to 15,000 USD  |
  

## Rules 

* You must not compromise users' personal data. 
* You must not use the vulnerability to receive any reward or monetary gain outside of the bug bounty program or allow anyone else to profit outside the bug bounty program.
* out-of-scope submissions still have a chance to be rewarded, but it's at our discretion.
* You may not be the first one who identifies the vulnerability. But you are still eligible for bounty rewards if you propose a novel solution to it.
* It’s at OmniBOLT's discretion to decide whether a bug is significant enough to be eligible for a reward and its severity.
* The vulnerability is not exploited on OmniBOLT mainnet (though the mainnet is not online yet).
* To be added.

 
