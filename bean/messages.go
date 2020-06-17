package bean

import (
	"fmt"
	"github.com/omnilaboratory/obd/bean/chainhash"
	"github.com/omnilaboratory/obd/bean/enum"
	"time"
)

// 设置时间格式
const (
	timeFormat = "2006-01-02"
)

// 自定义类型
type JsonDate time.Time

// JsonDate反序列化
func (t *JsonDate) UnmarshalJSON(data []byte) (err error) {
	newTime, err := time.ParseInLocation("\""+timeFormat+"\"", string(data), time.Local)
	*t = JsonDate(newTime)
	return
}

// JsonDate序列化
func (t JsonDate) MarshalJSON() ([]byte, error) {
	timeStr := fmt.Sprintf("\"%s\"", time.Time(t).Format(timeFormat))
	return []byte(timeStr), nil
}

// string方法
func (t JsonDate) String() string {
	return time.Time(t).Format(timeFormat)
}

//obd客户端请求消息体
type RequestMessage struct {
	Type                enum.MsgType `json:"type"`
	SenderNodePeerId    string       `json:"sender_node_peer_id"`
	SenderUserPeerId    string       `json:"sender_user_peer_id"`
	RecipientUserPeerId string       `json:"recipient_user_peer_id"`
	RecipientNodePeerId string       `json:"recipient_node_peer_id"`
	Data                string       `json:"data"`
}

//obd答复消息体
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

//TLV 附加消息
type TypeLengthValue struct {
	ValueType string      `json:"value_type"`
	Length    int         `json:"length"`
	Value     interface{} `json:"value"`
}

// -100032
type SendChannelOpen struct {
	FundingPubKey string `json:"funding_pubkey"`
	TypeLengthValue
}

//https://github.com/obdlayer/Omni-BOLT-spec/blob/master/OmniBOLT-03-RSMC-and-OmniLayer-Transactions.md
// type = -32 请求方的obd发给接收方的obd的通道开通请求
// type = -110032 接收方obd发给接收方客户端的消息
type RequestOpenChannel struct {
	SendChannelOpen
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
	MaxAcceptedHtlcs         uint16              `json:"max_accepted_htlcs"` //最多可以接受多少给hltc请求 500
	FunderNodeAddress        string              `json:"funder_node_address"`
	FunderPeerId             string              `json:"funder_peer_id"`
	FundingAddress           string              `json:"funding_address"`
	RevocationBasePoint      chainhash.Point     `json:"revocation_base_point"`
	PaymentBasePoint         chainhash.Point     `json:"payment_base_point"`
	DelayedPaymentBasePoint  chainhash.Point     `json:"delayed_payment_base_point"`
	HtlcBasePoint            chainhash.Point     `json:"htlc_base_point"`
}

// -100033 接收方给自己的obd发送回复开通通道的请求
type SendSignOpenChannel struct {
	FundingPubKey      string `json:"funding_pubkey"`
	TemporaryChannelId string `json:"temporary_channel_id"`
	Approval           bool   `json:"approval"`
	TypeLengthValue
}

//type: -38 (close_channel)
type CloseChannel struct {
	ChannelId string `json:"channel_id"`
	TypeLengthValue
}

//type: -39 (close_channel_sign)
type CloseChannelSign struct {
	ChannelId               string `json:"channel_id"`
	RequestCloseChannelHash string `json:"request_close_channel_hash"`
	Approval                bool   `json:"approval"` // true agree false disagree
	TypeLengthValue
}

//type: -35107 (SendBreachRemedyTransaction)
type SendBreachRemedyTransaction struct {
	ChannelId                string `json:"channel_id"`
	ChannelAddressPrivateKey string `json:"channel_address_private_key"` // openChannel address
	TypeLengthValue
}

