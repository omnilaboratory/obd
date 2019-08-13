package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/dao"
	"github.com/tidwall/gjson"
)

type FundingManager struct {
}

var FundingService FundingManager

func (service *FundingManager) CreateFunding(jsonData string) (node *bean.FundingCreated, err error) {
	db, e := dao.DB_Manager.GetDB()
	if e != nil {
		return nil, e
	}
	tempId, _ := ChannelService.getTemporaryChannelId()
	hashes, _ := chainhash.NewHashFromStr(gjson.Get(jsonData, "funderPubKey").String())
	node = &bean.FundingCreated{
		TemporaryChannelId: *tempId,
		FunderPubkey:       *hashes,
		PropertyId:         gjson.Get(jsonData, "propertyId").Int(),
		MaxAssets:          gjson.Get(jsonData, "maxAssets").Float(),
		AmountA:            gjson.Get(jsonData, "amountA").Float(),
	}

	return node, db.Save(node)
}
