package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/config"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"log"
)

type channelManager struct{}

var ChannelService = channelManager{}

// openChannel init data
func (c *channelManager) OpenChannel(jsonData string) (node *dao.OpenChannelInfo, err error) {
	var data = bean.OpenChannelInfo{}
	json.Unmarshal([]byte(jsonData), &data)
	if len(data.FundingPubKey) != 34 {
		return nil, errors.New("wrong FundingPubKey")
	}

	client := rpc.NewClient()
	address, err := client.GetNewAddress("serverBob")
	if err != nil {
		return nil, err
	}

	multiAddr, err := client.CreateMultiSig(2, []string{string(data.FundingPubKey), address})
	if err != nil {
		return nil, err
	}
	log.Println(multiAddr)

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	data.ChainHash = config.Init_node_chain_hash
	tempId := bean.ChannelIdService.NextTemporaryChanID()
	data.TemporaryChannelId = tempId
	node = &dao.OpenChannelInfo{}
	node.OpenChannelInfo = data
	err = db.Save(node)
	return node, err
}

// GetChannelByTemporaryChanId
func (c *channelManager) GetChannelByTemporaryChanId(jsonData string) (node *dao.OpenChannelInfo, err error) {
	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	var tempChanId chainhash.Hash
	for index, value := range array {
		tempChanId[index] = byte(value.Num)
	}

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.OpenChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", tempChanId)).First(node)
	return node, err
}

// DelChannelByTemporaryChanId
func (c *channelManager) DelChannelByTemporaryChanId(jsonData string) (node *dao.OpenChannelInfo, err error) {

	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	var tempChanId chainhash.Hash
	for index, value := range array {
		tempChanId[index] = byte(value.Num)
	}

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.OpenChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", tempChanId)).First(node)
	if err == nil {
		err = db.DeleteStruct(node)
	}
	return node, err
}

// TotalCount
func (c *channelManager) AllItem() (data []dao.OpenChannelInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	data = []dao.OpenChannelInfo{}
	err = db.All(&data)
	return data, err
}

// TotalCount
func (c *channelManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.OpenChannelInfo{})
}
