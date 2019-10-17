package dao

import (
	"LightningOnOmni/bean"
	"time"
)

type HTLCCommitmentTransaction CommitmentTransaction

//HT1a 锁住发起方的交易资金：锁住的意思是：把资金放到一个临时多签帐号（alice1&bob1）
type HTLCTimeoutTxA struct {
	Id                           int            `storm:"id,increment" json:"id" `
	ChannelId                    bean.ChannelID `json:"channel_id"`
	CommitmentTxId               int            `json:"commitment_tx_id"`
	PropertyId                   int64          `json:"property_id"`
	InputHex                     string         `json:"input_hex"`
	Timeout                      int            `json:"timeout"` // if 3days 432=3*24*6
	CurrBlockHeight              int            `json:"curr_block_height"`
	RSMCTempAddressPubKey        string         `json:"rsmc_temp_address_pub_key"`
	RSMCMultiAddress             string         `json:"rsmc_multi_address"`
	RSMCRedeemScript             string         `json:"rsmc_redeem_script"`
	RSMCMultiAddressScriptPubKey string         `json:"rsmc_multi_address_script_pub_key"`
	RSMCOutAmount                float64        `json:"rsmc_out_amount"`
	RSMCTxHash                   string         `json:"rsmc_tx_hash"`
	RSMCTxid                     string         `json:"rsmc_txid"`
	CurrState                    TxInfoState    `json:"curr_state"`
	Owner                        string         `json:"owner"`
	CreateBy                     string         `json:"create_by"`
	CreateAt                     time.Time      `json:"create_at"`
	SignAt                       time.Time      `json:"sign_at"`
	SendAt                       time.Time      `json:"send_at"`
}

//HTD1b 锁住发起方的交易资金：如果bob广播了，如果过了三天，如果R没有获取到，alice可以通过广播拿回资金
type HTLCTimeoutDeliveryTxB struct {
	Id              int            `storm:"id,increment" json:"id" `
	ChannelId       bean.ChannelID `json:"channel_id"`
	CommitmentTxId  int            `json:"commitment_tx_id"`
	InputTxid       string         `json:"input_txid"` //input txid  from commitTx aliceTempHtlc&bob multtaddr, so need  sign of aliceTempHtlc and bob
	InputVout       uint32         `json:"input_vout"` // input vout
	Timeout         int            `json:"timeout"`
	CurrBlockHeight int            `json:"curr_block_height"`
	InputAmount     float64        `json:"input_amount"`   //input amount
	OutputAddress   string         `json:"output_address"` //output Sender Alice(if alice is sender) or Bob(if bob is sender)
	OutAmount       float64        `json:"out_amount"`
	TxHash          string         `json:"tx_hash"`
	Txid            string         `json:"txid"`
	CurrState       TxInfoState    `json:"curr_state"`
	Owner           string         `json:"owner"`
	CreateAt        string         `json:"create_at"`
	SendAt          time.Time      `json:"send_at"`
	CreateBy        time.Time      `json:"create_by"`
}

//HED1a 如果bob返回了正确的R，就可以完成签名，标识这次的htlc可以成功了
type HTLCExecutionDeliveryA struct {
	Id             int         `storm:"id,increment" json:"id" `
	CommitmentTxId int         `json:"commitment_tx_id"`
	InputTxid      string      `json:"input_txid"`   // input txid  from commitTx aliceTempHtlc&bob multtaddr, so need  sign of aliceTempHtlc and bob
	InputVout      uint32      `json:"input_vout"`   // input vout
	InputAmount    float64     `json:"input_amount"` // input amount
	HtlcR          string      `json:"htlc_r"`
	OutputAddress  string      `json:"output_address"` //to Bob
	OutAmount      float64     `json:"out_amount"`
	TxHash         string      `json:"tx_hash"`
	Txid           string      `json:"txid"`
	CurrState      TxInfoState `json:"curr_state"`
	Owner          string      `json:"owner"`
	IsEnable       bool        `json:"is_enable"`
	CreateAt       time.Time   `json:"create_at"`
	SendAt         time.Time   `json:"send_at"`
	CreateBy       time.Time   `json:"create_by"`
}

//HE1b 如果bob获得了正确的R，就可以完成签名，标识这次的htlc可以成功了
type HTLCExecutionB HTLCExecutionDeliveryA

type NormalState int

const (
	NS_Create NormalState = 10
	NS_Finish NormalState = 20
	NS_Refuse NormalState = 30
)

type HtlcCreateRandHInfo struct {
	Id              int         `storm:"id,increment" json:"id" `
	SenderPeerId    string      `json:"sender_peer_id"`
	RecipientPeerId string      `json:"recipient_peer_id"`
	PropertyId      int         `json:"property_id"`
	Amount          float64     `json:"amount"`
	R               string      `json:"r"`
	H               string      `json:"h"`
	CurrState       NormalState `json:"curr_state"`
	RequestHash     string      `json:"request_hash"`
	CreateBy        string      `json:"create_by"`
	CreateAt        time.Time   `json:"create_at"`
	SignAt          time.Time   `json:"sign_at"`
	SignBy          string      `json:"sign_by"`
}

type HtlcSingleHopTxBaseInfo struct {
	Id                             int         `storm:"id,increment" json:"id" `
	HtlcCreateRandHInfoRequestHash string      `json:"htlc_create_rand_h_info_request_hash"`
	FirstChannelId                 int         `json:"alice_channel_id"`
	InterNodePeerId                string      `json:"inter_node_peer_id"`
	SecondChannelId                int         `json:"second_channel_id"`
	BobCurrRsmcTempPubKey          string      `json:"bob_curr_rsmc_temp_pub_key"`
	BobCurrHtlcTempPubKey          string      `json:"bob_curr_htlc_temp_pub_key"`
	BobCurrHtlcTempForHt1bPubKey   string      `json:"bob_curr_htlc_temp_for_ht1b_pub_key"`
	CurrState                      NormalState `json:"curr_state"`
	CreateBy                       string      `json:"create_by"`
	CreateAt                       time.Time   `json:"create_at"`
	SignAt                         time.Time   `json:"sign_at"`
	SignBy                         string      `json:"sign_by"`
}
