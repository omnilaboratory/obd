package dao

import (
	"LightningOnOmni/bean"
	"time"
)

type User struct {
	Id int `storm:"id,increment" `
	bean.User
}

type ChannelState int

const (
	ChannelState_Create            ChannelState = 10
	ChannelState_Accept            ChannelState = 20
	ChannelState_Close             ChannelState = 21
	ChannelState_OpenChannelDefuse ChannelState = 30
	ChannelState_FundingDefuse     ChannelState = 31
)

type ChannelInfo struct {
	bean.OpenChannelInfo
	Id                         int            `storm:"id,increment" json:"id"`
	ChannelId                  bean.ChannelID `json:"channel_id"`
	PeerIdA                    string         `json:"peer_id_a"`
	PubKeyA                    string         `json:"pub_key_a"`
	AddressA                   string         `json:"address_a"`
	PeerIdB                    string         `json:"peer_id_b"`
	PubKeyB                    string         `json:"pub_key_b"`
	AddressB                   string         `json:"address_b"`
	ChannelAddress             string         `json:"channel_address"`
	ChannelAddressRedeemScript string         `json:"channel_address_redeem_script"`
	ChannelAddressScriptPubKey string         `json:"channel_address_script_pub_key"`
	CurrState                  ChannelState   `json:"curr_state"`
	CreateBy                   string         `json:"create_by"`
	CreateAt                   time.Time      `json:"create_at"`
	AcceptAt                   time.Time      `json:"accept_at"`
	CloseAt                    time.Time      `json:"close_at"`
}

type CloseChannel struct {
	bean.CloseChannel
	Id       int       `storm:"id,increment" json:"id"`
	CreateAt time.Time `json:"create_at"`
}

type FundingTransactionState int

const (
	FundingTransactionState_Create FundingTransactionState = 10
	FundingTransactionState_Accept FundingTransactionState = 20
	FundingTransactionState_Defuse FundingTransactionState = 30
)

type FundingTransaction struct {
	Id                         int                     `storm:"id,increment" json:"id"`
	ChannelInfoId              int                     `json:"channel_info_id"`
	CurrState                  FundingTransactionState `json:"curr_state"`
	PeerIdA                    string                  `json:"peer_id_a"`
	PeerIdB                    string                  `json:"peer_id_b"`
	ChannelId                  bean.ChannelID          `json:"channel_id"`
	PropertyId                 int64                   `json:"property_id"`
	AmountA                    float64                 `json:"amount_a"`
	FunderPubKey2ForCommitment string                  `json:"funder_pub_key_2_for_commitment"`
	FundingTxHex               string                  `json:"funding_tx_hex"`
	FundingTxid                string                  `json:"funding_txid"`
	FundingOutputIndex         uint32                  `json:"funding_output_index"`
	AmountB                    float64                 `json:"amount_b"`
	CreateBy                   string                  `json:"create_by"`
	CreateAt                   time.Time               `json:"create_at"`
	FundeeSignAt               time.Time               `json:"fundee_sign_at"`
	TxId                       string                  `json:"tx_id"`
}

type TxInfoState int

const (
	TxInfoState_OtherSign  TxInfoState = 10
	TxInfoState_MyselfSign TxInfoState = 20
	TxInfoState_Abord      TxInfoState = 30
)

type CommitmentTxRequestInfo struct {
	Id        int `storm:"id,increment" json:"id" `
	ChannelId bean.ChannelID
	UserId    string
	bean.CommitmentTx
	LastTempAddressPubKey string
	CreateAt              time.Time
	IsEnable              bool
}

