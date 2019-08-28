package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"time"
)

type fundingTransactionManager struct{}

var FundingTransactionService fundingTransactionManager

//funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) CreateFundingTx(jsonData string, user *bean.User) (node *dao.FundingTransaction, err error) {
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
		if tool.CheckIsString(&data.FundingTxid) == false {
			return nil, errors.New("wrong FundingTxid")
		}

		channelInfo := &dao.ChannelInfo{}
		err = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
		if err != nil {
			return nil, err
		}
		node.TemporaryChannelId = data.TemporaryChannelId
		node.PropertyId = data.PropertyId
		node.PeerIdA = channelInfo.PeerIdA
		node.PeerIdB = channelInfo.PeerIdB

		node.ChannelPubKey = channelInfo.ChannelPubKey
		node.RedeemScript = channelInfo.RedeemScript

		if user.PeerId == channelInfo.PeerIdA {
			if data.FunderPubKey != channelInfo.PubKeyA {
				return nil, errors.New("invalid FunderPubKey")
			}
		}
		if user.PeerId == channelInfo.PeerIdB {
			if data.FunderPubKey != channelInfo.PubKeyB {
				return nil, errors.New("invalid FunderPubKey")
			}
		}

		if data.FunderPubKey == channelInfo.PubKeyA {
			node.FunderPubKey = channelInfo.PubKeyA
			node.FundeePubKey = channelInfo.PubKeyB
		} else {
			node.FunderPubKey = channelInfo.PubKeyB
			node.FundeePubKey = channelInfo.PubKeyA
		}
		node.AmountA = data.AmountA
		node.FundingTxid = data.FundingTxid
		node.FundingOutputIndex = data.FundingOutputIndex

		hash, _ := chainhash.NewHashFromStr(node.FundingTxid)
		op := &bean.OutPoint{
			Hash:  *hash,
			Index: node.FundingOutputIndex,
		}
		node.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
		channelInfo.ChannelId = node.ChannelId

		commitmentTxInfo, err := createCommitmentTx(channelInfo, data.PropertyId, data.AmountA, data.FunderPubKey2, user.PeerId)
		if err != nil {
			return nil, err
		}

		tx, _ := db.Begin(true)
		defer tx.Rollback()

		err = tx.Update(channelInfo)
		node.CurrState = dao.FundingTransactionState_Create
		node.CreateAt = time.Now()
		err = tx.Save(node)
		err = tx.Save(commitmentTxInfo)
		tx.Commit()
	} else {
		_ = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(node)
		err = nil
	}
	return node, err
}

func createCommitmentTx(channelInfo *dao.ChannelInfo, propertyId int64, amountA float64, alice2 string, funderPeerId string) (*dao.CommitmentTxInfo, error) {
	node := &dao.CommitmentTxInfo{}
	node.PeerIdA = channelInfo.PeerIdA
	node.PeerIdB = channelInfo.PeerIdB
	node.ChannelId = channelInfo.ChannelId
	node.PropertyId = propertyId

	node.AmountM = amountA
	node.PubKeyA = alice2
	var addressB = channelInfo.PubKeyA
	if channelInfo.PeerIdA == funderPeerId {
		addressB = channelInfo.PubKeyB
	}
	node.PubKeyB = addressB

	multiAddr, err := rpcClient.CreateMultiSig(2, []string{alice2, addressB})
	if err != nil {
		return nil, err
	}
	node.MultiAddress = gjson.Get(multiAddr, "address").String()
	node.RedeemScript = gjson.Get(multiAddr, "redeemScript").String()

	node.CurrState = dao.CommitmentTxInfoState_Create
	node.CreateAt = time.Now()
	node.LastEditTime = time.Now()

	return node, nil
}

func (service *fundingTransactionManager) FundingTransactionSign(jsonData string) (signed *dao.FundingTransaction, err error) {
	data := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if len(data.ChannelId) == 0 {
		return nil, errors.New("wrong ChannelId")
	}

	var node = &dao.FundingTransaction{}
	err = db.One("ChannelId", data.ChannelId, node)
	if err != nil {
		return nil, err
	}
	if data.Attitude {
		if len(data.FundeeSignature) == 0 {
			return nil, errors.New("wrong FundeeSignature")
		}
		node.AmountB = data.AmountB
		// temp storage private key of bob use to create C1
		node.FundeeSignature = data.FundeeSignature
		node.CurrState = dao.FundingTransactionState_Accept
	} else {
		node.CurrState = dao.FundingTransactionState_Defuse
	}
	node.FundeeSignAt = time.Now()

	tx, _ := db.Begin(true)
	defer tx.Rollback()

	if data.Attitude {
		channelInfo := &dao.ChannelInfo{}
		err = tx.Select(q.Eq("ChannelId", data.ChannelId)).First(channelInfo)
		if err != nil {
			return nil, err
		}

		commitmentTxInfo := &dao.CommitmentTxInfo{}
		err := tx.Select(q.Eq("ChannelId", data.ChannelId)).OrderBy("CreateAt").Reverse().First(commitmentTxInfo)
		if err != nil {
			return nil, err
		}
		// create C1a tx
		txid, hex, err := rpcClient.BtcCreateAndSignRawTransaction(
			channelInfo.ChannelPubKey,
			[]string{
				data.FundeeSignature,
			},
			[]rpc.TransactionOutputItem{
				{commitmentTxInfo.MultiAddress, commitmentTxInfo.AmountM},
			},
			0,
			nil)
		commitmentTxInfo.Txid = txid
		commitmentTxInfo.TxHexFirstSign = hex
		commitmentTxInfo.FirstSignAt = time.Now()
		_ = tx.Update(commitmentTxInfo)

	}
	err = tx.Update(node)
	tx.Commit()
	return node, err
}

func (service *fundingTransactionManager) ItemByTempId(jsonData string) (node *dao.FundingTransaction, err error) {
	var tempChanId chainhash.Hash
	for index, item := range gjson.Parse(jsonData).Array() {
		tempChanId[index] = byte(item.Int())
	}
	return service.ItemByTempIdArray(tempChanId)
}

func (service *fundingTransactionManager) ItemByTempIdArray(tempId chainhash.Hash) (node *dao.FundingTransaction, err error) {
	var data = &dao.FundingTransaction{}
	err = db.One("TemporaryChannelId", tempId, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *fundingTransactionManager) AllItem(peerId string) (node []dao.FundingTransaction, err error) {
	var data []dao.FundingTransaction
	err = db.Select(q.Or(q.Eq("PeerIdB", peerId), q.Eq("PeerIdA", peerId))).OrderBy("CreateAt").Reverse().Find(&data)
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
}

func (service *fundingTransactionManager) Del(id int) (err error) {
	var data = &dao.FundingTransaction{}
	count, err := db.Select(q.Eq("Id", id)).Count(data)
	if err == nil && count == 1 {
		err = db.DeleteStruct(data)
	}
	return err
}
func (service *fundingTransactionManager) TotalCount(peerId string) (count int, err error) {
	return db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId))).Count(&dao.FundingTransaction{})
}
