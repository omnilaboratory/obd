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
	"log"
	"time"
)

type fundingTransactionManager struct{}

var FundingTransactionService fundingTransactionManager

//funder request to fund to the multiAddr (channel)
func (service *fundingTransactionManager) CreateFundingTxRequest(jsonData string, user *bean.User) (fundingTransaction *dao.FundingTransaction, err error) {
	reqData := &bean.FundingCreated{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		return nil, err
	}

	if len(reqData.TemporaryChannelId) == 0 {
		return nil, errors.New("wrong TemporaryChannelId")
	}

	fundingTransaction = &dao.FundingTransaction{}
	count, _ := db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId)).Count(fundingTransaction)
	if count == 0 {
		if tool.CheckIsString(&reqData.FundingTxid) == false {
			return nil, errors.New("wrong FundingTxid")
		}

		channelInfo := &dao.ChannelInfo{}
		err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId), q.Eq("CurrState", dao.ChannelState_Accept), q.Or(q.Eq("PeerIdA", user.PeerId), q.Eq("PeerIdB", user.PeerId))).OrderBy("CreateAt").Reverse().First(channelInfo)
		if err != nil {
			return nil, err
		}

		if bean.ChannelIdService.IsEmpty(channelInfo.ChannelId) == false {
			return nil, errors.New("Channel is used, can not funding again")
		}

		hash, _ := chainhash.NewHashFromStr(reqData.FundingTxid)
		op := &bean.OutPoint{
			Hash:  *hash,
			Index: reqData.FundingOutputIndex,
		}
		fundingTransaction.ChannelId = bean.ChannelIdService.NewChanIDFromOutPoint(op)
		channelInfo.ChannelId = fundingTransaction.ChannelId

		count, err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Or(q.Eq("CurrState", dao.ChannelState_Accept), q.Eq("CurrState", dao.ChannelState_Close))).Count(channelInfo)
		if err != nil || count != 0 {
			return nil, errors.New("FundingTxid and FundingOutputIndex have been used")
		}

		fundingTransaction.TemporaryChannelId = reqData.TemporaryChannelId
		fundingTransaction.PropertyId = reqData.PropertyId
		fundingTransaction.PeerIdA = channelInfo.PeerIdA
		fundingTransaction.PeerIdB = channelInfo.PeerIdB

		fundingTransaction.ChannelPubKey = channelInfo.ChannelPubKey
		fundingTransaction.RedeemScript = channelInfo.RedeemScript

		if user.PeerId == channelInfo.PeerIdA {
			if reqData.FunderPubKey != channelInfo.PubKeyA {
				return nil, errors.New("invalid FunderPubKey")
			}
		}
		if user.PeerId == channelInfo.PeerIdB {
			if reqData.FunderPubKey != channelInfo.PubKeyB {
				return nil, errors.New("invalid FunderPubKey")
			}
		}

		// if alice launch funding
		if user.PeerId == channelInfo.PeerIdA {
			fundingTransaction.FunderPubKey = channelInfo.PubKeyA
			fundingTransaction.FundeePubKey = channelInfo.PubKeyB
			fundingTransaction.AmountA = reqData.AmountA
		} else { // if bob launch funding
			fundingTransaction.FunderPubKey = channelInfo.PubKeyB
			fundingTransaction.FundeePubKey = channelInfo.PubKeyA
			fundingTransaction.AmountB = reqData.AmountA
		}
		fundingTransaction.FunderSignature = reqData.FundingTxHex
		fundingTransaction.FundingTxid = reqData.FundingTxid
		fundingTransaction.FundingOutputIndex = reqData.FundingOutputIndex
		fundingTransaction.FunderPubKey2ForCommitment = reqData.FunderPubKey2

		tx, _ := db.Begin(true)
		defer tx.Rollback()

		err = tx.Update(channelInfo)
		if err != nil {
			return nil, err
		}
		fundingTransaction.CurrState = dao.FundingTransactionState_Create
		fundingTransaction.CreateBy = user.PeerId
		fundingTransaction.CreateAt = time.Now()
		err = tx.Save(fundingTransaction)
		if err != nil {
			return nil, err
		}
		tx.Commit()
	} else {
		err = db.Select(q.Eq("TemporaryChannelId", reqData.TemporaryChannelId)).First(fundingTransaction)
	}
	return fundingTransaction, err
}

