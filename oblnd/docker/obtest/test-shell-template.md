Docker helps people to quickly interact with obd and omnicored via their terminals. OBD uses `docker-compose` to package `obd`, `omnicored`, and `btcd` together to make deploying these daemons easily. Please check the `docker-compose.yml` config file under this folder for all the configurations.

## prerequisites
Name    | Version 
--------|---------
docker-compose  | latest
docker          | latest

## startup 
```bash
#all test is in the docker/obtest directory
cd docker/obtest

./start-compose.sh
# ./reset-compose.sh will reset exists test env.

```
this will startup compose services;

## close
```shell
./clean-compose.sh
```


# command lines

A network example, each node uses a omnicore backend:  

```
      | omnicored proxy |             | omnicore proxy |                        
            |	  |	                      |	        
            |	  |	                      |	                  
        ----|       |-------                  |	     
        |                  |	                |   	     
        |                  |	                |   	      
    +---------+        +--------+         +---------+      
    |  Alice  | ------ |  Bob   | ------  |  Carol  |	     
    +---------+        +--------+         +---------+      

```  

alice cli: ./a-cli.sh
#more option   ./a-cli.sh -h

bob cli: ./b-cli.sh
#more option   ./b-cli.sh -h

omnicored cli: ./om-cli.sh

usage example

```shell
./a-cli.sh decodepayreq  obort2147483651:500u1p3vlcmzpp5ynph4j2tvmklqq3g4m6djqv6sjycr55t8uw5m9xvjsn7pf0yvdsqdqqcqzpgxqyz5vq3q8zqqqqqrsp54cq6k33jznf7l7uv6ypnrdccdmrkwe459h3m9spdy8un4hk5xavs9qyyssqcsup8urhhh5dvaszcq4zputt8z0gmrqt7uvxad9eug389wkewunx24tk95ng7kj9nevlj0u8euvk42e5cmzyuxlqftewrqsl4phuayqqzexztj
./a-cli.sh getinfo  
```

## omnicored
the omnicored-docker will not mine block automation; if you want  some blocks confirm, exec ./mine.sh manually.
available commands
```shell
./om-cli.sh  send_coin.sh alice_address #send_coin to alice , will send 1 btc and  100Asset to  alice, the assetId 2147483651
./mine.sh  mine 3 blocks
```

## faucet
as an **option** you can modify docker-compose.yml to use online omnicoreporxy for regtest
* omnicoreporxy is public prxoy omnicore-backand ,it can be access anonymous.
* faucet-swager-api: https://swagger.oblnd.top/?surl=https://faucet.oblnd.top/openapiv2/foo.swagger.json
* omnicoreporxy-server wallet addre ms5u6Wmc8xF8wFBo9w5HFouFNAmnWzkVa6 have enough test-coin to send you.
* omnicoreporxy have pre-created an asset which id is 2147483651;
*  the SendCoin api  every invoke will send you 1btc and 100 asset.
```shell
#send test coin form curl
export assetId=2147483651
curl -X 'GET' \
  'https://faucet.oblnd.top/api/SendCoin/$a_address?assetId=$assetId' \
  -H 'accept: application/json'
```

## newaddress
```shell
./a-cli.sh newaddress
{
    "address": "moUP1tmVjU8WmqkvkbpfjfmmfiTMSEh5w2"
}
./b-cli.sh newaddress
{
    "address": "mzrUe6zbrVJua6jtcjzMWXgq4yjLRqT1st"
}

export asset_id=2147483651
export a_address=moUP1tmVjU8WmqkvkbpfjfmmfiTMSEh5w2
export b_address=mzrUe6zbrVJua6jtcjzMWXgq4yjLRqT1st
```
## send_coin
```shell

#send_coin to alice , will send 1 btc and  100Asset to  alice, the assetId 2147483651
./om-cli.sh send_coin.sh $a_address
```

