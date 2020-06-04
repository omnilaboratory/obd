package dao

import (
	"github.com/omnilaboratory/obd/bean"
	"time"
)

type User struct {
	Id int `storm:"id,increment" `
	bean.User
	CurrState       int       `json:"curr_state"`
	CreateAt        time.Time `json:"create_at"`
	LatestLoginTime time.Time `json:"latest_login_time"`
}
type UserLoginLog struct {
	Id       int       `storm:"id,increment" `
	PeerId   string    `json:"peer_id"`
	LoginAt  time.Time `json:"login_at"`
	LogoutAt time.Time `json:"logout_at"`
}

type ChannelState int

const (
	ChannelState_Create            ChannelState = 10
	ChannelState_WaitFundAsset     ChannelState = 11
	ChannelState_CanUse            ChannelState = 20
	ChannelState_Close             ChannelState = 21
	ChannelState_HtlcTx            ChannelState = 22
	ChannelState_OpenChannelRefuse ChannelState = 30
)

type ChannelInfo struct {
	bean.OpenChannelInfo
	Id                         int          `storm:"id,increment" json:"id"`
	ChannelId                  string       `json:"channel_id"`
	PeerIdA                    string       `storm:"index" json:"peer_id_a"`
	PubKeyA                    string       `json:"pub_key_a"`
	AddressA                   string       `json:"address_a"`
	PeerIdB                    string       `storm:"index" json:"peer_id_b"`
	PubKeyB                    string       `json:"pub_key_b"`
	AddressB                   string       `json:"address_b"`
	ChannelAddress             string       `json:"channel_address"`
	ChannelAddressRedeemScript string       `json:"channel_address_redeem_script"`
	ChannelAddressScriptPubKey string       `json:"channel_address_script_pub_key"`
	PropertyId                 int64        `json:"property_id"`
	CurrState                  ChannelState `json:"curr_state"`
	RefuseReason               string       `json:"refuse_reason"`
	CreateBy                   string       `json:"create_by"`
	CreateAt                   time.Time    `json:"create_at"`
	AcceptAt                   time.Time    `json:"accept_at"`
	CloseAt                    time.Time    `json:"close_at"`
}

type CloseChannel struct {
	bean.CloseChannel
	Id             int       `storm:"id,increment" json:"id"`
	CommitmentTxId int       `json:"commitment_tx_id"`
	RequestHex     string    `json:"request_hex"`
	Owner          string    `json:"owner"`
	Approval       bool      `json:"approval"`
	CurrState      int       `json:"curr_state"` // 0: create 1 finish
	CreateAt       time.Time `json:"create_at"`
}

type FundingTransactionState int

const (
	FundingTransactionState_Create FundingTransactionState = 10
	FundingTransactionState_Accept FundingTransactionState = 20
	FundingTransactionState_Defuse FundingTransactionState = 30
)

type FundingTransaction struct {
	Id                         int                     `storm:"id,increment" json:"id"`
	TemporaryChannelId         string                  `json:"temporary_channel_id"`
	ChannelInfoId              int                     `json:"channel_info_id"`
	CurrState                  FundingTransactionState `json:"curr_state"`
	PeerIdA                    string                  `json:"peer_id_a"`
	PeerIdB                    string                  `json:"peer_id_b"`
	ChannelId                  string                  `json:"channel_id"`
	PropertyId                 int64                   `json:"property_id"`
	AmountA                    float64                 `json:"amount_a"`
	FunderPubKey2ForCommitment string                  `json:"funder_pub_key_2_for_commitment"`
	FundingTxHex               string                  `json:"funding_tx_hex"`
	FundingTxid                string                  `json:"funding_txid"`
	FundingOutputIndex         uint32                  `json:"funding_output_index"`
	AmountB                    float64                 `json:"amount_b"`
	RsmcTempAddressPubKey      string                  `json:"rsmc_temp_address_pub_key"` //为bob准备的
	FunderRsmcHex              string                  `json:"funder_rsmc_hex"`
	CreateBy                   string                  `json:"create_by"`
	Owner                      string                  `json:"owner"`
	FunderAddress              string                  `json:"funder_address"`
	CreateAt                   time.Time               `json:"create_at"`
	FundeeSignAt               time.Time               `json:"fundee_sign_at"`
}

type TxInfoState int

const (
	TxInfoState_Create              TxInfoState = 5
	TxInfoState_Htlc_WaitHTRD1aSign TxInfoState = 6 //等待bob创建htrd1a
	TxInfoState_CreateAndSign       TxInfoState = 10
	TxInfoState_Htlc_GetH           TxInfoState = 11 // 创建Htlc交易的时候的状态
	TxInfoState_Htlc_GetR           TxInfoState = 12 // 获取到R后的状态
	TxInfoState_SendHex             TxInfoState = 20
	TxInfoState_Abord               TxInfoState = 30
)

