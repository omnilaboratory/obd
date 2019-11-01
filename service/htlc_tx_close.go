package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/shopspring/decimal"
	"log"
	"strings"
	"sync"
	"time"
)

//close htlc or close channel
type htlcCloseTxManager struct {
	operationFlag sync.Mutex
}

// htlc 关闭当前htlc交易
var HtlcCloseTxService htlcCloseTxManager

// -48 request close htlc
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

// -49 close htlc
func (service *htlcCloseTxManager) SignCloseHtlc(msgData string, user bean.User) (outData map[string]interface{}, targetUser string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcSignCloseCurrTx{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// region check data
	if tool.CheckIsString(&reqData.RequestCloseHtlcHash) {
		err = errors.New("empty RequestCloseHtlcHash")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPubKey) {
		err = errors.New("empty CurrRsmcTempAddressPubKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPrivateKey) {
		err = errors.New("empty CurrRsmcTempAddressPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.LastHtlcTempAddressForHt1aPrivateKey) {
		err = errors.New("empty LastHtlcTempAddressForHt1aPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.LastHtlcTempAddressPrivateKey) {
		err = errors.New("empty LastHtlcTempAddressPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.LastRsmcTempAddressPrivateKey) {
		err = errors.New("empty LastRsmcTempAddressPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) {
		err = errors.New("empty ChannelAddressPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	// endregion

	// region query data
	dataFromCloseOpStarter := dao.HtlcRequestCloseCurrTxInfo{}
	err = db.Select(q.Eq("RequestHash", reqData.RequestCloseHtlcHash)).First(&dataFromCloseOpStarter)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	commitmentTxInfo, err := getHtlcLatestCommitmentTx(dataFromCloseOpStarter.ChannelId, user.PeerId)
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

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CurrState", dao.FundingTransactionState_Accept)).OrderBy("CreateAt").Reverse().First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	ht1aOrHe1b := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", commitmentTxInfo.Id), q.Eq("Owner", user.PeerId)).First(&ht1aOrHe1b)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	isAliceExecutionCloseOp := true
	targetUser = channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
		isAliceExecutionCloseOp = false
	}
	// endregion

	if channelInfo.PeerIdA == user.PeerId {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = reqData.ChannelAddressPrivateKey
	}

	dbTx, err := db.Begin(true)
	if err != nil {
		return nil, "", err
	}
	defer dbTx.Rollback()

	//  region create BR  BR2a HBR1a HTBR1a ; BR2b HBR1b HEBR1b
	lastCommitmentTxInfoA, err := createAliceSideBRTxs(dbTx, channelInfo, isAliceExecutionCloseOp, *reqData, *fundingTransaction, user)
	if err != nil {
		return nil, "", err
	}

	lastCommitmentTxInfoB, err := createBobSideBRTxs(dbTx, channelInfo, isAliceExecutionCloseOp, *reqData, *fundingTransaction, user)
	if err != nil {
		return nil, "", err
	}
	//  endregion

	// region create C3a and C3b
	newCommitmentTxInfoA, err := createAliceRsmcTxsForCloseHtlc(dbTx, channelInfo, isAliceExecutionCloseOp, *reqData, dataFromCloseOpStarter, lastCommitmentTxInfoA, *fundingTransaction, user)
	if err != nil {
		return nil, "", err
	}
	newCommitmentTxInfoB, err := createBobRsmcTxsForCloseHtlc(dbTx, channelInfo, isAliceExecutionCloseOp, *reqData, dataFromCloseOpStarter, lastCommitmentTxInfoB, *fundingTransaction, user)
	if err != nil {
		return nil, "", err
	}
	// endregion

	channelInfo.CurrState = dao.ChannelState_Accept
	_ = dbTx.Update(channelInfo)

	err = dbTx.Commit()
	if err != nil {
		return nil, "", err
	}

	// region delete cache
	delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	// endregion

	outData = make(map[string]interface{})
	outData["msg"] = "close htlc success"
	outData["newCommitmentTxInfoA"] = newCommitmentTxInfoA
	outData["newCommitmentTxInfoB"] = newCommitmentTxInfoB
	return outData, targetUser, nil
}

// C3a RD3a
func createAliceRsmcTxsForCloseHtlc(tx storm.Node, channelInfo dao.ChannelInfo, isAliceExecutionCloseOp bool, reqData bean.HtlcSignCloseCurrTx, dataFromCloseOpStarter dao.HtlcRequestCloseCurrTxInfo, lastCommitmentATx *dao.CommitmentTransaction, fundingTransaction dao.FundingTransaction, operator bean.User) (newCommitmentTxInfo *dao.CommitmentTransaction, err error) {

	owner := channelInfo.PeerIdA
	// 这里需要确认一个事情：在这个通道里面，这次的htlc到底是谁是转出方
	// rmsc的资产分配方案
	var outputBean = commitmentOutputBean{}
	if lastCommitmentATx.HtlcSender == channelInfo.PeerIdA {
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToRSMC).Sub(decimal.NewFromFloat(lastCommitmentATx.AmountToHtlc)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToOther).Add(decimal.NewFromFloat(lastCommitmentATx.AmountToHtlc)).Float64()
	} else {
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentATx.AmountToHtlc)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToOther).Sub(decimal.NewFromFloat(lastCommitmentATx.AmountToHtlc)).Float64()
	}
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
	outputBean.OppositeSideChannelAddress = channelInfo.AddressB

	if isAliceExecutionCloseOp {
		outputBean.RsmcTempPubKey = reqData.CurrRsmcTempAddressPubKey
	} else {
		outputBean.RsmcTempPubKey = dataFromCloseOpStarter.CurrRsmcTempAddressPubKey
	}

	commitmentTxInfo, err := createCommitmentTx(owner, &channelInfo, &fundingTransaction, outputBean, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Rsmc

	allUsedTxidTemp := ""
	// rsmc
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			commitmentTxInfo.RSMCMultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += usedTxid
		commitmentTxInfo.RSMCTxid = txid
		commitmentTxInfo.RSMCTxHash = hex
	}

	//create to Bob tx
	if commitmentTxInfo.AmountToOther > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTxToBob(
			channelInfo.ChannelAddress,
			allUsedTxidTemp,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			channelInfo.AddressB,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToOther,
			0,
			0, &channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		commitmentTxInfo.ToOtherTxid = txid
		commitmentTxInfo.ToOtherTxHash = hex
	}

	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	commitmentTxInfo.LastCommitmentTxId = lastCommitmentATx.Id
	commitmentTxInfo.LastHash = lastCommitmentATx.CurrHash
	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, nil
}

// C3b RD3b
func createBobRsmcTxsForCloseHtlc(tx storm.Node, channelInfo dao.ChannelInfo, isAliceExecutionCloseOp bool, reqData bean.HtlcSignCloseCurrTx, dataFromCloseOpStarter dao.HtlcRequestCloseCurrTxInfo, lastCommitmentBTx *dao.CommitmentTransaction, fundingTransaction dao.FundingTransaction, operator bean.User) (newCommitmentTxInfo *dao.CommitmentTransaction, err error) {

	owner := channelInfo.PeerIdB
	// 这里需要确认一个事情：在这个通道里面，这次的htlc到底是谁是转出方
	// rmsc的资产分配方案
	var outputBean = commitmentOutputBean{}
	if lastCommitmentBTx.HtlcSender == channelInfo.PeerIdA {
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentBTx.AmountToHtlc)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountToOther).Sub(decimal.NewFromFloat(lastCommitmentBTx.AmountToHtlc)).Float64()
	} else {
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountToRSMC).Sub(decimal.NewFromFloat(lastCommitmentBTx.AmountToHtlc)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountToOther).Add(decimal.NewFromFloat(lastCommitmentBTx.AmountToHtlc)).Float64()
	}
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
	outputBean.OppositeSideChannelAddress = channelInfo.AddressA

	if isAliceExecutionCloseOp {
		outputBean.RsmcTempPubKey = dataFromCloseOpStarter.CurrRsmcTempAddressPubKey
	} else {
		outputBean.RsmcTempPubKey = reqData.CurrRsmcTempAddressPubKey
	}

	commitmentTxInfo, err := createCommitmentTx(owner, &channelInfo, &fundingTransaction, outputBean, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Rsmc

	allUsedTxidTemp := ""
	// rsmc
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			commitmentTxInfo.RSMCMultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += usedTxid
		commitmentTxInfo.RSMCTxid = txid
		commitmentTxInfo.RSMCTxHash = hex
	}

	//create to alice tx
	if commitmentTxInfo.AmountToOther > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTxToBob(
			channelInfo.ChannelAddress,
			allUsedTxidTemp,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			channelInfo.AddressA,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToOther,
			0,
			0, &channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		commitmentTxInfo.ToOtherTxid = txid
		commitmentTxInfo.ToOtherTxHash = hex
	}

	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	commitmentTxInfo.LastCommitmentTxId = lastCommitmentBTx.Id
	commitmentTxInfo.LastHash = lastCommitmentBTx.CurrHash
	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, nil
}

