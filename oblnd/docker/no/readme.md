Neutrino mode backend docker

## prerequest
```shell
mkdir btc/bin/

download bitcoind binary from  https://bitcoincore.org/en/releases/  for version 22/23
uncompress bitcoind bitcoin-cli to btc/bin/ dir
```


## startup
docker compose up -d

## dokcers
| NodeType  |  RPC port   | cli         |
|-----------|-----|-------------|
| omnicored |  18332   | ./om-cli.sh |
| btcoind   |  18400   | ./btc-cli.sh   |

example command line
```shell
./om-cli.sh omnicore-cli omni_getinfo
./btc-cli.sh getblockchaininfo
```


## start a test lnd with neutrino mod
```shell
./start-with-neutrino.sh

./neutrino-cli.sh newaddress
mq7YT77M2dZmQZ6j56PhrsUztUMcU2VJwh


./om-cli.sh send_coin.sh mq7YT77M2dZmQZ6j56PhrsUztUMcU2VJwh

./neutrino-cli.sh assetsbalancebyaddress --address mq7YT77M2dZmQZ6j56PhrsUztUMcU2VJwh
{
    "list": [
        {
            "propertyid": "2147483651",
            "name": "ftoken",
            "balance": "10000000000",
            "reserved": "0.00000000",
            "frozen": "0.00000000"
        }
    ]
}

# verify send_coin res from oblnd
./docker/lnd/neutrino-cli.sh walletbalancebyaddress --address mq7YT77M2dZmQZ6j56PhrsUztUMcU2VJwh
{
    "total_balance": "100000546",
    "confirmed_balance": "100000546",
    "unconfirmed_balance": "0",
    "account_balance": {
        "default": {
            "confirmed_balance": "100000546",
            "unconfirmed_balance": "0"
        }
    },
    "address": "mq7YT77M2dZmQZ6j56PhrsUztUMcU2VJwh"
}
```

