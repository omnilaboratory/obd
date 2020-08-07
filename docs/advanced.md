# Advanced topic

This is an introduction on omnilayer, in order to quickly let users to know how to manage their own smart assets, which is quite useful in applying lightning technology to various scenarios.


## Close HTLC

Invoke **closeHTLC** function from [htlc.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/sdk/htlc.js) of SDK.

First parameter is `myUserID`. It's the user id of the client currently logged in.

Second and third parameter are `nodeID` and `remoteUserID`. Both returned by **logIn** function. Explanation is [here](https://omnilaboratory.github.io/obd/#/GUI-tool?id=step-3-login-using-mnemonic-words)

Final parameter is `CloseHtlcTxInfo object`. It contains a lot of information, see example code.

#### Example Code:

```js
let nodeID       = 'nodePeerId';
let remoteUserID = 'userPeerId';

let info                                         = new CloseHtlcTxInfo();
info.channel_id                                  = 'channel_id';
info.channel_address_private_key                 = 'channel_address_private_key';
info.last_rsmc_temp_address_private_key          = 'last_rsmc_temp_address_private_key';
info.last_htlc_temp_address_private_key          = 'last_htlc_temp_address_private_key';
info.last_htlc_temp_address_for_htnx_private_key = 'last_htlc_temp_address_for_htnx_private_key';
info.curr_rsmc_temp_address_pub_key              = 'curr_rsmc_temp_address_pub_key';
info.curr_rsmc_temp_address_private_key          = 'curr_rsmc_temp_address_private_key';

// SDK API
closeHTLC(myUserID, nodeID, remoteUserID, info);
```

Full example in GUI-tool you could be see [sdkCloseHTLC](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js) function.


## Close HTLC Signed

Invoke **closeHTLCSigned** function from [htlc.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/sdk/htlc.js) of SDK.

First parameter is `myUserID`. It's the user id of the client currently logged in.

Second and third parameter are `nodeID` and `remoteUserID`. Both returned by **logIn** function. Explanation is [here](https://omnilaboratory.github.io/obd/#/GUI-tool?id=step-3-login-using-mnemonic-words)

Final parameter is `CloseHtlcTxInfoSigned object`. It contains a lot of information, see example code.

#### Example Code:

```js
let nodeID       = 'nodePeerId';
let remoteUserID = 'userPeerId';

let info                                         = new CloseHtlcTxInfoSigned();
info.msg_hash                                    = 'msg_hash'; // returned by closeHTLC function
info.channel_address_private_key                 = 'channel_address_private_key';
info.last_rsmc_temp_address_private_key          = 'last_rsmc_temp_address_private_key';
info.last_htlc_temp_address_private_key          = 'last_htlc_temp_address_private_key';
info.last_htlc_temp_address_for_htnx_private_key = 'last_htlc_temp_address_for_htnx_private_key';
info.curr_rsmc_temp_address_pub_key              = 'curr_rsmc_temp_address_pub_key';
info.curr_rsmc_temp_address_private_key          = 'curr_rsmc_temp_address_private_key';

// SDK API
closeHTLCSigned(myUserID, nodeID, remoteUserID, info);
```

Full example in GUI-tool you could be see [sdkCloseHTLCSigned](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js) function.


## Managed and smart asset

### Issue new tokens with fixed supply

Invoke **issueFixedAmount** function from [manage_asset.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/sdk/manage_asset.js) of SDK.

Parameter is `IssueFixedAmountInfo object`. It contains a lot of information, see example code.

#### Example Code:

```js
let info            = new IssueFixedAmountInfo();
info.from_address   = 'the address to send from';
info.name           = 'the name of the new tokens to create';
info.ecosystem      = 'the ecosystem to create the tokens in'; // (1 for main ecosystem, 2 for test ecosystem)
info.divisible_type = 'the type of the tokens to create'; // (1 for indivisible tokens, 2 for divisible tokens)
info.data           = 'a description for the new tokens';
info.amount         = 'the number of tokens to create';

// SDK API
issueFixedAmount(info);
```

Full example in GUI-tool you could be see [sdkIssueFixedAmount](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js) function.


### Issue new tokens with manageable supply.

Invoke **issueManagedAmout** function from [manage_asset.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/sdk/manage_asset.js) of SDK.

Parameter is `IssueManagedAmoutInfo object`. It contains a lot of information, see example code.

#### Example Code:

```js
let info            = new IssueManagedAmoutInfo();
info.from_address   = 'the address to send from';
info.name           = 'the name of the new tokens to create';
info.ecosystem      = 'the ecosystem to create the tokens in'; // (1 for main ecosystem, 2 for test ecosystem)
info.divisible_type = 'the type of the tokens to create'; // (1 for indivisible tokens, 2 for divisible tokens)
info.data           = 'a description for the new tokens';

// SDK API
issueManagedAmout(info);
```

Full example in GUI-tool you could be see [sdkIssueManagedAmout](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js) function.


## Issue tokens

### Issue or grant new units of managed tokens.

Invoke **issueManagedAmout** function from [manage_asset.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/sdk/manage_asset.js) of SDK.

Parameter is `OmniSendGrant object`. It contains a lot of information, see example code.

#### Example Code:

```js
let info          = new OmniSendGrant();
info.from_address = 'the address to send from';
info.property_id  = 'the identifier of the tokens to grant';
info.amount       = 'the amount of tokens to create';
info.memo         = 'a text note attached to this transaction (none by default)';

// SDK API
sendGrant(info);
```

Full example in GUI-tool you could be see [sdkSendGrant](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js) function.


## Burn tokens

### Revoke units of managed tokens.

Invoke **sendRevoke** function from [manage_asset.js](https://github.com/omnilaboratory/DebuggingTool/blob/master/sdk/manage_asset.js) of SDK.

Parameter is `OmniSendRevoke object`. It contains a lot of information, see example code.

#### Example Code:

```js
let info          = new OmniSendRevoke();
info.from_address = 'the address to send from';
info.property_id  = 'the identifier of the tokens to revoke';
info.amount       = 'the amount of tokens to revoke';
info.memo         = 'a text note attached to this transaction (none by default)';

// SDK API
sendRevoke(info);
```

Full example in GUI-tool you could be see [sdkSendRevoke](https://github.com/omnilaboratory/DebuggingTool/blob/master/js/common.js) function.


## Transfer ownership

to be done

## Atomic swap
