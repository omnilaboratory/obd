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

	//  region create BR
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
	newCommitmentTxInfoA, err := createAliceRsmcTxs(dbTx, channelInfo, isAliceExecutionCloseOp, *reqData, dataFromCloseOpStarter, lastCommitmentTxInfoA, *fundingTransaction, user)
	if err != nil {
		return nil, "", err
	}
	newCommitmentTxInfoB, err := createBobRsmcTxs(dbTx, channelInfo, isAliceExecutionCloseOp, *reqData, dataFromCloseOpStarter, lastCommitmentTxInfoB, *fundingTransaction, user)
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
func createAliceRsmcTxs(tx storm.Node, channelInfo dao.ChannelInfo, isAliceExecutionCloseOp bool, reqData bean.HtlcSignCloseCurrTx, dataFromCloseOpStarter dao.HtlcRequestCloseCurrTxInfo, lastCommitmentATx *dao.CommitmentTransaction, fundingTransaction dao.FundingTransaction, operator bean.User) (newCommitmentTxInfo *dao.CommitmentTransaction, err error) {

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
	commitmentTxInfo.CurrState = dao.TxInfoState_Rsmc_CreateAndSign
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
func createBobRsmcTxs(tx storm.Node, channelInfo dao.ChannelInfo, isAliceExecutionCloseOp bool, reqData bean.HtlcSignCloseCurrTx, dataFromCloseOpStarter dao.HtlcRequestCloseCurrTxInfo, lastCommitmentBTx *dao.CommitmentTransaction, fundingTransaction dao.FundingTransaction, operator bean.User) (newCommitmentTxInfo *dao.CommitmentTransaction, err error) {

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
	commitmentTxInfo.CurrState = dao.TxInfoState_Rsmc_CreateAndSign
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

// BR2a,HBR1a,HTBr1a
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
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("CurrState", dao.TxInfoState_Rsmc_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
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
		br.CurrState = dao.TxInfoState_Rsmc_CreateAndSign
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
		hbr.CurrState = dao.TxInfoState_Rsmc_CreateAndSign
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
		htbr.CurrState = dao.TxInfoState_Rsmc_CreateAndSign
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
	err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id), q.Eq("CurrState", dao.TxInfoState_Rsmc_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
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
		br.CurrState = dao.TxInfoState_Rsmc_CreateAndSign
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
		hbr.CurrState = dao.TxInfoState_Rsmc_CreateAndSign
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
		htbr.CurrState = dao.TxInfoState_Rsmc_CreateAndSign
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

func (service *htlcCloseTxManager) RequestCloseChannel(msgData string, user bean.User) error {

	return nil
}

func (service *htlcCloseTxManager) SignCloseChannel(msgData string, user bean.User) error {

	return nil
}
