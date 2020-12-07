package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"sync"
	"time"
)

type commitmentTxSignedManager struct {
	operationFlag sync.Mutex
}

//var tempP2pData_353 map[string]bean.AliceSignedC2bTxDataP2p

var CommitmentTxSignedService commitmentTxSignedManager

// step 3 协议号：351 bob所在的obd接收到了alice的转账申请 提送110351
func (this *commitmentTxSignedManager) BeforeBobSignCommitmentTransactionAtBobSide(data string, user *bean.User) (retData *bean.PayerRequestCommitmentTxToBobClient, err error) {
	requestCreateCommitmentTx := &bean.AliceRequestToCreateCommitmentTxOfP2p{}
	_ = json.Unmarshal([]byte(data), requestCreateCommitmentTx)

	retData = &bean.PayerRequestCommitmentTxToBobClient{}
	retData.AliceRequestToCreateCommitmentTxOfP2p = *requestCreateCommitmentTx

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, retData.ChannelId, user.PeerId)
	if channelInfo == nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}

	channelInfo.CurrState = dao.ChannelState_NewTx
	_ = tx.Update(channelInfo)

	senderPeerId := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		senderPeerId = channelInfo.PeerIdB
	}

	// 351的p2p消息缓存到了msg的数据字段了
	messageHash := messageService.saveMsgUseTx(tx, senderPeerId, user.PeerId, data)
	retData.MsgHash = messageHash

	commitmentTxInfo, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if commitmentTxInfo.Id == 0 {
		commitmentTxInfo.Owner = user.PeerId
		commitmentTxInfo.AmountToRSMC = 0
		commitmentTxInfo.AmountToCounterparty = channelInfo.Amount
	}
	sendChannelStateToTracker(*channelInfo, *commitmentTxInfo)

	_ = tx.Commit()
	return retData, nil
}

