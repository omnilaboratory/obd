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

func (service *FundingManager) CreateFunding(jsonData string) (node *bean.Funding_created, err error) {
	db, e := dao.DB_Manager.GetDB()
	if e != nil {
		return nil, e
	}
	tempId, _ := ChannelService.getTemporaryChannelId()
	hashes, _ := chainhash.NewHashFromStr(gjson.Get(jsonData, "funderPubKey").String())
	node = &bean.Funding_created{
		Temporary_channel_id: *tempId,
		Funder_pubKey:        *hashes,
		Property_id:          gjson.Get(jsonData, "propertyId").Int(),
		Max_assets:           gjson.Get(jsonData, "maxAssets").Float(),
		Amount_a:             gjson.Get(jsonData, "amountA").Float(),
	}

	return node, db.Save(node)
}
