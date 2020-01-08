package dao

import (
	"LightningOnOmni/bean"
	"time"
)

type HTLCCommitmentTransaction CommitmentTransaction

//HT1a 锁住发起方的交易资金：锁住的意思是：把资金放到一个临时多签帐号（alice1&bob）
//HE1b 等获取到R后，锁住接收方的交易资金：锁住的意思是：把资金放到一个临时多签帐号（alice&bob6）
type HTLCTimeoutTxForAAndExecutionForB struct {
	Id                           int            `storm:"id,increment" json:"id" `
	ChannelId                    bean.ChannelID `json:"channel_id"`
	CommitmentTxId               int            `json:"commitment_tx_id"`
	PropertyId                   int64          `json:"property_id"`
	InputHex                     string         `json:"input_hex"`
	Timeout                      int            `json:"timeout"` // if 3days 432=3*24*6
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
	Id             int            `storm:"id,increment" json:"id" `
	ChannelId      bean.ChannelID `json:"channel_id"`
	CommitmentTxId int            `json:"commitment_tx_id"`
	PropertyId     int64          `json:"property_id"`
	InputHex       string         `json:"input_hex"`
	Timeout        int            `json:"timeout"`
	InputAmount    float64        `json:"input_amount"`   //input amount
	OutputAddress  string         `json:"output_address"` //output Sender Alice(if alice is sender) or Bob(if bob is sender)
	OutAmount      float64        `json:"out_amount"`
	TxHash         string         `json:"tx_hash"`
	Txid           string         `json:"txid"`
	Owner          string         `json:"owner"`
	CurrState      TxInfoState    `json:"curr_state"`
	CreateAt       time.Time      `json:"create_at"`
	SendAt         time.Time      `json:"send_at"`
	CreateBy       string         `json:"create_by"`
}

//HED1a 如果bob返回了正确的R，就可以完成签名，标识这次的htlc可以成功了
type HTLCExecutionDelivery struct {
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
type HTLCExecutionB HTLCExecutionDelivery

type NormalState int

const (
	NS_Create NormalState = 10
	NS_Finish NormalState = 20
	NS_Refuse NormalState = 30
)

type HtlcRAndHInfo struct {
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
	Memo            string      `json:"memo"`
}

type HtlcSingleHopPathInfoState int

const (
	SingleHopPathInfoState_Created            HtlcSingleHopPathInfoState = 0
	SingleHopPathInfoState_StepBegin          HtlcSingleHopPathInfoState = 10
	SingleHopPathInfoState_StepFinish         HtlcSingleHopPathInfoState = 11
	SingleHopPathInfoState_RefusedByInterNode HtlcSingleHopPathInfoState = -1
)

type HtlcSingleHopPathInfo struct {
	Id                           int                        `storm:"id,increment" json:"id" `
	HAndRInfoRequestHash         string                     `json:"h_and_r_info_request_hash"`
	H                            string                     `json:"h"`
	ChannelIdArr                 []int                      `json:"channel_id_arr"`
	CurrState                    HtlcSingleHopPathInfoState `json:"curr_state"`
	BeginBlockHeight             int                        `json:"begin_block_height"`
	TotalStep                    int                        `json:"total_step"`
	CurrStep                     int                        `json:"curr_step"`
	CreateBy                     string                     `json:"create_by"`
	CreateAt                     time.Time                  `json:"create_at"`
	BobCurrRsmcTempPubKey        string                     `json:"bob_curr_rsmc_temp_pub_key"`          // for cnb output1 temp data
	BobCurrHtlcTempPubKey        string                     `json:"bob_curr_htlc_temp_pub_key"`          // for cnb output2 temp data
	BobCurrHtlcTempForHt1bPubKey string                     `json:"bob_curr_htlc_temp_for_ht1b_pub_key"` // for he1b  temp data
}

// 为记录-48的关闭htlc的请求数据
type HtlcRequestCloseCurrTxInfo struct {
	Id                        int            `storm:"id,increment" json:"id" `
	RequestHash               string         `json:"request_hash"`
	ChannelId                 bean.ChannelID `json:"channel_id"`
	CurrRsmcTempAddressPubKey string         `json:"curr_rsmc_temp_address_pub_key"`
	CreateBy                  string         `json:"create_by"`
	CurrState                 NormalState    `json:"curr_state"`
	CreateAt                  time.Time      `json:"create_at"`
}

// to punish alice do not admit the latest commitment tx
type HTLCBreachRemedyTransaction BreachRemedyTransaction
type HTLCTimeoutBreachRemedyTransaction struct {
	Id                                  int            `storm:"id,increment" json:"id" `
	ChannelId                           bean.ChannelID `json:"channel_id"`
	CommitmentTxId                      int            `json:"commitment_tx_id"` // parent commitmentTx id
	HTLCTimeoutTxForAAndExecutionForBId int            `json:"htlc_timeout_tx_for_a_and_execution_for_b_id"`
	PropertyId                          int64          `json:"property_id"`
	InputHash                           string         `json:"input_hash"`
	Amount                              float64        `json:"amount"` // output bob amount
	TxHash                              string         `json:"tx_hash"`
	Txid                                string         `json:"txid"`
	CurrState                           TxInfoState    `json:"curr_state"`
	CreateBy                            string         `json:"create_by"`
	CreateAt                            time.Time      `json:"create_at"`
	SignAt                              time.Time      `json:"sign_at"`
	SendAt                              time.Time      `json:"send_at"`
	LastEditTime                        time.Time      `json:"last_edit_time"`
	Owner                               string         `json:"owner"`
}
