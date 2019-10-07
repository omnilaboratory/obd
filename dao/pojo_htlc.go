package dao

import (
	"LightningOnOmni/bean"
	"time"
)

type HTLCCommitmentTransaction CommitmentTransaction

type HTLCTimeoutTxA struct {
	Id                 int            `storm:"id,increment" json:"id" `
	ChannelId          bean.ChannelID `json:"channel_id"`
	CommitmentTxId     int            `json:"commitment_tx_id"`
	InputTxid          string         `json:"input_txid"` //input txid  from commitTx alice2&bob multtaddr, so need  sign of alice2 and bob
	InputVout          uint32         `json:"input_vout"` // input vout
	Timeout            int            `json:"timeout"`
	TxHash             string         `json:"tx_hash"`
	InputAmount        float64        `json:"input_amount"`   //input amount
	OutputAddress      string         `json:"output_address"` //output alice
	Sequence           int            `json:"sequence"`
	Amount             float64        `json:"amount"` // output alice amount
	TransactionSignHex string         `json:"transaction_sign_hex"`
	Txid               string         `json:"txid"`
	CurrState          TxInfoState    `json:"curr_state"`
	Owner              string         `json:"owner"`
	CreateAt           time.Time      `json:"create_at"`
	SendAt             time.Time      `json:"send_at"`
	CreateBy           time.Time      `json:"create_by"`
}