// step 4 协议号：100352 bob签收这次转账
func (this *commitmentTxSignedManager) RevokeAndAcknowledgeCommitmentTransaction(msg bean.RequestMessage, signer *bean.User) (retData interface{}, needSignC2b bool, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		err = errors.New(enum.Tips_common_empty + "msg.data")
		log.Println(err)
		return nil, false, err
	}

	reqData := &bean.PayeeSendSignCommitmentTx{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	if tool.CheckIsString(&reqData.MsgHash) == false {
		err = errors.New(enum.Tips_common_wrong + "msg_hash")
		log.Println(err)
		return nil, false, err
	}

	tx, err := signer.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	defer tx.Rollback()

	//region 确认是给自己的信息
	message, err := messageService.getMsgUseTx(tx, reqData.MsgHash)
	if err != nil {
		return nil, false, errors.New(enum.Tips_common_invilidMsgHash)
	}
	if message.Receiver != signer.PeerId {
		return nil, false, errors.New(enum.Tips_rsmc_notTargetUser)
	}

	c2aDataJson := &bean.AliceRequestToCreateCommitmentTxOfP2p{}
	_ = json.Unmarshal([]byte(message.Data), c2aDataJson)
	reqData.MsgHash = c2aDataJson.CommitmentTxHash
	//endregion

	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id")
		log.Println(err)
		return nil, false, err
	}

	channelInfo := getChannelInfoByChannelId(tx, reqData.ChannelId, signer.PeerId)
	if channelInfo == nil {
		err = errors.New(enum.Tips_funding_notFoundChannelByChannelId + reqData.ChannelId)
		log.Println(err)
		return nil, false, err
	}

	if channelInfo.CurrState < dao.ChannelState_NewTx {
		return nil, false, errors.New("do not finish funding")
	}

	payeeRevokeAndAcknowledgeCommitment := &dao.PayeeRevokeAndAcknowledgeCommitment{}
	_ = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxHash", reqData.MsgHash)).First(payeeRevokeAndAcknowledgeCommitment)
	if payeeRevokeAndAcknowledgeCommitment.Id > 0 {
		return nil, false, errors.New(enum.Tips_rsmc_notDoItAgain)
	}

	//Make sure who creates the transaction, who will sign the transaction.
	//The default creator is Alice, and Bob is the signer.
	//While if ALice is the signer, then Bob creates the transaction.
	targetUser := channelInfo.PeerIdA
	if signer.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	} else {
		targetUser = channelInfo.PeerIdB
	}

	if targetUser != msg.RecipientUserPeerId {
		return nil, false, errors.New(enum.Tips_rsmc_notTargetUser)
	}

	toAliceP2pData := bean.PayeeSignCommitmentTxOfP2p{}
	toAliceP2pData.ChannelId = channelInfo.ChannelId
	toAliceP2pData.CommitmentTxHash = reqData.MsgHash
	toAliceP2pData.Approval = reqData.Approval

	if reqData.Approval == false {

		if channelInfo.CurrState == dao.ChannelState_NewTx {
			channelInfo.CurrState = dao.ChannelState_CanUse
			_ = tx.Update(channelInfo)
		}

		payeeRevokeAndAcknowledgeCommitment.ChannelId = toAliceP2pData.ChannelId
		payeeRevokeAndAcknowledgeCommitment.CommitmentTxHash = toAliceP2pData.CommitmentTxHash
		payeeRevokeAndAcknowledgeCommitment.Approval = false
		_ = tx.Save(payeeRevokeAndAcknowledgeCommitment)

		_ = messageService.updateMsgStateUseTx(tx, message)
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
		return retData, false, nil
	}

	this.operationFlag.Lock()
	defer this.operationFlag.Unlock()

	needSignData := bean.NeedBobSignRawDataForC2b{}
	needSignData.ChannelId = channelInfo.ChannelId

	currNodeChannelPubKey := channelInfo.PubKeyB
	if signer.PeerId == channelInfo.PeerIdA {
		currNodeChannelPubKey = channelInfo.PubKeyA
	}

	//for rsmc
	if _, err = getAddressFromPubKey(reqData.CurrTempAddressPubKey); err != nil {
		err = errors.New(enum.Tips_common_wrong + "curr_temp_address_pub_key")
		log.Println(err)
		return nil, false, err
	}

	//for br
	creatorLastTempAddressPrivateKey := c2aDataJson.LastTempAddressPrivateKey
	if tool.CheckIsString(&creatorLastTempAddressPrivateKey) == false {
		err = errors.New(enum.Tips_common_empty + "the starter's last temp address private key")
		log.Println(err)
		return nil, false, err
	}

	// get the funding transaction
	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, signer.PeerId)
	if fundingTransaction == nil {
		return nil, false, errors.New(enum.Tips_common_notFound + "fundingTransaction")
	}

	toAliceP2pData.LastTempAddressPrivateKey = reqData.LastTempAddressPrivateKey
	toAliceP2pData.CurrTempAddressPubKey = reqData.CurrTempAddressPubKey

	//region 1、验证签名后的C2a的Rsmc和ToBob hex
	var signedRsmcHex, aliceRsmcTxId, c2aRsmcMultiAddress, c2aRsmcRedeemScript, c2aRsmcMultiAddressScriptPubKey string
	var c2aRsmcOutputs []bean.TransactionInputItem
	if tool.CheckIsString(&c2aDataJson.RsmcRawData.Hex) {
		signedRsmcHex = reqData.C2aRsmcSignedHex
		if pass, _ := rpcClient.CheckMultiSign(true, signedRsmcHex, 2); pass == false {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "c2a_rsmc_signed_hex"))
		}

		if verifyCompleteSignHex(c2aDataJson.RsmcRawData.Inputs, signedRsmcHex) != nil {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "c2a_rsmc_signed_hex"))
		}

		aliceRsmcTxId = rpcClient.GetTxId(signedRsmcHex)

		// region 根据alice的临时地址+bob的通道address,获取alice2+bob的多签地址，并得到AliceSignedRsmcHex签名后的交易的input，为创建alice的RD和bob的BR做准备
		c2aRsmcMultiAddr, err := omnicore.CreateMultiSig(2, []string{c2aDataJson.CurrTempAddressPubKey, currNodeChannelPubKey})
		if err != nil {
			return nil, false, err
		}
		c2aRsmcMultiAddress = gjson.Get(c2aRsmcMultiAddr, "address").String()
		c2aRsmcRedeemScript = gjson.Get(c2aRsmcMultiAddr, "redeemScript").String()
		c2aRsmcMultiAddressScriptPubKey = gjson.Get(c2aRsmcMultiAddr, "scriptPubKey").String()

		c2aRsmcOutputs, err = getInputsForNextTxByParseTxHashVout(signedRsmcHex, c2aRsmcMultiAddress, c2aRsmcMultiAddressScriptPubKey, c2aRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
	}
	toAliceP2pData.C2aSignedRsmcHex = signedRsmcHex
	//endregion

	// region 2、签名 ToCounterpartyTxHex
	signedToCounterpartyTxHex := reqData.C2aCounterpartySignedHex
	if tool.CheckIsString(&c2aDataJson.CounterpartyRawData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedToCounterpartyTxHex, 2); pass == false {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "to_counterparty_tx_hex"))
		}

		if verifyCompleteSignHex(c2aDataJson.CounterpartyRawData.Inputs, signedToCounterpartyTxHex) != nil {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "c2a_rsmc_signed_hex"))
		}
	}
	toAliceP2pData.C2aSignedToCounterpartyTxHex = signedToCounterpartyTxHex
	//endregion

	//获取bob最新的承诺交易
	latestCommitmentTxInfo, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, signer.PeerId)
	isFirstRequest := false
	if latestCommitmentTxInfo != nil && latestCommitmentTxInfo.Id > 0 {
		if latestCommitmentTxInfo.TxType != dao.CommitmentTransactionType_Rsmc {
			return nil, false, errors.New(enum.Tips_rsmc_errorCommitmentTxType + strconv.Itoa(int(latestCommitmentTxInfo.TxType)))
		}

		if latestCommitmentTxInfo.CurrState != dao.TxInfoState_CreateAndSign && latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
			return nil, false, errors.New(enum.Tips_rsmc_errorCommitmentTxState + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		}

		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Create && latestCommitmentTxInfo.LastCommitmentTxId > 0 {
			lastCommitmentTx := &dao.CommitmentTransaction{}
			err = tx.Select(q.Eq("Id", latestCommitmentTxInfo.LastCommitmentTxId)).First(lastCommitmentTx)
			if err != nil {
				return nil, false, errors.New(enum.Tips_common_notFound + "lastCommitmentTx")
			}
			_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey)
			if err != nil {
				return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, reqData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey))
			}
		}

		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign { //有上一次的承诺交易
			_, err = tool.GetPubKeyFromWifAndCheck(reqData.LastTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey)
			if err != nil {
				return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, reqData.LastTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
			}
			isFirstRequest = true
		}
	} else { // 因为没有充值，没有最初的承诺交易C1b
		isFirstRequest = true
	}

	var c2bAmountToCounterparty = 0.0
	var c2bCommitmentId = 0
	//如果是本轮的第一次请求交易
	if isFirstRequest {
		//region 3、根据对方传过来的上一个交易的临时rsmc私钥，签名最近的BR交易，保证对方确实放弃了上一个承诺交易
		err := signLastBR(tx, dao.BRType_Rmsc, *channelInfo, signer.PeerId, c2aDataJson.LastTempAddressPrivateKey, latestCommitmentTxInfo.Id)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
		//endregion

		//region 4、创建C2b
		commitmentTxRequest := &bean.RequestCreateCommitmentTx{}
		commitmentTxRequest.ChannelId = channelInfo.ChannelId
		commitmentTxRequest.Amount = c2aDataJson.Amount
		commitmentTxRequest.CurrTempAddressIndex = reqData.CurrTempAddressIndex
		commitmentTxRequest.CurrTempAddressPubKey = reqData.CurrTempAddressPubKey
		newCommitmentTxInfo, rawTx, err := createCommitmentTxHex(tx, false, commitmentTxRequest, channelInfo, latestCommitmentTxInfo, *signer)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
		c2bAmountToCounterparty = newCommitmentTxInfo.AmountToCounterparty
		c2bCommitmentId = newCommitmentTxInfo.Id

		toAliceP2pData.C2bRsmcTxData = rawTx.RsmcRawTxData
		toAliceP2pData.C2bCounterpartyTxData = rawTx.ToCounterpartyRawTxData

		//endregion
	} else {
		if reqData.CurrTempAddressPubKey != latestCommitmentTxInfo.RSMCTempAddressPubKey {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, reqData.CurrTempAddressPubKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
		}

		c2bAmountToCounterparty = latestCommitmentTxInfo.AmountToCounterparty
		c2bCommitmentId = latestCommitmentTxInfo.Id

		rawTx := &dao.CommitmentTxRawTx{}
		tx.Select(q.Eq("CommitmentTxId", latestCommitmentTxInfo.Id)).First(rawTx)
		if rawTx.Id == 0 {
			return nil, false, errors.New("not found rawTx")
		}

		toAliceP2pData.C2bRsmcTxData = rawTx.RsmcRawTxData
		toAliceP2pData.C2bCounterpartyTxData = rawTx.ToCounterpartyRawTxData
	}

	needSignData.C2bRsmcRawData = toAliceP2pData.C2bRsmcTxData
	needSignData.C2bCounterpartyRawData = toAliceP2pData.C2bCounterpartyTxData

	//region 5、根据签名后的C2a Rsmc创建alice的RD create RD tx for alice
	c2aRdOutputAddress := channelInfo.AddressA
	if signer.PeerId == channelInfo.PeerIdA {
		c2aRdOutputAddress = channelInfo.AddressB
	}
	if len(c2aRsmcOutputs) > 0 {
		c2aRdTx, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
			c2aRsmcMultiAddress,
			c2aRsmcOutputs,
			c2aRdOutputAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			c2bAmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&c2aRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_failToCreate, "RD raw transacation"))
		}

		c2aRdRawData := bean.NeedClientSignTxData{}
		c2aRdRawData.Hex = c2aRdTx["hex"].(string)
		c2aRdRawData.Inputs = c2aRdTx["inputs"]
		c2aRdRawData.PubKeyA = c2aDataJson.CurrTempAddressPubKey
		c2aRdRawData.PubKeyB = currNodeChannelPubKey
		c2aRdRawData.IsMultisig = true
		toAliceP2pData.C2aRdTxData = c2aRdRawData

		needSignData.C2aRdRawData = c2aRdRawData
	}
	//endregion create RD tx for alice

	// region 6、根据alice的Rsmc，创建对应的BR,为下一个交易做准备，create BR2b tx  for bob
	if len(c2aRsmcOutputs) > 0 {
		var myAddress = channelInfo.AddressB
		if signer.PeerId == channelInfo.PeerIdA {
			myAddress = channelInfo.AddressA
		}
		senderCommitmentTx := &dao.CommitmentTransaction{}
		senderCommitmentTx.Id = c2bCommitmentId
		senderCommitmentTx.PropertyId = fundingTransaction.PropertyId
		senderCommitmentTx.RSMCTempAddressPubKey = c2aDataJson.CurrTempAddressPubKey
		senderCommitmentTx.RSMCMultiAddress = c2aRsmcMultiAddress
		senderCommitmentTx.RSMCRedeemScript = c2aRsmcRedeemScript
		senderCommitmentTx.RSMCMultiAddressScriptPubKey = c2aRsmcMultiAddressScriptPubKey
		senderCommitmentTx.RSMCTxHex = signedRsmcHex
		senderCommitmentTx.RSMCTxid = aliceRsmcTxId
		senderCommitmentTx.AmountToRSMC = c2bAmountToCounterparty
		c2aBrHexData, err := createCurrCommitmentTxRawBR(tx, dao.BRType_Rmsc, channelInfo, senderCommitmentTx, c2aRsmcOutputs, myAddress, *signer)
		if err != nil {
			log.Println(err)
			return nil, false, err
		}
		C2aBrRawData := bean.NeedClientSignRawBRTxData{}
		C2aBrRawData.Hex = c2aBrHexData["hex"].(string)
		C2aBrRawData.Inputs = c2aBrHexData["inputs"]
		C2aBrRawData.BrId = c2aBrHexData["br_id"].(int)
		C2aBrRawData.PubKeyA = c2aDataJson.CurrTempAddressPubKey
		C2aBrRawData.PubKeyB = currNodeChannelPubKey
		C2aBrRawData.IsMultisig = true
		needSignData.C2aBrRawData = C2aBrRawData
	}
	//endregion

	_ = messageService.updateMsgStateUseTx(tx, message)

	// 缓存数据
	cacheDataForTx := &dao.CacheDataForTx{}
	cacheDataForTx.KeyName = signer.PeerId + "_" + channelInfo.ChannelId
	_ = tx.Select(q.Eq("KeyName", cacheDataForTx.KeyName)).First(cacheDataForTx)
	if cacheDataForTx.Id != 0 {
		_ = tx.DeleteStruct(cacheDataForTx)
	}
	if cacheDataForTx.Id != 0 {
		_ = tx.DeleteStruct(cacheDataForTx)
	}
	bytes, _ := json.Marshal(&toAliceP2pData)
	cacheDataForTx.Data = bytes
	_ = tx.Save(cacheDataForTx)

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	return needSignData, true, err
}