// type: -100340
type SendRequestFundingBtc struct {
	TemporaryChannelId       string `json:"temporary_channel_id"`
	FundingTxHex             string `json:"funding_tx_hex"`
	ChannelAddressPrivateKey string `json:"channel_address_private_key"`
	TypeLengthValue
}

// type: -340
// type: -110340
type FundingBtcOfP2p struct {
	TemporaryChannelId string `json:"temporary_channel_id"`
	FundingTxid        string `json:"funding_txid"`
	FundingBtcHex      string `json:"funding_btc_hex"`
	FundingRedeemHex   string `json:"funding_redeem_hex"`
	FunderNodeAddress  string `json:"funder_node_address"`
	FunderPeerId       string `json:"funder_peer_id"`
}

//type: -100350 (SendSignFundingBtc)
type SendSignFundingBtc struct {
	TemporaryChannelId       string `json:"temporary_channel_id"`
	FundingTxid              string `json:"funding_txid"`
	ChannelAddressPrivateKey string `json:"channel_address_private_key"`
	Approval                 bool   `json:"approval"`
	TypeLengthValue
}

// -100034
type SendRequestAssetFunding struct {
	TemporaryChannelId       string  `json:"temporary_channel_id"`
	PropertyId               int64   `json:"property_id"`
	MaxAssets                float64 `json:"max_assets"`
	AmountA                  float64 `json:"amount_a"`
	FundingTxHex             string  `json:"funding_tx_hex"`
	TempAddressPubKey        string  `json:"temp_address_pub_key"`
	TempAddressPrivateKey    string  `json:"temp_address_private_key"`
	ChannelAddressPrivateKey string  `json:"channel_address_private_key"`
	TypeLengthValue
}

// type: -340
// type: -110034
type FundingAssetOfP2p struct {
	TemporaryChannelId    string `json:"temporary_channel_id"`
	FundingOmniHex        string `json:"funding_omni_hex"`
	C1aRsmcHex            string `json:"c1a_rsmc_hex"`
	RsmcTempAddressPubKey string `json:"rsmc_temp_address_pub_key"`
	FunderNodeAddress     string `json:"funder_node_address"`
	FunderPeerId          string `json:"funder_peer_id"`
}

//type: -35 (funding_signed)
type SignAssetFunding struct {
	TemporaryChannelId string `json:"temporary_channel_id"`
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
	//Approval    bool   `json:"approval"`
	TypeLengthValue
}

//type: -100351 (commitment_tx)
type SendRequestCommitmentTx struct {
	ChannelId                 string  `json:"channel_id"` //the global channel id.
	Amount                    float64 `json:"amount"`     //amount of the payment
	ChannelAddressPrivateKey  string  `json:"channel_address_private_key"`
	LastTempAddressPrivateKey string  `json:"last_temp_address_private_key"`
	CurrTempAddressPubKey     string  `json:"curr_temp_address_pub_key"`
	CurrTempAddressPrivateKey string  `json:"curr_temp_address_private_key"`
	TypeLengthValue
}

//p2p 351
type PayerRequestCommitmentTxOfP2p struct {
	ChannelId                 string  `json:"channel_id"` //the global channel id.
	CommitmentTxHash          string  `json:"commitment_tx_hash"`
	Amount                    float64 `json:"amount"` //amount of the payment
	ToCounterpartyTxHex       string  `json:"to_counterparty_tx_hex"`
	RsmcHex                   string  `json:"rsmc_hex"`
	LastTempAddressPrivateKey string  `json:"last_temp_address_private_key"`
	CurrTempAddressPubKey     string  `json:"curr_temp_address_pub_key"`
	PayerNodeAddress          string  `json:"payer_node_address"`
	PayerPeerId               string  `json:"payer_peer_id"`
}

// -110351
type PayerRequestCommitmentTxToBobClient struct {
	PayerRequestCommitmentTxOfP2p
	MsgHash string `json:"msg_hash"`
}

