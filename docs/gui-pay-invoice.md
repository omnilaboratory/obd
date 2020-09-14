## pay invoice

After a channel has been funded, Alice or Bob are able to pay to each other or any one else in the network. Here we let Bob to create an invoice:

<p align="center">
  <img width="500" alt="connectNode" src="assets/createInvoice.png">
</p>

1. switch to Bob's window;  
2. click "createInvoice";  
3. input the `property_id`, `amount`, `h` `expiry_time` and short memo, where `h` is the locker (`hash(r)`) that Bob uses it to lock a payment, only Bob can unlock it by the secrete `r`;  
4. click "invoke API", Bob will see the beth32 encoded invoice string and QR code are created;  


Share ths invoice string or QR code to Alice:   
 

<p align="center">
  <img width="750" alt="Connect Screenshot" src="assets/payInvoice.png">
</p>

On Alice's screen:  
1. click "payInvoice";  
2. past the invoice string in the box;  
3. click "invoke API", then Bob's obd and Alice's obd will communicate with each other to finish the payment process automatically. 

The invoice system simplifies the complex process of multi hop HTLC payment. Users don't need to manually response every incomming message, but it you want to go deeper to see what happens during a payment life cycle, we suggest you to keep reading:  
 
 
## inside the process of invoice payment



```
    +-------+                                    +-------+
    |       |---(1)---   HTLCFindPath   -------->|       |
    |       |---(2)---     addHTLC      -------->|       |
    |       |                                    |       |
    | Alice |<--------   HTLCSigned     ---(3)---|  Bob  |
    |       |<--------    forwardR      ---(4)---|       |
    |       |                                    |       |
    |       |---(5)---      signR       -------->|       |
    |       | 				         |       |
    |       |   either Alice or Bob can close    |       |
    |       |                                    |       |
    |       |<--------    closeHTLC     ---(6)---|       |
    |       |                                    |       |
    |       |---(7)---  closeHTLCSigned -------->|       |
    |       |                                    |       |
    |       |                 or                 |       |
    |       |                                    |       |
    |       |---(6)---     closeHTLC    -------->|       |
    |       |                                    |       |
    |       |<--------  closeHTLCSigned ---(7)---|       |
    +-------+                                    +-------+

    - where node Alice is the 'payer' and node Bob is the 'payee' >.  

```

### find a payment path

to be done

### add HTLC

to be done

### sign HTLC

to be done

### forword R to unlock HTLC

to be done

### sign R to accept unlocking HTLC

to be done  

### close HTLC 

to be done 


### sign to close HTLC

Click "closeHTLCSigned", sign to agree closing the HTLC, then this HTLC will be closed and the balance of both sides will be updated accordingly. The resource occupied by this HTLC will be released, and the channel will be available for other operations. 
