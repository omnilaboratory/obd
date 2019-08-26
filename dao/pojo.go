package dao

import (
	"LightningOnOmni/bean"
	"time"
)

type User struct {
	Id int `storm:"id,increment" `
	bean.User
}

type OpenChannelInfo struct {
	Id           int    `storm:"id,increment" json:"id"`
	FunderPeerId string `json:"funder_peer_id"`
	FundeePeerId string `json:"fundee_peer_id"`
	bean.OpenChannelInfo
	FunderPubKey  string    `json:"funder_pubkey"`
	FundeePubKey  string    `json:"fundee_pubkey"`
	ChannelPubKey string    `json:"channel_pubkey"`
	RedeemScript  string    `json:"redeem_script"`
	CreateAt      time.Time `json:"create_at"`
	AcceptAt      time.Time `json:"accept_at"`
}
type AcceptChannelInfo struct {
	Id int `storm:"id,increment" json:"id"`
	bean.AcceptChannelInfo
	CreateAt time.Time `json:"create_at"`
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