## getinfo
```shell
./a-cli.sh getinfo
{
    "version": "0.14.2-beta commit=v0.14.2-beta-45-g15186093b-dirty",
    "commit_hash": "15186093b656830ef8e74485925ed80d4d96ea16",
    "identity_pubkey": "0236e0de4e5d15038db6ba6a872811055fef6c16a9b28fcdc315320c5f48b03ae7",
    ......
    ......
    ......
    "uris": [
        "0236e0de4e5d15038db6ba6a872811055fef6c16a9b28fcdc315320c5f48b03ae7@172.26.0.4:9735"
    ],
    "features": {
        "0": {
            "name": "data-loss-protect",
            "is_required": true,
            "is_known": true
        },
        "5": {
            "name": "upfront-shutdown-script",
            "is_required": false,
            "is_known": true
        },
        "7": {
            "name": "gossip-queries",
            "is_required": false,
            "is_known": true
        },
        "9": {
            "name": "tlv-onion",
            "is_required": false,
            "is_known": true
        },
        "12": {
            "name": "static-remote-key",
            "is_required": true,
            "is_known": true
        },
        "14": {
            "name": "payment-addr",
            "is_required": true,
            "is_known": true
        },
        "17": {
            "name": "multi-path-payments",
            "is_required": false,
            "is_known": true
        },
        "23": {
            "name": "anchors-zero-fee-htlc-tx",
            "is_required": false,
            "is_known": true
        },
        "30": {
            "name": "amp",
            "is_required": true,
            "is_known": true
        },
        "31": {
            "name": "amp",
            "is_required": false,
            "is_known": true
        },
        "45": {
            "name": "explicit-commitment-type",
            "is_required": false,
            "is_known": true
        },
        "2023": {
            "name": "script-enforced-lease",
            "is_required": false,
            "is_known": true
        }
    }
}

./b-cli.sh getinfo
{
    "version": "0.14.2-beta commit=v0.14.2-beta-45-g15186093b-dirty",
    "commit_hash": "15186093b656830ef8e74485925ed80d4d96ea16",
    "identity_pubkey": "02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1",
    "alias": "bob",
       ......
    ......
    ......
    "uris": [
        "02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1@172.26.0.3:9735"
    ],
        ......
    ......
    ......
    }
}

export a_pubkey="024f1adf8eab75f576e7499e702654f07b1b29c9ea0aca33b8a54695e859f7b61f"
export a_url="$a_pubkey@alice:9735"
export b_pubkey=$(./b-cli.sh getinfo |jq -r .identity_pubkey)
export b_url="$b_pubkey@bob:9735"

```

## connect
```shell
# connect to bob
./a-cli.sh connect $b_url
sleep 3
./a-cli.sh listpeers
{
    "peers": [
        {
            "pub_key": "02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1",
            "address": "172.21.0.3:9735",
            "bytes_sent": "391",
            "bytes_recv": "391",
            .....
            .....
        }
    ]
```

