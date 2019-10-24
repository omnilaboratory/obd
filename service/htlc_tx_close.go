package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"log"
	"sync"
	"time"
)

//close htlc or close channel
type htlcCloseTxManager struct {
	operationFlag sync.Mutex
}

// htlc 关闭htlc交易
var HtlcCloseTxService htlcCloseTxManager

// -48
func (service *htlcCloseTxManager) RequestCloseHtlc(msgData string, user bean.User) (outData interface{}, targetUser string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcRequestCloseCurrTx{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		err = errors.New("wrong channel Id")
		log.Println(err)
		return nil, "", err
	}

	commitmentTxInfo, err := getHtlcLatestCommitmentTx(reqData.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(commitmentTxInfo)

	channelInfo := dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", commitmentTxInfo.ChannelId), q.Eq("", dao.ChannelState_HtlcBegin)).First(&channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(channelInfo)

	ht1aOrHe1b := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", commitmentTxInfo.Id), q.Eq("Owner", user.PeerId)).First(&ht1aOrHe1b)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if user.PeerId == channelInfo.PeerIdA {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
		targetUser = channelInfo.PeerIdB
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = reqData.ChannelAddressPrivateKey
		targetUser = channelInfo.PeerIdA
	}
	tempAddrPrivateKeyMap[commitmentTxInfo.RSMCTempAddressPubKey] = reqData.LastRsmcTempAddressPrivateKey
	tempAddrPrivateKeyMap[commitmentTxInfo.HTLCTempAddressPubKey] = reqData.LastHtlcTempAddressPrivateKey
	tempAddrPrivateKeyMap[ht1aOrHe1b.RSMCTempAddressPubKey] = reqData.LastHtlcTempAddressForHt1aPrivateKey
	tempAddrPrivateKeyMap[reqData.CurrRsmcTempAddressPubKey] = reqData.CurrRsmcTempAddressPrivateKey

	info := dao.HtlcRequestCloseCurrTxInfo{}
	info.ChannelId = commitmentTxInfo.ChannelId
	info.CurrRsmcTempAddressPubKey = reqData.CurrRsmcTempAddressPubKey
	info.CreateAt = time.Now()
	info.CreateBy = user.PeerId
	infoBytes, _ := json.Marshal(info)
	requestHash := tool.SignMsgWithSha256(infoBytes)
	info.RequestHash = requestHash
	err = db.Save(&info)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	return info, targetUser, nil
}

