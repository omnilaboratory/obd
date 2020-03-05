package bean

import (
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/bean/enum"
	"github.com/tyler-smith/go-bip32"
)

type RequestMessage struct {
	Type            enum.MsgType `json:"type"`
	SenderPeerId    string       `json:"sender_peer_id"`
	RecipientPeerId string       `json:"recipient_peer_id"`
	Data            string       `json:"data"`
	PubKey          string       `json:"pub_key"`
	Signature       string       `json:"signature"`
}
type ReplyMessage struct {
	Type   enum.MsgType `json:"type"`
	Status bool         `json:"status"`
	From   string       `json:"from"`
	To     string       `json:"to"`
	Result interface{}  `json:"result"`
}

type UserState int

const (
	UserState_ErrorState UserState = -1
	UserState_Offline    UserState = 0
	UserState_OnLine     UserState = 1
)

//type = 1
type User struct {
	PeerId        string    `json:"peer_id"`
	Mnemonic      string    `json:"mnemonic"`
	State         UserState `json:"state"`
	ChangeExtKey  *bip32.Key
	CurrAddrIndex int `json:"curr_addr_index"`
}

//https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/OmniBOLT-03-RSMC-and-OmniLayer-Transactions.md
//type = -32
type OpenChannelInfo struct {
	ChainHash                chainhash.ChainHash `json:"chain_hash"`
	TemporaryChannelId       string              `json:"temporary_channel_id"`
	FundingSatoshis          uint64              `json:"funding_satoshis"`
	PushMsat                 uint64              `json:"push_msat"`
	DustLimitSatoshis        uint64              `json:"dust_limit_satoshis"`
	MaxHtlcValueInFlightMsat uint64              `json:"max_htlc_value_in_flight_msat"`
	ChannelReserveSatoshis   uint64              `json:"channel_reserve_satoshis"`
	HtlcMinimumMsat          uint64              `json:"htlc_minimum_msat"`
	FeeRatePerKw             uint32              `json:"fee_rate_per_kw"`
	ToSelfDelay              uint16              `json:"to_self_delay"`
	MaxAcceptedHtlcs         uint16              `json:"max_accepted_htlcs"`
	FundingPubKey            string              `json:"funding_pubkey"`
	FundingAddress           string              `json:"funding_address"`
	RevocationBasePoint      chainhash.Point     `json:"revocation_base_point"`
	PaymentBasePoint         chainhash.Point     `json:"payment_base_point"`
	DelayedPaymentBasePoint  chainhash.Point     `json:"delayed_payment_base_point"`
	HtlcBasePoint            chainhash.Point     `json:"htlc_base_point"`
}

//type = -33
type AcceptChannelInfo struct {
	OpenChannelInfo
	Approval bool `json:"approval"`
}

//type: -38 (close_channel)
type CloseChannel struct {
	ChannelId string `json:"channel_id"`
}

//type: -39 (close_channel_sign)
type CloseChannelSign struct {
	ChannelId               string `json:"channel_id"`
	RequestCloseChannelHash string `json:"request_close_channel_hash"`
	Approval                bool   `json:"approval"` // true agree false disagree
}

//type: -35107 (SendBreachRemedyTransaction)
type SendBreachRemedyTransaction struct {
	ChannelId                string `json:"channel_id"`
	ChannelAddressPrivateKey string `json:"channel_address_private_key"` // openChannel address
}

//type: -34 (funding_created)
type FundingCreated struct {
	TemporaryChannelId       string  `json:"temporary_channel_id"`
	PropertyId               int64   `json:"property_id"`
	MaxAssets                float64 `json:"max_assets"`
	AmountA                  float64 `json:"amount_a"`
	FundingTxHex             string  `json:"funding_tx_hex"`
	TempAddressPubKey        string  `json:"temp_address_pub_key"`
	TempAddressPrivateKey    string  `json:"temp_address_private_key"`
	ChannelAddressPrivateKey string  `json:"channel_address_private_key"`
}

//type: -3400 (FundingBtcCreated)
type FundingBtcCreated struct {
	TemporaryChannelId       string  `json:"temporary_channel_id"`
	Amount                   float64 `json:"amount"`
	FundingTxHex             string  `json:"funding_tx_hex"`
	ChannelAddressPrivateKey string  `json:"channel_address_private_key"`
}

//type: -3500 (FundingBtcSigned)
type FundingBtcSigned struct {
	TemporaryChannelId       string `json:"temporary_channel_id"`
	FundingTxid              string `json:"funding_txid"`
	ChannelAddressPrivateKey string `json:"channel_address_private_key"`
	Approval                 bool   `json:"approval"`
}

