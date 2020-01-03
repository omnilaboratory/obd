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

type htlcBackwardTxManager struct {
	operationFlag sync.Mutex
}

// HTLC Reverse pass the R (Preimage R)
var HtlcBackwardTxService htlcBackwardTxManager

// SendRToPreviousNode
//
// Process type -46: Send R to Previous Node (middleman).
//  * R is <Preimage_R>
func (service *htlcBackwardTxManager) SendRToPreviousNode(msgData string,
	user bean.User) (data map[string]interface{}, previousNode string, err error) {

	// region Parse data inputed from websocket client of sender.
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcSendR{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion

	// region Check data inputed from websocket client of sender.
	if tool.CheckIsString(&reqData.RequestHash) == false {
		err = errors.New("empty request_hash")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_cnb_private_key is empty")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressForHE1bPubKey) == false {
		err = errors.New("curr_htlc_temp_address_for_he1b_pub_key is empty")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressForHE1bPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_he1b_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	// endregion

	// region Check out if the input R is correct.
	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = db.Select(
		q.Eq("RequestHash", reqData.RequestHash),
		q.Eq("R", reqData.R), // R from websocket client of sender
		q.Eq("CurrState", dao.NS_Finish)).First(rAndHInfo)

	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion

	// region Get peerId of previous node.
	pathInfo := dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("HAndRInfoRequestHash", reqData.RequestHash)).First(&pathInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// Currently solution is Alice to Bob to Carol.
	if pathInfo.CurrStep < int(pathInfo.TotalStep/2) {
		return nil, "", errors.New("The transfer H has not completed yet.")
	} else if pathInfo.CurrStep > (int(pathInfo.TotalStep/2) + 1) {
		return nil, "", errors.New("The transfer R has completed.")
	}

	currBlockHeight, err := rpcClient.GetBlockCount()
	if err != nil {
		return nil, "", errors.New("fail to get blockHeight ,please try again later")
	}
	needDayCount := (pathInfo.CurrStep + 1) - int(pathInfo.TotalStep/2)
	maxHeight := pathInfo.BeginBlockHeight + needDayCount*singleHopPerHopDuration
	if currBlockHeight > maxHeight {
		return nil, "", errors.New("timeout, can't transfer the R")
	}

	// If CurrStep = 2, that indicate the transfer H has completed.
	currChannelIndex := pathInfo.TotalStep - pathInfo.CurrStep - 1
	if currChannelIndex < -1 || currChannelIndex > len(pathInfo.ChannelIdArr) {
		return nil, "", errors.New("err channel id")
	}

	currChannel := &dao.ChannelInfo{}
	err = db.One("Id", pathInfo.ChannelIdArr[currChannelIndex], currChannel)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}

	if currChannel.PeerIdA != user.PeerId && currChannel.PeerIdB != user.PeerId {
		return nil, "", errors.New("error user.")
	}

	if user.PeerId == currChannel.PeerIdA {
		previousNode = currChannel.PeerIdB
	} else {
		previousNode = currChannel.PeerIdA
	}

	err = FindUserIsOnline(previousNode)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// Transfer H or R increase step.
	// endregion

	// region Save private key to memory.
	//
	// PeerIdA of the channel is the sender of transfer R.
	// ChannelAddressPrivateKey is a private key of the sender.
	//
	// Example: When Bob transfer R to Alice, Bob is the sender.
	// Alice create HED1a need the private key of Bob for sign that.
	if user.PeerId == currChannel.PeerIdA {
		tempAddrPrivateKeyMap[currChannel.PubKeyA] = reqData.ChannelAddressPrivateKey
	} else { // PeerIdB is the sender of transfer R.
		tempAddrPrivateKeyMap[currChannel.PubKeyB] = reqData.ChannelAddressPrivateKey
	}
	tempAddrPrivateKeyMap[reqData.CurrHtlcTempAddressForHE1bPubKey] = reqData.CurrHtlcTempAddressForHE1bPrivateKey

	commitmentTxInfo, err := getLatestCommitmentTx(currChannel.ChannelId, user.PeerId)
	if err != nil {
		return nil, "", err
	}

	if commitmentTxInfo.TxType != dao.CommitmentTransactionType_Htlc {
		return nil, "", errors.New("error tx type")
	}
	tempAddrPrivateKeyMap[commitmentTxInfo.HTLCTempAddressPubKey] = reqData.CurrHtlcTempAddressPrivateKey

	// Save pubkey to database.
	dataChange := false
	if pathInfo.BobCurrHtlcTempForHt1bPubKey != reqData.CurrHtlcTempAddressForHE1bPubKey {
		pathInfo.BobCurrHtlcTempForHt1bPubKey = reqData.CurrHtlcTempAddressForHE1bPubKey
		dataChange = true
	}
	if pathInfo.CurrState != dao.SingleHopPathInfoState_StepBegin {
		pathInfo.CurrState = dao.SingleHopPathInfoState_StepBegin
		dataChange = true
	}

	if dataChange {
		err = db.Update(&pathInfo)
		if err != nil {
			log.Println(err.Error())
			return nil, "", err
		}
	}
	// endregion

	// Generate response message.
	// If no error, the response data is displayed in websocket client of previous node.
	// Otherwise, it is displayed in websocket client of myself.
	responseData := make(map[string]interface{})
	responseData["id"] = rAndHInfo.Id
	responseData["request_hash"] = rAndHInfo.RequestHash
	responseData["r"] = rAndHInfo.R

	return responseData, previousNode, nil
}

// CheckRAndCreateTxs
//
// Process type -47: Middleman node Check out if R is correct
// and create commitment transactions.
//  * R is <Preimage_R>
func (service *htlcBackwardTxManager) CheckRAndCreateTxs(msgData string, user bean.User) (
	data map[string]interface{}, recipientUser string, err error) {

	// region Parse data inputed from websocket client of middleman node.
	if tool.CheckIsString(&msgData) == false {
		return nil, "", errors.New("empty json data")
	}

	reqData := &bean.HtlcCheckRAndCreateTx{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion

	// region Check data inputed from websocket client of  middleman node.
	if tool.CheckIsString(&reqData.RequestHash) == false {
		err = errors.New("empty request_hash")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("channel_address_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	if tool.CheckIsString(&reqData.R) == false {
		err = errors.New("R is empty")
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.CurrHtlcTempAddressPrivateKey) == false {
		err = errors.New("curr_htlc_temp_address_for_cna_private_key is empty")
		log.Println(err)
		return nil, "", err
	}
	// endregion

	// region Check out if the request hash is correct.
	rAndHInfo := &dao.HtlcRAndHInfo{}
	err = db.Select(
		q.Eq("RequestHash", reqData.RequestHash),
		q.Eq("R", reqData.R),
		q.Eq("CurrState", dao.NS_Finish)).First(rAndHInfo)

	if err != nil {
		log.Println(err.Error())
		return nil, "", err
	}
	// endregion

	// region Create HED1a commitment transaction.
	// launch database transaction, if anything goes wrong, roll back.
	dbTx, err := db.Begin(true)
	if err != nil {
		return nil, "", err
	}
	defer dbTx.Rollback()

	// prepare the data
	pathInfo := dao.HtlcSingleHopPathInfo{}
	err = db.Select(q.Eq("HAndRInfoRequestHash",
		reqData.RequestHash)).First(&pathInfo)

	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	currChannelIndex := pathInfo.TotalStep - pathInfo.CurrStep - 1
	if currChannelIndex < -1 || currChannelIndex > len(pathInfo.ChannelIdArr) {
		return nil, "", errors.New("err channel id")
	}

	channelInfo := dao.ChannelInfo{}
	err = dbTx.One("Id", pathInfo.ChannelIdArr[currChannelIndex], &channelInfo)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if user.PeerId == channelInfo.PeerIdA {
		tempAddrPrivateKeyMap[channelInfo.PubKeyA] = reqData.ChannelAddressPrivateKey
	} else {
		tempAddrPrivateKeyMap[channelInfo.PubKeyB] = reqData.ChannelAddressPrivateKey
	}
	defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyA)
	defer delete(tempAddrPrivateKeyMap, channelInfo.PubKeyB)
	defer delete(tempAddrPrivateKeyMap, pathInfo.BobCurrHtlcTempForHt1bPubKey)

	// 判断自己是否有作为发送方的交易
	// 只有每个通道的转账发送方才能去创建关于R的交易
	commitmentTransaction := dao.CommitmentTransaction{}
	err = dbTx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("TxType", dao.CommitmentTransactionType_Htlc),
		q.Eq("HtlcH", rAndHInfo.H),
		q.Eq("HtlcSender", user.PeerId),
		q.Eq("Owner", user.PeerId),
		q.Eq("CurrState", dao.TxInfoState_Htlc_GetH)).
		First(&commitmentTransaction)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// get the funding transaction
	var fundingTransaction = &dao.FundingTransaction{}
	err = dbTx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("CurrState", dao.FundingTransactionState_Accept)).
		OrderBy("CreateAt").
		Reverse().
		First(fundingTransaction)

	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	// endregion

	var aliceIsSender bool
	// 如果通道的PeerIdA（概念中的Alice）作为发送方
	if channelInfo.PeerIdA == user.PeerId {
		aliceIsSender = true
		recipientUser = channelInfo.PeerIdB
	} else {
		aliceIsSender = false
		recipientUser = channelInfo.PeerIdA
	}

	// 如果转账发送方是PeerIdA（Alice），也就是Alice转账给bob，
	// 那就创建HED1a:Alice这边直接给钱bob，不需要时间锁定;
	// HE1b,HERD1b，Bob因为是收款方，他自己的钱需要RSMC的方式锁在通道.
	// 如果转账发送方是PeerIdB（Bob），也就是Bob转账给Alice，那就创建HE1a,HERD1a：
	// Alice作为收款方，她得到的钱就需要RSMC锁定;
	// HED1b：bob是发送方，他这边给Alice的钱是不需要锁定

	// region	HED1x
	_, err = htlcCreateExecutionDelivery(dbTx, channelInfo, *fundingTransaction, commitmentTransaction, *reqData, aliceIsSender)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	// endregion

	// region 对方的两个交易 he1x herd1x
	commitmentTransactionB := dao.CommitmentTransaction{}
	err = dbTx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("TxType", dao.CommitmentTransactionType_Htlc),
		q.Eq("HtlcH", rAndHInfo.H),
		q.Eq("HtlcSender", user.PeerId),
		q.Eq("Owner", recipientUser),
		q.Eq("CurrState", dao.TxInfoState_Htlc_GetH)).First(&commitmentTransactionB)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// HE1x 一个新的Rsmc
	he1x, err := createHtlcExecution(dbTx, channelInfo, *fundingTransaction, commitmentTransactionB, *reqData, pathInfo, aliceIsSender, user)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	// htrd1x
	herd1x, err := createHtlcRDForR(dbTx, channelInfo, *fundingTransaction, he1x, pathInfo, *reqData, aliceIsSender, user)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	log.Println(herd1x)
	//endregion

	commitmentTransaction.HtlcR = rAndHInfo.R
	commitmentTransaction.CurrState = dao.TxInfoState_Htlc_GetR
	commitmentTransaction.LastEditTime = time.Now()
	_ = dbTx.Update(&commitmentTransaction)
	commitmentTransactionB.HtlcR = rAndHInfo.R
	commitmentTransactionB.CurrState = dao.TxInfoState_Htlc_GetR
	commitmentTransactionB.LastEditTime = time.Now()
	_ = dbTx.Update(&commitmentTransactionB)

	pathInfo.CurrStep += 1
	pathInfo.CurrState = dao.SingleHopPathInfoState_StepFinish
	err = dbTx.Update(&pathInfo)

	err = dbTx.Commit()
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	data = make(map[string]interface{})
	data["r"] = commitmentTransaction.HtlcR
	data["request_hash"] = rAndHInfo.RequestHash
	return data, recipientUser, nil
}

