package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/config"
	"LightningOnOmni/dao"
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
func (c *channelManager) OpenChannel(msg bean.RequestMessage, peerIdA string) (data *bean.OpenChannelInfo, err error) {
	data = &bean.OpenChannelInfo{}
	json.Unmarshal([]byte(msg.Data), &data)
	if len(data.FundingPubKey) != 34 {
		return nil, errors.New("wrong fundingPubKey")
	}

	isMine, err := rpcClient.Validateaddress(data.FundingPubKey)
	if err != nil {
		return nil, err
	}
	if isMine == false {
		return nil, errors.New("invalid fundingPubKey")
	}

	data.ChainHash = config.Init_node_chain_hash
	tempId := bean.ChannelIdService.NextTemporaryChanID()
	data.TemporaryChannelId = tempId

	node := &dao.ChannelInfo{}
	node.OpenChannelInfo = *data
	node.PeerIdA = peerIdA
	node.PeerIdB = msg.RecipientPeerId
	node.PubKeyA = data.FundingPubKey
	node.CurrState = dao.ChannelState_Create
	node.CreateAt = time.Now()

	db.Save(node)
	return data, err
}

func (c *channelManager) AcceptChannel(jsonData string, peerIdB string) (node *dao.ChannelInfo, err error) {
	data := &bean.AcceptChannelInfo{}
	err = json.Unmarshal([]byte(jsonData), &data)

	if err != nil {
		return nil, err
	}

	if len(data.TemporaryChannelId) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	if data.Attitude {
		if len(data.FundingPubKey) != 34 {
			return nil, errors.New("wrong PubKeyB")
		}
		isMine, err := rpcClient.Validateaddress(data.FundingPubKey)
		if err != nil {
			return nil, err
		}
		if isMine == false {
			return nil, errors.New("invalid fundingPubKey")
		}
	}

	node = &dao.ChannelInfo{}
	db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(node)

	if &node.Id == nil {
		return nil, errors.New("invalid TemporaryChannelId")
	}
	if node.CurrState != dao.ChannelState_Create {
		return nil, errors.New("invalid ChannelState " + strconv.Itoa(int(node.CurrState)))
	}

	if data.Attitude {
		node.PubKeyB = data.FundingPubKey
		multiSig, err := rpcClient.CreateMultiSig(2, []string{node.PubKeyA, node.PubKeyB})
		if err != nil {
			return nil, err
		}
		node.ChannelPubKey = gjson.Get(multiSig, "address").String()
		node.RedeemScript = gjson.Get(multiSig, "redeemScript").String()
	}
	node.AcceptAt = time.Now()
	if data.Attitude {
		node.CurrState = dao.ChannelState_Accept
	} else {
		node.CurrState = dao.ChannelState_Defuse
	}
	err = db.Update(node)

	return node, err
}

// GetChannelByTemporaryChanId
func (c *channelManager) GetChannelByTemporaryChanId(jsonData string) (node *dao.ChannelInfo, err error) {
	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	var tempChanId chainhash.Hash
	for index, value := range array {
		tempChanId[index] = byte(value.Num)
	}
	node = &dao.ChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", tempChanId)).First(node)
	return node, err
}

// DelChannelByTemporaryChanId
func (c *channelManager) DelChannelByTemporaryChanId(jsonData string) (node *dao.ChannelInfo, err error) {

	array := gjson.Parse(jsonData).Array()
	if len(array) != 32 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	var tempChanId chainhash.Hash
	for index, value := range array {
		tempChanId[index] = byte(value.Num)
	}
	node = &dao.ChannelInfo{}
	err = db.Select(q.Eq("TemporaryChannelId", tempChanId)).First(node)
	if err == nil {
		err = db.DeleteStruct(node)
	}
	return node, err
}

// AllItem
func (c *channelManager) AllItem(peerId string) (data []dao.ChannelInfo, err error) {
	var infos []dao.ChannelInfo
	err = db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).OrderBy("CreateAt").Reverse().Find(&infos)
	return infos, err
}

// TotalCount
func (c *channelManager) TotalCount(peerId string) (count int, err error) {
	return db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).Count(&dao.ChannelInfo{})
}
