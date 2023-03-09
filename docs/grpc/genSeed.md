## genSeed

GenSeed is the first method that should be used to instantiate a new lnd instance. This method allows a caller to generate a new aezeed cipher seed given an optional passphrase. If provided, the passphrase will be necessary to decrypt the cipherseed to expose the internal wallet seed.

Once the cipherseed is obtained and verified by the user, the InitWallet method should be used to commit the newly generated seed, and create the wallet 

#### Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |  
| aezeed_passphrase	     |	bytes		  |	      aezeed_passphrase is an optional user provided passphrase that will be used to encrypt the generated aezeed cipher seed. When using REST, this field must be encoded as base64.|  
| seed_entropy   |	bytes		  |       seed_entropy is an optional 16-bytes generated via CSPRNG. If not specified, then a fresh set of randomness will be used to create the seed. When using REST, this field must be encoded as base64.          | 


#### Response:
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| cipher_seed_mnemonic|	string[]	    |cipher_seed_mnemonic is a 24-word mnemonic that encodes a prior aezeed cipher seed obtained by the user. This field is optional, as if not provided, then the daemon will generate a new cipher seed for the user. Otherwise, then the daemon will attempt to recover the wallet state linked to this cipher seed.|  
| enciphered_seed  |	string		  |enciphered_seed are the raw aezeed cipher seed bytes. This is the raw cipher text before run through our mnemonic encoding scheme.|

#### Example:

<!--
java code example
-->

```java
Walletunlocker.GenSeedRequest genSeedRequest = Walletunlocker.GenSeedRequest.newBuilder().build();
Obdmobile.genSeed(genSeedRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {
    }

    @Override
    public void onResponse(byte[] bytes) {
        try{
            Walletunlocker.GenSeedResponse genSeedResponse = Walletunlocker.GenSeedResponse.parseFrom(bytes);
            List genSeedResponseList =  genSeedResponse.getCipherSeedMnemonicList();
        }catch (InvalidProtocolBufferException e){
            e.printStackTrace();
        }
    }
});
```

<!--
下面放例子的返回结果 
-->
response:
```
{
cipher_seed_mnemonic: "absorb"
    cipher_seed_mnemonic: "spread"
    cipher_seed_mnemonic: "interest"
    cipher_seed_mnemonic: "coconut"
    cipher_seed_mnemonic: "charge"
    cipher_seed_mnemonic: "treat"
    cipher_seed_mnemonic: "light"
    cipher_seed_mnemonic: "thunder"
    cipher_seed_mnemonic: "antenna"
    cipher_seed_mnemonic: "pluck"
    cipher_seed_mnemonic: "blossom"
    cipher_seed_mnemonic: "confirm"
    cipher_seed_mnemonic: "dragon"
    cipher_seed_mnemonic: "devote"
    cipher_seed_mnemonic: "actor"
    cipher_seed_mnemonic: "rug"
    cipher_seed_mnemonic: "smart"
    cipher_seed_mnemonic: "swallow"
    cipher_seed_mnemonic: "domain"
    cipher_seed_mnemonic: "brush"
    cipher_seed_mnemonic: "orbit"
    cipher_seed_mnemonic: "kit"
    cipher_seed_mnemonic: "crane"
    cipher_seed_mnemonic: "spatial"
    enciphered_seed: "\000\332a\326\226ri\317\240g\f\t\324\324`\027t y\200\255\350\314\233a\003\216\231\276\365\314\226\204"}
```


