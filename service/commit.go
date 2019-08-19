package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
)

type commitTxManager struct {
}

var CommitTxService commitTxManager

func (service *commitTxManager) Edit(jsonData string) (node *dao.CommitmentTx, err error) {
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

func (service *commitTxManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTx, count *int, err error) {
	var chanId bean.ChannelID

	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, nil, errors.New("wrong ChannelId")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	pageIndex := gjson.Get(jsonData, "page_index").Int()
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize := gjson.Get(jsonData, "page_size").Int()
	if pageSize <= 0 {
		pageSize = 10
	}
	skip := (pageIndex - 1) * pageSize

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, nil, err
	}

	nodes = []dao.CommitmentTx{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId)).Count(&dao.CommitmentTx{})
	if err != nil {
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId)).Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitTxManager) GetItemById(id int) (node *dao.CommitmentTx, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTx{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitTxManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTx{})
}

func (service *commitTxManager) Del(id int) (node *dao.CommitmentTx, err error) {
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