//创建hed1a  此交易要修改创建时机，等到bob拿到R的时候，再来创建，这个时候就需要广播交易（关闭通道），那么在很多情况下，其实是不用创建的
func htlcCreateExecutionDelivery(tx storm.Node, channelInfo dao.ChannelInfo, fundingTransaction dao.FundingTransaction,
	commitmentTxInfo dao.CommitmentTransaction, reqData bean.HtlcCheckRAndCreateTx, aliceIsSender bool) (he1x *dao.HTLCExecutionDelivery, err error) {

	otherSideChannelAddress := channelInfo.AddressB
	otherSideChannelPubKey := channelInfo.PubKeyB
	owner := channelInfo.PeerIdB
	if aliceIsSender == false {
		otherSideChannelAddress = channelInfo.AddressA
		otherSideChannelPubKey = channelInfo.PubKeyA
		owner = channelInfo.PeerIdA
	}

	he1x = &dao.HTLCExecutionDelivery{}
	he1x.Owner = owner
	he1x.OutputAddress = otherSideChannelAddress
	he1x.OutAmount = commitmentTxInfo.AmountToHtlc

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.HtlcTxHash, commitmentTxInfo.HTLCMultiAddress, commitmentTxInfo.HTLCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		commitmentTxInfo.HTLCMultiAddress,
		[]string{
			reqData.CurrHtlcTempAddressPrivateKey,
			tempAddrPrivateKeyMap[otherSideChannelPubKey],
		},
		inputs,
		he1x.OutputAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		he1x.OutAmount,
		0,
		0,
		&commitmentTxInfo.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	he1x.Txid = txid
	he1x.TxHash = hex
	he1x.CreateAt = time.Now()
	he1x.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(he1x)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return he1x, nil
}

