package bean

type MsgType int

const (
	MsgType_Error_0      MsgType = 0
	MsgType_Connect_1    MsgType = 1
	MsgType_HeartBeat_2  MsgType = 2
	MsgType_NodeLogin_3  MsgType = 3
	MsgType_UserLogin_4  MsgType = 4
	MsgType_UserLogout_5 MsgType = 5
)

func CheckExist(msgType MsgType) bool {
	switch msgType {
	case MsgType_Error_0:
		return true
	case MsgType_Connect_1:
		return true
	case MsgType_HeartBeat_2:
		return true
	case MsgType_NodeLogin_3:
		return true
	case MsgType_UserLogin_4:
		return true
	case MsgType_UserLogout_5:
		return true
	}
	return false
}
