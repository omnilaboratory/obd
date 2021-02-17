package dao

import "time"

type RDTxWaitingSend struct {
	Id                int       `storm:"id,increment" json:"id" `
	TransactionHex    string    `json:"transaction_hex"`
	Type              int       `json:"type"`                   // 0: RD 1000, 1:HT1a  2:htd1b
	HtnxIdAndHtnxRdId []int     `json:"htnx_id_and_htnx_rd_id"` // for ht1a later logic
	IsEnable          bool      `json:"is_enable"`
	CreateAt          time.Time `json:"create_at"`
	FinishAt          time.Time `json:"finish_at"`
}

type ObdConfig struct {
	Id              int    `storm:"id,increment" json:"id" `
	InitHashCode    string `json:"init_hash_code"`
	AdminLoginToken string `json:"admin_login_token"`
}