//type: -100352 (commitment_tx_signed)
type PayeeSendSignCommitmentTx struct {
	ChannelId                 string `json:"channel_id"`
	MsgHash                   string `json:"msg_hash"`
	ChannelAddressPrivateKey  string `json:"channel_address_private_key"`   // bob private key
	LastTempAddressPrivateKey string `json:"last_temp_address_private_key"` // bob2's private key
	CurrTempAddressPubKey     string `json:"curr_temp_address_pub_key"`     // bob3 or alice3
	CurrTempAddressPrivateKey string `json:"curr_temp_address_private_key"`
	Approval                  bool   `json:"approval"` // true agree false disagree
	TypeLengthValue
}

//p2p -100352 -> 353
type PayeeSignCommitmentTxOfP2p struct {
	ChannelId                 string `json:"channel_id"` //the global channel id.
	CommitmentTxHash          string `json:"commitment_tx_hash"`
	Approval                  bool   `json:"approval"`
	ToCounterpartyTxHex       string `json:"to_counterparty_tx_hex"`
	RsmcHex                   string `json:"rsmc_hex"`
	LastTempAddressPrivateKey string `json:"last_temp_address_private_key"`
	CurrTempAddressPubKey     string `json:"curr_temp_address_pub_key"`
	SignedRsmcHex             string `json:"signed_rsmc_hex"`
	SignedToCounterpartyTxHex string `json:"signed_to_counterparty_tx_hex"`
	PayerRdHex                string `json:"payer_rd_hex"`
	TypeLengthValue
}

//type -100402: invoice
type HtlcRequestInvoice struct {
	NetType             string   `json:"net_type"`               //解析用
	RecipientNodePeerId string   `json:"recipient_node_peer_id"` //解析用
	RecipientUserPeerId string   `json:"recipient_user_peer_id"` //解析用
	H                   string   `json:"h"`
	ExpiryTime          JsonDate `json:"expiry_time"`
	PropertyId          int64    `json:"property_id"`
	Amount              float64  `json:"amount"`
	Description         string   `json:"description"`
	TypeLengthValue
}

//type --100401: alice tell carl ,she wanna transfer some money to Carl
type HtlcRequestFindPath struct {
	Invoice string `json:"invoice"`
	TypeLengthValue
}

// type 40 payer start htlc tx
type AddHtlcRequest struct {
	PropertyId                           int64   `json:"property_id"`
	Amount                               float64 `json:"amount"`
	Memo                                 string  `json:"memo"`
	H                                    string  `json:"h"`
	CltvExpiry                           int     `json:"cltv_expiry"` //发起者设定的总的等待的区块个数
	RoutingPacket                        string  `json:"routing_packet"`
	ChannelAddressPrivateKey             string  `json:"channel_address_private_key"`                 //	开通通道用到的地址的私钥
	LastTempAddressPrivateKey            string  `json:"last_temp_address_private_key"`               //	上个RSMC委托交易用到的临时地址的私钥
	CurrRsmcTempAddressPubKey            string  `json:"curr_rsmc_temp_address_pub_key"`              //	创建Cnx中的toRsmc的部分使用的临时地址的公钥
	CurrRsmcTempAddressPrivateKey        string  `json:"curr_rsmc_temp_address_private_key"`          //	创建Cnx中的toRsmc的部分使用的临时地址的私钥
	CurrHtlcTempAddressPubKey            string  `json:"curr_htlc_temp_address_pub_key"`              //	创建Cnx中的toHtlc的部分使用的临时地址的公钥
	CurrHtlcTempAddressPrivateKey        string  `json:"curr_htlc_temp_address_private_key"`          //	创建Cnx中的toHtlc的部分使用的临时地址的私钥
	CurrHtlcTempAddressForHt1aPubKey     string  `json:"curr_htlc_temp_address_for_ht1a_pub_key"`     //	创建Ht1a中生成ht1a的输出的Rmsc的临时地址的公钥
	CurrHtlcTempAddressForHt1aPrivateKey string  `json:"curr_htlc_temp_address_for_ht1a_private_key"` //	创建Ht1a中生成ht1a的输出的Rmsc的临时地址的私钥
	TypeLengthValue
}