//type: -35 (funding_signed)
type FundingSigned struct {
	ChannelId string `json:"channel_id"`
	//the omni address of funder Alice
	FunderPubKey string `json:"funder_pub_key"`
	// the id of the Omni asset
	PropertyId int `json:"property_id"`
	//amount of the asset on Alice side
	AmountA float64 `json:"amount_a"`
	//the omni address of fundee Bob
	FundeePubKey string `json:"fundee_pub_key"`
	//amount of the asset on Bob side
	AmountB float64 `json:"amount_b"`
	//signature of fundee Bob
	FundeeChannelAddressPrivateKey string `json:"fundee_channel_address_private_key"`
	//redeem script used to generate P2SH address
	RedeemScript string `json:"redeem_script"`
	//hash of redeemScript
	P2shAddress string `json:"p2sh_address"`
	Approval    bool   `json:"approval"`
}

//type: -351 (commitment_tx)
type CommitmentTx struct {
	ChannelId                 string  `json:"channel_id"` //the global channel id.
	RequestCommitmentHash     string  `json:"request_commitment_hash"`
	PropertyId                int     `json:"property_id"` //the id of the Omni asset
	Amount                    float64 `json:"amount"`      //amount of the payment
	ChannelAddressPrivateKey  string  `json:"channel_address_private_key"`
	LastTempAddressPrivateKey string  `json:"last_temp_address_private_key"`
	CurrTempAddressPubKey     string  `json:"curr_temp_address_pub_key"`
	CurrTempAddressPrivateKey string  `json:"curr_temp_address_private_key"`
}

//type: -352 (commitment_tx_signed)
type CommitmentTxSigned struct {
	ChannelId                 string `json:"channel_id"`
	RequestCommitmentHash     string `json:"request_commitment_hash"`
	ChannelAddressPrivateKey  string `json:"channel_address_private_key"`   // bob private key
	LastTempAddressPrivateKey string `json:"last_temp_address_private_key"` // bob2's private key
	CurrTempAddressPubKey     string `json:"curr_temp_address_pub_key"`     // bob3 or alice3
	CurrTempAddressPrivateKey string `json:"curr_temp_address_private_key"`
	Approval                  bool   `json:"approval"` // true agree false disagree
}

//type: -353 (get_balance_request)
type GetBalanceRequest struct {
	//the global channel id.
	ChannelId string `json:"channel_id"`
	//the p2sh address generated in funding_signed message.
	P2shAddress string `json:"p2sh_address"`
	// the channel owner, Alice or Bob, can query the balance.
	Who chainhash.Hash `json:"who"`
	//the signature of Alice or Bob
	Signature chainhash.Signature `json:"signature"`
}

//type: -354 (get_balance_respond)
type GetBalanceRespond struct {
	//the global channel id.
	ChannelId string `json:"channel_id"`
	//the asset id generated by Omnilayer protocol.
	PropertyId int `json:"property_id"`
	//the name of the asset.
	Name string `json:"name"`
	//balance in this channel
	Balance float64 `json:"balance"`
	//currently not in use
	Reserved float64 `json:"reserved"`
	//currently not in use
	Frozen float64 `json:"frozen"`
}

//type -40: alice tell carl ,she wanna transfer some money to Carl
type HtlcHRequest struct {
	PropertyId      int     `json:"property_id"`
	Amount          float64 `json:"amount"`
	RecipientPeerId string  `json:"recipient_peer_id"`
}

//type -41: carl tell alice the H,and he ca
type HtlcHRespond struct {
	RequestHash string  `json:"request_hash"`
	Approval    bool    `json:"approval"` // true agree false disagree
	PropertyId  int     `json:"property_id"`
	Amount      float64 `json:"amount"`
	H           string  `json:"h"` // pubKey
	R           string  `json:"r"` // privateKey
}

//type -42: alice request ask bob whether agree to be the interNode
type HtlcRequestFindPathAndSendH struct {
	H               string  `json:"h"`
	RecipientPeerId string  `json:"recipient_peer_id"`
	PropertyId      int     `json:"property_id"`
	Amount          float64 `json:"amount"`
	Memo            string  `json:"memo"`
}

//type -43: bob send h to carl
type HtlcSendH struct {
	H                    string `json:"h"`
	HAndRInfoRequestHash string `json:"h_and_r_info_request_hash"`
}

//type -44: bob sign the request for the interNode
type HtlcSignGetH struct {
	RequestHash                      string `json:"request_hash"`
	Approval                         bool   `json:"approval"`                                // true agree false disagree ,最后的收款节点，必须是true
	ChannelAddressPrivateKey         string `json:"channel_address_private_key"`             //	开通通道用到的私钥
	LastTempAddressPrivateKey        string `json:"last_temp_address_private_key"`           //	上个RSMC委托交易用到的临时私钥
	CurrRsmcTempAddressPubKey        string `json:"curr_rsmc_temp_address_pub_key"`          //	创建Cnx中的toRsmc的部分使用的临时地址的公钥
	CurrRsmcTempAddressPrivateKey    string `json:"curr_rsmc_temp_address_private_key"`      //	创建Cnx中的toRsmc的部分使用的临时地址的私钥
	CurrHtlcTempAddressPubKey        string `json:"curr_htlc_temp_address_pub_key"`          //	创建Cnx中的toHtlc的部分使用的临时地址的公钥
	CurrHtlcTempAddressPrivateKey    string `json:"curr_htlc_temp_address_private_key"`      //	创建Cnx中的toHtlc的部分使用的临时地址的私钥
	CurrHtlcTempAddressHe1bOfHPubKey string `json:"curr_htlc_temp_address_he1b_ofh_pub_key"` //	创建he1b,锁住支付资金的三签地址临时地址的公钥
}

