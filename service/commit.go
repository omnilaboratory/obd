package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"time"
)

type commitTxManager struct{}

var CommitmentTxService commitTxManager

func (service *commitTxManager) Edit(jsonData string) (node *dao.CommitmentTxInfo, err error) {
	if len(jsonData) == 0 {
		return nil, errors.New("empty json data")
	}
	data := &bean.CommitmentTx{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}
	if len(data.ChannelId) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	node = &dao.CommitmentTxInfo{}
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node.CreateAt = time.Now()
	err = db.Save(node)
	return node, err
}
func (service *commitTxManager) GetNewestCommitmentTxByChannelId(jsonData string, user *bean.User) (node *dao.CommitmentTxInfo, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}
	node = &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}

func (service *commitTxManager) GetNewestRDTxByChannelId(jsonData string, user *bean.User) (node *dao.RevocableDeliveryTransaction, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}
	node = &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}
func (service *commitTxManager) GetNewestBRTxByChannelId(jsonData string, user *bean.User) (node *dao.BreachRemedyTransaction, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}
	node = &dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}

func (service *commitTxManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTxInfo, count *int, err error) {
	var chanId bean.ChannelID

	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, nil, errors.New("wrong channel_id")
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

	nodes = []dao.CommitmentTxInfo{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId)).Count(&dao.CommitmentTxInfo{})
	if err != nil {
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId)).OrderBy("CreateAt").Reverse().Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitTxManager) GetItemById(id int) (node *dao.CommitmentTxInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitTxManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTxInfo{})
}

func (service *commitTxManager) Del(id int) (node *dao.CommitmentTxInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTxInfo{}
	err = db.One("Id", id, node)
	if err != nil {
		return nil, err
	}
	err = db.DeleteStruct(node)
	return node, err
}

type commitTxSignedManager struct{}

var CommitTxSignedService commitTxSignedManager

func (service *commitTxSignedManager) Edit(jsonData string) (node *dao.CommitmentTxInfo, err error) {
	if len(jsonData) == 0 {
		return nil, errors.New("empty json data")
	}
	data := &bean.CommitmentTxSigned{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}
	if len(data.ChannelId) != 32 {
		return nil, errors.New("wrong ChannelId")
	}
	node = &dao.CommitmentTxInfo{}
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node.CreateAt = time.Now()
	err = db.Save(node)
	return node, err
}

func (service *commitTxSignedManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTxInfo, count *int, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()

	if len(array) != 32 {
		return nil, nil, errors.New("wrong channel_id")
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

	nodes = []dao.CommitmentTxInfo{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId)).Count(&dao.CommitmentTxInfo{})
	if err != nil {
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId)).Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitTxSignedManager) GetItemById(id int) (node *dao.CommitmentTxInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTxInfo{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitTxSignedManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTxInfo{})
}
