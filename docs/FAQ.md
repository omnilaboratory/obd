# FAQ

* `Contact`: Neo Carmack(neocarmack@omnilab.online), Ben Fei(benfei@omnilab.online)

## 1. What is OmniBOLT lightning network? 

OmniBOLT is a big extension of BOLT (Basis of Lightning Technology), the major parts of this specification are p2p instant payment for crypto-assets. Simply, OmniBOLT is a lightning network for smart assets, issued on OmniBOLT via OmniLayer.

Further reading: [OmniBOLT: Facilitates smart assets lightning transactions](https://omnilaboratory.github.io/obd/#/README?id=omnibolt-facilitates-smart-assets-lightning-transactions)

## 2. Is it another incompatible lightning network?

Based on the fundamental theory of Lightning network, OmniBOLT specification describes how to enable OmniLayer assets to be circulated via lightning channels and how can smart assets benefit from this novel instant payment theory. Specifically, OmniBOLT applies lightning network theory. It's a lightning network for smart assets, and not just a LND network for bitcoin only.  

## 3. Why OmniBOLT?
According to the layer-2 protocol BOLT (Basis of Lightning Technology) specification for off-chain bitcoin transfer, we need a protocol to support wider range of assets for upper layer applications: payment, game, finance or stable coins.As you might know, Omnilayer is an onchain smart assets issuance technology which is proven secure and stable. Constructing lightning channels on top of it automatically acquires the ability of issuing assets, temper resistant, and on-chain settlement. This is where OmniBOLT is built upon.

## 4. Advantages of OmniBolt?

* Instant payment of smart assets issued on OmniLayer. 
* Cross channel atomic swap for various crypto assets.
* Decentralized exchange on top of stable coin enabled lightning channels.  
* Automatic market maker and liquidity pool for DEX.

## 5. Is OmniBOLT compatible with current lightning network?
OmniBOLT is able to communicate and exchange information with the LN network. The bottom layer is the  existing lightning technology, so they are naturally compatible with each other. The only difference is that current lightning network is designed for bitcoin only, whereas OmniBOLT adds assets on top of the basic layer. Using atomic swap technology, it is easy to exchange tokens or any other assets against bitcoin. 

## 6. Will OmniBOLT split the liquidity of bitcoin lightning network?
No. In fact, it is not possible for any token to use the liquidity of bitcoin unless that token is exchanged to btc at first and then the BTC is transferred to the destination and exchange back to the original token. This solution suffers the volatility of btc-token exchange rate. For example, if the price of bitcoin rises, the tokens received in the destination will be less than the amount that was initially sent.  

## 7. Is it possible that a token uses bitcoin liquidity to transfer?
Yes. using some sort of exchange mechanism, tokens are able to use bitcoin liquidity to reach their destination. But the shortcoming is that this solution suffers the volatility of btc-token exchange rate. For example, if the price of bitcoin rises, the tokens received in the destination will be less than the amount that was initially sent. 

## 8. Is OmniBOLT be able to issue token?
Though OmniBOLT itself doesn't issue token by itself but it uses Omnilayer to issue assets/tokens on bitcoin main-chain.

## 9. How multiple assets are supported by OmniBOLT?
Each asset has its own logical channel network. Currently, one channel supports one asset. 

## 10. How does this compare to bitcoin lightning network?
OmniBOLT can be seen as an extension to the bitcoin lightning network. It focuses on transferring tokens quicker and cheaper. OmniBOLT not only focuses on bitcoin but also integrate other smart assets on lightning network. 

## 11. Does this support cross exchange of assets against bitcoin, or any token pair trading?
Yes. Using atomic swap technology, it is easy to build a dex on top of omnibolt and lnd network. It will enable the exchange of one cryptocurrency for another without using centralised intermediaries, such as exchanges. 

## 12. Will OmniBOLT support Wumbo channel?
yes, actually OmniBOLT doesn't has limits on channel size. 

 