// BR2a,HBR1a,HTBR1a
func createAliceSideBRTxs(tx storm.Node, channelInfo dao.ChannelInfo, isAliceExecutionCloseOp bool, reqData bean.HtlcSignCloseCurrTx, fundingTransaction dao.FundingTransaction, user bean.User) (lastCommitmentTxInfo *dao.CommitmentTransaction, err error) {
	owner := channelInfo.PeerIdA
	brOwner := channelInfo.PeerIdB
	lastCommitmentTxInfo = &dao.CommitmentTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_Htlc_GetR)).OrderBy("CreateAt").Reverse().First(lastCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	lastRDTransaction := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htOrHeTx := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("Owner", user.PeerId)).First(&htOrHeTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	htrd := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", htOrHeTx.Id), q.Eq("Owner", owner), q.Eq("RDType", 1)).First(htrd)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// region create BR1a
	count, _ := tx.Select(q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("Owner", owner)).Count(&dao.BreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already exist BreachRemedyTransaction ")
		return nil, err
	}

	br, err := createBRTx(brOwner, &channelInfo, lastCommitmentTxInfo, &user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if br.Amount > 0 {
		lastTempAddressPrivateKey := ""
		if isAliceExecutionCloseOp {
			lastTempAddressPrivateKey = reqData.LastRsmcTempAddressPrivateKey
		} else {
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentTxInfo.RSMCTempAddressPubKey]
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return nil, err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTxInfo.RSMCTxHash, lastCommitmentTxInfo.RSMCMultiAddress, lastCommitmentTxInfo.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
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
			return nil, err
		}
		br.Txid = txid
		br.TransactionSignHex = hex
		br.SignAt = time.Now()
		br.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(br)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	// endregion

	// region create HBR1a
	count, _ = tx.Select(q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("Owner", owner)).Count(&dao.HTLCBreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already create HTLCBreachRemedyTransaction")
		return nil, err
	}
	hbr, err := createHtlcBRTx(brOwner, &channelInfo, lastCommitmentTxInfo, &user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if hbr.Amount > 0 {
		lastTempAddressPrivateKey := ""
		// 如果当前操作用户是PeerIdA方（概念中的Alice方），则取当前操作人传入的数据
		if isAliceExecutionCloseOp {
			lastTempAddressPrivateKey = reqData.LastHtlcTempAddressPrivateKey
		} else {
			// 如果当前操作用户是PeerIdB方，而我们现在正在处理Alice方，所以我们要取另一方的数据
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentTxInfo.HTLCTempAddressPubKey]
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return nil, err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTxInfo.RSMCTxHash, lastCommitmentTxInfo.RSMCMultiAddress, lastCommitmentTxInfo.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
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
			return nil, err
		}
		hbr.Txid = txid
		hbr.TransactionSignHex = hex
		hbr.SignAt = time.Now()
		hbr.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(hbr)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	// endregion

	// region create HTBR1a
	count, _ = tx.Select(q.Eq("CommitmentTxId", htOrHeTx.Id), q.Eq("Owner", owner)).Count(&dao.HTLCTimeoutBreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already create HTLCBreachRemedyTransaction")
		return nil, err
	}

	htbr, err := createHtlcTimeoutBRTx(brOwner, &channelInfo, htOrHeTx, &user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	htbr.CommitmentTxId = lastCommitmentTxInfo.Id

	if htbr.Amount > 0 {
		lastTempAddressPrivateKey := ""
		// 如果当前操作用户是PeerIdA方（概念中的Alice方），则取当前操作人传入的数据
		if isAliceExecutionCloseOp {
			lastTempAddressPrivateKey = reqData.LastHtlcTempAddressForHt1aPrivateKey
		} else {
			// 如果当前操作用户是PeerIdB方，而我们现在正在处理Alice方，所以我们要取另一方的数据
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[htOrHeTx.RSMCTempAddressPubKey]
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return nil, err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(htOrHeTx.RSMCTxHash, htOrHeTx.RSMCMultiAddress, htOrHeTx.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
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
			return nil, err
		}
		htbr.Txid = txid
		htbr.TxHash = hex
		htbr.SignAt = time.Now()
		htbr.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(htbr)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	// endregion

	lastCommitmentTxInfo.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(lastRDTransaction)
	lastRDTransaction.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(lastRDTransaction)
	htbr.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(htbr)

	return lastCommitmentTxInfo, nil
}

// BR2b,HBR1b,HEBR1b
func createBobSideBRTxs(tx storm.Node, channelInfo dao.ChannelInfo, isAliceExecutionCloseOp bool, reqData bean.HtlcSignCloseCurrTx, fundingTransaction dao.FundingTransaction, user bean.User) (lastCommitmentTxInfo *dao.CommitmentTransaction, err error) {

	owner := channelInfo.PeerIdB
	brOwner := channelInfo.PeerIdA
	lastCommitmentTxInfo = &dao.CommitmentTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_Htlc_GetR)).OrderBy("CreateAt").Reverse().First(lastCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(lastCommitmentTxInfo)

	lastRDTransaction := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htOrHeTx := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("Owner", user.PeerId)).First(&htOrHeTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	htrd := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", htOrHeTx.Id), q.Eq("Owner", owner), q.Eq("RDType", 1)).First(htrd)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// region create BR1b

	//如果已经创建过了，return
	count, _ := tx.Select(q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("Owner", owner)).Count(&dao.BreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already create BreachRemedyTransaction ")
		return nil, err
	}

	br, err := createBRTx(brOwner, &channelInfo, lastCommitmentTxInfo, &user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//如果金额大于0
	if br.Amount > 0 {
		lastTempAddressPrivateKey := ""
		if isAliceExecutionCloseOp {
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentTxInfo.RSMCTempAddressPubKey]
		} else {
			lastTempAddressPrivateKey = reqData.LastRsmcTempAddressPrivateKey
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return nil, err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTxInfo.RSMCTxHash, lastCommitmentTxInfo.RSMCMultiAddress, lastCommitmentTxInfo.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
			lastCommitmentTxInfo.RSMCMultiAddress,
			[]string{
				lastTempAddressPrivateKey,
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
			},
			inputs,
			channelInfo.AddressA,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			br.Amount,
			0,
			0,
			&lastCommitmentTxInfo.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		br.Txid = txid
		br.TransactionSignHex = hex
		br.SignAt = time.Now()
		br.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(br)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	// endregion

	// region create HBR1b
	count, _ = tx.Select(q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("Owner", owner)).Count(&dao.HTLCBreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already create HTLCBreachRemedyTransaction")
		return nil, err
	}

	hbr, err := createHtlcBRTx(brOwner, &channelInfo, lastCommitmentTxInfo, &user)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if hbr.Amount > 0 {
		lastTempAddressPrivateKey := ""
		if isAliceExecutionCloseOp {
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentTxInfo.HTLCTempAddressPubKey]
		} else {
			lastTempAddressPrivateKey = reqData.LastHtlcTempAddressPrivateKey
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return nil, err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTxInfo.RSMCTxHash, lastCommitmentTxInfo.RSMCMultiAddress, lastCommitmentTxInfo.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
			lastCommitmentTxInfo.HTLCMultiAddress,
			[]string{
				lastTempAddressPrivateKey,
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
			},
			inputs,
			channelInfo.AddressA,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			hbr.Amount,
			0,
			0,
			&lastCommitmentTxInfo.HTLCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		hbr.Txid = txid
		hbr.TransactionSignHex = hex
		hbr.SignAt = time.Now()
		hbr.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(hbr)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	// endregion

	// region create HEBR1b
	count, _ = tx.Select(q.Eq("CommitmentTxId", htOrHeTx.Id), q.Eq("Owner", owner)).Count(&dao.HTLCTimeoutBreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already create HTLCBreachRemedyTransaction")
		return nil, err
	}

	htbr, err := createHtlcTimeoutBRTx(brOwner, &channelInfo, htOrHeTx, &user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	htbr.CommitmentTxId = lastCommitmentTxInfo.Id

	if htbr.Amount > 0 {
		lastTempAddressPrivateKey := ""
		if isAliceExecutionCloseOp {
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[htOrHeTx.RSMCTempAddressPubKey]
		} else {
			lastTempAddressPrivateKey = reqData.LastHtlcTempAddressForHt1aPrivateKey
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return nil, err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(htOrHeTx.RSMCTxHash, htOrHeTx.RSMCMultiAddress, htOrHeTx.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
			htOrHeTx.RSMCMultiAddress,
			[]string{
				lastTempAddressPrivateKey,
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
			},
			inputs,
			channelInfo.AddressA,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			htbr.Amount,
			0,
			0,
			&htOrHeTx.RSMCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		htbr.Txid = txid
		htbr.TxHash = hex
		htbr.SignAt = time.Now()
		htbr.CurrState = dao.TxInfoState_CreateAndSign
		err = tx.Save(htbr)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	// endregion

	lastCommitmentTxInfo.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(lastRDTransaction)
	lastRDTransaction.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(lastRDTransaction)
	htbr.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(htbr)

	return lastCommitmentTxInfo, nil
}

// -50 请求关闭通道
func (service *htlcCloseTxManager) RequestCloseChannel(msgData string, user bean.User) (outData map[string]interface{}, targetUser string, err error) {

	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty inputData")
	}
	reqData := &bean.HtlcCloseChannelReq{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_HtlcBegin)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	targetUser = channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	}

	lastCommitmentTx := &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("TxType", dao.CommitmentTransactionType_Htlc), q.Eq("Owner", user.PeerId)).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	closeChannel := &dao.CloseChannel{}
	closeChannel.ChannelId = reqData.ChannelId
	closeChannel.Owner = user.PeerId
	closeChannel.CurrState = 0
	closeChannel.CreateAt = time.Now()
	dataBytes, _ := json.Marshal(closeChannel)
	closeChannel.RequestHex = tool.SignMsgWithSha256(dataBytes)
	err = db.Save(closeChannel)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	outData = make(map[string]interface{})
	outData["channel_id"] = reqData.ChannelId
	outData["request_close_channel_hash"] = closeChannel.RequestHex
	return outData, targetUser, nil
}

// -51 对通道关闭进行表态 谁请求，就以谁的身份关闭通道
func (service *htlcCloseTxManager) SignCloseChannel(msgData string, user bean.User) (outData interface{}, closeOpStarter string, err error) {
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty inputData")
	}
	reqData := &bean.HtlcCloseChannelSign{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// region check input data
	if bean.ChannelIdService.IsEmpty(reqData.ChannelId) {
		err = errors.New("empty channel_id")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.RequestCloseChannelHash) == false {
		err = errors.New("empty request_close_channel_hash")
		log.Println(err)
		return nil, "", err
	}
	// endregion

	//region query data from db
	closeOpStaterReqData := &dao.CloseChannel{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", 0), q.Eq("RequestHex", reqData.RequestCloseChannelHash)).First(closeOpStaterReqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(q.Eq("ChannelId", reqData.ChannelId), q.Eq("CurrState", dao.ChannelState_HtlcBegin)).First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	// endregion

	//region Cnx RDnx

	// 提现操作的发起者
	closeOpStarter = channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		closeOpStarter = channelInfo.PeerIdA
	}

	lastCommitmentTx := &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("TxType", dao.CommitmentTransactionType_Htlc), q.Eq("Owner", closeOpStarter)).OrderBy("CreateAt").Reverse().First(lastCommitmentTx)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&lastCommitmentTx.RSMCTxHash) {
		commitmentTxid, err := rpcClient.SendRawTransaction(lastCommitmentTx.RSMCTxHash)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxid)
	}

	if tool.CheckIsString(&lastCommitmentTx.ToOtherTxHash) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(lastCommitmentTx.ToOtherTxHash)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxidToBob)
	}

	lastRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
	err = db.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", closeOpStarter)).OrderBy("CreateAt").Reverse().First(lastRevocableDeliveryTx)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	revocableDeliveryTxid, err := rpcClient.SendRawTransaction(lastRevocableDeliveryTx.TxHash)
	if err != nil {
		log.Println(err)
		msg := err.Error()
		if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
			return nil, "", err
		}
	}
	log.Println(revocableDeliveryTxid)

	if tool.CheckIsString(&lastCommitmentTx.HtlcTxHash) {
		commitmentTxidToHtlc, err := rpcClient.SendRawTransaction(lastCommitmentTx.HtlcTxHash)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxidToHtlc)
	}

	// endregion

	// 提现人是否是这次htlc的转账发起者
	var starterIsHtlcSender bool
	if lastCommitmentTx.HtlcSender == closeOpStarter {
		starterIsHtlcSender = true
	} else {
		starterIsHtlcSender = false
	}

	tx, err := db.Begin(true)
	defer tx.Rollback()

	// region htlc的相关交易广播
	isRsmcTx := true
	// 提现人是这次htlc的转账发起者
	if starterIsHtlcSender {
		// 如果已经得到R，直接广播HED1a
		if lastCommitmentTx.CurrState == dao.TxInfoState_Htlc_GetR {
			isRsmcTx = false
			hednx := &dao.HTLCExecutionDeliveryA{}
			err = tx.Select(q.Eq("CommitmentTxId", lastCommitmentTx.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign), q.Eq("Owner", closeOpStarter)).First(hednx)
			if err == nil {
				if tool.CheckIsString(&hednx.TxHash) {
					_, err := rpcClient.SendRawTransaction(hednx.TxHash)
					if err != nil {
						log.Println(err)
						return nil, "", err
					}
					hednx.CurrState = dao.TxInfoState_SendHex
					hednx.SendAt = time.Now()
					_ = tx.Update(hednx)
				}
			}
		}
	} else { // 提现人是这次htlc的转账接收者
		//如果还没有获取到R 执行HTD1b
		if lastCommitmentTx.CurrState == dao.TxInfoState_Htlc_GetH {
			isRsmcTx = false
			htdnx := &dao.HTLCTimeoutDeliveryTxB{}
			err = tx.Select(q.Eq("CommitmentTxId", lastCommitmentTx.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign), q.Eq("Owner", closeOpStarter)).First(htdnx)
			if err == nil {
				if tool.CheckIsString(&htdnx.TxHash) {
					_, err := rpcClient.SendRawTransaction(htdnx.TxHash)
					if err != nil {
						log.Println(err)
						msg := err.Error()
						if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
							return nil, "", err
						}
					}
					_ = addHTDnxTxToWaitDB(tx, htdnx)
					htdnx.CurrState = dao.TxInfoState_SendHex
					htdnx.SendAt = time.Now()
					_ = tx.Update(htdnx)
				}
			}
		}
	}

	//如果转账方在超时后还没有得到R,或者接收方得到R后想直接体现
	if isRsmcTx {
		htnx := &dao.HTLCTimeoutTxForAAndExecutionForB{}
		err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", lastCommitmentTx.Id), q.Eq("Owner", closeOpStarter), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).First(htnx)
		if err == nil {
			htrd := &dao.RevocableDeliveryTransaction{}
			err = tx.Select(q.Eq("CommitmentTxId", htnx.Id), q.Eq("Owner", closeOpStarter), q.Eq("RDType", 1), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).First(htrd)
			if err == nil {
				if tool.CheckIsString(&htnx.RSMCTxHash) {
					_, err := rpcClient.SendRawTransaction(htnx.RSMCTxHash)
					if err == nil {
						if tool.CheckIsString(&htrd.TxHash) {
							_, err = rpcClient.SendRawTransaction(htrd.TxHash)
							if err != nil {
								log.Println(err)
								msg := err.Error()
								if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
									return nil, "", err
								}
							}
							_ = addRDTxToWaitDB(tx, htrd)
							htnx.CurrState = dao.TxInfoState_SendHex
							htnx.SendAt = time.Now()
							_ = tx.Update(htnx)

							htrd.CurrState = dao.TxInfoState_SendHex
							htrd.SendAt = time.Now()
							_ = tx.Update(htnx)
						}
					}
				}
			}
		}
	}
	// endregion

	// region update obj state to db
	lastCommitmentTx.CurrState = dao.TxInfoState_SendHex
	lastCommitmentTx.SendAt = time.Now()
	err = tx.Update(lastCommitmentTx)
	if err != nil {
		return nil, "", err
	}

	lastRevocableDeliveryTx.CurrState = dao.TxInfoState_SendHex
	lastRevocableDeliveryTx.SendAt = time.Now()
	err = tx.Update(lastRevocableDeliveryTx)
	if err != nil {
		return nil, "", err
	}

	err = addRDTxToWaitDB(tx, lastRevocableDeliveryTx)
	if err != nil {
		return nil, "", err
	}

	channelInfo.CurrState = dao.ChannelState_Close
	channelInfo.CloseAt = time.Now()
	err = tx.Update(channelInfo)
	if err != nil {
		return nil, "", err
	}

	closeOpStaterReqData.CurrState = 1
	_ = tx.Update(closeOpStaterReqData)
	//endregion

	err = tx.Commit()
	if err != nil {
		return nil, "", err
	}
	return nil, closeOpStarter, nil
}

//htlc timeout Delivery 1b
func addHTDnxTxToWaitDB(tx storm.Node, txInfo *dao.HTLCTimeoutDeliveryTxB) (err error) {
	node := &dao.RDTxWaitingSend{}
	count, err := tx.Select(q.Eq("TransactionHex", txInfo.TxHash)).Count(node)
	if err == nil {
		return err
	}
	if count > 0 {
		return errors.New("always save")
	}
	node.TransactionHex = txInfo.TxHash
	node.IsEnable = true
	node.CreateAt = time.Now()
	err = tx.Save(node)
	if err != nil {
		return err
	}
	return nil
}
