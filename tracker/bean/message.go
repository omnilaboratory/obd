package bean

import (
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
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

//节点登录
type ObdNodeLoginRequest struct {
	NodeId string `json:"node_id"`
}

//节点的用户登录
type ObdNodeUserLoginRequest struct {
	UserId string `json:"user_id"`
}

//更新通道
type ChannelInfoRequest struct {
	ChannelId  string           `json:"channel_id"`
	PropertyId int64            `json:"property_id"`
	CurrState  dao.ChannelState `json:"curr_state"`
	IsAlice    bool             `json:"is_alice"` //是否是alice方的节点
	PeerIdA    string           `json:"peer_ida"`
	PeerIdB    string           `json:"peer_idb"`
	AmountA    float64          `json:"amount_a"`
	AmountB    float64          `json:"amount_b"`
}

//请求获取htlc的path
type HtlcPathRequest struct {
	RealPayerPeerId string  `json:"real_payer_peer_id"`
	PayeePeerId     string  `json:"payee_peer_id"`
	PropertyId      int64   `json:"property_id"`
	Amount          float64 `json:"amount"`
}

const (
	HtlcTxState_PayMoney        = 0
	HtlcTxState_ConfirmPayMoney = 1
)

//newHtlcTx
type GetHtlcTxStateRequest struct {
	Path string `json:"path"`
	H    string `json:"h"`
}

//newHtlcTx
type UpdateHtlcTxStateRequest struct {
	GetHtlcTxStateRequest
	R string `json:"r"`
	//0 0:forward h 1:backword
	DirectionFlag int    `json:"direction_flag"`
	CurrChannelId string `json:"curr_channel_id"`
}