func createHtlcExecution(tx storm.Node, channelInfo dao.ChannelInfo, fundingTransaction dao.FundingTransaction,
	commitmentTxInfo dao.CommitmentTransaction, reqData bean.HtlcCheckRAndCreateTx,
	pathInfo dao.HtlcSingleHopPathInfo, aliceIsSender bool, operator bean.User) (he1x *dao.HTLCTimeoutTxForAAndExecutionForB, err error) {

	outputBean := commitmentOutputBean{}
	outputBean.AmountToRsmc = commitmentTxInfo.AmountToHtlc
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
	outputBean.RsmcTempPubKey = pathInfo.BobCurrHtlcTempPubKey
	owner := channelInfo.PeerIdB
	if aliceIsSender == false {
		owner = channelInfo.PeerIdA
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
	}

	he1x, err = createHtlcTimeoutTxObj(tx, owner, channelInfo, commitmentTxInfo, outputBean, 0, operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.HtlcTxHash, commitmentTxInfo.HTLCMultiAddress, commitmentTxInfo.HTLCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		commitmentTxInfo.HTLCMultiAddress,
		[]string{
			reqData.ChannelAddressPrivateKey,
			tempAddrPrivateKeyMap[commitmentTxInfo.HTLCTempAddressPubKey],
		},
		inputs,
		he1x.RSMCMultiAddress,
		he1x.RSMCMultiAddress,
		fundingTransaction.PropertyId,
		he1x.RSMCOutAmount,
		0,
		he1x.Timeout,
		&commitmentTxInfo.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	he1x.RSMCTxid = txid
	he1x.RSMCTxHash = hex
	he1x.SignAt = time.Now()
	he1x.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(he1x)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return he1x, nil
}

func createHtlcRDForR(tx storm.Node, channelInfo dao.ChannelInfo,
	fundingTransaction dao.FundingTransaction, he1x *dao.HTLCTimeoutTxForAAndExecutionForB,
	pathInfo dao.HtlcSingleHopPathInfo, reqData bean.HtlcCheckRAndCreateTx,
	aliceIsSender bool, operator bean.User) (*dao.RevocableDeliveryTransaction, error) {

	owner := channelInfo.PeerIdB
	outAddress := channelInfo.AddressB
	if aliceIsSender == false {
		owner = channelInfo.PeerIdA
		outAddress = channelInfo.AddressA
	}

	count, _ := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", he1x.Id), q.Eq("Owner", owner), q.Eq("RDType", 1)).Count(&dao.RevocableDeliveryTransaction{})
	if count > 0 {
		return nil, errors.New("already create")
	}

	rdTransaction, err := createHtlcRDTxObj(owner, &channelInfo, he1x, outAddress, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(he1x.RSMCTxHash, he1x.RSMCMultiAddress, he1x.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		he1x.RSMCMultiAddress,
		[]string{
			reqData.ChannelAddressPrivateKey,
			tempAddrPrivateKeyMap[pathInfo.BobCurrHtlcTempForHt1bPubKey],
		},
		inputs,
		rdTransaction.OutputAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		rdTransaction.Amount,
		0,
		rdTransaction.Sequence,
		&he1x.RSMCRedeemScript)
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
	return rdTransaction, nil
}
