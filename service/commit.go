package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
)

type commnitTxManager struct {
}

var CommnitTxService commnitTxManager

func (service *commnitTxManager) Edit(jsonData string) (node *dao.CommitmentTx, err error) {
	if len(jsonData) == 0 {
		return nil, errors.New("empty json data")
	}
	data := &bean.CommitmentTx{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}
	if len(data.ChannelId) != 32 {
		return nil, errors.New("wrong ChannelId")
	}
	node = &dao.CommitmentTx{}
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node.CommitmentTx = *data
	err = db.Save(node)
	return node, err
}

func (service *commnitTxManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTx, err error) {
	var chanId bean.ChannelID
	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong ChannelId")
	}

	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	nodes = []dao.CommitmentTx{}
	err = db.Select(q.Eq("ChannelId", chanId)).Find(&nodes)
	return nodes, nil
}

func (service *commnitTxManager) GetItemById(id int) (node *dao.CommitmentTx, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTx{}
	err = db.Select(q.Eq("Id", id)).Find(&node)
	return node, nil
}

func (service *commnitTxManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTx{})
}

func (service *commnitTxManager) Del(id int) (node *dao.CommitmentTx, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTx{}
	err = db.One("Id", id, node)
	if err != nil {
		return nil, err
	}
	err = db.DeleteStruct(node)
	return node, err
}
