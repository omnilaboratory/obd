package dao

import (
	"time"
)

type HtlcHAndRImage struct {
	Id       int       `storm:"id,increment" json:"id" `
	H        string    `json:"h"`
	Index    int       `json:"index"`
	R        string    `json:"r"`
	CreateAt time.Time `json:"create_at"`
}

type AddHtlcRequestInfo struct {
	Id                               int         `storm:"id,increment" json:"id" `
	ChannelId                        string      `json:"channel_id"`
	RecipientUserPeerId              string      `json:"recipient_user_peer_id"`
	PropertyId                       int64       `json:"property_id"`
	Amount                           float64     `json:"amount"`
	Memo                             string      `json:"memo"`
	H                                string      `json:"h"`
	RoutingPacket                    string      `json:"routing_packet"`
	LastTempAddressPrivateKey        string      `json:"last_temp_address_private_key"`  //	支付方的上一个RSMC委托交易用到的临时地址的私钥 存储在bob这边的请求
	CurrRsmcTempAddressPubKey        string      `json:"curr_rsmc_temp_address_pub_key"` //	创建Cnx中的toRsmc的部分使用的临时地址的公钥
	CurrHtlcTempAddressPubKey        string      `json:"curr_htlc_temp_address_pub_key"` //	创建Cnx中的toHtlc的部分使用的临时地址的公钥
	CurrHtlcTempAddressForHt1aIndex  int         `json:"curr_htlc_temp_address_for_ht1a_index"`
	CurrHtlcTempAddressForHt1aPubKey string      `json:"curr_htlc_temp_address_for_ht1a_pub_key"` //	创建Ht1a中生成ht1a的输出的Rmsc的临时地址的公钥
	CurrState                        NormalState `json:"curr_state"`
	CreateBy                         string      `json:"create_by"`
	CreateAt                         time.Time   `json:"create_at"`
	FinishAt                         time.Time   `json:"finish_at"`
}

//HT1a 锁住发起方的交易资金：锁住的意思是：把资金放到一个临时多签帐号（alice1&bob）
//HE1b 等获取到R后，锁住接收方的交易资金：锁住的意思是：把资金放到一个临时多签帐号（alice&bob6）
type HTLCTimeoutTxForAAndExecutionForB struct {
	Id                           int         `storm:"id,increment" json:"id" `
	ChannelId                    string      `json:"channel_id"`
	CommitmentTxId               int         `json:"commitment_tx_id"`
	PropertyId                   int64       `json:"property_id"`
	InputTxid                    string      `json:"input_txid"`   // input txid  from commitTx aliceTempHtlc&bob multtaddr, so need  sign of aliceTempHtlc and bob
	InputAmount                  float64     `json:"input_amount"` // input amount
	InputHex                     string      `json:"input_hex"`
	Timeout                      int         `json:"timeout"` // if 3days 432=3*24*6
	RSMCTempAddressIndex         int         `json:"rsmc_temp_address_index"`
	RSMCTempAddressPubKey        string      `json:"rsmc_temp_address_pub_key"`
	RSMCMultiAddress             string      `json:"rsmc_multi_address"`
	RSMCRedeemScript             string      `json:"rsmc_redeem_script"`
	RSMCMultiAddressScriptPubKey string      `json:"rsmc_multi_address_script_pub_key"`
	RSMCOutAmount                float64     `json:"rsmc_out_amount"`
	RSMCTxHex                    string      `json:"rsmc_tx_hex"`
	RSMCTxid                     string      `json:"rsmc_txid"`
	CurrState                    TxInfoState `json:"curr_state"`
	Owner                        string      `json:"owner"`
	CreateBy                     string      `json:"create_by"`
	CreateAt                     time.Time   `json:"create_at"`
	SignAt                       time.Time   `json:"sign_at"`
	SendAt                       time.Time   `json:"send_at"`
}

