## start 

Start is used to start the local wallet and node after entering the app. 

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |  
| params	     |	string		  |	      Node startup related parameters.|


## Response:
This response has no parameters.

## Example:

<!--
java code example
-->

```java
String params = "--lnddir=" +            
         "/storage/emulated/0/Android/data/com.omni.wallet/cache/data/chain/bitcoin/testnet/" +
        "--trickledelay=5000 --debuglevel=debug --alias=alice\n" +
        "--autopilot.active --maxpendingchannels=100 " +
        "--bitcoin.active --bitcoin.testnet --bitcoin.node=neutrino " +
        "--enable-upfront-shutdown " +
        "--tlsdisableautofill " +
        "--norest "+
        "--neutrino.connect=192.144.199.67" +
        " --omnicoreproxy.rpchost=192.144.199.67:18332";
Obdmobile.start(params , new Callback() {
    @Override
    public void onError(Exception e) {
        if (e.getMessage().contains("lnd already started")) {

        } else if (e.getMessage().contains("unable to start server: unable to unpack single backups: chacha20poly1305: message authentication failed")) {

        } else if(e.getMessage().contains("error creating wallet config: unable to initialize neutrino backend: unable to create neutrino database: cannot allocate memory")){
                    
        }
    }

    @Override
    public void onResponse(byte[] bytes) {
    }
});
```

<!--
下面放例子的返回结果 
-->
response:
```
This response has no parameters.
```



