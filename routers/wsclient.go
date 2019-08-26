package routers

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/service"
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
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Printf("send data: %v \n", string(_order))
			c.Socket.WriteMessage(websocket.TextMessage, _order)
		}
	}
}

func (c *Client) Read() {
	defer func() {
		service.UserService.UserLogout(c.User)
		GlobalWsClientManager.Unregister <- c
		c.Socket.Close()
		log.Println("socket closed after reading...")
	}()

	for {
		_, dataReq, err := c.Socket.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var msg bean.RequestMessage
		log.Println(string(dataReq))
		parse := gjson.Parse(string(dataReq))

		if parse.Exists() == false {
			log.Println("wrong json input")
			continue
		}

		msg.Type = enum.MsgType(parse.Get("type").Int())
		msg.Data = parse.Get("data").String()
		msg.Sender = parse.Get("sender").String()
		msg.Recipient = parse.Get("recipient").String()

		var sender, receiver bean.User
		json.Unmarshal([]byte(msg.Sender), &sender)
		json.Unmarshal([]byte(msg.Recipient), &receiver)

		var sendType = enum.SendTargetType_SendToNone
		data := ""
		status := false
		if msg.Type < 1000 && msg.Type >= 0 {
			sendType, dataReq, status = c.userModule(msg, sendType, dataReq, &data, status, receiver, err)
		}
		if msg.Type > 1000 {
			sendType, dataReq, status = c.omniCoreModule(msg, sendType, dataReq, &data, status, receiver, err)
		}

		typeStr := strconv.Itoa(int(msg.Type))
		//-32 -3201 -3202 -3203 -3204
		if strings.HasPrefix(typeStr, "-32") {
			sendType, dataReq, status, data = c.channelModule(msg, sendType, dataReq, status, receiver, err)
		}
		//-33 -3301 -3302 -3303 -3304
		if strings.HasPrefix(typeStr, "-33") {
			sendType, dataReq, status, data = c.channelModule(msg, sendType, dataReq, status, receiver, err)
		}
		//-34 -3401 -3402 -3403 -3404
		if strings.HasPrefix(typeStr, "-34") {
			sendType, dataReq, status, data = c.fundCreateModule(msg, sendType, dataReq, status, receiver, err)
		}

		if strings.HasPrefix(typeStr, "-35") {
			//-351 -35101 -35102 -35103 -35104
			if strings.HasPrefix(typeStr, "-351") {
				sendType, dataReq, status, data = c.commitmentTxModule(msg, sendType, dataReq, status, receiver, err)
			} else
			//-352 -35201 -35202 -35203 -35204
			if strings.HasPrefix(typeStr, "-352") {
				sendType, dataReq, status, data = c.commitmentTxSignModule(msg, sendType, dataReq, status, receiver, err)
			} else
			//-353 -35301 -35302 -35303 -35304
			if strings.HasPrefix(typeStr, "-353") {
				sendType, dataReq, status, data = c.otherModule(msg, sendType, dataReq, status, receiver, err)
			} else
			//-354 -35401 -35402 -35403 -35404
			if strings.HasPrefix(typeStr, "-354") {
				sendType, dataReq, status, data = c.otherModule(msg, sendType, dataReq, status, receiver, err)
			} else
			//-35 -3501 -3502 -3503 -3504
			{
				sendType, dataReq, status, data = c.fundSignModule(msg, sendType, dataReq, status, receiver, err)
			}
		}

		//broadcast except me
		if sendType == enum.SendTargetType_SendToExceptMe {
			for client := range GlobalWsClientManager.Clients_map {
				if client != c {
					jsonMessage := getReplyObj(string(dataReq), msg.Type, status, c)
					client.SendChannel <- jsonMessage
				}
			}
		}
		//broadcast to all
		if sendType == enum.SendTargetType_SendToAll {
			jsonMessage := getReplyObj(string(dataReq), msg.Type, status, c)
			GlobalWsClientManager.Broadcast <- jsonMessage
		}
	}
}

func getReplyObj(data string, msgType enum.MsgType, status bool, client *Client) []byte {
	var jsonMessage []byte
	node := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), node)
	if err == nil {
		parse := gjson.Parse(data)
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: client.Id, Result: parse.Value()})
	} else {
		jsonMessage, _ = json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, Sender: client.Id, Result: data})
	}
	return jsonMessage
}

func (c *Client) sendToMyself(msgType enum.MsgType, status bool, data string) {
	jsonMessage := getReplyObj(data, msgType, status, c)
	c.SendChannel <- jsonMessage
}

func (c *Client) sendToSomeone(msgType enum.MsgType, status bool, recipient *bean.User, data string) error {
	if recipient != nil {
		for client := range GlobalWsClientManager.Clients_map {
			if client.User.PeerId == recipient.PeerId {
				jsonMessage := getReplyObj(data, msgType, status, c)
				client.SendChannel <- jsonMessage
				return nil
			}
		}
	}
	return errors.New("recipient not exist")
}
