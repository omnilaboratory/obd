package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"github.com/tidwall/gjson"
)

func (client *Client) userModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	var data string
	switch msg.Type {
	case enum.MsgType_UserLogin_1:
		if client.User != nil {
			client.sendToMyself(msg.Type, true, "already login")
			sendType = enum.SendTargetType_SendToSomeone
		} else {
			user := bean.User{
				PeerId: gjson.Get(msg.Data, "peer_id").String(),
			}
			if len(user.PeerId) > 0 {
				client.User = &user
				_ = service.UserService.UserLogin(&user)
				if client.User != nil {
					GlobalWsClientManager.OnlineUserMap[user.PeerId] = client
					service.OnlineUserMap[user.PeerId] = true
				}
				data = client.User.PeerId + " login"
				sendType = enum.SendTargetType_SendToAll
				status = true
			} else {
				client.sendToMyself(msg.Type, true, "error peer_id")
				sendType = enum.SendTargetType_SendToSomeone
			}
		}
	case enum.MsgType_UserLogout_2:
		if client.User != nil {
			data = client.User.PeerId + " logout"
			sendType = enum.SendTargetType_SendToAll
			if client.User != nil {
				delete(GlobalWsClientManager.OnlineUserMap, client.User.PeerId)
				delete(service.OnlineUserMap, client.User.PeerId)
			}
			client.User = nil
			status = true
		} else {
			client.sendToMyself(msg.Type, true, "please login")
			sendType = enum.SendTargetType_SendToSomeone
		}
	}
	return sendType, []byte(data), status
}
