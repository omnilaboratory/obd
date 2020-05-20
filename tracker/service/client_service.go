package service

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/tidwall/gjson"
	"log"
)

var tracker *ObdNode

func init() {
	tracker = &ObdNode{Id: tool.GetTrackerNodeId()}
}

type ObdNode struct {
	Id          string
	Socket      *websocket.Conn
	SendChannel chan []byte
	IsLogin     bool
}

func (this *ObdNode) sendMsgBackToSender(msgType bean.MsgType, status bool, data string) {
	jsonMessage := getReplyObj(data, msgType, status, tracker, this)
	this.SendChannel <- jsonMessage
}

func (this *ObdNode) findSomeNode(nodeId *string) (*ObdNode, error) {
	if tool.CheckIsString(nodeId) {
		itemClient := ObdNodeManager.ObdNodeMap[*nodeId]
		if itemClient != nil {
			return itemClient, nil
		}
	}
	return nil, errors.New("user not exist or online")
}

func (this *ObdNode) Read() {
	defer func() {
		ObdNodeManager.Disconnected <- this
		_ = this.Socket.Close()
		log.Println("socket closed after reading...")
	}()

	for {
		_, dataReq, err := this.Socket.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("get data from client: ", string(dataReq))
		reqDataJson := gjson.Parse(string(dataReq))
		msgType := bean.MsgType(reqDataJson.Get("type").Int())
		if bean.CheckExist(msgType) == false {
			this.sendMsgBackToSender(msgType, false, "not exist the msg type")
			continue
		}
		if msgType > 3 && this.IsLogin == false {
			this.sendMsgBackToSender(msgType, false, "obd need login")
			continue
		}

		msgData := reqDataJson.Get("data").String()
		switch msgType {
		case bean.MsgType_HeartBeat_2:
			this.sendMsgBackToSender(msgType, true, "echo")
		case bean.MsgType_NodeLogin_3:
			retData, err := nodeAccountService.login(this, msgData)
			sendDataBackToSender(this, msgType, retData, err)
		case bean.MsgType_UserLogin_4:
			retData, err := nodeAccountService.userLogin(this, msgData)
			sendDataBackToSender(this, msgType, retData, err)
		case bean.MsgType_UserLogout_5:
			_ = nodeAccountService.userLogout(this, msgData)
		}
	}
}

func sendDataBackToSender(this *ObdNode, msgType bean.MsgType, retData interface{}, err error) {
	var data interface{}
	status := false
	if err != nil {
		data = err.Error()
	} else {
		status = true
		data = retData
	}
	bytes, _ := json.Marshal(data)
	this.sendMsgBackToSender(msgType, status, string(bytes))
}

func (this *ObdNode) Write() {
	defer func() {
		e := this.Socket.Close()
		if e != nil {
			log.Println(e)
		} else {
			log.Println("socket closed after writing...")
		}
	}()

	for {
		select {
		case data, ok := <-this.SendChannel:
			if !ok {
				_ = this.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := this.Socket.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("fail to send data ", string(data))
			} else {
				log.Println("send data", string(data))
			}
		}
	}
}

func sendToSomeObdNode(msgType bean.MsgType, status bool, recipientObdId string, data string) error {
	if tool.CheckIsString(&recipientObdId) {
		recipientNode := ObdNodeManager.ObdNodeMap[recipientObdId]
		if recipientNode != nil {
			jsonMessage := getReplyObj(data, msgType, status, tracker, recipientNode)
			recipientNode.SendChannel <- jsonMessage
			return nil
		}
	}
	return errors.New("recipient not exist or online")
}

func getReplyObj(data string, msgType bean.MsgType, status bool, fromClient, toClient *ObdNode) []byte {
	parse := gjson.Parse(data)
	result := parse.Value()
	if parse.Exists() == false {
		result = data
	}
	jsonMessage, _ := json.Marshal(&bean.ReplyMessage{Type: msgType, Status: status, From: fromClient.Id, To: toClient.Id, Result: result})
	return jsonMessage
}

type obdNodeManager struct {
	Broadcast    chan []byte
	Connected    chan *ObdNode
	Disconnected chan *ObdNode
	ClientsMap   map[*ObdNode]bool
	ObdNodeMap   map[string]*ObdNode
}

var ObdNodeManager = obdNodeManager{
	Broadcast:    make(chan []byte),
	Connected:    make(chan *ObdNode),
	Disconnected: make(chan *ObdNode),
	ClientsMap:   make(map[*ObdNode]bool),
	ObdNodeMap:   make(map[string]*ObdNode),
}

func (endManager *obdNodeManager) TrackerStart() {
	for {
		select {
		case newConn := <-endManager.Connected:
			endManager.ClientsMap[newConn] = true
			endManager.ObdNodeMap[newConn.Id] = newConn
			newConn.sendMsgBackToSender(bean.MsgType_Connect_1, true, "connect server successfully")
		case currConn := <-endManager.Disconnected:
			if _, ok := endManager.ClientsMap[currConn]; ok {
				currConn.sendMsgBackToSender(bean.MsgType_Connect_1, true, "disconnect from server successfully")
				_ = nodeAccountService.logout(currConn)
				delete(endManager.ClientsMap, currConn)
				delete(endManager.ObdNodeMap, currConn.Id)
				close(currConn.SendChannel)
			}
		case msg := <-endManager.Broadcast:
			for conn := range endManager.ClientsMap {
				select {
				case conn.SendChannel <- msg:
				default:
					close(conn.SendChannel)
					delete(endManager.ClientsMap, conn)
				}
			}
		}
	}
}
