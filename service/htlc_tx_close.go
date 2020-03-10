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

	// region check data
	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New("wrong channel Id")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New("empty CurrRsmcTempAddressPubKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPrivateKey) == false {
		err = errors.New("empty CurrRsmcTempAddressPrivateKey")
		log.Println(err)
		return nil, "", err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.CurrRsmcTempAddressPrivateKey, reqData.CurrRsmcTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("CurrRsmcTempAddressPrivateKey is wrong")
	}

	if tool.CheckIsString(&reqData.LastHtlcTempAddressForHtnxPrivateKey) == false {
		err = errors.New("empty LastHtlcTempAddressForHtnxPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.LastHtlcTempAddressPrivateKey) == false {
		err = errors.New("empty LastHtlcTempAddressPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.LastRsmcTempAddressPrivateKey) == false {
		err = errors.New("empty LastRsmcTempAddressPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("empty ChannelAddressPrivateKey")
		log.Println(err)
		return nil, "", err
	}
	// endregion

	commitmentTxInfo, err := getHtlcLatestCommitmentTx(reqData.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastRsmcTempAddressPrivateKey, commitmentTxInfo.RSMCTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("LastRsmcTempAddressPrivateKey is wrong")
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressPrivateKey, commitmentTxInfo.HTLCTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("LastHtlcTempAddressPrivateKey is wrong")
	}

	channelInfo := dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", commitmentTxInfo.ChannelId),
		q.Eq("CurrState", dao.ChannelState_HtlcBegin)).
		First(&channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(channelInfo)

	ht1aOrHe1b := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id),
		q.Eq("Owner", user.PeerId)).
		First(&ht1aOrHe1b)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressForHtnxPrivateKey, ht1aOrHe1b.RSMCTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("LastHtlcTempAddressForHtnxPrivateKey is wrong")
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
	tempAddrPrivateKeyMap[ht1aOrHe1b.RSMCTempAddressPubKey] = reqData.LastHtlcTempAddressForHtnxPrivateKey
	tempAddrPrivateKeyMap[reqData.CurrRsmcTempAddressPubKey] = reqData.CurrRsmcTempAddressPrivateKey

	info := dao.HtlcRequestCloseCurrTxInfo{}
	info.ChannelId = commitmentTxInfo.ChannelId
	info.CurrRsmcTempAddressPubKey = reqData.CurrRsmcTempAddressPubKey
	info.CreateBy = user.PeerId
	info.CurrState = dao.NS_Create
	infoBytes, _ := json.Marshal(info)
	requestHash := tool.SignMsgWithSha256(infoBytes)
	info.CreateAt = time.Now()
	info.RequestHash = requestHash
	count, _ := db.Select(
		q.Eq("ChannelId", info.ChannelId),
		q.Eq("CreateBy", info.CreateBy),
		q.Eq("CurrState", info.CurrState)).
		Count(&dao.HtlcRequestCloseCurrTxInfo{})
	if count == 0 {
		err = db.Save(&info)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
	}

	msgHash := MessageService.saveMsg(user.PeerId, targetUser, info.RequestHash)
	if tool.CheckIsString(&msgHash) == false {
		return nil, "", errors.New("fail to save msgHash")
	}
	info.RequestHash = msgHash

	return info, targetUser, nil
}

