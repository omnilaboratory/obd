package bean

import (
	"github.com/omnilaboratory/obd/bean/enum"
)

type RequestMessage struct {
	Type             enum.MsgType `json:"type"`
	SenderNodePeerId string       `json:"sender_node_peer_id"`
	Data             interface{}  `json:"data"`
}

type ReplyMessage struct {
	Type   enum.MsgType `json:"type"`
	Status bool         `json:"status"`
	From   string       `json:"from"`
	To     string       `json:"to"`
	Result interface{}  `json:"result"`
}

//请求获取htlc的path
type HtlcPathRequest struct {
	PayerObdNodeId  string  `json:"payer_obd_node_id"`
	RealPayerPeerId string  `json:"real_payer_peer_id"`
	PayeePeerId     string  `json:"payee_peer_id"`
	PropertyId      int64   `json:"property_id"`
	H               string  `json:"h"`
	Amount          float64 `json:"amount"`
}

const (
	HtlcTxState_PayMoney        = 0
	HtlcTxState_ConfirmPayMoney = 1
)

//GetHtlcTxStateRequest
type GetHtlcTxStateRequest struct {
	Path string `json:"path"`
	H    string `json:"h"`
}

//GetChannelStateRequest
type GetChannelStateRequest struct {
	ChannelId string `json:"channel_id"`
}

//newHtlcTx
type UpdateHtlcTxStateRequest struct {
	GetHtlcTxStateRequest
	R string `json:"r"`
	//0 0:forward h 1:backword
	DirectionFlag int    `json:"direction_flag"`
	CurrChannelId string `json:"curr_channel_id"`
}

//GetChannelStateRequest
type GetOmniBalanceRequest struct {
	Address    string `json:"address"`
	PropertyId int    `json:"property_id"`
}