//type -100041: bob sign the request for the interNode
type HtlcSignGetH struct {
	AliceCommitmentTxHash         string `json:"alice_commitment_tx_hash"`
	ChannelAddressPrivateKey      string `json:"channel_address_private_key"`        //	开通通道用到的私钥
	LastTempAddressPrivateKey     string `json:"last_temp_address_private_key"`      //	上个RSMC委托交易用到的临时私钥
	CurrRsmcTempAddressPubKey     string `json:"curr_rsmc_temp_address_pub_key"`     //	创建Cnx中的toRsmc的部分使用的临时地址的公钥
	CurrRsmcTempAddressPrivateKey string `json:"curr_rsmc_temp_address_private_key"` //	创建Cnx中的toRsmc的部分使用的临时地址的私钥
	CurrHtlcTempAddressPubKey     string `json:"curr_htlc_temp_address_pub_key"`     //	创建Cnx中的toHtlc的部分使用的临时地址的公钥
	CurrHtlcTempAddressPrivateKey string `json:"curr_htlc_temp_address_private_key"` //	创建Cnx中的toHtlc的部分使用的临时地址的私钥
	TypeLengthValue
}

// -42 msg
type AfterBobSignAddHtlcToAlice struct {
	ChannelId                      string `json:"channel_id"`
	PayerCommitmentTxHash          string `json:"payer_commitment_tx_hash"`
	PayerSignedRsmcHex             string `json:"payer_signed_rsmc_hex"`
	PayerSignedToCounterpartyHex   string `json:"payer_signed_to_counterparty_hex"`
	PayerSignedHtlcHex             string `json:"payer_signed_htlc_hex"`
	PayerRsmcRdHex                 string `json:"payer_rsmc_rd_hex"`
	PayerLockByHForBobHex          string `json:"payer_lock_by_h_for_bob_hex"`
	PayerHt1aHex                   string `json:"payer_ht_1_a_hex"`
	PayeeLastTempAddressPrivateKey string `json:"payee_last_temp_address_private_key"`
	PayeeCurrRsmcTempAddressPubKey string `json:"payee_curr_rsmc_temp_address_pub_key"`
	PayeeCurrHtlcTempAddressPubKey string `json:"payee_curr_htlc_temp_address_pub_key"`
	PayeeCommitmentTxHash          string `json:"payee_commitment_tx_hash"`
	PayeeRsmcHex                   string `json:"payee_rsmc_hex"`
	PayeeToCounterpartyTxHex       string `json:"payee_to_counterparty_tx_hex"`
	PayeeHtlcHex                   string `json:"payee_htlc_hex"`
}

// -43 付款人签名收款人的承诺交易的三个hex及创建对应的子交易
type AfterAliceSignAddHtlcToBob struct {
	PayerCommitmentTxHash                 string `json:"payer_commitment_tx_hash"`
	PayerCurrHtlcTempAddressForHt1aPubKey string `json:"payer_curr_htlc_temp_address_for_ht1a_pub_key"`
	PayerHt1aSignedHex                    string `json:"payer_ht1a_signed_hex"`
	PayeeCommitmentTxHash                 string `json:"payee_commitment_tx_hash"`
	PayeeSignedRsmcHex                    string `json:"payee_signed_rsmc_hex"`
	PayeeRsmcRdHex                        string `json:"payee_rsmc_rd_hex"`
	PayeeSignedToCounterpartyHex          string `json:"payee_signed_to_counterparty_hex"`
	PayeeSignedHtlcHex                    string `json:"payee_signed_htlc_hex"`
	PayeeHtd1bHex                         string `json:"payee_htd1b_hex"`
	PayeeHlockHex                         string `json:"payee_hlock_hex"`
}