// -49 close htlc
func (service *htlcCloseTxManager) CloseHTLCSigned(msgData string, user bean.User) (outData map[string]interface{}, targetUser string, err error) {
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
	if tool.CheckIsString(&reqData.RequestCloseHtlcHash) == false {
		err = errors.New("empty request_close_htlc_hash")
		log.Println(err)
		return nil, "", err
	}

	message, err := MessageService.getMsg(reqData.RequestCloseHtlcHash)
	if err != nil {
		return nil, "", errors.New("wrong request_close_htlc_hash")
	}
	if message.Receiver != user.PeerId {
		return nil, "", errors.New("you are not the operator")
	}
	reqData.RequestCloseHtlcHash = message.Data

	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPubKey) == false {
		err = errors.New("empty curr_rsmc_temp_address_pub_key")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.CurrRsmcTempAddressPrivateKey) == false {
		err = errors.New("empty curr_rsmc_temp_address_private_key")
		log.Println(err)
		return nil, "", err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.CurrRsmcTempAddressPrivateKey, reqData.CurrRsmcTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("CurrRsmcTempAddressPrivateKey is wrong")
	}

	if tool.CheckIsString(&reqData.LastHtlcTempAddressForHtnxPrivateKey) == false {
		err = errors.New("empty last_htlc_temp_address_for_htnx_private_key")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.LastHtlcTempAddressPrivateKey) == false {
		err = errors.New("empty last_htlc_temp_address_private_key")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.LastRsmcTempAddressPrivateKey) == false {
		err = errors.New("empty last_rsmc_temp_address_private_key")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("empty channel_address_private_key")
		log.Println(err)
		return nil, "", err
	}
	// endregion

	// region query data
	dataFromCloseOpStarter := dao.HtlcRequestCloseCurrTxInfo{}
	err = db.Select(
		q.Eq("RequestHash", reqData.RequestCloseHtlcHash),
		q.Eq("CurrState", dao.NS_Create)).
		First(&dataFromCloseOpStarter)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	commitmentTxInfo, err := getHtlcLatestCommitmentTx(dataFromCloseOpStarter.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastRsmcTempAddressPrivateKey, commitmentTxInfo.RSMCTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("LastRsmcTempAddressPrivateKey is wrong")
	}
	_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressPrivateKey, commitmentTxInfo.HTLCTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("LastHtlcTempAddressPrivateKey is wrong")
	}

	channelInfo := dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", commitmentTxInfo.ChannelId),
		q.Eq("CurrState", dao.ChannelState_HtlcBegin)).
		First(&channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId)),
		q.Eq("CurrState", dao.FundingTransactionState_Accept)).
		OrderBy("CreateAt").Reverse().
		First(fundingTransaction)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	ht1aOrHe1b := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", commitmentTxInfo.Id),
		q.Eq("Owner", user.PeerId)).
		First(&ht1aOrHe1b)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastHtlcTempAddressForHtnxPrivateKey, ht1aOrHe1b.RSMCTempAddressPubKey)
	if err != nil {
		return nil, "", errors.New("LastHtlcTempAddressForHtnxPrivateKey is wrong")
	}

	isAliceExecutionCloseOp := true
	targetUser = channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
		isAliceExecutionCloseOp = false
	}
	// endregion

	currNodeChannelPubKey := ""
	otherSideChannelPubKey := ""
	if channelInfo.PeerIdA == user.PeerId {
		currNodeChannelPubKey = channelInfo.PubKeyA
		otherSideChannelPubKey = channelInfo.PubKeyB
	} else {
		currNodeChannelPubKey = channelInfo.PubKeyB
		otherSideChannelPubKey = channelInfo.PubKeyA
	}

	otherSideChannelPrivateKey := tempAddrPrivateKeyMap[otherSideChannelPubKey]
	if tool.CheckIsString(&otherSideChannelPrivateKey) == false {
		return nil, targetUser, errors.New("sender private key is miss,send 48 again")
	}

	_, err = tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, currNodeChannelPubKey)
	if err != nil {
		return nil, "", errors.New("ChannelAddressPrivateKey is wrong")
	}

	tempAddrPrivateKeyMap[currNodeChannelPubKey] = reqData.ChannelAddressPrivateKey
	dbTx, err := db.Begin(true)
	if err != nil {
		return nil, "", err
	}
	defer dbTx.Rollback()

	//  region create BR  BR2a HBR1a HTBR1a ; BR2b HBR1b HEBR1b
	/// HBR1a的存在意义是，在规定时间内完成了交易，创建了后续的C3x，ht1a是不会被广播的，那么如果alice这个时候反悔了，就用这个Br去惩罚她
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
	log.Println(newCommitmentTxInfoA)
	newCommitmentTxInfoB, err := createBobRsmcTxsForCloseHtlc(dbTx, channelInfo, isAliceExecutionCloseOp, *reqData, dataFromCloseOpStarter, lastCommitmentTxInfoB, *fundingTransaction, user)
	if err != nil {
		return nil, "", err
	}
	log.Println(newCommitmentTxInfoB)
	// endregion

	channelInfo.CurrState = dao.ChannelState_CanUse
	_ = dbTx.Update(&channelInfo)

	dataFromCloseOpStarter.CurrState = dao.NS_Finish
	_ = dbTx.Update(&dataFromCloseOpStarter)

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
	return outData, targetUser, nil
}