//HTD1b 锁住发起方的交易资金：如果bob广播了，如果过了三天，如果R没有获取到，alice可以通过广播拿回资金
type HTLCTimeoutDeliveryTxB struct {
	Id             int         `storm:"id,increment" json:"id" `
	ChannelId      string      `json:"channel_id"`
	CommitmentTxId int         `json:"commitment_tx_id"`
	PropertyId     int64       `json:"property_id"`
	InputTxid      string      `json:"input_txid"`
	InputHex       string      `json:"input_hex"`
	Timeout        int         `json:"timeout"`
	OutputAddress  string      `json:"output_address"` //output Sender Alice(if alice is sender) or Bob(if bob is sender)
	OutAmount      float64     `json:"out_amount"`
	TxHex          string      `json:"tx_hex"`
	Txid           string      `json:"txid"`
	Owner          string      `json:"owner"`
	CurrState      TxInfoState `json:"curr_state"`
	CreateAt       time.Time   `json:"create_at"`
	SendAt         time.Time   `json:"send_at"`
	CreateBy       string      `json:"create_by"`
}

//HED1a when get H , alice 得到H（公钥），生成三签地址，锁住给中间节点的钱，当得到R（私钥）的时候，完成三签地址的签名，生成最终支付交易
type HtlcLockTxByH struct {
	Id                 int         `storm:"id,increment" json:"id" `
	ChannelId          string      `json:"channel_id"`
	PropertyId         int64       `json:"property_id"`
	CommitmentTxId     int         `json:"commitment_tx_id"`
	InputHex           string      `json:"input_hex"`
	InputTxid          string      `json:"input_txid"` // input txid  from commitTx aliceTempHtlc&bob multtaddr, so need  sign of aliceTempHtlc and bob
	HtlcH              string      `json:"htlc_h"`     // H(公钥，双签地址之一)
	PayeeChannelPubKey string      `json:"payee_channel_pub_key"`
	OutputAddress      string      `json:"output_address"` // 双签地址 锁定支付资金
	RedeemScript       string      `json:"redeem_script"`  // 双签地址对应的赎回脚本
	ScriptPubKey       string      `json:"script_pub_key"`
	OutAmount          float64     `json:"out_amount"`
	Timeout            int         `json:"timeout"`
	TxHex              string      `json:"tx_hex"`
	Txid               string      `json:"txid"`
	CurrState          TxInfoState `json:"curr_state"`
	Owner              string      `json:"owner"`
	CreateAt           time.Time   `json:"create_at"`
	SendAt             time.Time   `json:"send_at"`
	CreateBy           string      `json:"create_by"`
}

//HED1a when get R 如果bob返回了正确的R，就可以完成签名，标识这次的htlc可以成功了
type HTLCExecutionDeliveryOfR struct {
	Id             int         `storm:"id,increment" json:"id" `
	ChannelId      string      `json:"channel_id"`
	CommitmentTxId int         `json:"commitment_tx_id"`
	HLockTxId      int         `json:"hlock_tx_id"`
	InputHex       string      `json:"input_hex"`
	InputTxid      string      `json:"input_txid"`   // input txid  from commitTx aliceTempHtlc&bob multtaddr, so need  sign of aliceTempHtlc and bob
	InputAmount    float64     `json:"input_amount"` // input amount
	HtlcR          string      `json:"htlc_r"`
	OutputAddress  string      `json:"output_address"` //to Bob
	OutAmount      float64     `json:"out_amount"`
	TxHex          string      `json:"tx_hex"`
	Txid           string      `json:"txid"`
	CurrState      TxInfoState `json:"curr_state"`
	Owner          string      `json:"owner"`
	CreateAt       time.Time   `json:"create_at"`
	SendAt         time.Time   `json:"send_at"`
	CreateBy       string      `json:"create_by"`
}

type NormalState int

const (
	NS_Create NormalState = 10
	NS_Finish NormalState = 20
	NS_Refuse NormalState = 30
)