func (service *fundingTransactionManager) FundingTransactionSign(jsonData string, signer *bean.User) (signed *dao.FundingTransaction, err error) {
	data := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return nil, err
	}

	if bean.ChannelIdService.IsEmpty(data.ChannelId) {
		return nil, errors.New("wrong ChannelId")
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", data.ChannelId), q.Eq("CurrState", dao.ChannelState_Accept)).First(channelInfo)
	if err != nil {
		log.Println("channel not find")
		return nil, err
	}

	var creatorSide = 0
	var isAliceFunder = true
	// if alice launch the funding,signer is bob
	if signer.PeerId == channelInfo.PeerIdB {
		isAliceFunder = true
		creatorSide = 0
	} else {
		isAliceFunder = false
		creatorSide = 1
	}

	var fundingTransaction = &dao.FundingTransaction{}
	err = db.One("ChannelId", data.ChannelId, fundingTransaction)
	if err != nil {
		return nil, err
	}

	if data.Attitude {
		if tool.CheckIsString(&data.FundeeSignature) == false {
			return nil, errors.New("wrong FundeeSignature")
		}

		if isAliceFunder {
			fundingTransaction.AmountB = data.AmountB
		} else {
			fundingTransaction.AmountA = data.AmountB
		}
		fundingTransaction.CurrState = dao.FundingTransactionState_Accept
	} else {
		fundingTransaction.CurrState = dao.FundingTransactionState_Defuse
		channelInfo.CurrState = dao.ChannelState_FundingDefuse
	}
	fundingTransaction.FundeeSignAt = time.Now()

	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if data.Attitude {

		var outputBean = commitmentOutputBean{}
		outputBean.TempAddress = fundingTransaction.FunderPubKey2ForCommitment
		if isAliceFunder {
			outputBean.ToAddressB = channelInfo.PubKeyB
			outputBean.AmountM = fundingTransaction.AmountA
			outputBean.AmountB = fundingTransaction.AmountB
		} else {
			outputBean.ToAddressB = channelInfo.PubKeyA
			outputBean.AmountM = fundingTransaction.AmountB
			outputBean.AmountB = fundingTransaction.AmountA
		}

		// create C1a tx
		commitmentTxInfo, err := createCommitmentTx(creatorSide, channelInfo, fundingTransaction, outputBean, signer)
		if err != nil {
			return nil, err
		}

		txid, hex, err := rpcClient.BtcCreateAndSignRawTransactionForUnsendTx(
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
		err = tx.Save(commitmentTxInfo)
		if err != nil {
			return nil, err
		}

		// create RDa tx
		rdTransaction, err := createRDTx(creatorSide, channelInfo, commitmentTxInfo, channelInfo.PubKeyA, signer)
		if err != nil {
			return nil, err
		}

		txid, hex, err = rpcClient.BtcCreateAndSignRawTransactionForUnsendTx(
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
		err = tx.Save(rdTransaction)
		if err != nil {
			return nil, err
		}
	}
	err = tx.Update(fundingTransaction)
	if err != nil {
		return nil, err
	}

	if data.Attitude {
		// if agree,send the fundingtx to chain network
		_, err := rpcClient.SendRawTransaction(fundingTransaction.FunderSignature)
		if err != nil {
			return nil, err
		}
	}

	if data.Attitude == false {
		err = tx.Update(channelInfo)
		if err != nil {
			return nil, err
		}
	}
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