## openchannel
```shell
# alice createchannel : {channel_btc_cap: 20000 sat, channel_asset_cap: 1000000000(omniAmount)==10(assetUnit)}
./a-cli.sh  openchannel --asset_id $asset_id --node_key $b_pubkey --local_btc_amt=20000 --local_asset_amt=1000000000
{
        "funding_txid": "ee46bb1fbf4c4f780afb4413b15c249855743d9dbb8f6d9f0d657450391e2b10"
}

export funding_txid=ee46bb1fbf4c4f780afb4413b15c249855743d9dbb8f6d9f0d657450391e2b10

./mine.sh
[
  "1ad4f88c278456f8206f772c31dcf4a72ca4a4222ed6dc2e08fef048d06fa9c6",
  "765faa1c829b6ff86db073cdce01e506d5b90d0c8dd64b5cf856e7613b299cf2",
  "31eda0b808a03f8aedaf51bb79d2e0c7839f6b2aeabe66642142576ece7c26a1"
]

./a-cli.sh listchannels
{
    "channels": [
        {
            "active": true,
            "remote_pubkey": "02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1",
            "channel_point": "9bd84e28ee53e0b0e0e78ed09e29d13965f7fe74d17ba8ee6d6090c0334ece89:1",
            "chan_id": "173722837254145",
            "btc_capacity": "20000",
            "asset_capacity": "1000000000",
            "local_balance": "16530",
            "remote_balance": "0",
            "local_asset_balance": "1000000000",
            "remote_asset_balance": "0",
            "asset_id": 2147483651,
            "commit_fee": "0",
            "commit_weight": "816",
            "fee_per_kw": "2500",
            "unsettled_balance": "0",
            "total_satoshis_sent": "0",
            "total_satoshis_received": "0",
            "num_updates": "0",
            "pending_htlcs": [
            ],
            "csv_delay": 2016,
            "private": false,
            "initiator": true,
            "chan_status_flags": "ChanStatusDefault",
            "local_chan_reserve_sat": "10000000",
            "remote_chan_reserve_sat": "10000000",
            "static_remote_key": false,
            "commitment_type": "ANCHORS",
            "lifetime": "7",
            "uptime": "7",
            "close_address": "",
            "push_btc_amount_sat": "0",
            "push_asset_amount_sat": "0",
            "thaw_height": 0,
            "local_constraints": {
                "csv_delay": 2016,
                "chan_reserve_sat": "10000000",
                "dust_limit_sat": "354",
                "max_pending_amt_msat": "990000000",
                "min_htlc_msat": "10",
                "max_accepted_htlcs": 483
            },
            "remote_constraints": {
                "csv_delay": 2016,
                "chan_reserve_sat": "10000000",
                "dust_limit_sat": "354",
                "max_pending_amt_msat": "990000000",
                "min_htlc_msat": "10",
                "max_accepted_htlcs": 483
            }
        }
    ]
}

 export channel_id=180319907020801
 ./a-cli.sh getchaninfo $channel_id

```


## addinvoice
```shell
./b-cli.sh addinvoice  --asset_id $asset_id --amount 100000000
{
    "r_hash": "8894dbbc6fc13ec2d33f4de1d17bda178ad42db7f542dc7138a8c6bfbf4137ee",
    "payment_request": "obort2147483651:11p3dqrjqpp53z2dh0r0cylv95elfhsaz776z79dgtdh74pdcufc4rrtl06pxlhqdqqcqzpgxqyz5vq3q8zqqqqqrsp5jyj08qn8lgpndgkp4cv65g0djm7jhqj8wu2vw8u5xydlc0rlcsks9qyyssqxj9k6hhkqjax0lkxcw6ar5u7j9k7k8ddamudvslep2fk84hpkcm3xcehuckv6nu0ypwpevm9hv9x9gyxlh3mmtwj85a0j6mkrayjtmsq98avmg",
    "add_index": "1",
    "payment_addr": "9124f38267fa0336a2c1ae19aa21ed96fd2b82477714c71f94311bfc3c7fc42d"
}
export pay_req="obort2147483651:11p3dqrjqpp53z2dh0r0cylv95elfhsaz776z79dgtdh74pdcufc4rrtl06pxlhqdqqcqzpgxqyz5vq3q8zqqqqqrsp5jyj08qn8lgpndgkp4cv65g0djm7jhqj8wu2vw8u5xydlc0rlcsks9qyyssqxj9k6hhkqjax0lkxcw6ar5u7j9k7k8ddamudvslep2fk84hpkcm3xcehuckv6nu0ypwpevm9hv9x9gyxlh3mmtwj85a0j6mkrayjtmsq98avmg"
```

