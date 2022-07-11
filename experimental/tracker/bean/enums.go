package bean

import (
	"github.com/omnilaboratory/obd/bean/enum"
)

func CheckExist(msgType enum.MsgType) bool {
	switch msgType {
	case enum.MsgType_Error_0:
		return true
	case enum.MsgType_Tracker_Connect_301:
		return true
	case enum.MsgType_Tracker_HeartBeat_302:
		return true
	case enum.MsgType_Tracker_NodeLogin_303:
		return true
	case enum.MsgType_Tracker_UserLogin_304:
		return true
	case enum.MsgType_Tracker_UserLogout_305:
		return true
	case enum.MsgType_Tracker_UpdateChannelInfo_350:
		return true
	case enum.MsgType_Tracker_GetHtlcPath_351:
		return true
	case enum.MsgType_Tracker_UpdateHtlcTxState_352:
		return true
	case enum.MsgType_Tracker_UpdateUserInfo_353:
		return true
	}
	return false
}
