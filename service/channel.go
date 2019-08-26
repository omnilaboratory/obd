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
	"strconv"
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
	node.FunderPeerId = funderId
	node.FunderPubKey = data.FundingPubKey
	node.CurrState = dao.OpenChannelState_Create
	node.CreateAt = time.Now()

	db.Save(node)
	return data, err
}

func (c *channelManager) AcceptChannel(jsonData string, fundeeId string) (node *dao.OpenChannelInfo, err error) {
	data := &bean.AcceptChannelInfo{}
	err = json.Unmarshal([]byte(jsonData), &data)

	if err != nil {
		return nil, err
	}

	if len(data.TemporaryChannelId) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	client := rpc.NewClient()

	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}

	if data.Attitude {
		if len(data.FundingPubKey) != 34 {
			return nil, errors.New("wrong FundingPubKey")
		}
		isMine, err := client.Validateaddress(data.FundingPubKey)
		if err != nil {
			return nil, err
		}
		if isMine == false {
			return nil, errors.New("invalid fundingPubKey")
		}
	}

	node = &dao.OpenChannelInfo{}
	db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(node)

	if &node.Id == nil {
		return nil, errors.New("invalid TemporaryChannelId")
	}
	if node.CurrState != dao.OpenChannelState_Create {
		return nil, errors.New("invalid ChannelState " + strconv.Itoa(int(node.CurrState)))
	}

	if data.Attitude {
		node.FundeePubKey = data.FundingPubKey
		node.FundeePeerId = fundeeId
		multiSig, err := client.CreateMultiSig(2, []string{node.FunderPubKey, node.FundeePubKey})
		if err != nil {
			return nil, err
		}
		node.ChannelPubKey = gjson.Get(multiSig, "address").String()
		node.RedeemScript = gjson.Get(multiSig, "redeemScript").String()
	}
	node.AcceptAt = time.Now()
	if data.Attitude {
		node.CurrState = dao.OpenChannelState_Accept
	} else {
		node.CurrState = dao.OpenChannelState_Defuse
	}
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

// AllItem
func (c *channelManager) AllItem(peerId string) (data []dao.OpenChannelInfo, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return nil, err
	}
	infos := []dao.OpenChannelInfo{}
	err = db.Select(q.Or(q.Eq("FundeePeerId", peerId), q.Eq("FunderPeerId", peerId))).OrderBy("CreateAt").Reverse().Find(&infos)
	return infos, err
}

// TotalCount
func (c *channelManager) TotalCount() (count int, err error) {
	db, err := dao.DBService.GetDB()
	if err != nil {
		return 0, err
	}
	return db.Count(&dao.OpenChannelInfo{})
}