## decodepayreq
```shell
./b-cli.sh decodepayreq $pay_req
{
    "destination": "02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1",
    "payment_hash": "8894dbbc6fc13ec2d33f4de1d17bda178ad42db7f542dc7138a8c6bfbf4137ee",
    "timestamp": "1657802304",
    "expiry": "86400",
    "description": "",
    "description_hash": "",
    "fallback_addr": "",
    "cltv_expiry": "40",
    "route_hints": [
    ],
    "payment_addr": "9124f38267fa0336a2c1ae19aa21ed96fd2b82477714c71f94311bfc3c7fc42d",
    "amt_msat": "0",
    "amount": "100000000",
    "asset_id": 2147483651,
    "features": {
        "9": {
            "name": "tlv-onion",
            "is_required": false,
            "is_known": true
        },
        "14": {
            "name": "payment-addr",
            "is_required": true,
            "is_known": true
        },
        "17": {
            "name": "multi-path-payments",
            "is_required": false,
            "is_known": true
        }
    }
}
```


## sendpayment 
```shell
 ./a-cli.sh sendpayment --asset_id 2147483651 --pay_req $pay_req
 
 Payment hash: 8894dbbc6fc13ec2d33f4de1d17bda178ad42db7f542dc7138a8c6bfbf4137ee
Description: 
AssetId : 2147483651
Amount (in satoshis): 100000000
Fee limit (in satoshis): 5000000
Destination: 02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1
Confirm payment (yes/no): yes
+------------+--------------+--------------+--------------+-----+----------+-----------------+-------+
| HTLC_STATE | ATTEMPT_TIME | RESOLVE_TIME | RECEIVER_AMT | FEE | TIMELOCK | CHAN_OUT        | ROUTE |
+------------+--------------+--------------+--------------+-----+----------+-----------------+-------+
| SUCCEEDED  |        0.015 |        0.117 | 100000       | 0   |      203 | 173722837254145 |       |
+------------+--------------+--------------+--------------+-----+----------+-----------------+-------+
Amount + fee:   100000 + 0 sat
Payment hash:   8894dbbc6fc13ec2d33f4de1d17bda178ad42db7f542dc7138a8c6bfbf4137ee
Payment status: SUCCEEDED, preimage: 11d85d63bb85ba2c3d9d6ef22666a485057651e4f3879dd4ab4a87e05076f53b


./mine.sh 
[
  "16f7dc8d366e64bca38510f8f6f7b6491c68bc1cc892414f6d9837168ba350e8",
  "2d5a0db44b4d6e127cd2a6f6fcad9a87eeb9cc3f8fcdfd634aa26804849a2ab5",
  "55eceb96b90a3b0d054bac8bc7448b7f1234a36514383af66d62b9fa59f71ee9"
]

 ./a-cli.sh getchaninfo $channel_id
 
 {
    "channel_id": "173722837254145",
    "chan_point": "9bd84e28ee53e0b0e0e78ed09e29d13965f7fe74d17ba8ee6d6090c0334ece89:1",
    "last_update": 1657802218,
    "node1_pub": "0236e0de4e5d15038db6ba6a872811055fef6c16a9b28fcdc315320c5f48b03ae7",
    "node2_pub": "02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1",
    "capacity": "1000000000",
    "asset_id": 2147483651,
    "node1_policy": {
        "time_lock_delta": 40,
        "min_htlc": "10000",
        "fee_base_msat": "1000",
        "fee_rate_milli_msat": "100",
        "disabled": false,
        "max_htlc_msat": "990000000",
        "last_update": 1657802218,
        "asset_id": 2147483651
    },
    "node2_policy": {
        "time_lock_delta": 40,
        "min_htlc": "10000",
        "fee_base_msat": "1000",
        "fee_rate_milli_msat": "100",
        "disabled": false,
        "max_htlc_msat": "990000000",
        "last_update": 1657802218,
        "asset_id": 2147483651
    }
}

```


