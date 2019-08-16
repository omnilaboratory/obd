package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/config"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"log"
)

type ChannelManager struct{}

var ChannelService = ChannelManager{}

// openChannel init data
func (c *ChannelManager) OpenChannel(jsonData string) (node *dao.OpenChannelInfo, err error) {
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

	db, err := dao.DB_Manager.GetDB()
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
func (c *ChannelManager) GetChannelByTemporaryChanId(jsonData string) (node *dao.OpenChannelInfo, err error) {

	var data = bean.OpenChannelInfo{}
	json.Unmarshal([]byte(jsonData), &data)
	if len(data.TemporaryChannelId) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	db, err := dao.DB_Manager.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.OpenChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(node)
	return node, err
}

// DelChannelByTemporaryChanId
func (c *ChannelManager) DelChannelByTemporaryChanId(jsonData string) (node *dao.OpenChannelInfo, err error) {

	var data = bean.OpenChannelInfo{}
	json.Unmarshal([]byte(jsonData), &data)
	if len(data.TemporaryChannelId) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	db, err := dao.DB_Manager.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.OpenChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(node)
	if err == nil {
		err = db.DeleteStruct(node)
	}
	return node, err
}

// TotalCount
func (c *ChannelManager) AllItem() (data []dao.OpenChannelInfo, err error) {
	db, err := dao.DB_Manager.GetDB()
	if err != nil {
		return nil, err
	}
	data = []dao.OpenChannelInfo{}
	err = db.All(&data)
	return data, err
}

// TotalCount
func (c *ChannelManager) TotalCount() (count int, err error) {
	db, err := dao.DB_Manager.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.OpenChannelInfo{})
}