//type -45: sender request obd  to open htlc tx
type HtlcRequestOpen struct {
	RequestHash                          string `json:"request_hash"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"`                  //	开通通道用到的私钥
	LastTempAddressPrivateKey            string `json:"last_temp_address_private_key"`                //	上个RSMC委托交易用到的临时地址的私钥
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`               //	创建Cnx中的toRsmc的部分使用的临时地址的公钥
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`           //	创建Cnx中的toRsmc的部分使用的临时地址的私钥
	CurrHtlcTempAddressPubKey            string `json:"curr_htlc_temp_address_pub_key"`               //	创建Cnx中的toHtlc的部分使用的临时地址的公钥
	CurrHtlcTempAddressPrivateKey        string `json:"curr_htlc_temp_address_private_key"`           //	创建Cnx中的toHtlc的部分使用的临时地址的私钥
	CurrHtlcTempAddressForHt1aPubKey     string `json:"curr_htlc_temp_address_for_ht1a_pub_key"`      //	创建Ht1a中生成ht1a的输出的Rmsc的临时地址的公钥
	CurrHtlcTempAddressForHt1aPrivateKey string `json:"curr_htlc_temp_address_for_ht1a_private_key"`  //	创建Ht1a中生成ht1a的输出的Rmsc的临时地址的私钥
	CurrHtlcTempAddressForHed1aOfHPubKey string `json:"curr_htlc_temp_address_for_hed1a_ofh_pub_key"` //	创建hed1a,锁住支付资金的三签地址临时地址的公钥
}

//type -46: Send R to previous node.
type HtlcSendR struct {
	RequestHash string `json:"request_hash"`
	R           string `json:"r"`

	// The key of Sender. Example Bob send R to Alice, the Sender is Bob.
	ChannelAddressPrivateKey string `json:"channel_address_private_key"`

	// The key of Cnb NO.3 output. Example Bob send R to Alice, that is Bob2's.
	CurrHtlcTempAddressHe1bOfHPrivateKey string `json:"curr_htlc_temp_address_he1b_ofh_private_key"`

	// These keys of HE1b output. Example Bob send R to Alice, these is Bob3's.
	CurrHtlcTempAddressForHE1bPubKey     string `json:"curr_htlc_temp_address_for_he1b_pub_key"`
	CurrHtlcTempAddressForHE1bPrivateKey string `json:"curr_htlc_temp_address_for_he1b_private_key"`
}

//type -47: Middleman node check out if R is correct
// and create commitment transactions.
type HtlcCheckRAndCreateTx struct {
	RequestHash string `json:"request_hash"`
	R           string `json:"r"`

	// The key of creator tx. Example Bob send R to Alice, that is Alice's.
	ChannelAddressPrivateKey string `json:"channel_address_private_key"`

	// The key of Cna NO.3 output. Example Bob send R to Alice, that is Alice2's.
	CurrHtlcTempAddressForHed1aOfHPrivateKey string `json:"curr_htlc_temp_address_for_hed1a_ofh_private_key"`
}

//type -48: user wanna close htlc tx when tx is on getH state
type HtlcRequestCloseCurrTx struct {
	ChannelId                            string `json:"channel_id"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"` //	开通通道用到的私钥
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`
}

//type -49: receiver sign the close request
type HtlcSignCloseCurrTx struct {
	RequestCloseHtlcHash                 string `json:"request_close_htlc_hash"`     //-48的操作，临时存储一下
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"` //	开通通道用到的私钥
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`
}

type HtlcCloseChannelReq CloseChannel
type HtlcCloseChannelSign CloseChannelSign
type ChannelIdReq CloseChannel

//type -80: MsgType_Atomic_Swap_N80
type AtomicSwapRequest struct {
	ChannelIdFrom    string  `json:"channel_id_from"`
	ChannelIdTo      string  `json:"channel_id_to"`
	RecipientPeerId  string  `json:"recipient_peer_id"`
	PropertySent     int64   `json:"property_sent"`
	Amount           float64 `json:"amount"`
	ExchangeRate     float64 `json:"exchange_rate"`
	PropertyReceived int64   `json:"property_received"`
	TransactionId    string  `json:"transaction_id"`
	TimeLocker       uint32  `json:"time_locker"`
}

//type -81: MsgType_Atomic_Swap_Accept_N81
type AtomicSwapAccepted struct {
	AtomicSwapRequest
	TargetTransactionId string `json:"target_transaction_id"` // 针对的目标交易id
}