// -44 收款人更加签名后的ht1a，创建这个交易的RD
type PayeeCreateHt1aRDForPayer struct {
	PayerCommitmentTxHash string `json:"payer_commitment_tx_hash"`
	PayerHt1aRDHex        string `json:"payer_ht1a_rd_hex"`
}

//type -45: Send R to previous node. and create commitment transactions.
type HtlcSendR struct {
	ChannelId                            string `json:"channel_id"`
	R                                    string `json:"r"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"`             // The key of Sender. Example Bob send R to Alice, the Sender is Bob.
	CurrHtlcTempAddressForHE1bPubKey     string `json:"curr_htlc_temp_address_for_he1b_pub_key"` // These keys of HE1b output. Example Bob send R to Alice, these is Bob3's.
	CurrHtlcTempAddressForHE1bPrivateKey string `json:"curr_htlc_temp_address_for_he1b_private_key"`
	TypeLengthValue
}

//type -46: Middleman node check out if R is correct
type HtlcCheckRAndCreateTx struct {
	ChannelId                string `json:"channel_id"`
	R                        string `json:"r"`
	MsgHash                  string `json:"msg_hash"`
	ChannelAddressPrivateKey string `json:"channel_address_private_key"` // The key of creator tx. Example Bob send R to Alice, that is Alice's.
	TypeLengthValue
}

// -47
type HtlcRPayerVerifyRInfoToPayee struct {
	ChannelId            string `json:"channel_id"`
	PayerHlockTxHex      string `json:"payer_hlock_tx_hex"`
	PayerHed1aHex        string `json:"payer_hed1a_hex"`
	PayeeSignedHerd1bHex string `json:"payee_signed_herd1b_hex"`
}

// -48
type HtlcRPayeeSignHed1aToPayer struct {
	ChannelId           string `json:"channel_id"`
	PayerSignedHed1aHex string `json:"payer_signed_hed1a_hex"`
}

