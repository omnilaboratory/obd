package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/rpc"
	"LightningOnOmni/service"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"strings"
)

type Client struct {
	Id          string
	User        *bean.User
	Socket      *websocket.Conn
	SendChannel chan []byte
}

func (c *Client) Write() {
	defer func() {
		e := c.Socket.Close()
		if e != nil {
			log.Println(e)
		} else {
			log.Println("socket closed after writing...")
		}
	}()

	for {
		select {
		case _order, ok := <-c.SendChannel:
			if !ok {
				_ = c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Printf("send data: %v \n", string(_order))
			_ = c.Socket.WriteMessage(websocket.TextMessage, _order)
		}
	}
}

func (c *Client) Read() {
	defer func() {
		_ = service.UserService.UserLogout(c.User)
		GlobalWsClientManager.Unregister <- c
		_ = c.Socket.Close()
		log.Println("socket closed after reading...")
	}()

	for {
		_, dataReq, err := c.Socket.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var msg bean.RequestMessage
		log.Println("request data: ", string(dataReq))
		parse := gjson.Parse(string(dataReq))

		if parse.Exists() == false {
			log.Println("wrong json input")
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
			_, err := c.findUser(&msg.RecipientPeerId)
			if err != nil {
				c.sendToMyself(msg.Type, true, "can not find target user")
				continue
			}
		}

		// check the data whether is right signature
		if tool.CheckIsString(&msg.PubKey) && tool.CheckIsString(&msg.Signature) {
			client := rpc.NewClient()
			result, err := client.VerifyMessage(msg.PubKey, msg.Signature, msg.Data)
			if err != nil {
				c.sendToMyself(msg.Type, false, err.Error())
				continue
			}
			if gjson.Parse(result).Bool() == false {
				c.sendToMyself(msg.Type, false, "error signature")
				continue
			}
		}

		var sendType = enum.SendTargetType_SendToNone
		status := false
		var dataOut []byte
		var flag = true
		if msg.Type < 1000 && msg.Type >= 0 {
			sendType, dataOut, status = c.userModule(msg)
			flag = false
		}

		if msg.Type > 1000 {
			sendType, dataOut, status = c.omniCoreModule(msg)
			flag = false
		}

		if flag {
			if c.User == nil {
				c.sendToMyself(msg.Type, false, "please login")
				continue
			}
		}

		typeStr := strconv.Itoa(int(msg.Type))
		//-32 -3201 -3202 -3203 -3204
		if strings.HasPrefix(typeStr, "-32") {
			sendType, dataOut, status = c.channelModule(msg)
		}
		//-33 -3301 -3302 -3303 -3304
		if strings.HasPrefix(typeStr, "-33") {
			sendType, dataOut, status = c.channelModule(msg)
		}
		//-34 -3401 -3402 -3403 -3404
		if strings.HasPrefix(typeStr, "-34") {
			sendType, dataOut, status = c.fundingTransactionModule(msg)
		}

		//-35
		if msg.Type == enum.MsgType_FundingSign_Edit {
			sendType, dataOut, status = c.fundingSignModule(msg)
		}

		if strings.HasPrefix(typeStr, "-35") {
			//-351 -35101 -35102 -35103 -35104
			if strings.HasPrefix(typeStr, "-351") {
				sendType, dataOut, status = c.commitmentTxModule(msg)
			} else
			//-352 -35201 -35202 -35203 -35204
			if strings.HasPrefix(typeStr, "-352") {
				sendType, dataOut, status = c.commitmentTxSignModule(msg)
			} else
			//-353 -35301 -35302 -35303 -35304
			if strings.HasPrefix(typeStr, "-353") {
				sendType, dataOut, status = c.otherModule(msg)
			} else
			//-354 -35401 -35402 -35403 -35404
			if strings.HasPrefix(typeStr, "-354") {
				sendType, dataOut, status = c.otherModule(msg)
			}
		}

		if len(dataOut) == 0 {
			dataOut = dataReq
		}

		//broadcast except me
		if sendType == enum.SendTargetType_SendToExceptMe {
			for client := range GlobalWsClientManager.Clients_map {
				if client != c {
					jsonMessage := getReplyObj(string(dataOut), msg.Type, status, c)
					client.SendChannel <- jsonMessage
				}
			}
		}
		//broadcast to all
		if sendType == enum.SendTargetType_SendToAll {
			jsonMessage := getReplyObj(string(dataOut), msg.Type, status, c)
			GlobalWsClientManager.Broadcast <- jsonMessage
		}
	}
}

func getReplyObj(data string, msgType enum.MsgType, status bool, client *Client) []byte {
	var jsonMessage []byte

	clientId := client.Id
	if client.User != nil {
		clientId = client.User.PeerId
	}

	node := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &node)
	if err == nil {
		parse := gjson.Parse(data)
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: clientId, Result: parse.Value()})
	} else {
		if strings.Contains(err.Error(), " array into Go value of type map") {
			parse := gjson.Parse(data)
			jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: clientId, Result: parse.Value()})
		} else {
			jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: clientId, Result: data})
		}
	}
	return jsonMessage
}

func (c *Client) sendToMyself(msgType enum.MsgType, status bool, data string) {
	jsonMessage := getReplyObj(data, msgType, status, c)
	c.SendChannel <- jsonMessage
}

func (c *Client) sendToSomeone(msgType enum.MsgType, status bool, recipientPeerId string, data string) error {
	if &recipientPeerId != nil {
		for client := range GlobalWsClientManager.Clients_map {
			if client.User != nil && client.User.PeerId == recipientPeerId {
				jsonMessage := getReplyObj(data, msgType, status, c)
				client.SendChannel <- jsonMessage
				return nil
			}
		}
	}
	return errors.New("recipient not exist")
}
func (c *Client) findUser(peerId *string) (client *Client, err error) {
	if tool.CheckIsString(peerId) {
		for client := range GlobalWsClientManager.Clients_map {
			if client.User != nil && client.User.PeerId == *peerId && GlobalWsClientManager.Clients_map[client] {
				return client, nil
			}
		}
	}
	return nil, errors.New("user not exist")
}