// C3a RD3a
func createAliceRsmcTxsForCloseHtlc(tx storm.Node, channelInfo dao.ChannelInfo, isAliceExecutionCloseOp bool,
	reqData bean.HtlcSignCloseCurrTx, dataFromCloseOpStarter dao.HtlcRequestCloseCurrTxInfo,
	lastCommitmentATx *dao.CommitmentTransaction, fundingTransaction dao.FundingTransaction, operator bean.User) (newCommitmentTxInfo *dao.CommitmentTransaction, err error) {

	owner := channelInfo.PeerIdA
	// 这里需要确认一个事情：在这个通道里面，这次的htlc到底是谁是转出方
	// rmsc的资产分配方案
	var outputBean = commitmentOutputBean{}
	if lastCommitmentATx.CurrState == dao.TxInfoState_Htlc_GetH { //转账失败 退钱回去
		if lastCommitmentATx.HtlcSender == owner {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentATx.AmountToHtlc)).Float64()
			outputBean.AmountToOther = lastCommitmentATx.AmountToOther
		} else {
			outputBean.AmountToRsmc = lastCommitmentATx.AmountToRSMC
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToOther).Add(decimal.NewFromFloat(lastCommitmentATx.AmountToHtlc)).Float64()
		}
	} else {
		if lastCommitmentATx.HtlcSender == owner {
			outputBean.AmountToRsmc = lastCommitmentATx.AmountToRSMC
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToOther).Add(decimal.NewFromFloat(lastCommitmentATx.AmountToHtlc)).Float64()
		} else {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentATx.AmountToHtlc)).Float64()
			outputBean.AmountToOther = lastCommitmentATx.AmountToOther
		}
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
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUserSingleInput(
			int(commitmentTxInfo.TxType),
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
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseRestInput(
			int(commitmentTxInfo.TxType),
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

	// create RDna tx
	rdTransaction, err := createRDTx(owner, &channelInfo, commitmentTxInfo, channelInfo.AddressA, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if isAliceExecutionCloseOp {
		currTempAddressPrivateKey = reqData.CurrRsmcTempAddressPrivateKey
	} else {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCloseOpStarter.CurrRsmcTempAddressPubKey]
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHash, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		commitmentTxInfo.RSMCMultiAddress,
		[]string{
			currTempAddressPrivateKey,
			tempAddrPrivateKeyMap[channelInfo.PubKeyB],
		},
		inputs,
		rdTransaction.OutputAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		rdTransaction.Amount,
		0,
		rdTransaction.Sequence,
		&commitmentTxInfo.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	rdTransaction.Txid = txid
	rdTransaction.TxHash = hex
	rdTransaction.SignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
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
	if lastCommitmentBTx.CurrState == dao.TxInfoState_Htlc_GetH { //转账失败
		if lastCommitmentBTx.HtlcSender == owner {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentBTx.AmountToHtlc)).Float64()
			outputBean.AmountToOther = lastCommitmentBTx.AmountToOther
		} else { // 是收款方
			outputBean.AmountToRsmc = lastCommitmentBTx.AmountToRSMC
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountToOther).Add(decimal.NewFromFloat(lastCommitmentBTx.AmountToHtlc)).Float64()
		}
	} else { //转账成功
		if lastCommitmentBTx.HtlcSender == owner { // 我是转账方
			outputBean.AmountToRsmc = lastCommitmentBTx.AmountToRSMC
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountToOther).Add(decimal.NewFromFloat(lastCommitmentBTx.AmountToHtlc)).Float64()
		} else { // 我是收款方
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentBTx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentBTx.AmountToHtlc)).Float64()
			outputBean.AmountToOther = lastCommitmentBTx.AmountToOther
		}
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
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUserSingleInput(
			int(commitmentTxInfo.TxType),
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
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseRestInput(
			int(commitmentTxInfo.TxType),
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

	// create RDb tx
	rdTransaction, err := createRDTx(owner, &channelInfo, commitmentTxInfo, channelInfo.AddressB, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if isAliceExecutionCloseOp {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCloseOpStarter.CurrRsmcTempAddressPubKey]
	} else {
		currTempAddressPrivateKey = reqData.CurrRsmcTempAddressPrivateKey
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHash, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		commitmentTxInfo.RSMCMultiAddress,
		[]string{
			tempAddrPrivateKeyMap[channelInfo.PubKeyA],
			currTempAddressPrivateKey,
		},
		inputs,
		rdTransaction.OutputAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		rdTransaction.Amount,
		0,
		rdTransaction.Sequence,
		&commitmentTxInfo.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	rdTransaction.Txid = txid
	rdTransaction.TxHash = hex
	rdTransaction.SignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
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
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", owner)).
		OrderBy("CreateAt").Reverse().
		First(lastCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if lastCommitmentTxInfo.TxType != dao.CommitmentTransactionType_Htlc {
		return nil, errors.New("latest commitment tx type is wrong")
	}

	lastRDTransaction := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", owner),
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
		OrderBy("CreateAt").Reverse().
		First(lastRDTransaction)
	if err != nil {
		lastRDTransaction = nil
	}

	htOrHeTx := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("Owner", owner),
	).First(&htOrHeTx)
	if err != nil {
		err = errors.New("not found HT1a from db")
		log.Println(err)
		return nil, err
	}
	htrd := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", htOrHeTx.Id),
		q.Eq("Owner", owner),
		q.Eq("RDType", 1)).
		First(htrd)
	if err != nil {
		err = errors.New("not found HTRD1a from db")
		log.Println(err)
		return nil, err
	}

	// region create BR1a
	count, _ := tx.Select(
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("Owner", owner)).
		Count(&dao.BreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already exist BreachRemedyTransaction ")
		log.Println(err)
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

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
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
	count, _ = tx.Select(
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("Owner", owner)).
		Count(&dao.HTLCBreachRemedyTransaction{})
	if count > 0 {
		err = errors.New("already create HTLCBreachRemedyTransaction")
		return nil, err
	}
	hbr, err := createHtlcBRTx(brOwner, &channelInfo, lastCommitmentTxInfo, &user)
	if err != nil {
		err = errors.New("fail to create the HtlcBR")
		log.Println(err)
		return nil, err
	}

	if hbr.Amount > 0 {
		lastTempAddressPrivateKey := ""
		// 如果当前操作用户是PeerIdA方（概念中的Alice方），则取当前操作人传入的数据
		if isAliceExecutionCloseOp {
			lastTempAddressPrivateKey = reqData.LastHtlcTempAddressPrivateKey
		} else {
			// 如果当前操作用户是PeerIdB方(概念中bob方)，而我们现在正在处理Alice方，所以我们要取另一方（alice）的数据
			lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentTxInfo.HTLCTempAddressPubKey]
		}
		if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
			err = errors.New("fail to get the lastTempAddressPrivateKey")
			log.Println(err)
			return nil, err
		}

		inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTxInfo.HtlcTxHash, lastCommitmentTxInfo.HTLCMultiAddress, lastCommitmentTxInfo.HTLCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
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
	count, _ = tx.Select(
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("HTLCTimeoutTxForAAndExecutionForBId", htOrHeTx.Id),
		q.Eq("Owner", owner)).Count(&dao.HTLCTimeoutBreachRemedyTransaction{})
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
			lastTempAddressPrivateKey = reqData.LastHtlcTempAddressForHtnxPrivateKey
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

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
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
	_ = tx.Update(lastCommitmentTxInfo)
	if lastRDTransaction != nil {
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		_ = tx.Update(lastRDTransaction)
	}
	htrd.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(htrd)

	return lastCommitmentTxInfo, nil
}

// BR2b,HBR1b,HEBR1b
func createBobSideBRTxs(tx storm.Node, channelInfo dao.ChannelInfo, isAliceExecutionCloseOp bool, reqData bean.HtlcSignCloseCurrTx, fundingTransaction dao.FundingTransaction, user bean.User) (lastCommitmentTxInfo *dao.CommitmentTransaction, err error) {

	owner := channelInfo.PeerIdB
	brOwner := channelInfo.PeerIdA
	lastCommitmentTxInfo = &dao.CommitmentTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", owner)).
		OrderBy("CreateAt").Reverse().
		First(lastCommitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if lastCommitmentTxInfo.TxType != dao.CommitmentTransactionType_Htlc {
		return nil, errors.New("latest commitment tx type is wrong")
	}

	lastRDTransaction := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("Owner", owner),
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
		OrderBy("CreateAt").Reverse().
		First(lastRDTransaction)
	if err != nil {
		lastRDTransaction = nil
	}

	htOrHeTx := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("Owner", owner),
	).First(&htOrHeTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	htrd := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", htOrHeTx.Id),
		q.Eq("Owner", owner),
		q.Eq("RDType", 1)).
		First(htrd)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// region create BR1b

	//如果已经创建过了，return
	count, _ := tx.Select(
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("Owner", owner)).
		Count(&dao.BreachRemedyTransaction{})
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

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
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
	count, _ = tx.Select(
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("Owner", owner)).
		Count(&dao.HTLCBreachRemedyTransaction{})
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

		inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTxInfo.HtlcTxHash, lastCommitmentTxInfo.HTLCMultiAddress, lastCommitmentTxInfo.HTLCRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
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
	count, _ = tx.Select(
		q.Eq("CommitmentTxId", lastCommitmentTxInfo.Id),
		q.Eq("HTLCTimeoutTxForAAndExecutionForBId", htOrHeTx.Id),
		q.Eq("Owner", owner)).Count(&dao.HTLCTimeoutBreachRemedyTransaction{})
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
			lastTempAddressPrivateKey = reqData.LastHtlcTempAddressForHtnxPrivateKey
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

		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
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
	_ = tx.Update(lastCommitmentTxInfo)
	if lastRDTransaction != nil {
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		_ = tx.Update(lastRDTransaction)
	}
	htrd.CurrState = dao.TxInfoState_Abord
	_ = tx.Update(htrd)

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
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", dao.ChannelState_HtlcBegin)).
		First(channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	targetUser = channelInfo.PeerIdB
	if user.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	}

	lastCommitmentTx := &dao.CommitmentTransaction{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("TxType", dao.CommitmentTransactionType_Htlc),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		First(lastCommitmentTx)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	closeChannel := &dao.CloseChannel{}
	closeChannel.ChannelId = reqData.ChannelId
	closeChannel.Owner = user.PeerId
	closeChannel.CurrState = 0
	count, _ := db.Select(
		q.Eq("ChannelId", closeChannel.ChannelId),
		q.Eq("Owner", closeChannel.Owner),
		q.Eq("CurrState", closeChannel.CurrState)).
		Count(closeChannel)
	if count == 0 {
		dataBytes, _ := json.Marshal(closeChannel)
		closeChannel.RequestHex = tool.SignMsgWithSha256(dataBytes)
		closeChannel.CreateAt = time.Now()
		err = db.Save(closeChannel)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
	}

	outData = make(map[string]interface{})
	outData["channel_id"] = reqData.ChannelId
	outData["request_close_channel_hash"] = closeChannel.RequestHex
	return outData, targetUser, nil
}

// -51 对通道关闭进行表态 谁请求，就以谁的身份关闭通道
func (service *htlcCloseTxManager) CloseHtlcChannelSigned(msgData string, user bean.User) (outData interface{}, closeOpStarter string, err error) {
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
	if tool.CheckIsString(&reqData.ChannelId) == false {
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
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", 0),
		q.Eq("RequestHex", reqData.RequestCloseChannelHash)).
		First(closeOpStaterReqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	channelInfo := &dao.ChannelInfo{}
	err = db.Select(
		q.Eq("ChannelId", reqData.ChannelId),
		q.Eq("CurrState", dao.ChannelState_HtlcBegin)).
		First(channelInfo)
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

	latestCommitmentTx, err := getHtlcLatestCommitmentTx(channelInfo.ChannelId, user.PeerId)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&latestCommitmentTx.RSMCTxHash) {
		commitmentTxid, err := rpcClient.SendRawTransaction(latestCommitmentTx.RSMCTxHash)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxid)
	}

	if tool.CheckIsString(&latestCommitmentTx.ToOtherTxHash) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(latestCommitmentTx.ToOtherTxHash)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxidToBob)
	}

	latestRsmcRD := &dao.RevocableDeliveryTransaction{}
	err = db.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CommitmentTxId", latestCommitmentTx.Id),
		q.Eq("RDType", 0),
		q.Eq("Owner", closeOpStarter)).
		OrderBy("CreateAt").Reverse().
		First(latestRsmcRD)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	latestRsmcRDTxid, err := rpcClient.SendRawTransaction(latestRsmcRD.TxHash)
	if err != nil {
		log.Println(err)
		msg := err.Error()
		if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
			return nil, "", err
		}
	}
	log.Println(latestRsmcRDTxid)

	if tool.CheckIsString(&latestCommitmentTx.HtlcTxHash) {
		commitmentTxidToHtlc, err := rpcClient.SendRawTransaction(latestCommitmentTx.HtlcTxHash)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		log.Println(commitmentTxidToHtlc)
	}

	// endregion

	// 提现人是否是这次htlc的转账发起者
	var withdrawerIsHtlcSender bool
	if latestCommitmentTx.HtlcSender == closeOpStarter {
		withdrawerIsHtlcSender = true
	} else {
		withdrawerIsHtlcSender = false
	}

	tx, err := db.Begin(true)
	defer tx.Rollback()

	// region htlc的相关交易广播
	isRsmcTx := true
	// 提现人是这次htlc的转账发起者
	if withdrawerIsHtlcSender {
		// 如果已经得到R，直接广播HED1a
		if latestCommitmentTx.CurrState == dao.TxInfoState_Htlc_GetR {
			isRsmcTx = false
			hednx := &dao.HTLCExecutionDeliveryOfR{}
			err = tx.Select(
				q.Eq("CommitmentTxId", latestCommitmentTx.Id),
				q.Eq("CurrState", dao.TxInfoState_CreateAndSign),
				q.Eq("Owner", closeOpStarter)).
				First(hednx)
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
		if latestCommitmentTx.CurrState == dao.TxInfoState_Htlc_GetH {
			isRsmcTx = false
			htdnx := &dao.HTLCTimeoutDeliveryTxB{}
			err = tx.Select(
				q.Eq("CommitmentTxId", latestCommitmentTx.Id),
				q.Eq("CurrState", dao.TxInfoState_CreateAndSign),
				q.Eq("Owner", closeOpStarter)).
				First(htdnx)
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

	//如果转账方在超时后还没有得到R,或者接收方得到R后想直接提现
	if isRsmcTx {
		htnx := &dao.HTLCTimeoutTxForAAndExecutionForB{}
		err = tx.Select(
			q.Eq("ChannelId", channelInfo.ChannelId),
			q.Eq("CommitmentTxId", latestCommitmentTx.Id),
			q.Eq("Owner", closeOpStarter),
			q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
			First(htnx)
		if err == nil {
			htrd := &dao.RevocableDeliveryTransaction{}
			err = tx.Select(
				q.Eq("CommitmentTxId", htnx.Id),
				q.Eq("Owner", closeOpStarter), q.Eq("RDType", 1),
				q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
				First(htrd)
			if err == nil {
				if tool.CheckIsString(&htnx.RSMCTxHash) {
					_, err := rpcClient.SendRawTransaction(htnx.RSMCTxHash)
					if err == nil { //如果已经超时
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
						}
					} else {
						if err != nil { // 如果是超时内的正常提现
							log.Println(err)
							if htnx.Timeout == 0 {
								return nil, "", err
							}
							msg := err.Error()
							if strings.Contains(msg, "non-BIP68-final (code 64)") == false {
								return nil, "", err
							}
						}
						_ = addHT1aTxToWaitDB(tx, htnx, htrd)
					}
				}
			}
		}
	}
	// endregion

	// region update obj state to db
	latestCommitmentTx.CurrState = dao.TxInfoState_SendHex
	latestCommitmentTx.SendAt = time.Now()
	err = tx.Update(latestCommitmentTx)
	if err != nil {
		return nil, "", err
	}

	latestRsmcRD.CurrState = dao.TxInfoState_SendHex
	latestRsmcRD.SendAt = time.Now()
	err = tx.Update(latestRsmcRD)
	if err != nil {
		return nil, "", err
	}

	err = addRDTxToWaitDB(tx, latestRsmcRD)
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

func addHT1aTxToWaitDB(tx storm.Node, htnx *dao.HTLCTimeoutTxForAAndExecutionForB, htrd *dao.RevocableDeliveryTransaction) error {
	node := &dao.RDTxWaitingSend{}
	count, err := tx.Select(
		q.Eq("TransactionHex", htnx.RSMCTxHash)).
		Count(node)
	if err == nil {
		return err
	}
	if count > 0 {
		return errors.New("already save")
	}
	node.TransactionHex = htnx.RSMCTxHash
	node.Type = 1
	node.IsEnable = true
	node.CreateAt = time.Now()
	node.HtnxIdAndHtnxRdId = make([]int, 2)
	node.HtnxIdAndHtnxRdId[0] = htnx.Id
	node.HtnxIdAndHtnxRdId[1] = htrd.Id
	err = tx.Save(node)
	if err != nil {
		return err
	}
	return nil
}

func addHTRD1aTxToWaitDB(htnxIdAndHtnxRdId []int) error {
	htnxId := htnxIdAndHtnxRdId[0]
	htrdId := htnxIdAndHtnxRdId[1]
	htnx := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err := db.One("Id", htnxId, &htnx)
	if err != nil {
		return err
	}

	htrd := dao.RevocableDeliveryTransaction{}
	err = db.One("Id", htrdId, &htrd)
	if err != nil {
		return err
	}

	node := &dao.RDTxWaitingSend{}
	count, err := db.Select(
		q.Eq("TransactionHex", htrd.TxHash)).
		Count(node)
	if err == nil {
		return err
	}
	if count > 0 {
		return errors.New("already save")
	}

	node.TransactionHex = htrd.TxHash
	node.Type = 0
	node.IsEnable = true
	node.CreateAt = time.Now()
	err = db.Save(node)
	if err != nil {
		return err
	}

	htnx.CurrState = dao.TxInfoState_SendHex
	htnx.SendAt = time.Now()
	_ = db.Update(htnx)

	return nil
}

//htlc timeout Delivery 1b
func addHTDnxTxToWaitDB(tx storm.Node, txInfo *dao.HTLCTimeoutDeliveryTxB) (err error) {
	node := &dao.RDTxWaitingSend{}
	count, err := tx.Select(
		q.Eq("TransactionHex", txInfo.TxHash)).
		Count(node)
	if err == nil {
		return err
	}
	if count > 0 {
		return errors.New("already save")
	}
	node.TransactionHex = txInfo.TxHash
	node.Type = 2
	node.IsEnable = true
	node.CreateAt = time.Now()
	err = tx.Save(node)
	if err != nil {
		return err
	}
	return nil
}