//type -49: user wanna close htlc tx when tx is on getH state
type HtlcRequestCloseCurrTx struct {
	ChannelId                            string `json:"channel_id"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"` //	开通通道用到的私钥
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`
	TypeLengthValue
}

//type -50: receiver sign the close request
type HtlcSignCloseCurrTx struct {
	MsgHash                              string `json:"msg_hash"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"` //	开通通道用到的私钥
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`
	TypeLengthValue
}

//50->51
type HtlcCloseCloseeSignedInfoToCloser struct {
	ChannelId                                  string `json:"channel_id"`
	CloseeLastRsmcTempAddressPrivateKey        string `json:"closee_last_rsmc_temp_address_private_key"`
	CloseeLastHtlcTempAddressPrivateKey        string `json:"closee_last_htlc_temp_address_private_key"`
	CloseeLastHtlcTempAddressForHtnxPrivateKey string `json:"closee_last_htlc_temp_address_for_htnx_private_key"`
	CloseeCurrRsmcTempAddressPubKey            string `json:"closee_curr_rsmc_temp_address_pub_key"`
	CloseeRsmcHex                              string `json:"closee_rsmc_hex"`
	CloseeToCounterpartyTxHex                  string `json:"closee_to_counterparty_tx_hex"`
	CloserCommitmentTxHash                     string `json:"closer_commitment_tx_hash"`
	CloserSignedRsmcHex                        string `json:"closer_signed_rsmc_hex"`
	CloserRsmcRdHex                            string `json:"closer_rsmc_rd_hex"`
	CloserSignedToCounterpartyTxHex            string `json:"closer_signed_to_counterparty_tx_hex"`
}

//51->52
type HtlcCloseCloserSignTxInfoToClosee struct {
	ChannelId                       string `json:"channel_id"`
	CloseeSignedRsmcHex             string `json:"closee_signed_rsmc_hex"`
	CloseeRsmcRdHex                 string `json:"closee_rsmc_rd_hex"`
	CloseeSignedToCounterpartyTxHex string `json:"closee_signed_to_counterparty_tx_hex"`
}

// 付款人的obd发给收款人的obd的消息体 在请求htlc交易 40->41
type AliceRequestAddHtlc struct {
	ChannelId                        string  `json:"channel_id"`
	Amount                           float64 `json:"amount"`
	RoutingPacket                    string  `json:"routing_packet"`
	CltvExpiry                       int     `json:"cltv_expiry"` //发起者设定的总的等待的区块个数
	PayerCommitmentTxHash            string  `json:"payer_commitment_tx_hash"`
	Memo                             string  `json:"memo"`
	H                                string  `json:"h"`
	ToCounterpartyTxHex              string  `json:"to_counterparty_tx_hex"`
	HtlcTxHex                        string  `json:"htlc_tx_hex"`
	RsmcTxHex                        string  `json:"rsmc_tx_hex"`
	LastTempAddressPrivateKey        string  `json:"last_temp_address_private_key"`
	CurrRsmcTempAddressPubKey        string  `json:"curr_rsmc_temp_address_pub_key"`
	CurrHtlcTempAddressPubKey        string  `json:"curr_htlc_temp_address_pub_key"`
	CurrHtlcTempAddressForHt1aPubKey string  `json:"curr_htlc_temp_address_for_ht1a_pub_key"`
	PayerNodeAddress                 string  `json:"payer_node_address"`
	PayerPeerId                      string  `json:"payer_peer_id"`
}

//  p2p消息 收款人的obd发给付款人的obd的消息体 在获得R后
// -100045 -> -45
type BobSendROfP2p struct {
	ChannelId        string `json:"channel_id"`
	R                string `json:"r"`
	He1bTxHex        string `json:"he1b_tx_hex"`
	He1bTempPubKey   string `json:"he1b_temp_pub_key"`
	Herd1bTxHex      string `json:"herd1b_tx_hex"`
	PayeeNodeAddress string `json:"payee_node_address"`
	PayeePeerId      string `json:"payee_peer_id"`
}

// ws消息 收款人的obd发给付款人的obd的消息体 在获得R后
type BobSendROfWs struct {
	BobSendROfP2p
	MsgHash string `json:"msg_hash"`
}

//  p2p消息 请求关闭htlc交易
type RequestCloseHtlcTxOfP2p struct {
	ChannelId                            string `json:"channel_id"`
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	RsmcHex                              string `json:"rsmc_hex"`
	ToCounterpartyTxHex                  string `json:"to_counterparty_tx_hex"`
	CommitmentTxHash                     string `json:"commitment_tx_hash"`
	SenderNodeAddress                    string `json:"sender_node_address"`
	SenderPeerId                         string `json:"sender_peer_id"`
}

// ws消息 收款人的obd发给付款人的obd的消息体 在获得R后
// -110050
type RequestCloseHtlcTxOfWs struct {
	RequestCloseHtlcTxOfP2p
	MsgHash string `json:"msg_hash"`
}

//type -80: MsgType_Atomic_Swap_80
type AtomicSwapRequest struct {
	ChannelIdFrom       string  `json:"channel_id_from"`
	ChannelIdTo         string  `json:"channel_id_to"`
	RecipientUserPeerId string  `json:"recipient_user_peer_id"`
	PropertySent        int64   `json:"property_sent"`
	Amount              float64 `json:"amount"`
	ExchangeRate        float64 `json:"exchange_rate"`
	PropertyReceived    int64   `json:"property_received"`
	TransactionId       string  `json:"transaction_id"`
	TimeLocker          uint32  `json:"time_locker"`
	TypeLengthValue
}

//type -81: MsgType_Atomic_Swap_Accept_N81
type AtomicSwapAccepted struct {
	AtomicSwapRequest
	TargetTransactionId string `json:"target_transaction_id"` // 针对的目标交易id
}
