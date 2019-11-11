package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
	"LightningOnOmni/tool"
	"encoding/json"
	"strconv"
)

func (client *Client) channelModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToNone
	data := ""
	switch msg.Type {
	//get openChannelReq from funder then send to fundee
	case enum.MsgType_ChannelOpen_N32:
		if tool.CheckIsString(&msg.RecipientPeerId) == false {
			data = "no target user"
		} else {
			if msg.RecipientPeerId == client.User.PeerId {
				data = "can not open channel to yourself"
			} else {
				node, err := service.ChannelService.AliceOpenChannel(msg, client.User.PeerId)
				if err != nil {
					data = err.Error()
				} else {
					bytes, err := json.Marshal(node)
					if err != nil {
						data = err.Error()
					} else {
						data = string(bytes)
						status = true
					}
				}
			}
		}
		_ = client.sendToSomeone(msg.Type, status, msg.RecipientPeerId, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ChannelOpen_ItemByTempId_N3201:
		node, err := service.ChannelService.GetChannelByTemporaryChanId(msg.Data)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ChannelOpen_DelItemByTempId_N3204:
		node, err := service.ChannelService.DelChannelByTemporaryChanId(msg.Data)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ChannelOpen_AllItem_N3202:
		nodes, err := service.ChannelService.AllItem(client.User.PeerId)
		if err != nil {
			data = err.Error()
		} else {
			page := make(map[string]interface{})
			page["count"] = len(nodes)
			page["body"] = nodes
			bytes, err := json.Marshal(page)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ChannelOpen_Count_N3203:
		node, err := service.ChannelService.TotalCount(client.User.PeerId)
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(node)
			status = true
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_ForceCloseChannel_N3205:
		node, err := service.ChannelService.ForceCloseChannel(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	//get acceptChannelReq from fundee then send to funder
	case enum.MsgType_ChannelAccept_N33:
		node, err := service.ChannelService.BobAcceptChannel(msg.Data, client.User.PeerId)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		if status {
			_ = client.sendToSomeone(msg.Type, status, node.PeerIdA, data)
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CloseChannelRequest_N38:
		node, targetUser, err := service.ChannelService.RequestCloseChannel(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		if status {
			if targetClient, err := client.FindUser(targetUser); err == nil {
				_ = client.sendToSomeone(msg.Type, status, targetClient.User.PeerId, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	case enum.MsgType_CloseChannelSign_N39:
		node, targetUser, err := service.ChannelService.CloseChannelSign(msg.Data, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		if status {
			if targetClient, err := client.FindUser(targetUser); err == nil {
				_ = client.sendToSomeone(msg.Type, status, targetClient.User.PeerId, data)
			}
		}
		client.sendToMyself(msg.Type, status, data)
		sendType = enum.SendTargetType_SendToSomeone
	}
	return sendType, []byte(data), status
}