## add third user carl
copy bob section in  docker-compose.yml, updates port and name int docker-compose.yml.
```shell
mkdir volumes/lnd/carl
docker compose up carl -d

alias c-cli.sh='docker compose exec -u 1000 carl lncli-debug -n regtest '
{
    "version": "0.14.2-beta commit=v0.14.2-beta-45-g15186093b-dirty",
    "commit_hash": "15186093b656830ef8e74485925ed80d4d96ea16",
    "identity_pubkey": "0296b66f30dbb4bd53f85bbc1886e0948ef21a9a7b68bc59e5e9d9eacb7b1ef35e",
    "alias": "carl",
    .......
    .......
    .......
    "uris": [
        "0296b66f30dbb4bd53f85bbc1886e0948ef21a9a7b68bc59e5e9d9eacb7b1ef35e@172.26.0.3:9735"
    ],
    .......
    .......
    .......
}

c-cli.sh newaddress
{
    "address": "mg3gqc4chXjNPZWy7jZt4YGBRN9fNgGqAn"
}
export c_address=mg3gqc4chXjNPZWy7jZt4YGBRN9fNgGqAn
export c_pubkey=03a1bde60ae2a3eb010e5212b7a874657a9e2a9dd8149b6af14e5270860f705941
export c_url=$c_pubkey@carl:9735


```

### bob connect to carl 
```shell
 ./b-cli.sh connect $c_url
 
 #will show two node 
  ./b-cli.sh listpeers 
  
```

### bob create channel with carl
```shell
# faucet to bob
docker compose exec  -u 1000 omnicored send_coin.sh $b_address

#bob openchannel with carl
./b-cli.sh  openchannel --asset_id $asset_id --node_key $c_pubkey --local_btc_amt=20000 --local_asset_amt=1000000000
{
        "funding_txid": "986f49f39871310abd47aaee9832f8975f4675413bdb880bbae95580b5546c6b"
}

export funding_txid_bc=986f49f39871310abd47aaee9832f8975f4675413bdb880bbae95580b5546c6b
#mine 3 bloks,for channel confirm
./mine.sh 
 c-cli.sh listchannels
{
    "channels": [
        {
            "active": true,
            "remote_pubkey": "02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1",
            "channel_point": "6212327b2ca335c498e57a8e66b296049ebaff274bb9bc73ad1bf22631a25ddf:1",
            "chan_id": "190215511670785",
            .....
            "asset_id": 2147483651,
            "commit_fee": "0",
            "commit_weight": "816",
            "fee_per_kw": "2500",
           .....
            "pending_htlcs": [
            ],
            .....
            .....
            .....
            "local_constraints": {
                "csv_delay": 2016,
                "chan_reserve_sat": "10000000",
                "dust_limit_sat": "354",
                "max_pending_amt_msat": "990000000",
                "min_htlc_msat": "10",
                "max_accepted_htlcs": 483
            },
            "remote_constraints": {
                "csv_delay": 2016,
                "chan_reserve_sat": "10000000",
                "dust_limit_sat": "354",
                "max_pending_amt_msat": "990000000",
                "min_htlc_msat": "10",
                "max_accepted_htlcs": 483
            }
        }
    ]
}

export channel_id_bc=186916976787457

#mine 3 blocks for  channel announcement
./mine.sh
```

### carl addinvoice
```shell
c-cli.sh addinvoice  --asset_id $asset_id --amount 10000000
{
    "r_hash": "60953e40751e2f264826f260e4b101b5d8bddde043285c787b782d87f2f1ab2d",
    "payment_request": "obort2147483651:100m1p3wykqcpp5vz2nusr4rchjvjpx7fswfvgpkhvtmh0qgv59c7rm0qkc0uh34vksdqqcqzpgxqyz5vq3q8zqqqqqrsp582wkwaj7cqfeddu0qrl6j60lvtcjtdp5fh2z33wljd5ddksykcrq9qyyssq0dkn9mrwvhvwhnn8rwqn3kny0f2vxumaxzk6qcqadc342kyuaq4r0qc4vru9hgjnrayw5lrk02qlkf0chquty06fjxmpkasy9fdz4wsq7r6drp",
    "add_index": "1",
    "payment_addr": "3a9d67765ec01396b78f00ffa969ff62f125b4344dd428c5df9368d6da04b606"
}
export pay_req_c=obort2147483651:100m1p3wykqcpp5vz2nusr4rchjvjpx7fswfvgpkhvtmh0qgv59c7rm0qkc0uh34vksdqqcqzpgxqyz5vq3q8zqqqqqrsp582wkwaj7cqfeddu0qrl6j60lvtcjtdp5fh2z33wljd5ddksykcrq9qyyssq0dkn9mrwvhvwhnn8rwqn3kny0f2vxumaxzk6qcqadc342kyuaq4r0qc4vru9hgjnrayw5lrk02qlkf0chquty06fjxmpkasy9fdz4wsq7r6drp

c-cli.sh decodepayreq $pay_req_c
  ......
  ......
  ......
  
```