// -49
func (service *htlcCloseTxManager) SignCloseHtlc(msgData string, user bean.User) (outData interface{}, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, errors.New("empty json data")
	}

	reqData := &bean.HtlcSignCloseCurrTx{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcRequestCloseCurrTxInfo := dao.HtlcRequestCloseCurrTxInfo{}
	err = db.Select(q.Eq("RequestHash", reqData.RequestCloseHtlcHash)).First(&htlcRequestCloseCurrTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	commitmentTxInfo, err := getHtlcLatestCommitmentTx(htlcRequestCloseCurrTxInfo.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(commitmentTxInfo)

	channelInfo := dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", commitmentTxInfo.ChannelId), q.Eq("", dao.ChannelState_HtlcBegin)).First(&channelInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CurrState", dao.FundingTransactionState_Accept)).OrderBy("CreateAt").Reverse().First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ht1aOrHe1b := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", commitmentTxInfo.Id), q.Eq("Owner", user.PeerId)).First(&ht1aOrHe1b)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	isAliceOperateClose := true
	if user.PeerId == channelInfo.PeerIdB {
		isAliceOperateClose = false
	}

	dbTx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer dbTx.Rollback()

	err = createAliceSideBRTxs(dbTx, channelInfo, isAliceOperateClose, *reqData, *fundingTransaction, user)
	if err != nil {
		return nil, err
	}
	err = createBobSideBRTxs(dbTx, channelInfo, isAliceOperateClose, *reqData, *fundingTransaction, user)
	if err != nil {
		return nil, err
	}

	err = dbTx.Commit()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BR2a,HBR1a,HTBr1a
func createAliceSideBRTxs(tx storm.Node, channelInfo dao.ChannelInfo, isAliceOperateClose bool, requestData bean.HtlcSignCloseCurrTx, fundingTransaction dao.FundingTransaction, user bean.User) error {
	owner := channelInfo.PeerIdA
	lastCommitmentTxInfo := &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_Htlc_GetR)).OrderBy("CreateAt").Reverse().First(lastCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(lastCommitmentTxInfo)

	lastRDTransaction := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
	if err != nil {
		log.Println(err)
		return err
	}

	htOrHeTx := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("Owner", user.PeerId)).First(&htOrHeTx)
	if err != nil {
		log.Println(err)
		return err
	}
	htrd := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", htOrHeTx.Id), q.Eq("Owner", owner), q.Eq("RDType", 1)).First(htrd)
	if err != nil {
		log.Println(err)
		return err
	}
	// region create BR1a begin

	//如果已经创建过了，return
	count, _ := tx.Select(q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id)).Count(&dao.BreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already exist BreachRemedyTransaction ")
		return err
	}

	br, err := createBRTx(channelInfo.PeerIdB, &channelInfo, lastCommitmentTxInfo, &user)
	if err != nil {
		log.Println(err)
		return err
	}

	//如果金额大于0
	if br.Amount > 0 {
		lastTempAddressPrivateKey := ""
		// 如果当前操作用户是PeerIdA方（概念中的Alice方），则取当前操作人传入的数据
		if isAliceOperateClose {
			lastTempAddressPrivateKey = requestData.LastRsmcTempAddressPrivateKey
		} else {
			// 如果当前操作用户是PeerIdB方，而我们现在正在处理Alice方，所以我们要取另一方的数据
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentTxInfo.RSMCTempAddressPubKey]
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTxInfo.RSMCTxHash, lastCommitmentTxInfo.RSMCMultiAddress, lastCommitmentTxInfo.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return err
		}

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
			lastCommitmentTxInfo.RSMCMultiAddress,
			[]string{
				lastTempAddressPrivateKey,
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			inputs,
			channelInfo.AddressB,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			br.Amount,
			0,
			0,
			&lastCommitmentTxInfo.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return err
		}
		br.Txid = txid
		br.TransactionSignHex = hex
		br.SignAt = time.Now()
		br.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(br)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	// endregion

	// region create HBR1a begin
	hbr, err := createHtlcBRTx(channelInfo.PeerIdB, &channelInfo, lastCommitmentTxInfo, &user)
	if err != nil {
		log.Println(err)
		return err
	}

	if hbr.Amount > 0 {
		lastTempAddressPrivateKey := ""
		// 如果当前操作用户是PeerIdA方（概念中的Alice方），则取当前操作人传入的数据
		if isAliceOperateClose {
			lastTempAddressPrivateKey = requestData.LastHtlcTempAddressPrivateKey
		} else {
			// 如果当前操作用户是PeerIdB方，而我们现在正在处理Alice方，所以我们要取另一方的数据
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentTxInfo.HTLCTempAddressPubKey]
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTxInfo.RSMCTxHash, lastCommitmentTxInfo.RSMCMultiAddress, lastCommitmentTxInfo.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return err
		}

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
			lastCommitmentTxInfo.HTLCMultiAddress,
			[]string{
				lastTempAddressPrivateKey,
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			inputs,
			channelInfo.AddressB,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			hbr.Amount,
			0,
			0,
			&lastCommitmentTxInfo.HTLCRedeemScript)
		if err != nil {
			log.Println(err)
			return err
		}
		hbr.Txid = txid
		hbr.TransactionSignHex = hex
		hbr.SignAt = time.Now()
		hbr.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(hbr)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	// endregion create HBR1a end

	// region create HTBR1a begin

	htbr, err := createHtlcTimeoutBRTx(channelInfo.PeerIdB, &channelInfo, htOrHeTx, &user)
	if err != nil {
		log.Println(err)
		return err
	}
	htbr.CommitmentTxId = lastCommitmentTxInfo.Id

	if htbr.Amount > 0 {
		lastTempAddressPrivateKey := ""
		// 如果当前操作用户是PeerIdA方（概念中的Alice方），则取当前操作人传入的数据
		if isAliceOperateClose {
			lastTempAddressPrivateKey = requestData.LastHtlcTempAddressForHt1aPrivateKey
		} else {
			// 如果当前操作用户是PeerIdB方，而我们现在正在处理Alice方，所以我们要取另一方的数据
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[htOrHeTx.RSMCTempAddressPubKey]
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(htOrHeTx.RSMCTxHash, htOrHeTx.RSMCMultiAddress, htOrHeTx.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return err
		}

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
			htOrHeTx.RSMCMultiAddress,
			[]string{
				lastTempAddressPrivateKey,
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			inputs,
			channelInfo.AddressB,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			htbr.Amount,
			0,
			0,
			&htOrHeTx.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return err
		}
		htbr.Txid = txid
		htbr.TxHash = hex
		htbr.SignAt = time.Now()
		htbr.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(htbr)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	// endregion create HTBR1a end

	lastCommitmentTxInfo.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(lastRDTransaction)
	lastRDTransaction.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(lastRDTransaction)
	htbr.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(htbr)

	return nil
}

func createBobSideBRTxs(tx storm.Node, channelInfo dao.ChannelInfo, isAliceOperateClose bool, requestData bean.HtlcSignCloseCurrTx, fundingTransaction dao.FundingTransaction, user bean.User) error {

	owner := channelInfo.PeerIdB
	lastCommitmentTxInfo := &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_Htlc_GetR)).OrderBy("CreateAt").Reverse().First(lastCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(lastCommitmentTxInfo)

	lastRDTransaction := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
	if err != nil {
		log.Println(err)
		return err
	}

	htOrHeTx := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("Owner", user.PeerId)).First(&htOrHeTx)
	if err != nil {
		log.Println(err)
		return err
	}
	htrd := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", htOrHeTx.Id), q.Eq("Owner", owner), q.Eq("RDType", 1)).First(htrd)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (service *htlcCloseTxManager) RequestCloseChannel(msgData string, user bean.User) error {

	return nil
}

func (service *htlcCloseTxManager) SignCloseChannel(msgData string, user bean.User) error {

	return nil
}
