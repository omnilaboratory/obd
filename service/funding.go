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
func (service *fundingTransactionManager) CreateFundingTx(jsonData string, user *bean.User) (fundingTransaction *dao.FundingTransaction, err error) {
	data := &bean.FundingCreated{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if len(data.TemporaryChannelId) == 0 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	fundingTransaction = &dao.FundingTransaction{}
	count, _ := db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).Count(fundingTransaction)
	if count == 0 {
		if tool.CheckIsString(&data.FundingTxid) == false {
			return nil, errors.New("wrong FundingTxid")
		}

		channelInfo := &dao.ChannelInfo{}
		err = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).First(channelInfo)
		if err != nil {
			return nil, err
		}
		fundingTransaction.TemporaryChannelId = data.TemporaryChannelId
		fundingTransaction.PropertyId = data.PropertyId
		fundingTransaction.PeerIdA = channelInfo.PeerIdA
		fundingTransaction.PeerIdB = channelInfo.PeerIdB

		fundingTransaction.ChannelPubKey = channelInfo.ChannelPubKey
		fundingTransaction.RedeemScript = channelInfo.RedeemScript

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
			fundingTransaction.FunderPubKey = channelInfo.PubKeyA
			fundingTransaction.FundeePubKey = channelInfo.PubKeyB
		} else {
			fundingTransaction.FunderPubKey = channelInfo.PubKeyB
			fundingTransaction.FundeePubKey = channelInfo.PubKeyA
		}
		fundingTransaction.AmountA = data.AmountA
		fundingTransaction.FundingTxid = data.FundingTxid
		fundingTransaction.FundingOutputIndex = data.FundingOutputIndex
		fundingTransaction.FunderPubKey2ForCommitment = data.FunderPubKey2

		hash, _ := chainhash.NewHashFromStr(fundingTransaction.FundingTxid)
		op := &bean.OutPoint{
			Hash:  *hash,
			Index: fundingTransaction.FundingOutputIndex,
		}
		fundingTransaction.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
		channelInfo.ChannelId = fundingTransaction.ChannelId

		tx, _ := db.Begin(true)
		defer tx.Rollback()

		err = tx.Update(channelInfo)
		fundingTransaction.CurrState = dao.FundingTransactionState_Create
		fundingTransaction.CreateBy = user.PeerId
		fundingTransaction.CreateAt = time.Now()
		err = tx.Save(fundingTransaction)
		tx.Commit()
	} else {
		_ = db.Select(q.Eq("TemporaryChannelId", data.TemporaryChannelId)).First(fundingTransaction)
		err = nil
	}
	return fundingTransaction, err
}

func (service *fundingTransactionManager) FundingTransactionSign(jsonData string, user *bean.User) (signed *dao.FundingTransaction, err error) {
	data := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if len(data.ChannelId) == 0 {
		return nil, errors.New("wrong ChannelId")
	}

	var fundingTransaction = &dao.FundingTransaction{}
	err = db.One("ChannelId", data.ChannelId, fundingTransaction)
	if err != nil {
		return nil, err
	}
	if data.Attitude {
		if len(data.FundeeSignature) == 0 {
			return nil, errors.New("wrong FundeeSignature")
		}
		fundingTransaction.AmountB = data.AmountB
		// temp storage private key of bob use to create C1
		fundingTransaction.FundeeSignature = data.FundeeSignature
		fundingTransaction.CurrState = dao.FundingTransactionState_Accept
	} else {
		fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
	}
	fundingTransaction.FundeeSignAt = time.Now()

	tx, _ := db.Begin(true)
	defer tx.Rollback()

	if data.Attitude {
		channelInfo := &dao.ChannelInfo{}
		err = tx.Select(q.Eq("ChannelId", data.ChannelId)).First(channelInfo)
		if err != nil {
			return nil, err
		}

		fundingTransaction := &dao.FundingTransaction{}
		err = tx.Select(q.Eq("ChannelId", data.ChannelId)).First(fundingTransaction)
		if err != nil {
			return nil, err
		}

		var outputBean = commitmentOutputBean{}
		outputBean.TempAddress = fundingTransaction.FunderPubKey2ForCommitment
		outputBean.ToAddressB = channelInfo.PubKeyB
		outputBean.AmountM = fundingTransaction.AmountA
		outputBean.AmountB = fundingTransaction.AmountB
		// create C1a tx
		commitmentTxInfo, err := createCommitmentATx(0, channelInfo, fundingTransaction, outputBean, user)
		if err != nil {
			return nil, err
		}

		txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
			channelInfo.ChannelPubKey,
			[]string{
				data.FundeeSignature,
			},
			[]rpc.TransactionInputItem{
				{commitmentTxInfo.InputTxid, commitmentTxInfo.InputVout, commitmentTxInfo.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{commitmentTxInfo.MultiAddress, commitmentTxInfo.AmountM},
				{commitmentTxInfo.PubKeyB, commitmentTxInfo.AmountB},
			},
			0,
			0)
		if err != nil {
			return nil, err
		}
		commitmentTxInfo.Txid = txid
		commitmentTxInfo.TxHexFirstSign = hex
		commitmentTxInfo.FirstSignAt = time.Now()
		commitmentTxInfo.CurrState = dao.TxInfoState_OtherSign
		_ = tx.Save(commitmentTxInfo)

		// create RDa tx
		rdTransaction, err := createRDaTx(0, channelInfo, commitmentTxInfo, channelInfo.PubKeyA, user)
		if err != nil {
			return nil, err
		}

		txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionFromUnsendTx(
			commitmentTxInfo.MultiAddress,
			[]string{
				data.FundeeSignature,
			},
			[]rpc.TransactionInputItem{
				{rdTransaction.InputTxid, rdTransaction.InputVout, rdTransaction.InputAmount},
			},
			[]rpc.TransactionOutputItem{
				{rdTransaction.PubKeyA, rdTransaction.Amount},
			},
			0,
			rdTransaction.Sequnence)
		if err != nil {
			return nil, err
		}
		rdTransaction.Txid = txid
		rdTransaction.TxHexFirstSign = hex
		rdTransaction.FirstSignAt = time.Now()
		rdTransaction.CurrState = dao.TxInfoState_OtherSign
		_ = tx.Save(rdTransaction)
	}
	err = tx.Update(fundingTransaction)
	tx.Commit()
	return fundingTransaction, err
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