### alice pay $pay_req_c
```shell
 ./a-cli.sh sendpayment --asset_id $asset_id --pay_req $pay_req_c
 
```



export a_address=mrjQTt5xMkZP3cHijSWTeFvDaXASTuFQHu
export b_address=mmRAA1txRcptDiDeCBUvL3z7Yno3jqyWUw
export a_url="0236e0de4e5d15038db6ba6a872811055fef6c16a9b28fcdc315320c5f48b03ae7@alice:9735"
export b_url="02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1@bob:9735"
export b_pubkey="02b3c11c620a977dbb63e972cc9ebee41043071e8226802d1f8e0d28cb02f2c3d1"
export asset_id=2147483651
export funding_txid=9bd84e28ee53e0b0e0e78ed09e29d13965f7fe74d17ba8ee6d6090c0334ece89
export channel_id=173722837254145
export pay_req="obort2147483651:11p3dqrjqpp53z2dh0r0cylv95elfhsaz776z79dgtdh74pdcufc4rrtl06pxlhqdqqcqzpgxqyz5vq3q8zqqqqqrsp5jyj08qn8lgpndgkp4cv65g0djm7jhqj8wu2vw8u5xydlc0rlcsks9qyyssqxj9k6hhkqjax0lkxcw6ar5u7j9k7k8ddamudvslep2fk84hpkcm3xcehuckv6nu0ypwpevm9hv9x9gyxlh3mmtwj85a0j6mkrayjtmsq98avmg"

alias c-cli.sh='docker compose exec -u 1000 carl lncli-debug -n regtest '
export c_address=mtL3vm36CjTdWTjmFvHLtZzbwnQRHB2brz
export c_url=0296b66f30dbb4bd53f85bbc1886e0948ef21a9a7b68bc59e5e9d9eacb7b1ef35e@carl:9735
export c_pubkey=0296b66f30dbb4bd53f85bbc1886e0948ef21a9a7b68bc59e5e9d9eacb7b1ef35e

export channel_id_bc=190215511670785
export funding_txid_bc=6212327b2ca335c498e57a8e66b296049ebaff274bb9bc73ad1bf22631a25ddf

export channel_id_bc1=200111116320769
export funding_txid_bc1=827c6e6a91571947a1e211f16dfc5bb205c5b194e55edf713ebde3285bcb7eb7

### david
```shell
alias d-cli.sh='~/git/lnd/lncli-debug --rpcserver=localhost:10009 --lnddir=~/apps/oblnd-tmp -n regtest '
export d_address=mx8koaQmwjF11uGpZPQeNMNfxecJY6EUq3
export d_pubkey=02703509d18efe3f2f9ad901e3cc7cb3be174d632d57da4d6eacb37ce1cb0daa27
export d_url=$d_pubkey@10.1.1.50:9735

#connect to bob
export a_url=$a_pubkey@172.20.0.3:9735
export b_url=$b_pubkey@172.20.0.4:9735

export channel_id_bd=190215511670785
export funding_txid_bd=6212327b2ca335c498e57a8e66b296049ebaff274bb9bc73ad1bf22631a25ddf
d-cli.sh describegraph
```
