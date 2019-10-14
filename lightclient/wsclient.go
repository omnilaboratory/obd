package lightclient

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/rpc"
	"LightningOnOmni/service"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type Client struct {
	Id          string
	User        *bean.User
	Socket      *websocket.Conn
	SendChannel chan []byte
}

func (client *Client) Write() {
	defer func() {
		e := client.Socket.Close()
		if e != nil {
			log.Println(e)
		} else {
			log.Println("socket closed after writing...")
		}
	}()

	for {
		select {
		case data, ok := <-client.SendChannel:
			if !ok {
				_ = client.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Println("send data", string(data))
			_ = client.Socket.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (client *Client) Read() {
	defer func() {
		_ = service.UserService.UserLogout(client.User)
		if client.User != nil {
			delete(GlobalWsClientManager.UserMap, client.User.PeerId)
		}
		GlobalWsClientManager.Disconnected <- client
		_ = client.Socket.Close()
		log.Println("socket closed after reading...")
	}()

	for {
		_, dataReq, err := client.Socket.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var msg bean.RequestMessage
		log.Println("request data: ", string(dataReq))
		parse := gjson.Parse(string(dataReq))

		if parse.Exists() == false {
			log.Println("wrong json input")
			client.sendToMyself(enum.MsgType_Error, false, string(dataReq))
			continue
		}

		msg.Type = enum.MsgType(parse.Get("type").Int())
		msg.Data = parse.Get("data").String()
		msg.SenderPeerId = parse.Get("sender_peer_id").String()
		msg.RecipientPeerId = parse.Get("recipient_peer_id").String()
		msg.PubKey = parse.Get("pub_key").String()
		msg.Signature = parse.Get("signature").String()

		// check the Recipient is online
		if tool.CheckIsString(&msg.RecipientPeerId) {
			_, err := client.FindUser(&msg.RecipientPeerId)
			if err != nil {
				client.sendToMyself(msg.Type, true, "can not find target user")
				continue
			}
		}

		// check the data whether is right signature
		if tool.CheckIsString(&msg.PubKey) && tool.CheckIsString(&msg.Signature) {
			rpcClient := rpc.NewClient()
			result, err := rpcClient.VerifyMessage(msg.PubKey, msg.Signature, msg.Data)
			if err != nil {
				client.sendToMyself(msg.Type, false, err.Error())
				continue
			}
			if gjson.Parse(result).Bool() == false {
				client.sendToMyself(msg.Type, false, "error signature")
				continue
			}
		}

		var sendType = enum.SendTargetType_SendToNone
		status := false
		var dataOut []byte
		var flag = true
		if msg.Type < 1000 && msg.Type >= 0 {
			sendType, dataOut, status = client.userModule(msg)
			flag = false
		}

		if msg.Type > 1000 {
			sendType, dataOut, status = client.omniCoreModule(msg)
			flag = false
		}

		if flag {
			if client.User == nil {
				client.sendToMyself(msg.Type, false, "please login")
				continue
			}
		}

		typeStr := strconv.Itoa(int(msg.Type))
		//-32 -3201 -3202 -3203 -3204
		if strings.HasPrefix(typeStr, "-32") {
			sendType, dataOut, status = client.channelModule(msg)
		}
		//-33 -3301 -3302 -3303 -3304
		if strings.HasPrefix(typeStr, "-33") {
			sendType, dataOut, status = client.channelModule(msg)
		}
		//-34 -3400 -3401 -3402 -3403 -3404
		if strings.HasPrefix(typeStr, "-34") {
			sendType, dataOut, status = client.fundingTransactionModule(msg)
		}

		//-35 -3500
		if msg.Type == enum.MsgType_FundingSign_OmniSign ||
			msg.Type == enum.MsgType_FundingSign_BtcSign {
			sendType, dataOut, status = client.fundingSignModule(msg)
		}
		//-38
		if msg.Type == enum.MsgType_CloseChannelRequest || msg.Type == enum.MsgType_CloseChannelSign {
			sendType, dataOut, status = client.channelModule(msg)
		}

		if strings.HasPrefix(typeStr, "-35") {
			//-351 -35101 -35102 -35103 -35104
			if strings.HasPrefix(typeStr, "-351") {
				sendType, dataOut, status = client.commitmentTxModule(msg)
			} else
			//-352 -35201 -35202 -35203 -35204
			if strings.HasPrefix(typeStr, "-352") {
				sendType, dataOut, status = client.commitmentTxSignModule(msg)
			} else
			//-353 -35301 -35302 -35303 -35304
			if strings.HasPrefix(typeStr, "-353") {
				sendType, dataOut, status = client.otherModule(msg)
			} else
			//-354 -35401 -35402 -35403 -35404
			if strings.HasPrefix(typeStr, "-354") {
				sendType, dataOut, status = client.otherModule(msg)
			}
		}

		//-40
		if strings.HasPrefix(typeStr, "-40") {
			sendType, dataOut, status = client.htlcHDealModule(msg)
		}

		if len(dataOut) == 0 {
			dataOut = dataReq
		}

		//broadcast except me
		if sendType == enum.SendTargetType_SendToExceptMe {
			for itemClient := range GlobalWsClientManager.ClientsMap {
				if itemClient != itemClient {
					jsonMessage := getReplyObj(string(dataOut), msg.Type, status, client, itemClient)
					itemClient.SendChannel <- jsonMessage
				}
			}
		}
		//broadcast to all
		if sendType == enum.SendTargetType_SendToAll {
			jsonMessage := getReplyObj(string(dataOut), msg.Type, status, client, nil)
			GlobalWsClientManager.Broadcast <- jsonMessage
		}
	}
}

func getReplyObj(data string, msgType enum.MsgType, status bool, fromClient, toClient *Client) []byte {
	var jsonMessage []byte

	fromId := fromClient.Id
	if fromClient.User != nil {
		fromId = fromClient.User.PeerId
	}

	toClientId := "all"
	if toClient != nil {
		toClientId = toClient.Id
		if toClient.User != nil {
			toClientId = toClient.User.PeerId
		}
	}

	node := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &node)
	if err == nil {
		parse := gjson.Parse(data)
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: parse.Value()})
	} else {
		if strings.Contains(err.Error(), " array into Go value of type map") {
			parse := gjson.Parse(data)
			jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: parse.Value()})
		} else {
			jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromId, To: toClientId, Result: data})
		}
	}
	return jsonMessage
}

func (client *Client) sendToMyself(msgType enum.MsgType, status bool, data string) {
	jsonMessage := getReplyObj(data, msgType, status, client, client)
	client.SendChannel <- jsonMessage
}

func (client *Client) sendToSomeone(msgType enum.MsgType, status bool, recipientPeerId string, data string) error {
	if tool.CheckIsString(&recipientPeerId) {
		itemClient := GlobalWsClientManager.UserMap[recipientPeerId]
		if itemClient != nil && itemClient.User != nil {
			jsonMessage := getReplyObj(data, msgType, status, client, itemClient)
			itemClient.SendChannel <- jsonMessage
			return nil
		}
	}
	return errors.New("recipient not exist or online")
}
func (client *Client) FindUser(peerId *string) (*Client, error) {
	if tool.CheckIsString(peerId) {
		itemClient := GlobalWsClientManager.UserMap[*peerId]
		if itemClient != nil && client.User != nil {
			return itemClient, nil
		}
	}
	return nil, errors.New("user not exist or online")
}
