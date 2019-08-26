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
	"time"
)

type channelManager struct{}

var ChannelService = channelManager{}

// openChannel init data
func (c *channelManager) OpenChannel(jsonData string, funderId string) (data *bean.OpenChannelInfo, err error) {
	data = &bean.OpenChannelInfo{}
	json.Unmarshal([]byte(jsonData), &data)
	if len(data.FundingPubKey) != 34 {
		return nil, errors.New("wrong FundingPubKey")
	}
	client := rpc.NewClient()
	isMine, err := client.Validateaddress(data.FundingPubKey)
	if err != nil {
		return nil, err
	}
	if isMine == false {
		return nil, errors.New("invalid fundingPubKey")
	}

	data.ChainHash = config.Init_node_chain_hash
	tempId := bean.ChannelIdService.NextTemporaryChanID()
	data.TemporaryChannelId = tempId

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node := &dao.OpenChannelInfo{}
	node.OpenChannelInfo = *data
	node.FundeePeerId = funderId
	node.FunderPubKey = data.FundingPubKey

	db.Save(node)
	return data, err
}

func (c *channelManager) AcceptChannel(jsonData string, fundeeId string) (node *dao.OpenChannelInfo, err error) {
	data := &bean.OpenChannelInfo{}
	json.Unmarshal([]byte(jsonData), &data)
	if len(data.FundingPubKey) != 34 {
		return nil, errors.New("wrong FundingPubKey")
	}

	if len(data.TemporaryChannelId) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	client := rpc.NewClient()
	isMine, err := client.Validateaddress(data.FundingPubKey)
	if err != nil {
		return nil, err
	}
	if isMine == false {
		return nil, errors.New("invalid fundingPubKey")
	}

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	node = &dao.OpenChannelInfo{}
	db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(node)
	if &node.Id == nil {
		return nil, errors.New("invalid TemporaryChannelId")
	}
	node.FundeePubKey = data.FundingPubKey
	node.FundeePeerId = fundeeId

	multiSig, err := client.CreateMultiSig(2, []string{node.FunderPubKey, node.FundeePubKey})
	node.ChannelPubKey = gjson.Get(multiSig, "address").String()
	node.RedeemScript = gjson.Get(multiSig, "redeemScript").String()
	node.AcceptAt = time.Now()
	err = db.Update(node)
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
	db.Select().OrderBy("CreateAt").Find(&data)
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
