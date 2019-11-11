package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"log"
)

func (service *commitmentTxManager) GetLatestCommitmentTxByChannelId(jsonData string, user *bean.User) (node *dao.CommitmentTransaction, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty jsonData")
	}
	var reqData = &bean.GetBalanceRequest{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		return nil, errors.New("wrong channelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("Owner", user.PeerId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}

type RequestGetBroadcastCommitmentTx struct {
	ChannelId bean.ChannelID `json:"channel_id"`
}

func (service *commitmentTxManager) GetBroadcastCommitmentTxByChannelId(jsonData string, user *bean.User) (node *dao.CommitmentTransaction, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return nil, errors.New("empty jsonData")
	}
	var reqData = &RequestGetBroadcastCommitmentTx{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		return nil, errors.New("wrong channelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.TxInfoState_SendHex), q.Eq("Owner", user.PeerId)).First(node)
	return node, err
}

func (service *commitmentTxManager) GetBroadcastRDTxByChannelId(jsonData string, user *bean.User) (node interface{}, err error) {
	commitmentTx, err := service.GetBroadcastCommitmentTxByChannelId(jsonData, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	node = &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("CommitmentTxId", commitmentTx.Id), q.Eq("CurrState", dao.TxInfoState_SendHex), q.Eq("Owner", user.PeerId)).First(node)
	return node, err
}
func (service *commitmentTxManager) GetBroadcastBRTxByChannelId(jsonData string, user *bean.User) (node interface{}, err error) {
	commitmentTx, err := service.GetBroadcastCommitmentTxByChannelId(jsonData, user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	node = &dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("CommitmentTxId", commitmentTx.Id)).First(node)
	return node, err
}

func (service *commitmentTxManager) GetLatestRDTxByChannelId(jsonData string, user *bean.User) (node *dao.RevocableDeliveryTransaction, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", chanId)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}

func (service *commitmentTxManager) GetLatestAllRDByChannelId(jsonData string, user *bean.User) (nodes []dao.RevocableDeliveryTransaction, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", chanId)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	nodes = []dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().Find(&nodes)
	return nodes, err
}

func (service *commitmentTxManager) GetLatestBRTxByChannelId(jsonData string, user *bean.User) (node *dao.BreachRemedyTransaction, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", chanId)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(node)
	return node, err
}

func (service *commitmentTxManager) GetLatestAllBRByChannelId(jsonData string, user *bean.User) (nodes []dao.BreachRemedyTransaction, err error) {
	var chanId bean.ChannelID
	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", chanId)).First(channelInfo)
	if err != nil {
		return nil, err
	}

	nodes = []dao.BreachRemedyTransaction{}
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().Find(&nodes)
	return nodes, err
}

func (service *commitmentTxManager) GetItemsByChannelId(jsonData string, user *bean.User) (nodes []dao.CommitmentTransaction, count *int, err error) {
	var chanId bean.ChannelID

	array := gjson.Get(jsonData, "channel_id").Array()
	if len(array) != 32 {
		return nil, nil, errors.New("wrong channel_id")
	}
	for index, value := range array {
		chanId[index] = byte(value.Num)
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", chanId)).First(channelInfo)
	if err != nil {
		return nil, nil, err
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

	nodes = []dao.CommitmentTransaction{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId)).Count(&dao.CommitmentTransaction{})
	if err != nil {
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitmentTxManager) GetItemById(id int) (node *dao.CommitmentTransaction, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitmentTxManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.CommitmentTransaction{})
}

func (service *commitmentTxManager) Del(id int) (node *dao.CommitmentTransaction, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTransaction{}
	err = db.One("Id", id, node)
	if err != nil {
		return nil, err
	}
	err = db.DeleteStruct(node)
	return node, err
}

func (service *commitmentTxSignedManager) GetItemsByChannelId(jsonData string) (nodes []dao.CommitmentTransaction, count *int, err error) {
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
		log.Println(err)
		return nil, nil, err
	}

	nodes = []dao.CommitmentTransaction{}
	tempCount, err := db.Select(q.Eq("ChannelId", chanId)).Count(&dao.CommitmentTransaction{})
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	count = &tempCount
	err = db.Select(q.Eq("ChannelId", chanId)).Skip(int(skip)).Limit(int(pageSize)).Find(&nodes)
	return nodes, count, err
}

func (service *commitmentTxSignedManager) GetItemById(id int) (node *dao.CommitmentTransaction, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	node = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("Id", id)).First(node)
	return node, nil
}

func (service *commitmentTxSignedManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return db.Count(&dao.CommitmentTransaction{})
}
