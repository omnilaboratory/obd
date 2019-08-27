package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"log"
	"time"
)

var db *storm.DB

func init() {
	var err error
	db, err = dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
	}
}

type fundingTransactionManager struct{}

var FundingTransactionService fundingTransactionManager

//funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) CreateFundingTx(jsonData string) (node *dao.FundingTransaction, err error) {
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
		err = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(channelInfo)
		if err != nil {
			return nil, err
		}
		node.TemporaryChannelId = data.TemporaryChannelId
		node.PropertyId = data.PropertyId
		node.FunderPeerId = channelInfo.FunderPeerId
		node.FundeePeerId = channelInfo.FundeePeerId
		node.FunderPubKey = channelInfo.FunderPubKey
		node.FundeePubKey = channelInfo.FundeePubKey
		if data.FundingPubKey == node.FunderPubKey {
			node.AmountA = data.Amount
		} else {
			node.AmountB = data.Amount
		}
		node.CreateAt = time.Now()
		node.CurrState = dao.FundingTransaction_Create
		err = db.Save(node)
	} else {
		node = nil
		err = errors.New("request has send,please wait")
	}
	return node, err
}

func (service *fundingTransactionManager) ItemByTempId(jsonData string) (node *dao.FundingTransaction, err error) {
	var data = &dao.FundingTransaction{}
	err = db.One("TemporaryChannelId", gjson.Parse(jsonData), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (service *fundingTransactionManager) AllItem(peerId string) (node []dao.FundingTransaction, err error) {
	var data = []dao.FundingTransaction{}
	err = db.Select(q.Or(q.Eq("FundeePeerId", peerId), q.Eq("FunderPeerId", peerId))).OrderBy("CreateAt").Reverse().Find(&data)
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
	return nil
}

func (service *fundingTransactionManager) Del(id int) (err error) {
	var data = &dao.FundingTransaction{}
	count, err := db.Select(q.Eq("Id", id)).Count(data)
	if err == nil && count == 1 {
		err = db.DeleteStruct(data)
	}
	return err
}
func (service *fundingTransactionManager) TotalCount() (count int, err error) {
	var data = &dao.FundingTransaction{}
	return db.Count(data)
}

func (service *fundingTransactionManager) FundingTransactionSign(jsonData string) (signed *dao.FundingSigned, err error) {
	vo := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), vo)
	if err != nil {
		return nil, err
	}

	vo.TemporaryChannelId = bean.ChannelIdService.NextTemporaryChanID()
	node := &dao.FundingSigned{}
	//https://www.ctolib.com/storm.html
	err = db.Select(
		q.Eq("FundeePubKey", vo.FundeePubKey),
		q.Eq("FunderPubKey", vo.FunderPubKey),
		//q.And(
		//	q.Eq("FunderPubKey", vo.FunderPubKey),
		//),
	).First(node)
	node.FundingSigned = *vo
	if err != nil {
		err = db.Save(node)
	} else {
		err = db.Update(node)
	}
	return node, err
}