type FundingBtcRequest struct {
	Id                 int       `storm:"id,increment" json:"id" `
	Owner              string    `json:"owner"`
	TemporaryChannelId string    `json:"temporary_channel_id"`
	TxHash             string    `json:"tx_hash"`
	RedeemHex          string    `json:"redeem_hex"`
	TxId               string    `json:"tx_id"`
	Amount             float64   `json:"amount"`
	CreateAt           time.Time `json:"create_at"`
	SignAt             time.Time `json:"sign_at"`
	SignApproval       bool      `json:"sign_approval"`
	FinishAt           time.Time `json:"finish_at"`
	IsFinish           bool      `json:"is_finish"`
}

//redeem the btc fee
type MinerFeeRedeemTransaction struct {
	Id                 int       `storm:"id,increment" json:"id" `
	Owner              string    `json:"owner"`
	TemporaryChannelId string    `json:"temporary_channel_id"`
	FundingTxId        string    `json:"funding_tx_id"`
	Hex                string    `json:"hex"`
	Txid               string    `json:"txid"`
	CreateAt           time.Time `json:"create_at"`
}

type CommitmentTxRequestInfo struct {
	Id int `storm:"id,increment" json:"id" `
	bean.CommitmentTx
	ChannelId             string
	UserId                string
	LastTempAddressPubKey string
	CreateAt              time.Time
	IsEnable              bool
}

type CommitmentTransactionType int

const (
	CommitmentTransactionType_Rsmc = 0
	CommitmentTransactionType_Htlc = 1
)

//CommitmentTransaction
type CommitmentTransaction struct {
	Id                 int     `storm:"id,increment" json:"id" `
	LastCommitmentTxId int     `json:"last_commitment_tx_id"`
	LastHash           string  `json:"last_hash"`
	CurrHash           string  `json:"curr_hash"`
	PeerIdA            string  `json:"peer_id_a"`
	PeerIdB            string  `json:"peer_id_b"`
	ChannelId          string  `json:"channel_id"`
	PropertyId         int64   `json:"property_id"`
	InputTxid          string  `json:"input_txid"`   //input txid  from channelAddr: alice&bob multiAddr, so need  sign of alice and bob
	InputVout          uint32  `json:"input_vout"`   // input vout
	InputAmount        float64 `json:"input_amount"` //input amount

	TxType CommitmentTransactionType `json:"tx_type"` // 0 rsmc 1 htlc

	//RSMC
	RSMCTempAddressPubKey        string  `json:"rsmc_temp_address_pub_key"` //aliceTempRemc or bobTempRsmc
	RSMCMultiAddress             string  `json:"rsmc_multi_address"`        //output aliceTempRsmc&bob  or alice&bobTempRsmc  multiAddr
	RSMCRedeemScript             string  `json:"rsmc_redeem_script"`
	RSMCMultiAddressScriptPubKey string  `json:"rsmc_multi_address_script_pub_key"`
	AmountToRSMC                 float64 `json:"amount_to_rsmc"` // amount to multiAddr
	RsmcInputTxid                string  `json:"rsmc_input_txid"`
	RSMCTxHex                    string  `json:"rsmc_tx_hex"`
	RSMCTxid                     string  `json:"rsmc_txid"`
	//To CounterParty
	ToCounterpartyTxHex            string  `json:"to_counterparty_tx_hex"`
	ToCounterpartyTxid             string  `json:"to_counterparty_txid"`
	AmountToCounterparty           float64 `json:"amount_to_counterparty"`               //amount to bob(if Cna) or alice(if Cnb)
	FromCounterpartySideForMeTxHex string  `json:"from_counterparty_side_for_me_tx_hex"` //对方给自己的转账部分，防止对方不广播此交易
	//htlc
	HtlcRoutingPacket            string  `json:"htlc_routing_packet"`       //借道Path
	HtlcCltvExpiry               int     `json:"htlc_cltv_expiry"`          //借道的最大超时 分钟为单位
	BeginBlockHeight             int     `json:"begin_block_height"`        //借道时的区块高度
	HTLCTempAddressPubKey        string  `json:"htlc_temp_address_pub_key"` //alice for htlc or bob for htlc
	HTLCMultiAddress             string  `json:"htlc_multi_address"`        //output aliceTempHtlc&bob  or alice&bobTempHtlc  multiAddr
	HTLCRedeemScript             string  `json:"htlc_redeem_script"`
	HTLCMultiAddressScriptPubKey string  `json:"htlc_multi_address_script_pub_key"`
	AmountToHtlc                 float64 `json:"amount_to_htlc"`
	HtlcTxHex                    string  `json:"htlc_tx_hex"`
	HTLCTxid                     string  `json:"htlc_txid"`
	HtlcH                        string  `json:"htlc_h"`
	HtlcR                        string  `json:"htlc_r"`
	HtlcSender                   string  `json:"htlc_sender"`

	CurrState    TxInfoState `json:"curr_state"`
	CreateBy     string      `json:"create_by"`
	CreateAt     time.Time   `json:"create_at"`
	SignAt       time.Time   `json:"sign_at"`
	SendAt       time.Time   `json:"send_at"`
	LastEditTime time.Time   `json:"last_edit_time"`
	Owner        string      `json:"owner"`
}

