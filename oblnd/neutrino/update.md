this update fix the err below
```shell
2022-12-29 01:09:06.934 [DBG] BTCN: Sending getheaders (locator 5a2d3c297a17bd0831c595729e0236d5b562faa77f815e06fda8332892259c94, stop 0000000000000000000000000000000000000000000000000000000000000000) to 43.138.107.248:18444 (outbound)
2022-12-29 01:09:07.254 [DBG] BTCN: Received ping from 43.138.107.248:18444 (outbound)
2022-12-29 01:09:07.254 [DBG] BTCN: Sending pong to 43.138.107.248:18444 (outbound)
2022-12-29 01:09:07.255 [DBG] BTCN: Received feefilter from 43.138.107.248:18444 (outbound)
2022-12-29 01:09:07.255 [DBG] BTCN: Received addr (1 addr) from 43.138.107.248:18444 (outbound)
2022-12-29 01:09:09.760 [DBG] BTCN: Received headers (num 2000) from 43.138.107.248:18444 (outbound)
2022-12-29 01:09:09.763 [INF] BTCN: Processed 62 blocks in the last 17.57s (height 4001, 2022-12-05 18:26:01 +0800 CST)
2022-12-29 01:09:09.764 [DBG] BTCN: Difficulty retarget at block height 4032
2022-12-29 01:09:09.764 [DBG] BTCN: Old target 207fffff (7fffff0000000000000000000000000000000000000000000000000000000000)
2022-12-29 01:09:09.764 [DBG] BTCN: New target 201fffff (1fffff0000000000000000000000000000000000000000000000000000000000)
2022-12-29 01:09:09.764 [DBG] BTCN: Actual timespan 32h12m0s, adjusted timespan 84h0m0s, target timespan 336h0m0s
2022-12-29 01:09:09.764 [WRN] BTCN: Header doesn't pass sanity check: block target difficulty of 7fffff0000000000000000000000000000000000000000000000000000000000 is higher than max of 1fffff0000000000000000000000000000000000000000000000000000000000 -- disconnecting peer

```

* more info for the bug
: https://codesti.com/issue/lightninglabs/neutrino/255

*  the pull request for the bug :
https://github.com/lightninglabs/neutrino/pull/256

and the pull-request  256 used some the btcsuite v2 module,example "btcec/v2" ,"btcsuite/btcutil". all the v2 module not incompatibility with the oblnd now. 
so i clone the old version lightninglabs/neutrino, and checkout to hash 53b628ce175698d112ddff3b60856902556c2fc2 which use "btcsuite v1 module", then add the  pull-request-256's update to it.
