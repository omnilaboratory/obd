package dao

import (
	"LightningOnOmni/bean"
	"time"
)

type HTLCCommitmentTransaction CommitmentTransaction

//HT1a 锁住发起方的转移资金：锁住的意思是：把资金放到一个临时多签帐号（alice1&bob1）
type HTLCTimeoutTxA struct {
	Id                 int            `storm:"id,increment" json:"id" `
	ChannelId          bean.ChannelID `json:"channel_id"`
	CommitmentTxId     int            `json:"commitment_tx_id"`
	InputTxid          string         `json:"input_txid"` //input txid  from commitTx alice2&bob multtaddr, so need  sign of alice2 and bob
	InputVout          uint32         `json:"input_vout"` // input vout
	Timeout            int            `json:"timeout"`
	InputAmount        float64        `json:"input_amount"`   //input amount
	OutputAddress      string         `json:"output_address"` //output alice
	OutAmount          float64        `json:"out_amount"`
	TransactionSignHex string         `json:"transaction_sign_hex"`
	Txid               string         `json:"txid"`
	CurrState          TxInfoState    `json:"curr_state"`
	Owner              string         `json:"owner"`
	CreateAt           time.Time      `json:"create_at"`
	SendAt             time.Time      `json:"send_at"`
	CreateBy           time.Time      `json:"create_by"`
}

//HED1a 如果bob返回了正确的R，就可以完成签名，标识这次的htlc可以成功了
type HTLCExecutionDeliveryA struct {
	Id                 int         `storm:"id,increment" json:"id" `
	CommitmentTxId     int         `json:"commitment_tx_id"`
	InputTxid          string      `json:"input_txid"`   //
	InputVout          uint32      `json:"input_vout"`   // input vout
	InputAmount        float64     `json:"input_amount"` //input amount
	HtlcR              string      `json:"htlc_r"`
	OutputAddress      string      `json:"output_address"`
	OutAmount          float64     `json:"out_amount"`
	TransactionSignHex string      `json:"transaction_sign_hex"`
	Txid               string      `json:"txid"`
	CurrState          TxInfoState `json:"curr_state"`
	Owner              string      `json:"owner"`
	IsEnable           bool        `json:"is_enable"`
	CreateAt           time.Time   `json:"create_at"`
	SendAt             time.Time   `json:"send_at"`
	CreateBy           time.Time   `json:"create_by"`
}
