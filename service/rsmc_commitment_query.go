package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
)

type requestGetInfoByChannelId struct {
	ChannelId string `json:"channel_id"`
}

func getChannelIdFromJson(jsonData string) (channelId string, err error) {
	if tool.CheckIsString(&jsonData) == false {
		return "", errors.New(enum.Tips_common_empty + "input data")
	}
	var reqData = &requestGetInfoByChannelId{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return "", err
	}

	if tool.CheckIsString(&reqData.ChannelId) == false {
		return "", errors.New(enum.Tips_common_empty + "channelId")
	}
	return reqData.ChannelId, nil
}

func (this *commitmentTxManager) GetItemsByChannelId(jsonData string, user *bean.User) (nodes []dao.CommitmentTransaction, count *int, err error) {
	channelId, err := getChannelIdFromJson(jsonData)
	if err != nil {
		return nil, nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId)).
		First(channelInfo)
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

	nodes = []dao.CommitmentTransaction{}
	tempCount, err := user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", user.PeerId)).
		Count(&dao.CommitmentTransaction{})
	if err != nil {
		return nil, nil, err
	}
	count = &tempCount
	err = user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		Skip(int(skip)).
		Limit(int(pageSize)).
		Find(&nodes)
	return nodes, count, err
}

func (this *commitmentTxManager) GetItemById(id int, user bean.User) (node *dao.CommitmentTransaction, err error) {
	node = &dao.CommitmentTransaction{}
	err = user.Db.Select(
		q.Eq("Id", id)).
		First(node)
	return node, nil
}

func (this *commitmentTxManager) TotalCount(jsonData string, user bean.User) (count int, err error) {
	channelId, err := getChannelIdFromJson(jsonData)
	if err != nil {
		return 0, err
	}
	return user.Db.Select(q.Eq("ChannelId", channelId)).Count(&dao.CommitmentTransaction{})
}

func (this *commitmentTxManager) GetLatestCommitmentTxByChannelId(jsonData string, user *bean.User) (node *dao.CommitmentTransaction, err error) {
	channelId, err := getChannelIdFromJson(jsonData)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId)).
		First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.CommitmentTransaction{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", user.PeerId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(node)
	return node, err
}

func (this *commitmentTxManager) GetLatestRDTxByChannelId(jsonData string, user *bean.User) (node *dao.RevocableDeliveryTransaction, err error) {
	channelId, err := getChannelIdFromJson(jsonData)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId)).
		First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.RevocableDeliveryTransaction{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		First(node)
	return node, err
}

func (this *commitmentTxManager) GetAllRDByChannelId(jsonData string, user *bean.User) (nodes []dao.RevocableDeliveryTransaction, err error) {
	channelId, err := getChannelIdFromJson(jsonData)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId)).
		First(channelInfo)
	if err != nil {
		return nil, err
	}

	nodes = []dao.RevocableDeliveryTransaction{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		Find(&nodes)
	return nodes, err
}

func (this *commitmentTxManager) GetAllBRByChannelId(jsonData string, user *bean.User) (nodes []dao.BreachRemedyTransaction, err error) {
	channelId, err := getChannelIdFromJson(jsonData)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId)).
		First(channelInfo)
	if err != nil {
		return nil, err
	}

	nodes = []dao.BreachRemedyTransaction{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		Find(&nodes)
	return nodes, err
}

func (this *commitmentTxManager) GetLatestBRTxByChannelId(jsonData string, user *bean.User) (node *dao.BreachRemedyTransaction, err error) {
	channelId, err := getChannelIdFromJson(jsonData)
	if err != nil {
		return nil, err
	}

	channelInfo := &dao.ChannelInfo{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId)).
		First(channelInfo)
	if err != nil {
		return nil, err
	}

	node = &dao.BreachRemedyTransaction{}
	err = user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		First(node)
	return node, err
}
