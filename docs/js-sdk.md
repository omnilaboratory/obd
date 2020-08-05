## JS SDK

The JS SDK is located under [/SDK](https://github.com/omnilaboratory/DebuggingTool/tree/master/sdk). It implments a complete set of APIs for HD wallets and client applications. It manages pub/priv keys genertion and helps developers automatically fill in the OBD arguments which is hardly to be manually finished.

## How to use APIs

The [GUI tool](https://github.com/omnilaboratory/DebuggingTool/) is the best reference to the use of js sdk. 

For extensive explanation of APIs, you will seek help via [online documentation](https://api.omnilab.online). 

## 5 minutes to build a basic wallet app

First, you should put all of [JS SDK](https://github.com/omnilaboratory/DebuggingTool/tree/master/sdk) files 
to your path of project.

And, following these steps below in your code:

* [Step 1: connect to an OBD node](#step-1-connect-to-an-obd-node)
* [Step 2: signup a new user](#step-2-signup-a-new-user)
* [Step 3: login using mnemonic words](#step-3-login-using-mnemonic-words)
* [Step 4: connect another user](#step-4-connect-another-user)
* [Step 5: open channel](#step-5-open-channel)
* [Step 6: create an invoice](#step-6-create-an-invoice)
* [Step 7: channel operations](#step-7-channel-operations)

### Step 1: connect to an OBD node

Invoke **connectToServer** function from [wallet.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/sdk/wallet.js) of SDK.

First parameter is `nodeAddress`. It's a URL of the OBD node like `ws://62.234.216.108:60020/wstest`.

Second parameter is `callback`. It's a callback function could be used to process the return data.

Third parameter is `globalCallback`. It's a callback function could be used to process the global messages.

#### Example Code:

```js

// URL of an OBD node
let nodeAddress = 'ws://62.234.216.108:60020/wstest';

// SDK API
connectToServer(nodeAddress, function(response) {
    console.info('Print the callback = ' + response);

    // Your code to process the callback data.
    // Example: Display the success or fail message on app screen.

}, function(globalResponse) {

    // Your code to process the global callback data.

});
```

Full example in GUI-tool you could be see [sdkConnect2OBD](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js) function.


### Step 2: signup a new user

Invoke **genMnemonic** function from [wallet.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/sdk/wallet.js) of SDK.

`genMnemonic` function is used to sign up a new user by hirarchecal deterministic wallet system integrated in the local client. Client generates mnemonic words and the hash of the mnemonic words as the UserID.

#### Example Code:

```js

// SDK API
let mnemonicWords = genMnemonic();

// Your code to process the data.
// Example: Display the mnemonic words on app screen.
```

Full example in GUI-tool you could be see [sdkGenMnemonic](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js) function.