// close channel , alice or bob wait 1000 sequence to drawback the balance
type RevocableDeliveryTransaction struct {
	Id             int         `storm:"id,increment" json:"id" `
	CommitmentTxId int         `json:"commitment_tx_id"`
	PeerIdA        string      `json:"peer_id_a"`
	PeerIdB        string      `json:"peer_id_b"`
	ChannelId      string      `json:"channel_id"`
	PropertyId     int64       `json:"property_id"`
	InputTxHex     string      `json:"input_tx_hex"`
	InputTxid      string      `json:"input_txid"`     //input txid  from commitTx alice2&bob multtaddr, so need  sign of alice2 and bob
	InputVout      uint32      `json:"input_vout"`     // input vout
	InputAmount    float64     `json:"input_amount"`   //input amount
	OutputAddress  string      `json:"output_address"` //output alice
	Sequence       int         `json:"sequence"`
	RDType         int         `json:"rd_type"` // default 0 for rsmc Rd,1 for htrd
	Amount         float64     `json:"amount"`  // output alice amount
	TxHex          string      `json:"tx_hex"`
	Txid           string      `json:"txid"`
	CurrState      TxInfoState `json:"curr_state"`
	CreateBy       string      `json:"create_by"`
	CreateAt       time.Time   `json:"create_at"`
	SignAt         time.Time   `json:"sign_at"`
	SendAt         time.Time   `json:"send_at"`
	LastEditTime   time.Time   `json:"last_edit_time"`
	Owner          string      `json:"owner"`
}

// rd tx of waiting 1000 sequence
type RDTxWaitingSend struct {
	Id                int       `storm:"id,increment" json:"id" `
	TransactionHex    string    `json:"transaction_hex"`
	Type              int       `json:"type"`                   // 0: RD 1000, 1:HT1a  2:htd1b
	HtnxIdAndHtnxRdId []int     `json:"htnx_id_and_htnx_rd_id"` // for ht1a later logic
	IsEnable          bool      `json:"is_enable"`
	CreateAt          time.Time `json:"create_at"`
	FinishAt          time.Time `json:"finish_at"`
}
type BRType int

const (
	BRType_Rmsc BRType = 0 //来源RSMC
	BRType_Htlc BRType = 1 //来源Htlc
	BRType_Ht1a BRType = 2 //来源Ht1a
	BRType_HE1b BRType = 3 //来源HE1b
)

// to punish alice do not admit the latest commitment tx
type BreachRemedyTransaction struct {
	Id                       int         `storm:"id,increment" json:"id" `
	CommitmentTxId           int         `json:"commitment_tx_id"` // parent commitmentTx id
	PeerIdA                  string      `json:"peer_id_a"`
	PeerIdB                  string      `json:"peer_id_b"`
	ChannelId                string      `json:"channel_id"`
	PropertyId               int64       `json:"property_id"`
	TempPubKey               string      `json:"temp_pub_key"`
	InputAddress             string      `json:"input_address"`
	InputAddressScriptPubKey string      `json:"input_address_script_pub_key"`
	InputRedeemScript        string      `json:"input_redeem_script"`
	InputTxHex               string      `json:"input_tx_hex"`
	InputTxid                string      `json:"input_txid"`   //input txid  from commitTx alice2&bob multtAddr, so need  sign of alice2 and bob
	InputVout                uint32      `json:"input_vout"`   // input vout
	InputAmount              float64     `json:"input_amount"` //input amount
	OutAddress               string      `json:"out_address"`
	Amount                   float64     `json:"amount"`    // output bob amount
	BrTxHex                  string      `json:"br_tx_hex"` // first alice2 sign
	Txid                     string      `json:"txid"`
	Type                     BRType      `json:"type"` //:0:rmsc ,1:htlc,2:ht1a 3:hebr1b
	CurrState                TxInfoState `json:"curr_state"`
	CreateBy                 string      `json:"create_by"`
	CreateAt                 time.Time   `json:"create_at"`
	SignAt                   time.Time   `json:"sign_at"`
	SendAt                   time.Time   `json:"send_at"`
	LastEditTime             time.Time   `json:"last_edit_time"`
	Owner                    string      `json:"owner"`
}

type AtomicSwapInfo struct {
	bean.AtomicSwapRequest
	Id           int       `storm:"id,increment" json:"id" `
	CreateBy     string    `json:"create_by"`
	CreateAt     time.Time `json:"create_at"`
	LatestEditAt time.Time `json:"latest_edit_at"`
}
type AtomicSwapAcceptedInfo struct {
	bean.AtomicSwapAccepted
	Id           int       `storm:"id,increment" json:"id" `
	CreateBy     string    `json:"create_by"`
	CreateAt     time.Time `json:"create_at"`
	LatestEditAt time.Time `json:"latest_edit_at"`
}
