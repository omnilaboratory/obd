package service

import (
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"github.com/asdine/storm/q"
	"time"
)

type Message struct {
	Id        int             `storm:"id,increment" json:"id" `
	HashValue string          `json:"hash_value"`
	Sender    string          `json:"sender"`
	Receiver  string          `json:"receiver"`
	Data      string          `json:"data"`
	CurrState dao.NormalState `json:"curr_state"`
	CreateAt  time.Time       `json:"create_at"`
}

type messageManage struct {
}

var MessageService messageManage

func (this *messageManage) saveMsg(sender, receiver, data string) string {
	tool.CheckIsString(&data)
	msg := &Message{
		Sender:    sender,
		Receiver:  receiver,
		Data:      data,
		CreateAt:  time.Now(),
		CurrState: dao.NS_Create,
	}

	bytes, err := json.Marshal(msg)
	if err == nil {
		msg.HashValue = tool.SignMsgWithSha256(bytes)
		err = db.Save(msg)
		if err == nil {
			return msg.HashValue
		}
	}
	return ""
}
func (this *messageManage) getMsg(hashValue string) (*Message, error) {
	msg := &Message{}
	err := db.Select(q.Eq("HashValue", hashValue)).First(msg)
	if err == nil {
		msg.CurrState = dao.NS_Finish
		_ = db.Update(msg)
		return msg, nil
	}
	return nil, err
}