// step 5 协议号：100361 完成后推送352协议 bob对C2b的Rsmc，toBob和c2a的Rd和C2a的Br进行签名
func (this *commitmentTxSignedManager) OnBobSignC2bTransactionAtBobSide(data string, user *bean.User) (toBob, retData interface{}, err error) {
	if tool.CheckIsString(&data) == false {
		err = errors.New(enum.Tips_common_empty + "msg.data")
		log.Println(err)
		return nil, nil, err
	}
	signedDataForC2b := bean.BobSignedRsmcDataForC2b{}
	_ = json.Unmarshal([]byte(data), &signedDataForC2b)

	if tool.CheckIsString(&signedDataForC2b.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id")
		log.Println(err)
		return nil, nil, err
	}

	//得到step4缓存的数据
	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	cacheDataForTx := &dao.CacheDataForTx{}
	tx.Select(q.Eq("KeyName", user.PeerId+"_"+signedDataForC2b.ChannelId)).First(cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}
	p2pData := &bean.PayeeSignCommitmentTxOfP2p{}
	_ = json.Unmarshal(cacheDataForTx.Data, p2pData)

	if len(p2pData.ChannelId) == 0 {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}

	if tool.CheckIsString(&p2pData.C2bRsmcTxData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedDataForC2b.C2bRsmcSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c2b_rsmc_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&p2pData.C2bCounterpartyTxData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedDataForC2b.C2bCounterpartySignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c2b_counterparty_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&p2pData.C2aRdTxData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(false, signedDataForC2b.C2aRdSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c2a_rd_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&signedDataForC2b.C2aBrSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(false, signedDataForC2b.C2aBrSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "c2a_br_signed_hex")
			log.Println(err)
			return nil, nil, err
		}

		if signedDataForC2b.C2aBrId == 0 {
			err = errors.New(enum.Tips_common_wrong + "c2a_br_id")
			log.Println(err)
			return nil, nil, err
		}
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, signedDataForC2b.ChannelId, user.PeerId)
	if err != nil {
		return nil, nil, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
	}

	if len(signedDataForC2b.C2bRsmcSignedHex) > 0 {
		latestCommitmentTxInfo.RSMCTxHex = signedDataForC2b.C2bRsmcSignedHex
		latestCommitmentTxInfo.RSMCTxid = rpcClient.GetTxId(signedDataForC2b.C2bRsmcSignedHex)
	}

	if len(signedDataForC2b.C2bCounterpartySignedHex) > 0 {
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedDataForC2b.C2bCounterpartySignedHex
		latestCommitmentTxInfo.ToCounterpartyTxid = rpcClient.GetTxId(signedDataForC2b.C2bCounterpartySignedHex)
	}

	if len(signedDataForC2b.C2aBrSignedHex) > 0 {
		err = updateCurrCommitmentTxRawBR(tx, signedDataForC2b.C2aBrId, signedDataForC2b.C2aBrSignedHex, *user)
		if err != nil {
			return nil, nil, err
		}
	}
	_ = tx.Update(latestCommitmentTxInfo)

	_ = tx.Commit()

	p2pData.C2bRsmcTxData.Hex = signedDataForC2b.C2bRsmcSignedHex
	p2pData.C2bCounterpartyTxData.Hex = signedDataForC2b.C2bCounterpartySignedHex
	p2pData.C2aRdTxData.Hex = signedDataForC2b.C2aRdSignedHex

	toBobData := bean.BobSignedRsmcDataForC2bResult{}
	toBobData.ChannelId = p2pData.ChannelId
	toBobData.CommitmentTxHash = p2pData.CommitmentTxHash
	toBobData.Approval = p2pData.Approval
	return toBobData, p2pData, nil
}

