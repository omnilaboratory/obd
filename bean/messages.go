package bean

import (
	"github.com/omnilaboratory/obd/bean/chainhash"
	"github.com/omnilaboratory/obd/bean/enum"
)

var CurrObdNodeInfo obdNodeInfo

type obdNodeInfo struct {
	ChainNetworkType string `json:"chain_network_type"`
	OmniCoreVersion  string `json:"omni_core_version"`
	BtcCoreVersion   string `json:"btc_core_version"`
	TrackerNodeId    string `json:"tracker_node_id"`
	P2pAddress       string `json:"p2p_address"`
	WebsocketLink    string `json:"websocket_link"`
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
type typeLengthValue struct {
	ValueType string      `json:"value_type"`
	Length    int         `json:"length"`
	Value     interface{} `json:"value"`
}

// -100032
type SendChannelOpen struct {
	//充值的pubKey
	FundingPubKey      string `json:"funding_pubkey"`
	FunderAddressIndex int    `json:"funder_address_index"`
	IsPrivate          bool   `json:"is_private"` // channel is a private channel, can not use htlc hop
	typeLengthValue
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
	FundeeAddressIndex int    `json:"fundee_address_index"`
	TemporaryChannelId string `json:"temporary_channel_id"`
	Approval           bool   `json:"approval"`
	typeLengthValue
}

//type: -38 (close_channel)
type CloseChannel struct {
	ChannelId string `json:"channel_id"`
	typeLengthValue
}

//type: -39 (close_channel_sign)
type CloseChannelSign struct {
	ChannelId               string `json:"channel_id"`
	RequestCloseChannelHash string `json:"request_close_channel_hash"`
	Approval                bool   `json:"approval"` // true agree false disagree
	typeLengthValue
}

// type: -100340
type SendRequestFundingBtc struct {
	TemporaryChannelId string `json:"temporary_channel_id"`
	FundingTxHex       string `json:"funding_tx_hex"`
	typeLengthValue
}

// type: -340
// type: -110340
type FundingBtcOfP2p struct {
	TemporaryChannelId string                `json:"temporary_channel_id"`
	FundingTxid        string                `json:"funding_txid"`
	FundingBtcHex      string                `json:"funding_btc_hex"`
	FundingRedeemHex   string                `json:"funding_redeem_hex"`
	FunderNodeAddress  string                `json:"funder_node_address"`
	FunderPeerId       string                `json:"funder_peer_id"`
	SignData           NeedClientSignHexData `json:"sign_data"`
}

//type: -100350 (SendSignFundingBtc)
type SendSignFundingBtc struct {
	TemporaryChannelId              string `json:"temporary_channel_id"`
	FundingTxid                     string `json:"funding_txid"`
	SignedMinerRedeemTransactionHex string `json:"signed_miner_redeem_transaction_hex"`
	Approval                        bool   `json:"approval"`
	typeLengthValue
}

// type: -100034
type FundingAssetOfP2p struct {
	//临时通道id
	TemporaryChannelId string `json:"temporary_channel_id"`
	//充值hex
	FundingOmniHex        string                `json:"funding_omni_hex"`
	C1aRsmcHex            string                `json:"c1a_rsmc_hex"`
	RsmcTempAddressPubKey string                `json:"rsmc_temp_address_pub_key"`
	SignData              NeedClientSignHexData `json:"sign_data"`
	FunderNodeAddress     string                `json:"funder_node_address"`
	FunderPeerId          string                `json:"funder_peer_id"`
}

// -100034
type SendRequestAssetFunding struct {
	TemporaryChannelId string `json:"temporary_channel_id"`
	FundingTxHex       string `json:"funding_tx_hex"`
	TempAddressIndex   int    `json:"temp_address_index"`
	TempAddressPubKey  string `json:"temp_address_pub_key"`
	typeLengthValue
}

// -101034
type AliceSignC1aOfAssetFunding struct {
	SignedC1aHex string `json:"signed_c1a_hex"`
	typeLengthValue
}

//type: -100035 (funding_signed)
type SignAssetFunding struct {
	TemporaryChannelId string `json:"temporary_channel_id"`
	SignedAliceRsmcHex string `json:"signed_alice_rsmc_hex"`
	typeLengthValue
}

//type: -101035 (funding_signed)
type SignRdAndBrOfAssetFunding struct {
	TemporaryChannelId string `json:"temporary_channel_id"`
	BrId               int64  `json:"br_id"`
	RdSignedHex        string `json:"rd_signed_hex"`
	BrSignedHex        string `json:"br_signed_hex"`
	typeLengthValue
}

// -101134
type AliceSignRDOfAssetFunding struct {
	ChannelId   string `json:"channel_id"`
	RdSignedHex string `json:"rd_signed_hex"`
	typeLengthValue
}

//type -49: user wanna close htlc tx when tx is on getH state
type HtlcRequestCloseCurrTx struct {
	ChannelId                            string `json:"channel_id"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"` //	开通通道用到的私钥
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressIndex             int    `json:"curr_rsmc_temp_address_index"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`
	typeLengthValue
}

//type -50: receiver sign the close request
type HtlcSignCloseCurrTx struct {
	MsgHash                              string `json:"msg_hash"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"` //	开通通道用到的私钥
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressIndex             int    `json:"curr_rsmc_temp_address_index"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`
	typeLengthValue
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
	typeLengthValue
}

//type -81: MsgType_Atomic_Swap_Accept_N81
type AtomicSwapAccepted struct {
	AtomicSwapRequest
	TargetTransactionId string `json:"target_transaction_id"` // 针对的目标交易id
}

type NeedClientSignHexData struct {
	TemporaryChannelId string      `json:"temporary_channel_id"`
	Hex                string      `json:"hex"`
	Inputs             interface{} `json:"inputs"`
	IsMultisig         bool        `json:"is_multisig"`
	PubKeyA            string      `json:"pub_key_a"`
	PubKeyB            string      `json:"pub_key_b"`
	TotalInAmount      float64     `json:"total_in_amount"`
	TotalOutAmount     float64     `json:"total_out_amount"`
}

// 正式通道的需要客户端签名的信息体
type NeedClientSignRawTxData struct {
	Hex        string      `json:"hex"`
	Inputs     interface{} `json:"inputs"`
	IsMultisig bool        `json:"is_multisig"`
	PubKeyA    string      `json:"pub_key_a"`
	PubKeyB    string      `json:"pub_key_b"`
}
type NeedClientSignRawBRTxData struct {
	NeedClientSignRawTxData
	BrId int `json:"br_id"`
}
