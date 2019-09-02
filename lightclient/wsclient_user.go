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
	case enum.MsgType_UserLogin:
		if client.User != nil {
			client.sendToMyself(msg.Type, true, "already login")
			sendType = enum.SendTargetType_SendToSomeone
		} else {
			user := bean.User{
				PeerId: gjson.Get(msg.Data, "peer_id").String(),
			}
			if len(user.PeerId) > 0 {
				client.User = &user
				service.UserService.UserLogin(&user)
				data = client.User.PeerId + " login"
				sendType = enum.SendTargetType_SendToAll
				status = true
			} else {
				client.sendToMyself(msg.Type, true, "error peer_id")
				sendType = enum.SendTargetType_SendToSomeone
			}
		}
	case enum.MsgType_UserLogout:
		if client.User != nil {
			data = client.User.PeerId + " logout"
			sendType = enum.SendTargetType_SendToAll
			client.User = nil
			status = true
		} else {
			client.sendToMyself(msg.Type, true, "please login")
			sendType = enum.SendTargetType_SendToSomeone
		}
	}
	return sendType, []byte(data), status
}