//CommitmentTransaction
type CommitmentTransaction struct {
	Id                 int            `storm:"id,increment" json:"id" `
	LastCommitmentTxId int            `json:"last_commitment_tx_id"`
	LastHash           string         `json:"last_hash"`
	CurrHash           string         `json:"curr_hash"`
	PeerIdA            string         `json:"peer_id_a"`
	PeerIdB            string         `json:"peer_id_b"`
	ChannelId          bean.ChannelID `json:"channel_id"`
	PropertyId         int64          `json:"property_id"`
	InputTxid          string         `json:"input_txid"`           //input txid  from channelAddr: alice&bob multiAddr, so need  sign of alice and bob
	InputVout          uint32         `json:"input_vout"`           // input vout
	InputAmount        float64        `json:"input_amount"`         //input amount
	TempAddressPubKey  string         `json:"temp_address_pub_key"` //output alice2 or bob2
	MultiAddress       string         `json:"multi_address"`        //output alice2&bob  or alice&bob2  multiAddr
	RedeemScript       string         `json:"redeem_script"`
	ScriptPubKey       string         `json:"script_pub_key"`
	AmountM            float64        `json:"amount_m"` // amount to multiAddr
	AmountB            float64        `json:"amount_b"` //amount to bob(if Cna) or alice(if Cnb)
	TransactionSignHex string         `json:"transaction_sign_hex"`
	Txid               string         `json:"txid"`
	CurrState          TxInfoState    `json:"curr_state"`
	CreateBy           string         `json:"create_by"`
	CreateAt           time.Time      `json:"create_at"`
	SignAt             time.Time      `json:"sign_at"`
	SendAt             time.Time      `json:"send_at"`
	LastEditTime       time.Time      `json:"last_edit_time"`
	Owner              string         `json:"owner"`
}

// close channel , alice or bob wait 1000 sequence to drawback the balance
type RevocableDeliveryTransaction struct {
	Id                 int            `storm:"id,increment" json:"id" `
	CommitmentTxId     int            `json:"commitment_tx_id"`
	PeerIdA            string         `json:"peer_id_a"`
	PeerIdB            string         `json:"peer_id_b"`
	ChannelId          bean.ChannelID `json:"channel_id"`
	PropertyId         int64          `json:"property_id"`
	InputTxid          string         `json:"input_txid"`     //input txid  from commitTx alice2&bob multtaddr, so need  sign of alice2 and bob
	InputVout          uint32         `json:"input_vout"`     // input vout
	InputAmount        float64        `json:"input_amount"`   //input amount
	OutputAddress      string         `json:"output_address"` //output alice
	Sequence           int            `json:"sequence"`
	Amount             float64        `json:"amount"` // output alice amount
	TransactionSignHex string         `json:"transaction_sign_hex"`
	Txid               string         `json:"txid"`
	CurrState          TxInfoState    `json:"curr_state"`
	CreateBy           string         `json:"create_by"`
	CreateAt           time.Time      `json:"create_at"`
	SignAt             time.Time      `json:"sign_at"`
	SendAt             time.Time      `json:"send_at"`
	LastEditTime       time.Time      `json:"last_edit_time"`
	Owner              string         `json:"owner"`
}

// to punish alice do not admit the latest commitment tx
type BreachRemedyTransaction struct {
	Id                 int            `storm:"id,increment" json:"id" `
	CommitmentTxId     int            `json:"commitment_tx_id"` // parent commitmentTx id
	PeerIdA            string         `json:"peer_id_a"`
	PeerIdB            string         `json:"peer_id_b"`
	ChannelId          bean.ChannelID `json:"channel_id"`
	PropertyId         int64          `json:"property_id"`
	InputTxid          string         `json:"input_txid"`           //input txid  from commitTx alice2&bob multtAddr, so need  sign of alice2 and bob
	InputVout          uint32         `json:"input_vout"`           // input vout
	InputAmount        float64        `json:"input_amount"`         //input amount
	Amount             float64        `json:"amount"`               // output boob amount
	TransactionSignHex string         `json:"transaction_sign_hex"` // first alice2 sign
	Txid               string         `json:"txid"`
	CurrState          TxInfoState    `json:"curr_state"`
	CreateBy           string         `json:"create_by"`
	CreateAt           time.Time      `json:"create_at"`
	SignAt             time.Time      `json:"sign_at"`
	SendAt             time.Time      `json:"send_at"`
	LastEditTime       time.Time      `json:"last_edit_time"`
	Owner              string         `json:"owner"`
}