// step 9 协议号：响应353 obd节点接收到alice的二次签名信息 推送110353信息给bob
func (this *commitmentTxSignedManager) OnGetAliceSignC2bTransactionAtBobSide(data string, user *bean.User) (retData bean.NeedBobSignRdTxForC2b, err error) {
	aliceSignedC2bTxDataP2p := bean.AliceSignedC2bTxDataP2p{}
	err = json.Unmarshal([]byte(data), &aliceSignedC2bTxDataP2p)
	if err != nil {
		return retData, err
	}

	cacheDataForTx := &dao.CacheDataForTx{}
	user.Db.Select(q.Eq("KeyName", user.PeerId+"_"+aliceSignedC2bTxDataP2p.ChannelId)).First(cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		return retData, errors.New(enum.Tips_common_wrong + "channel_id")
	}
	cacheDataFrom352 := &bean.PayeeSignCommitmentTxOfP2p{}
	_ = json.Unmarshal(cacheDataForTx.Data, cacheDataFrom352)

	if tool.CheckIsString(&cacheDataFrom352.C2bRsmcTxData.Hex) {
		if verifyCompleteSignHex(cacheDataFrom352.C2bRsmcTxData.Inputs, aliceSignedC2bTxDataP2p.C2bRsmcSignedHex) != nil {
			return retData, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "c2b_rsmc_signed_hex"))
		}
	}

	if tool.CheckIsString(&cacheDataFrom352.C2bCounterpartyTxData.Hex) {
		if verifyCompleteSignHex(cacheDataFrom352.C2bCounterpartyTxData.Inputs, aliceSignedC2bTxDataP2p.C2bCounterpartySignedHex) != nil {
			return retData, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "c2b_signed_to_counterparty_tx_hex"))
		}
	}

	cacheDataForTx = &dao.CacheDataForTx{}
	cacheDataForTx.KeyName = user.PeerId + "_353_" + aliceSignedC2bTxDataP2p.ChannelId
	_ = user.Db.Select(q.Eq("KeyName", cacheDataForTx.KeyName)).First(cacheDataForTx)
	if cacheDataForTx.Id != 0 {
		_ = user.Db.DeleteStruct(cacheDataForTx)
	}
	if cacheDataForTx.Id != 0 {
		_ = user.Db.DeleteStruct(cacheDataForTx)
	}
	bytes, _ := json.Marshal(&aliceSignedC2bTxDataP2p)
	cacheDataForTx.Data = bytes
	_ = user.Db.Save(cacheDataForTx)

	needBobSignRdTxForC2b := bean.NeedBobSignRdTxForC2b{}
	needBobSignRdTxForC2b.ChannelId = aliceSignedC2bTxDataP2p.ChannelId
	needBobSignRdTxForC2b.C2bRdPartialData = aliceSignedC2bTxDataP2p.C2bRdPartialData
	return needBobSignRdTxForC2b, nil
}

