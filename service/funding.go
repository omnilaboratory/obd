package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/dao"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"time"
)

type fundingTransactionManager struct{}

var FundingTransactionService fundingTransactionManager

//funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) CreateFundingTx(jsonData string, user *bean.User) (node *dao.FundingTransaction, err error) {
	data := &bean.FundingCreated{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if len(data.TemporaryChannelId) == 0 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	node = &dao.FundingTransaction{}
	count, _ := db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).Count(node)
	if count == 0 {
		channelInfo := &dao.ChannelInfo{}
		err = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
		if err != nil {
			return nil, err
		}

		node.TemporaryChannelId = data.TemporaryChannelId
		node.PropertyId = data.PropertyId

		node.PeerIdA = channelInfo.PeerIdA
		node.PeerIdB = channelInfo.PeerIdB
		node.ChannelPubKey = channelInfo.ChannelPubKey
		node.RedeemScript = channelInfo.RedeemScript

		if user.PeerId == channelInfo.PeerIdA {
			if data.FunderPubKey != channelInfo.PubKeyA {
				return nil, errors.New("invalid FunderPubKey")
			}
		}
		if user.PeerId == channelInfo.PeerIdB {
			if data.FunderPubKey != channelInfo.PubKeyB {
				return nil, errors.New("invalid FunderPubKey")
			}
		}

		if data.FunderPubKey == channelInfo.PubKeyA {
			node.FunderPubKey = channelInfo.PubKeyA
			node.FundeePubKey = channelInfo.PubKeyB
		} else {
			node.FunderPubKey = channelInfo.PubKeyB
			node.FundeePubKey = channelInfo.PubKeyA
		}
		node.AmountA = data.AmountA

		node.CurrState = dao.FundingTransaction_Create
		node.CreateAt = time.Now()
		err = db.Save(node)
	} else {
		node = nil
		err = errors.New("request has send,please wait")
	}
	return node, err
}

func (service *fundingTransactionManager) FundingTransactionSign(jsonData string) (signed *dao.FundingTransaction, err error) {
	data := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if len(data.TemporaryChannelId) == 0 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	var node = &dao.FundingTransaction{}
	//https://www.ctolib.com/storm.html
	err = db.One("TemporaryChannelId", data.TemporaryChannelId, node)
	if err != nil {
		return nil, err
	}
	if data.Attitude {
		node.AmountB = data.AmountB
		node.CurrState = dao.FundingTransactionState_Accept
	} else {
		node.CurrState = dao.FundingTransactionState_Defuse
	}
	node.FundeeSignAt = time.Now()

	err = db.Update(node)
	return node, err
}

func (service *fundingTransactionManager) ItemByTempId(jsonData string) (node *dao.FundingTransaction, err error) {
	var data = &dao.FundingTransaction{}
	var tempChanId chainhash.Hash
	for index, item := range gjson.Parse(jsonData).Array() {
		tempChanId[index] = byte(item.Int())
	}
	err = db.One("TemporaryChannelId", tempChanId, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (service *fundingTransactionManager) AllItem(peerId string) (node []dao.FundingTransaction, err error) {
	var data = []dao.FundingTransaction{}
	err = db.Select(q.Or(q.Eq("PeerIdB", peerId), q.Eq("PeerIdA", peerId))).OrderBy("CreateAt").Reverse().Find(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) ItemById(id int) (node *dao.FundingTransaction, err error) {
	var data = &dao.FundingTransaction{}
	err = db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) DelAll() (err error) {
	var data = &dao.FundingTransaction{}
	return db.Drop(data)
}

func (service *fundingTransactionManager) Del(id int) (err error) {
	var data = &dao.FundingTransaction{}
	count, err := db.Select(q.Eq("Id", id)).Count(data)
	if err == nil && count == 1 {
		err = db.DeleteStruct(data)
	}
	return err
}
func (service *fundingTransactionManager) TotalCount(peerId string) (count int, err error) {
	return db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).Count(&dao.FundingTransaction{})
}
