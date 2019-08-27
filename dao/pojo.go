package dao

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"time"
)

type User struct {
	Id int `storm:"id,increment" `
	bean.User
}

type ChannelState int

const (
	ChannelState_Create ChannelState = 10
	ChannelState_Accept ChannelState = 20
	ChannelState_Defuse ChannelState = 30
)

type ChannelInfo struct {
	bean.OpenChannelInfo
	Id            int            `storm:"id,increment" json:"id"`
	PeerIdA       string         `json:"peer_id_a"`
	PeerIdB       string         `json:"peer_id_b"`
	PubKeyA       string         `json:"pub_key_a"`
	PubKeyB       string         `json:"pub_key_b"`
	ChannelPubKey string         `json:"channel_pub_key"`
	RedeemScript  string         `json:"redeem_script"`
	ChannelID     bean.ChannelID `json:"channel_id"`
	CurrState     ChannelState   `json:"curr_state"`
	CreateAt      time.Time      `json:"create_at"`
	AcceptAt      time.Time      `json:"accept_at"`
}

type CloseChannel struct {
	bean.CloseChannel
	Id       int       `storm:"id,increment" json:"id"`
	CreateAt time.Time `json:"create_at"`
}

type FundingTransactionState int

const (
	FundingTransaction_Create      FundingTransactionState = 10
	FundingTransactionState_Accept FundingTransactionState = 20
	FundingTransactionState_Defuse FundingTransactionState = 30
)

type FundingTransaction struct {
	Id                 int                     `storm:"id,increment" `
	PeerIdA            string                  `json:"peer_id_a"`
	PeerIdB            string                  `json:"peer_id_b"`
	TemporaryChannelId chainhash.Hash          `json:"temporary_channel_id"`
	ChannelID          bean.ChannelID          `json:"channel_id"`
	PropertyId         int64                   `json:"property_id"`
	FunderPubKey       string                  `json:"funder_pub_key"`
	AmountA            float64                 `json:"amount_a"`
	FunderSignature    chainhash.Signature     `json:"funder_signature"`
	FundeePubKey       string                  `json:"fundee_pub_key"`
	AmountB            float64                 `json:"amount_b"`
	FundeeSignature    chainhash.Signature     `json:"fundee_signature"`
	RedeemScript       string                  `json:"redeem_script"`
	ChannelPubKey      string                  `json:"channel_pub_key"`
	CreateAt           time.Time               `json:"create_at"`
	FundeeSignAt       time.Time               `json:"fundee_sign_at"`
	TxId               string                  `json:"tx_id"`
	CurrState          FundingTransactionState `json:"curr_state"`
}

type CommitmentTx struct {
	Id int `storm:"id,increment" `
	bean.CommitmentTx
	CreateAt time.Time `json:"create_at"`
}
type CommitmentTxSigned struct {
	Id int `storm:"id,increment" `
	bean.CommitmentTxSigned
	CreateAt time.Time `json:"create_at"`
}

type GetBalanceRequest struct {
	Id int `storm:"id,increment" `
	bean.GetBalanceRequest
	CreateAt time.Time `json:"create_at"`
}

type GetBalanceRespond struct {
	Id int `storm:"id,increment" `
	bean.GetBalanceRespond
	CreateAt time.Time `json:"create_at"`
}