// step 10 协议号：100364 响应bob的才c2b的Rd的签名
func (this *commitmentTxSignedManager) BobSignC2b_RdAtBobSide(data string, user *bean.User) (retData interface{}, err error) {
	bobSignedRdTxForC2b := bean.BobSignedRdTxForC2b{}
	_ = json.Unmarshal([]byte(data), &bobSignedRdTxForC2b)

	cacheDataForTx := &dao.CacheDataForTx{}
	user.Db.Select(q.Eq("KeyName", user.PeerId+"_353_"+bobSignedRdTxForC2b.ChannelId)).First(cacheDataForTx)
	if cacheDataForTx.Id == 0 {
		return nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}
	dataFrom353 := &bean.AliceSignedC2bTxDataP2p{}
	_ = json.Unmarshal(cacheDataForTx.Data, dataFrom353)
	if len(dataFrom353.ChannelId) == 0 {
		return nil, errors.New(enum.Tips_common_empty + "channel_id")
	}

	if tool.CheckIsString(&dataFrom353.C2bRdPartialData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(false, bobSignedRdTxForC2b.C2bRdSignedHex, 2); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c2b_rd_signed_hex")
			log.Println(err)
			return nil, err
		}

		if tool.CheckIsString(&dataFrom353.C2bRdPartialData.Hex) {
			if verifyCompleteSignHex(dataFrom353.C2bRdPartialData.Inputs, bobSignedRdTxForC2b.C2bRdSignedHex) != nil {
				return retData, errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "c2b_rd_signed_hex"))
			}
		}
	}

	var channelId = dataFrom353.ChannelId
	var signedRsmcHex = dataFrom353.C2bRsmcSignedHex
	var signedToCounterpartyTxHex = dataFrom353.C2bCounterpartySignedHex
	var c2bSignedRdHex = bobSignedRdTxForC2b.C2bRdSignedHex
	var c2aInitTxHash = dataFrom353.C2aCommitmentTxHash

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, channelId, user.PeerId)
	if channelInfo == nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find payer's commitmentTxInfo")
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, err
	}

	myChannelAddress := channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdA {
		myChannelAddress = channelInfo.AddressA
	}

	if tool.CheckIsString(&signedRsmcHex) {
		latestCommitmentTxInfo.RSMCTxHex = signedRsmcHex
		latestCommitmentTxInfo.RSMCTxid = rpcClient.GetTxId(signedRsmcHex)
		err = saveRdTx(tx, channelInfo, signedRsmcHex, c2bSignedRdHex, latestCommitmentTxInfo, myChannelAddress, user)
		if err != nil {
			return nil, err
		}
	}

	if tool.CheckIsString(&signedToCounterpartyTxHex) {
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedToCounterpartyTxHex
		latestCommitmentTxInfo.ToCounterpartyTxid = rpcClient.GetTxId(signedToCounterpartyTxHex)
	}

	latestCommitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	latestCommitmentTxInfo.SignAt = time.Now()

	bytes, err := json.Marshal(latestCommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	latestCommitmentTxInfo.CurrHash = msgHash
	_ = tx.Update(latestCommitmentTxInfo)

	lastCommitmentTxInfo := dao.CommitmentTransaction{}
	_ = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, &lastCommitmentTxInfo)
	if lastCommitmentTxInfo.Id > 0 {
		lastCommitmentTxInfo.CurrState = dao.TxInfoState_Abord
		_ = tx.Update(lastCommitmentTxInfo)
	}

	payeeRevokeAndAcknowledgeCommitment := &dao.PayeeRevokeAndAcknowledgeCommitment{}
	payeeRevokeAndAcknowledgeCommitment.ChannelId = channelId
	payeeRevokeAndAcknowledgeCommitment.CommitmentTxHash = c2aInitTxHash
	payeeRevokeAndAcknowledgeCommitment.Approval = true
	_ = tx.Save(payeeRevokeAndAcknowledgeCommitment)

	channelInfo.CurrState = dao.ChannelState_CanUse
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	//同步通道信息到tracker
	sendChannelStateToTracker(*channelInfo, *latestCommitmentTxInfo)
	log.Println("end rsmc step 10", time.Now())
	return latestCommitmentTxInfo, nil
}
