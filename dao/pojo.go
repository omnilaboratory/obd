package dao

import "LightningOnOmni/bean"

type User struct {
	Id int `storm:"id,increment" `
	bean.User
}

type OpenChannelInfo struct {
	bean.OpenChannelInfo
	TemporaryChannelIdStr string `json:"temporary_channel_id_str"`
}
type AcceptChannelInfo struct {
	bean.AcceptChannelInfo
}
type CloseChannel struct {
	bean.CloseChannel
}
type FundingCreated struct {
	Id int `storm:"id,increment" `
	bean.FundingCreated
	TemporaryChannelIdStr string `json:"temporary_channel_id_str"`
}
type FundingSigned struct {
	Id int `storm:"id,increment" json:"id"`
	bean.FundingSigned
}

type CommitmentTx struct {
	Id int `storm:"id,increment" `
	bean.CommitmentTx
}
type CommitmentTxSigned struct {
	Id int `storm:"id,increment" `
	bean.CommitmentTxSigned
}

type GetBalanceRequest struct {
	Id int `storm:"id,increment" `
	bean.GetBalanceRequest
}

type GetBalanceRespond struct {
	Id int `storm:"id,increment" `
	bean.GetBalanceRespond
}
