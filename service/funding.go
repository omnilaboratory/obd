package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/dao"
	"encoding/json"
	"github.com/tidwall/gjson"
)

type FundingManager struct{}

var FundingCreateService FundingManager

func (service *FundingManager) CreateFunding(jsonData string) (node *dao.FundingCreated, err error) {
	node = &dao.FundingCreated{}

	sum, tempId, _ := ChannelService.getTemporaryChannelId()
	node.TemporaryChannelIdStr = sum
	node.FunderPubKeyStr = gjson.Get(jsonData, "funderPubKey").String()
	hashes, _ := chainhash.NewHashFromStr(node.FunderPubKeyStr)
	data := &bean.FundingCreated{
		TemporaryChannelId: *tempId,
		FunderPubKey:       *hashes,
		PropertyId:         gjson.Get(jsonData, "propertyId").Int(),
		MaxAssets:          gjson.Get(jsonData, "maxAssets").Float(),
		AmountA:            gjson.Get(jsonData, "amountA").Float(),
	}
	node.FundingCreated = *data

	db, _ := dao.DB_Manager.GetDB()
	err = db.Save(node)
	return node, err
}

func (service *FundingManager) GetFundingTx(id int) (node *dao.FundingCreated, err error) {
	db, _ := dao.DB_Manager.GetDB()
	var data = &dao.FundingCreated{}
	err = db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *FundingManager) DeleteTable() (err error) {
	//db, _ := dao.DB_Manager.GetDB()
	//var data = &dao.FundingCreated{}
	//return db.Drop(data)
	return nil
}
func (service *FundingManager) DeleteItem(id int) (err error) {
	db, _ := dao.DB_Manager.GetDB()
	var data = &dao.FundingCreated{}
	db.One("Id", id, data)
	err = db.DeleteStruct(data)
	return err
}
func (service *FundingManager) TotalCount() (count int, err error) {
	db, _ := dao.DB_Manager.GetDB()
	var data = &dao.FundingCreated{}
	return db.Count(data)
}

type FundingSignManager struct{}

var FundingSignService FundingSignManager

func (service *FundingSignManager) Edit(jsonData string) (signed *dao.FundingSigned, err error) {
	vo := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), vo)
	if err != nil {
		return nil, err
	}
	return signed, nil
}
