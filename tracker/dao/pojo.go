package dao

import (
	cbean "github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/tracker/bean"
	"time"
)

//ObdNodeInfo
type ObdNodeInfo struct {
	Id              int       `storm:"id,increment" json:"id"`
	IsOnline        bool      `json:"is_online"`
	LatestLoginIp   string    `json:"latest_login_ip"`
	LatestLoginAt   time.Time `json:"latest_login_at"`
	LatestOfflineAt time.Time `json:"latest_offline_at"`
	cbean.ObdNodeLoginRequest
}

//ObdNodeLoginLog
type ObdNodeLoginLog struct {
	Id        int       `storm:"id,increment" json:"id"`
	ObdId     string    `json:"obd_id"`
	LoginIp   string    `json:"login_ip"`
	LoginTime time.Time `json:"login_time"`
}

type UserInfo struct {
	Id           int       `storm:"id,increment" json:"id"`
	ObdP2pNodeId string    `json:"obd_p2p_node_id"`
	ObdNodeId    string    `json:"obd_node_id"`
	IsOnline     bool      `json:"is_online"`
	OfflineAt    time.Time `json:"offline_at"`
	cbean.ObdNodeUserLoginRequest
}
type ChannelInfo struct {
	Id           int                `storm:"id,increment" json:"id"`
	ObdNodeIdA   string             `json:"obd_node_ida"`
	ObdNodeIdB   string             `json:"obd_node_idb"`
	ChannelId    string             `json:"channel_id"`
	PropertyId   int64              `json:"property_id"`
	CurrState    cbean.ChannelState `json:"curr_state"`
	PeerIdA      string             `json:"peer_ida"`
	PeerIdB      string             `json:"peer_idb"`
	AmountA      float64            `json:"amount_a"`
	AmountB      float64            `json:"amount_b"`
	LatestEditAt time.Time          `json:"latest_edit_at"`
	CreateAt     time.Time          `json:"create_at"`
}

type HtlcTxInfo struct {
	Id       int       `storm:"id,increment" json:"id"`
	CreateAt time.Time `json:"create_at"`
	bean.UpdateHtlcTxStateRequest
}

type LockHtlcPath struct {
	Id        int       `storm:"id,increment" json:"id"`
	Path      []string  `json:"path"`
	CurrState int       `json:"curr_state"` // 0 创建，1，标记作废，2 标记通知完成
	CreateAt  time.Time `json:"create_at"`
}
