package dao

import (
	"github.com/omnilaboratory/obd/tracker/bean"
	"time"
)

//ObdNodeInfo
type ObdNodeInfo struct {
	Id        int       `storm:"id,increment" json:"id"`
	IsOnline  bool      `json:"is_online"`
	OfflineAt time.Time `json:"offline_at"`
	bean.ObdNodeLoginRequest
}

//ObdNodeLoginLog
type ObdNodeLoginLog struct {
	Id        int       `storm:"id,increment" json:"id"`
	LoginIp   string    `json:"login_ip"`
	LoginTime time.Time `json:"login_time"`
}

type UserInfo struct {
	Id        int       `storm:"id,increment" json:"id"`
	ObdNodeId string    `json:"obd_node_id"`
	IsOnline  bool      `json:"is_online"`
	OfflineAt time.Time `json:"offline_at"`
	bean.ObdNodeUserLoginRequest
}
type ChannelInfo struct {
	Id         int       `storm:"id,increment" json:"id"`
	ObdNodeIdA string    `json:"obd_node_ida"`
	ObdNodeIdB string    `json:"obd_node_idb"`
	ChannelId  string    `json:"channel_id"`
	PropertyId int64     `json:"property_id"`
	CurrState  int       `json:"curr_state"`
	PeerIdA    string    `json:"peer_ida"`
	PeerIdB    string    `json:"peer_idb"`
	AmountA    float64   `json:"amount_a"`
	AmountB    float64   `json:"amount_b"`
	CreateAt   time.Time `json:"create_at"`
}
