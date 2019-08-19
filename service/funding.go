package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
)

type fundingCreateManager struct {
}

var FundingCreateService fundingCreateManager

func (service *fundingCreateManager) Edit(jsonData string) (node *dao.FundingCreated, err error) {

	data := &bean.FundingCreated{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if len(data.TemporaryChannelId) == 0 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	node = &dao.FundingCreated{}
	count, _ := db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).Count(node)

	node.FundingCreated = *data

	if count == 0 {
		err = db.Save(node)
	} else {
		err = db.Update(node)
	}
	return node, err
}

func (service *fundingCreateManager) Item(id int) (node *dao.FundingCreated, err error) {
	db, _ := dao.DBService.GetDB()
	var data = &dao.FundingCreated{}
	err = db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingCreateManager) DelAll() (err error) {
	db, _ := dao.DBService.GetDB()
	var data = &dao.FundingCreated{}
	return db.Drop(data)
	return nil
}
func (service *fundingCreateManager) Del(id int) (err error) {
	db, _ := dao.DBService.GetDB()
	var data = &dao.FundingCreated{}
	count, err := db.Select(q.Eq("Id", id)).Count(data)
	if err == nil && count == 1 {
		err = db.DeleteStruct(data)
	}
	return err
}
func (service *fundingCreateManager) TotalCount() (count int, err error) {
	db, _ := dao.DBService.GetDB()
	var data = &dao.FundingCreated{}
	return db.Count(data)
}

type fundingSignManager struct{}

var FundingSignService fundingSignManager

func (service *fundingSignManager) Edit(jsonData string) (signed *dao.FundingSigned, err error) {
	vo := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), vo)
	if err != nil {
		return nil, err
	}

	vo.TemporaryChannelId = bean.ChannelIdService.NextTemporaryChanID()
	db, _ := dao.DBService.GetDB()
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
func (service *fundingSignManager) Item(id int) (signed *dao.FundingSigned, err error) {
	node := &dao.FundingSigned{}
	db, _ := dao.DBService.GetDB()
	err = db.One("Id", id, node)
	return node, err
}
func (service *fundingSignManager) Del(id int) (signed *dao.FundingSigned, err error) {
	db, _ := dao.DBService.GetDB()
	node := &dao.FundingSigned{}
	err = db.One("Id", id, node)
	if err == nil {
		err = db.DeleteStruct(node)
	}
	return node, err
}
func (service *fundingSignManager) DelAll() (err error) {
	db, _ := dao.DBService.GetDB()
	err = db.Drop(&dao.FundingSigned{})
	return err
}

func (service *fundingSignManager) TotalCount() (count int, err error) {
	db, _ := dao.DBService.GetDB()
	return db.Count(&dao.FundingSigned{})
}
