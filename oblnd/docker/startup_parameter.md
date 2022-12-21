## 后端连接启动参数说明．
下面只列关键参数列表;完整参数列表使用"-h"参数启动，就会打印地完整列表


### btc 后端连接模式
支持的网络类型有 regnet testnet mainnet; 每种网络类型可支持bitcoind omnicoreproxy neutrino 三种节点模式．

#### 网络类型指定参数格式　--bitcoin.xxxxxx　
```shell
#regtest网络下
--bitcoin.regtest
#testnet网络下
--bitcoin.testnet
#mainnet网络下
--bitcoin.mainnet
```
#### 三种节点模式指定参数格式 --bitcoin.node=[bitcoind|omnicoreproxy|neutrino|]
目前只用到这３种模式，每种模式连接btc节点的地址和验证方式不同．
* bitcoind
```shell
  --bitcoind.rpchost="$BTC_HOST_ADDRESS_PORT"
  --bitcoind.rpcuser="$RPCUSER"
  --bitcoind.rpcpass="$RPCPASS" 
  --bitcoind.zmqpubrawblock=tcp://"$BTC_HOST_ADDRESS":28332 
  --bitcoind.zmqpubrawtx=tcp://"$BTC_HOST_ADDRESS":28333
```
* omnicoreproxy
```shell
    --omnicoreproxy.rpchost="$OMNI_HOST_ADDRESS_PORT" 
    --omnicoreproxy.zmqpubrawblock=tcp://"$OMNI_HOST_ADDRESS_PORT":28332 
    --omnicoreproxy.zmqpubrawtx=tcp://"$OMNI_HOST_ADDRESS_PORT":28333
```

* neutrino
```shell
    --neutrino.connect="$BTC_HOST_ADDRESS"
    --omnicoreproxy.rpchost="$OMNI_HOST_ADDRESS_PORT"
```
#### 现部署有的每种网络类下的节点地址
每种网络类型下，我们不会把所有的节点模式全部部署一份．以文档为准
* regnet 出块时间：每2分钟3块
  * omnicoreproxy
    * 国内用：$OMNI_HOST_ADDRESS_PORT=43.138.107.248  $OMNI_HOST_ADDRESS_PORT=43.138.107.248:18332  
      * 水龙头：http://43.138.107.248:9090/swaggerTool/?surl=http://43.138.107.248:8090/openapiv2/foo.swagger.json 
      * 预创建token-propertyid 2147483651
    *  ~~国外用：国内国外regtest-node是独立的节点，无法使用相同的boostrap-dns-node, $OMNI_HOST_ADDRESS_PORT=regnet.oblnd.top  $OMNI_HOST_ADDRESS_PORT=regnet.oblnd.top:18332~~  
      * ~~水龙头：http://swagger.cn.oblnd.top:9090/?surl=surl=http://faucet.cn.oblnd.top:9090/openapiv2/foo.swagger.json~~ 
      * ~~预创建token-propertyid: 2147483651~~
* testnet 出块时间2-18分钟不等
  * neutrino
    * 国内用：$BTC_HOST_ADDRESS=192.144.199.67  $OMNI_HOST_ADDRESS_PORT=192.144.199.67:18332  
      * token水龙头：http://43.138.107.248:9090/swaggerTool/?surl=http://192.144.199.67:8090/openapiv2/foo.swagger.json 
      * 预创建token-propertyid: 2147485160 token-owner:mvd6r2KRoaMVr7Y9mDe8pDxe5a5iZLJHN9
    * 国外用：$BTC_HOST_ADDRESS=testnet.oblnd.top  $OMNI_HOST_ADDRESS_PORT=192.144.199.67:18332
      * token水龙头：http://43.138.107.248:9090/swaggerTool/?surl=http://192.144.199.67:8090/openapiv2/foo.swagger.json
      * 预创建token-propertyid: 2147485160 token-owner:mvd6r2KRoaMVr7Y9mDe8pDxe5a5iZLJHN9
    * btc-testnet水龙头: 可以google搜索＂btc　testnet　faucet＂查找到更多的水龙头，可用的不多；测试了一个可用的https://testnet-faucet.com/btc-testnet/　，每次只能发5000-10000 satoshi
    * neutrino.db 下载列表https://cache.oblnd.top/neutrino-testnet/　，按需下载相应的文件，下载时相的文件时，url要加日期参数,日期格式任意，如下载neutrino.db文件“https://cache.oblnd.top/neutrino-testnet/neutrino.db?date=2022-12-22”; gz扩展名为相应文件的gzip压缩版．
* mainnet 出块时间10分钟不等
　 * neutrino
  待部部署
 


### 公共参数
* --lnddir #lnd数据库目录
* --bitcoin.active #使用btc类型的后端
* --noseedbackup #可选：自动创建钱包/当需要用户手动保存助记词时，不能加这个参数
* --debuglevel=info/debug/trace #用于调试目的，打印不同级别的log，trace最详细，info是默认。
* --alias=alice #节点别名
* --autopilot.active ＃自动驾驶功能，要加上
* --maxpendingchannels=100 ＃pending正在建立，正在建立的通道的数量．
* --enable-upfront-shutdown #关闭通道时把coin返默认钱包地址，要加上
* --nobootstrap #可选：禁用启动时使用dns自动发现连接到公共节点

