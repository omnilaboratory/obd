package service

import (
	"encoding/json"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
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
	ReadAt    time.Time       `json:"read_at"`
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

func (this *messageManage) saveMsgUseTx(tx storm.Node, sender, receiver, data string) string {
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
		err = tx.Save(msg)
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
		msg.ReadAt = time.Now()
		_ = db.Update(msg)
		return msg, nil
	}
	return nil, err
}
func (this *messageManage) getMsgUseTx(tx storm.Node, hashValue string) (*Message, error) {
	msg := &Message{}
	err := tx.Select(q.Eq("HashValue", hashValue), q.Eq("CurrState", dao.NS_Create)).First(msg)
	if err == nil {
		msg.ReadAt = time.Now()
		_ = tx.Update(msg)
		return msg, nil
	}
	return nil, err
}
func (this *messageManage) updateMsgStateUseTx(tx storm.Node, msg *Message) error {
	msg.CurrState = dao.NS_Finish
	return tx.Update(msg)
}
