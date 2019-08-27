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
	ChannelState_Create ChannelState = 10
	ChannelState_Accept ChannelState = 20
	ChannelState_Defuse ChannelState = 30
)

type ChannelInfo struct {
	Id int `storm:"id,increment" json:"id"`
	bean.OpenChannelInfo
	FunderPeerId  string       `json:"funder_peer_id"`
	FundeePeerId  string       `json:"fundee_peer_id"`
	FunderPubKey  string       `json:"funder_pub_key"`
	FundeePubKey  string       `json:"fundee_pub_key"`
	ChannelPubKey string       `json:"channel_pub_key"`
	RedeemScript  string       `json:"redeem_script"`
	CurrState     ChannelState `json:"curr_state"`
	CreateAt      time.Time    `json:"create_at"`
	AcceptAt      time.Time    `json:"accept_at"`
}

type CloseChannel struct {
	Id int `storm:"id,increment" json:"id"`
	bean.CloseChannel
	CreateAt time.Time `json:"create_at"`
}
type FundingCreated struct {
	Id int `storm:"id,increment" `
	bean.FundingCreated
	TemporaryChannelIdStr string    `json:"temporary_channel_id_str"`
	CreateAt              time.Time `json:"create_at"`
}
type FundingSigned struct {
	Id int `storm:"id,increment" json:"id"`
	bean.FundingSigned
	CreateAt time.Time `json:"create_at"`
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
