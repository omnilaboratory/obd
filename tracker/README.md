# OmniBOLT Tracker | in Golang
[![](https://img.shields.io/badge/license-MIT-blue)](https://github.com/omnilaboratory/obd/blob/master/LICENSE) [![](https://img.shields.io/badge/golang-%3E%3D1.9.0-orange)](https://golang.org/dl/) [![](https://img.shields.io/badge/Spec-OmniLayer-brightgreen)](https://github.com/OmniLayer/spec) 
  

The architecture of tracker network is [here](https://omnilaboratory.github.io/obd/#/Architecture?id=tracker-network)

## Install Omnicore

[Install the omnicore from the project repository](https://github.com/OmniLayer/omnicore#installation) . After the installation is complete, the user needs to wait for a while to synchronize the full node data

## Config tracker

Open `tracker/config/conf.ini`, edit the omnicore address, username and password:

```
host = the omnicore ip address, for example: 62.234.216.108:18332
user = the user name
pass = the password
```

Omnicore has a whitelist, only the ip on the list can access it. We suggest user to run tracker and omnicore on the same device.  


## Run tracker

List unspent outputs for 1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc.
```
./tracker_server
```

 
