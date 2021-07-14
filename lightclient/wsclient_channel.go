package lightclient

import (
	"encoding/json"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/service"
	"strconv"
)

func (client *Client) ChannelModule(msg bean.RequestMessage) (enum.SendTargetType, []byte, bool) {
	status := false
	var sendType = enum.SendTargetType_SendToSomeone
	data := ""
	switch msg.Type {
	//get openChannelReq from funder then send to fundee
	case enum.MsgType_SendChannelOpen_32:
		if msg.RecipientUserPeerId == client.User.PeerId {
			data = "can not open channel to yourself"
		} else {
			node, err := service.ChannelService.AliceOpenChannel(msg, client.User)
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

		msg.Type = enum.MsgType_SendChannelOpen_32
		client.SendToMyself(msg.Type, status, data)

		if status {
			msg.Type = enum.MsgType_ChannelOpen_32
			err := client.sendDataToP2PUser(msg, status, data)
			if err != nil {
				data = err.Error()
				status = false
				msg.Type = enum.MsgType_RecvChannelAccept_33
				client.SendToMyself(msg.Type, status, data)
			}
		}

	case enum.MsgType_ChannelOpen_AllItem_3150:
		pageData, err := service.ChannelService.AllItem(msg.Data, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(pageData)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
			}
		}
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_ChannelOpen_ItemByTempId_3151:
		node, err := service.ChannelService.GetChannelByTemporaryChanId(msg.Data, *client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_ChannelOpen_Count_3152:
		node, err := service.ChannelService.TotalCount(*client.User)
		if err != nil {
			data = err.Error()
		} else {
			data = strconv.Itoa(node)
			status = true
		}
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_ChannelOpen_DelItemByTempId_3153:
		node, err := service.ChannelService.DelChannelByTemporaryChanId(msg.Data, *client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_GetChannelInfoByChannelId_3154:
		node, err := service.ChannelService.GetChannelInfoByChannelId(msg.Data, *client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_GetChannelInfoByDbId_3155:
		node, err := service.ChannelService.GetChannelInfoById(msg.Data, *client.User)
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
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_CheckChannelAddessExist_3156:
		node, err := service.ChannelService.BobCheckChannelAddressExist(msg.Data, client.User)
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
		client.SendToMyself(msg.Type, status, data)

	//get acceptChannelReq from fundee then send to funder
	case enum.MsgType_SendChannelAccept_33:
		node, err := service.ChannelService.BobAcceptChannel(msg, client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_ChannelAccept_33
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					data = err.Error()
					status = false
				}
			}
		}
		msg.Type = enum.MsgType_SendChannelAccept_33
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_SendCloseChannelRequest_38:
		node, err := service.ChannelService.ForceCloseChannel(msg, client.User)
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
			msg.Type = enum.MsgType_CloseChannelSign_39
			err = client.sendDataToP2PUser(msg, status, data)
		}
		msg.Type = enum.MsgType_SendCloseChannelRequest_38
		client.SendToMyself(msg.Type, status, data)

	case enum.MsgType_SendCloseChannelSign_39:
		node, err := service.ChannelService.SignCloseChannel(msg, *client.User)
		if err != nil {
			data = err.Error()
		} else {
			bytes, err := json.Marshal(node)
			if err != nil {
				data = err.Error()
			} else {
				data = string(bytes)
				status = true
				msg.Type = enum.MsgType_CloseChannelSign_39
				err = client.sendDataToP2PUser(msg, status, data)
				if err != nil {
					status = false
					data = err.Error()
				}
			}
		}
		msg.Type = enum.MsgType_SendCloseChannelSign_39
		client.SendToMyself(msg.Type, status, data)

	default:
		sendType = enum.SendTargetType_SendToNone
	}
	return sendType, []byte(data), status
}
