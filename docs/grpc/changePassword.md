## changePassword

ChangePassword changes the password of the encrypted wallet. This will automatically unlock the wallet database if successful.

## Arguments:
| Field		   |	gRPC Type		|	   Description  |
| -------- 	 |	---------   |    ---------    |  
| current_password	     |	bytes		  |current_password should be the current valid passphrase used to unlock the daemon. When using REST, this field must be encoded as base64.|
| new_password	     |	bytes		  |new_password should be the new passphrase that will be needed to unlock the daemon. When using REST, this field must be encoded as base64.|
| stateless_init	     |	bool		  |stateless_init is an optional argument instructing the daemon NOT to create any *.macaroon files in its file system.|
| new_macaroon_root_key	     |	bool		  |new_macaroon_root_key is an optional argument instructing the daemon to rotate the macaroon root key when set to true. This will invalidate all previously generated macaroons.|

## Response:
| Field		         |	gRPC Type		|	   Description  |
| -------- 	       |	---------   |    ---------    |  
| admin_macaroon|	bytes	    |The binary serialized admin macaroon that can be used to access the daemon after creating the wallet. If the stateless_init parameter was set to true, this is the ONLY copy of the macaroon and MUST be stored safely by the caller. Otherwise a copy of this macaroon is also persisted on disk by the daemon, together with other macaroon files.|

## Example:

<!--
java code example
-->

```java
String currentPassword = "12345678";
String newPassword = "876654321"
Walletunlocker.ChangePasswordRequest changePasswordRequest = Walletunlocker.ChangePasswordRequest.newBuilder()
                .setCurrentPassword(ByteString.copyFromUtf8(currentPassword ))
                .setNewPassword(ByteString.copyFromUtf8(newPassword ))
                .build();
Obdmobile.changePassword(changePasswordRequest.toByteArray(), new Callback() {
    @Override
    public void onError(Exception e) {
        e.printStackTrace();
    }

    @Override
    public void onResponse(byte[] bytes) {
        if(bytes == null){
            return;
        }
        try {
            Walletunlocker.ChangePasswordResponse changePasswordResponse = Walletunlocker.ChangePasswordResponse.parseFrom(bytes);
            String macaroon = changePasswordResponse.getAdminMacaroon().toString();

        } catch (InvalidProtocolBufferException e) {
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
admin_macaroon: "\002\001\003lnd\002\370\001\003\n\020O\2623*34\357\003\275B{D\030D\243\321\022\0010\032\026\n\aaddress\022\004read\022\005write\032\023\n\004info\022\004read\022\005write\032\027\n\binvoices\022\004read\022\005write\032!\n\bmacaroon\022\bgenerate\022\004read\022\005write\032\026\n\amessage\022\004read\022\005write\032\027\n\boffchain\022\004read\022\005write\032\026\n\aonchain\022\004read\022\005write\032\024\n\005peers\022\004read\022\005write\032\030\n\006signer\022\bgenerate\022\004read\000\000\006 \205\351\305\370)+\372\312\tG&\t\\(\367\020\2761R\252E\212\251\320.\242U\363\321%;\301"
}
```


