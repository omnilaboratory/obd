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
