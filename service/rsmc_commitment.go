package service

import (
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"obd/bean"
	"obd/dao"
	"obd/rpc"
	"obd/tool"
	"strconv"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

type commitmentTxManager struct {
	operationFlag sync.Mutex
}

var CommitmentTxService commitmentTxManager

func (this *commitmentTxManager) CommitmentTransactionCreated(msg bean.RequestMessage, creator *bean.User) (retData map[string]interface{}, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, errors.New("empty json reqData")
	}
	reqData := &bean.CommitmentTx{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		return nil, err
	}
	if tool.CheckIsString(&reqData.ChannelId) == false {
		return nil, errors.New("wrong channel_id")
	}

	if tool.CheckIsString(&reqData.LastTempAddressPrivateKey) == false {
		return nil, errors.New("wrong LastTempAddressPrivateKey")
	}

	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		return nil, errors.New("wrong ChannelAddressPrivateKey")
	}

	if reqData.Amount <= 0 {
		return nil, errors.New("wrong payment amount")
	}

	tx, err := creator.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, reqData.ChannelId, creator.PeerId)
	if channelInfo == nil {
		err = errors.New("not found channel " + reqData.ChannelId)
		log.Println(err)
		return nil, err
	}

	err = checkBtcFundFinish(channelInfo.ChannelAddress)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	senderPubKey := channelInfo.PubKeyA
	targetUser := channelInfo.PeerIdB
	if creator.PeerId == channelInfo.PeerIdB {
		senderPubKey = channelInfo.PubKeyB
		targetUser = channelInfo.PeerIdA
	}

	if targetUser != msg.RecipientPeerId {
		return nil, errors.New("error target user " + msg.RecipientPeerId)
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, creator.PeerId)
	if err != nil {
		return nil, errors.New("not find the lastCommitmentTxInfo")
	}

	if latestCommitmentTxInfo.TxType != dao.CommitmentTransactionType_Rsmc {
		return nil, errors.New("not find the lastCommitmentTxInfo at state " + strconv.Itoa(dao.CommitmentTransactionType_Rsmc))
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_CreateAndSign &&
		latestCommitmentTxInfo.CurrState != dao.TxInfoState_RsmcCreate {
		return nil, errors.New("not find the lastCommitmentTxInfo")
	}

	//region check input data 检测输入数据
	//如果是第一次发起请求

	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign {
		balance := latestCommitmentTxInfo.AmountToRSMC
		if balance < 0 {
			return nil, errors.New("not enough balance")
		}
		if balance < reqData.Amount {
			return nil, errors.New("not enough payment amount")
		}
		if _, err := tool.GetPubKeyFromWifAndCheck(reqData.LastTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey); err != nil {
			return nil, errors.New(reqData.LastTempAddressPrivateKey + " is wrong private key for the last RSMCTempAddressPubKey")
		}

	} else {
		lastCommitmentTx := &dao.CommitmentTransaction{}
		_ = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, lastCommitmentTx)
		if _, err := tool.GetPubKeyFromWifAndCheck(reqData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey); err != nil {
			return nil, errors.New(reqData.LastTempAddressPrivateKey + " is wrong private key for the last RSMCTempAddressPubKey")
		}
	}

	if _, err := tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, senderPubKey); err != nil {
		return nil, errors.New(reqData.ChannelAddressPrivateKey + " is wrong private key for the funder address")
	}
	tempAddrPrivateKeyMap[senderPubKey] = reqData.ChannelAddressPrivateKey

	if _, err := getAddressFromPubKey(reqData.CurrTempAddressPubKey); err != nil {
		return nil, errors.New("wrong CurrTempAddressPubKey")
	}
	if tool.CheckIsString(&reqData.CurrTempAddressPrivateKey) == false {
		return nil, errors.New("wrong CurrTempAddressPrivateKey")
	}
	if _, err := tool.GetPubKeyFromWifAndCheck(reqData.CurrTempAddressPrivateKey, reqData.CurrTempAddressPubKey); err != nil {
		return nil, errors.New(reqData.CurrTempAddressPrivateKey + " and " + reqData.CurrTempAddressPubKey + " not the pair key")
	}
	tempAddrPrivateKeyMap[reqData.CurrTempAddressPubKey] = reqData.CurrTempAddressPrivateKey
	//endregion

	retData = make(map[string]interface{})
	retData["channelId"] = channelInfo.ChannelId
	retData["amount"] = reqData.Amount
	retData["lastTempAddressPrivateKey"] = reqData.LastTempAddressPrivateKey
	retData["currTempAddressPubKey"] = reqData.CurrTempAddressPubKey
	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign {
		//创建c2a omni的交易不能一个输入，多个输出，所以就是两个交易
		newCommitmentTxInfo, err := createCommitmentTxHex(tx, true, reqData, channelInfo, latestCommitmentTxInfo, *creator)
		if err != nil {
			return nil, err
		}
		retData["rsmcHex"] = newCommitmentTxInfo.RSMCTxHex
		retData["toOtherHex"] = newCommitmentTxInfo.ToOtherTxHex
		retData["commitmentHash"] = newCommitmentTxInfo.CurrHash
	} else {
		retData["rsmcHex"] = latestCommitmentTxInfo.RSMCTxHex
		retData["toOtherHex"] = latestCommitmentTxInfo.ToOtherTxHex
		retData["commitmentHash"] = latestCommitmentTxInfo.CurrHash
	}
	_ = tx.Commit()

	return retData, err
}

//353 352的请求阶段完成，需要Alice这边签名C2b等相关的交易
func (this *commitmentTxManager) AfterBobSignCommitmentTranctionAtAliceSide(data string, user *bean.User) (retData map[string]interface{}, err error) {
	jsonObj := gjson.Parse(data)

	//region 检测传入数据
	var channelId = jsonObj.Get("channelId").String()
	if tool.CheckIsString(&channelId) == false {
		err = errors.New("wrong channelId")
		log.Println(err)
		return nil, err
	}
	var commitmentTxHash = jsonObj.Get("commitmentTxHash").String()
	if tool.CheckIsString(&commitmentTxHash) == false {
		err = errors.New("wrong commitmentTxHash")
		log.Println(err)
		return nil, err
	}

	var signedRsmcHex = jsonObj.Get("signedRsmcHex").String()
	if tool.CheckIsString(&signedRsmcHex) == false {
		err = errors.New("wrong signedRsmcHex")
		log.Println(err)
		return nil, err
	}
	_, err = rpcClient.TestMemPoolAccept(signedRsmcHex)
	if err != nil {
		err = errors.New("wrong signedRsmcHex")
		log.Println(err)
		return nil, err
	}

	var signedToOtherHex = jsonObj.Get("signedToOtherHex").String()
	if tool.CheckIsString(&signedToOtherHex) == false {
		err = errors.New("wrong signedToOtherHex")
		log.Println(err)
		return nil, err
	}
	_, err = rpcClient.TestMemPoolAccept(signedToOtherHex)
	if err != nil {
		err = errors.New("wrong signedToOtherHex")
		log.Println(err)
		return nil, err
	}

	var aliceRdHex = jsonObj.Get("senderRdHex").String()
	if tool.CheckIsString(&aliceRdHex) == false {
		err = errors.New("wrong senderRdHex")
		log.Println(err)
		return nil, err
	}
	_, err = rpcClient.TestMemPoolAccept(aliceRdHex)
	if err != nil {
		err = errors.New("wrong senderRdHex")
		log.Println(err)
		return nil, err
	}

	var bobRsmcHex = jsonObj.Get("rsmcHex").String()
	if tool.CheckIsString(&bobRsmcHex) == false {
		err = errors.New("wrong rsmcHex")
		log.Println(err)
		return nil, err
	}
	_, err = rpcClient.TestMemPoolAccept(bobRsmcHex)
	if err != nil {
		err = errors.New("wrong rsmcHex")
		log.Println(err)
		return nil, err
	}

	var bobCurrTempAddressPubKey = jsonObj.Get("currTempAddressPubKey").String()
	if tool.CheckIsString(&bobCurrTempAddressPubKey) == false {
		err = errors.New("wrong currTempAddressPubKey")
		log.Println(err)
		return nil, err
	}
	var bobToOtherHex = jsonObj.Get("toOtherHex").String()
	if tool.CheckIsString(&bobToOtherHex) == false {
		err = errors.New("wrong toOtherHex")
		log.Println(err)
		return nil, err
	}
	var bobLastTempAddressPrivateKey = jsonObj.Get("lastTempAddressPrivateKey").String()
	//endregion

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	retData = make(map[string]interface{})

	channelInfo := getChannelInfoByChannelId(tx, channelId, user.PeerId)
	if channelInfo == nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}

	fundingTransaction := getFundingTransactionByChannelId(tx, channelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, errors.New("not found fundingTransaction at targetSide")
	}

	latestCcommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find sender's commitmentTxInfo")
		log.Println(err)
		return nil, err
	}

	if latestCcommitmentTxInfo.CurrHash != commitmentTxHash {
		err = errors.New("wrong request hash")
		log.Println(err)
		return nil, err
	}

	if latestCcommitmentTxInfo.CurrState != dao.TxInfoState_RsmcCreate {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCcommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, err
	}

	var myChannelPubKey = channelInfo.PubKeyA
	var myChannelAddress = channelInfo.AddressA
	var partnerChannelAddress = channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdB {
		myChannelAddress = channelInfo.AddressB
		myChannelPubKey = channelInfo.PubKeyB
		partnerChannelAddress = channelInfo.AddressA
	}
	var myChannelPrivateKey = tempAddrPrivateKeyMap[myChannelPubKey]

	//region 根据对方传过来的上一个交易的临时rsmc私钥，签名上一次的BR交易，保证对方确实放弃了上一个承诺交易
	err = signLastBR(tx, *channelInfo, user.PeerId, bobLastTempAddressPrivateKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion

	// region 对自己的RD 二次签名
	err = signRdTx(tx, channelInfo, signedRsmcHex, aliceRdHex, *latestCcommitmentTxInfo, myChannelAddress, user)
	if err != nil {
		return nil, err
	}
	// endregion

	//更新alice的当前承诺交易
	latestCcommitmentTxInfo.SignAt = time.Now()
	latestCcommitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	latestCcommitmentTxInfo.RSMCTxHex = signedRsmcHex
	latestCcommitmentTxInfo.ToOtherTxHex = signedToOtherHex
	bytes, err := json.Marshal(latestCcommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	latestCcommitmentTxInfo.CurrHash = msgHash
	_ = tx.Update(latestCcommitmentTxInfo)

	//处理对方的数据
	//签名对方传过来的rsmcHex
	bobRsmcTxid, bobSignedRsmcHex, err := rpcClient.BtcSignRawTransaction(bobRsmcHex, myChannelPrivateKey)
	if err != nil {
		return nil, errors.New("fail to sign rsmc hex ")
	}
	testResult, err := rpcClient.TestMemPoolAccept(bobSignedRsmcHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}
	err = checkBobRemcData(bobSignedRsmcHex, latestCcommitmentTxInfo)
	if err != nil {
		return nil, err
	}
	retData["signedRsmcHex"] = bobSignedRsmcHex

	//region create RD tx for bob
	bobMultiAddr, err := rpcClient.CreateMultiSig(2, []string{bobCurrTempAddressPubKey, myChannelPubKey})
	if err != nil {
		return nil, err
	}
	bobRsmcMultiAddress := gjson.Get(bobMultiAddr, "address").String()
	bobRsmcRedeemScript := gjson.Get(bobMultiAddr, "redeemScript").String()
	addressJson, err := rpcClient.GetAddressInfo(bobRsmcMultiAddress)
	if err != nil {
		return nil, err
	}
	bobRsmcMultiAddressScriptPubKey := gjson.Get(addressJson, "scriptPubKey").String()

	inputs, err := getInputsForNextTxByParseTxHashVout(bobSignedRsmcHex, bobRsmcMultiAddress, bobRsmcMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, bobRdhex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		bobRsmcMultiAddress,
		[]string{
			myChannelPrivateKey,
		},
		inputs,
		partnerChannelAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		latestCcommitmentTxInfo.AmountToOther,
		0,
		1000,
		&bobRsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, errors.New("fail to create rd")
	}
	retData["rdHex"] = bobRdhex
	//endregion create RD tx for alice

	//region 根据对对方的Rsmc签名，生成惩罚对方，自己获益BR
	bobCommitmentTx := &dao.CommitmentTransaction{}
	bobCommitmentTx.Id = -1
	bobCommitmentTx.PropertyId = channelInfo.PropertyId
	bobCommitmentTx.RSMCMultiAddress = bobRsmcMultiAddress
	bobCommitmentTx.RSMCRedeemScript = bobRsmcRedeemScript
	bobCommitmentTx.RSMCTxHex = bobSignedRsmcHex
	bobCommitmentTx.RSMCTxid = bobRsmcTxid
	bobCommitmentTx.AmountToRSMC = latestCcommitmentTxInfo.AmountToOther
	err = createCurrBR(tx, channelInfo, bobCommitmentTx, inputs, myChannelAddress, fundingTransaction.FunderAddress, myChannelPrivateKey, *user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//endregion

	//签名对方传过来的toOtherHex
	_, bobSignedToOtherHex, err := rpcClient.BtcSignRawTransaction(bobToOtherHex, myChannelPrivateKey)
	if err != nil {
		return nil, errors.New("fail to sign toOther hex ")
	}
	testResult, err = rpcClient.TestMemPoolAccept(bobSignedToOtherHex)
	if err != nil {
		return nil, err
	}
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}
	retData["signedToOtherHex"] = bobSignedToOtherHex

	_ = tx.Commit()

	retData["channelId"] = channelId

	return retData, nil
}

func checkBobRemcData(rsmcHex string, commitmentTransaction *dao.CommitmentTransaction) error {
	result, err := rpcClient.OmniDecodeTransaction(rsmcHex)
	if err != nil {
		return errors.New("error rsmcHex")
	}
	parse := gjson.Parse(result)
	if parse.Get("propertyid").Int() != commitmentTransaction.PropertyId {
		return errors.New("error propertyid in rsmcHex ")
	}
	if parse.Get("amount").Float() != commitmentTransaction.AmountToOther {
		return errors.New("error amount in rsmcHex ")
	}
	return nil
}

func createCommitmentTxHex(dbTx storm.Node, isSender bool, reqData *bean.CommitmentTx, channelInfo *dao.ChannelInfo, lastCommitmentTx *dao.CommitmentTransaction, currUser bean.User) (commitmentTxInfo *dao.CommitmentTransaction, err error) {
	//1、转账给bob的交易：输入：通道其中一个input，输出：给bob
	//2、转账后的余额的交易：输入：通道总的一个input,输出：一个多签地址，这个钱又需要后续的RD才能赎回
	// create Cna tx
	fundingTransaction := getFundingTransactionByChannelId(dbTx, channelInfo.ChannelId, currUser.PeerId)
	if fundingTransaction == nil {
		err = errors.New("not found fundingTransaction")
		return nil, err
	}

	var outputBean = commitmentOutputBean{}
	outputBean.RsmcTempPubKey = reqData.CurrTempAddressPubKey
	if currUser.PeerId == channelInfo.PeerIdA {
		//default alice transfer to bob ,then alice minus money
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(reqData.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(reqData.Amount)).Float64()
		outputBean.OppositeSideChannelAddress = channelInfo.AddressB
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
	} else {
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(reqData.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(reqData.Amount)).Float64()
		outputBean.OppositeSideChannelAddress = channelInfo.AddressA
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
	}

	if lastCommitmentTx != nil && lastCommitmentTx.Id > 0 {
		if isSender {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(reqData.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Add(decimal.NewFromFloat(reqData.Amount)).Float64()
		} else {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(reqData.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Sub(decimal.NewFromFloat(reqData.Amount)).Float64()
		}
	}

	commitmentTxInfo, err = createCommitmentTx(currUser.PeerId, channelInfo, fundingTransaction, outputBean, &currUser)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Rsmc

	usedTxidTemp := ""
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUseSingleInput(
			int(commitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			[]string{
				reqData.ChannelAddressPrivateKey,
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

		usedTxidTemp = usedTxid
		commitmentTxInfo.RsmcInputTxid = usedTxid
		commitmentTxInfo.RSMCTxid = txid
		commitmentTxInfo.RSMCTxHex = hex
	}

	//create to other tx
	if commitmentTxInfo.AmountToOther > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseRestInput(
			int(commitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			usedTxidTemp,
			[]string{
				reqData.ChannelAddressPrivateKey,
			},
			outputBean.OppositeSideChannelAddress,
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
		commitmentTxInfo.ToOtherTxHex = hex
	}
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	if lastCommitmentTx != nil && lastCommitmentTx.Id > 0 {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentTx.Id
		commitmentTxInfo.LastHash = lastCommitmentTx.CurrHash
	}
	commitmentTxInfo.CurrState = dao.TxInfoState_RsmcCreate
	err = dbTx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = dbTx.Update(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, nil
}

type commitmentTxSignedManager struct {
	operationFlag sync.Mutex
}

var CommitmentTxSignedService commitmentTxSignedManager

func (this *commitmentTxSignedManager) BeforeBobSignCommitmentTranctionAtBobSide(data string, user *bean.User) (retData map[string]interface{}, err error) {

	jsonObj := gjson.Parse(data)
	retData = make(map[string]interface{})
	retData["channelId"] = jsonObj.Get("channelId").String()
	retData["amount"] = jsonObj.Get("amount").Float()
	retData["rsmcHex"] = jsonObj.Get("rsmcHex").String()
	retData["toOtherHex"] = jsonObj.Get("toOtherHex").String()

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, jsonObj.Get("channelId").String(), user.PeerId)
	if channelInfo == nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}

	senderPeerId := channelInfo.PeerIdA
	if user.PeerId == channelInfo.PeerIdA {
		senderPeerId = channelInfo.PeerIdB
	}
	messageHash := MessageService.saveMsgUseTx(tx, senderPeerId, user.PeerId, data)
	retData["msgHash"] = messageHash
	_ = tx.Commit()

	return retData, nil
}

func (this *commitmentTxSignedManager) RevokeAndAcknowledgeCommitmentTransaction(jsonData string, signer *bean.User) (retData map[string]interface{}, targetUser string, err error) {
	if tool.CheckIsString(&jsonData) == false {
		err := errors.New("empty json reqData")
		log.Println(err)
		return nil, "", err
	}

	reqData := &bean.CommitmentTxSigned{}
	err = json.Unmarshal([]byte(jsonData), reqData)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	if tool.CheckIsString(&reqData.RequestCommitmentHash) == false {
		err = errors.New("wrong RequestCommitmentHash")
		log.Println(err)
		return nil, "", err
	}

	tx, err := signer.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer tx.Rollback()
	//region 确认是给自己的信息
	message, err := MessageService.getMsgUseTx(tx, reqData.RequestCommitmentHash)
	if err != nil {
		return nil, "", errors.New("wrong request_hash")
	}
	if message.Receiver != signer.PeerId {
		return nil, "", errors.New("you are not the operator")
	}
	aliceDataJson := gjson.Parse(message.Data)
	reqData.RequestCommitmentHash = aliceDataJson.Get("commitmentHash").String()
	//endregion

	if tool.CheckIsString(&reqData.ChannelId) == false {
		err = errors.New("wrong ChannelId")
		log.Println(err)
		return nil, "", err
	}

	channelInfo := getChannelInfoByChannelId(tx, reqData.ChannelId, signer.PeerId)
	if channelInfo == nil {
		err = errors.New("not found the channel " + reqData.ChannelId)
		log.Println(err)
		return nil, "", err
	}
	err = checkBtcFundFinish(channelInfo.ChannelAddress)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	//Make sure who creates the transaction, who will sign the transaction.
	//The default creator is Alice, and Bob is the signer.
	//While if ALice is the signer, then Bob creates the transaction.
	targetUser = channelInfo.PeerIdA
	if signer.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	} else {
		targetUser = channelInfo.PeerIdB
	}

	if reqData.Approval == false {
		return nil, targetUser, errors.New("signer disagree transaction")
	}

	this.operationFlag.Lock()
	defer this.operationFlag.Unlock()

	//for c rd br
	if tool.CheckIsString(&reqData.ChannelAddressPrivateKey) == false {
		err = errors.New("fail to get the signer's channel address private key")
		log.Println(err)
		return nil, "", err
	}

	currNodeChannelPubKey := channelInfo.PubKeyB
	if signer.PeerId == channelInfo.PeerIdA {
		currNodeChannelPubKey = channelInfo.PubKeyA
	}

	if _, err := tool.GetPubKeyFromWifAndCheck(reqData.ChannelAddressPrivateKey, currNodeChannelPubKey); err != nil {
		return nil, "", errors.New(reqData.ChannelAddressPrivateKey + " is wrong private key for the fund address")
	}
	tempAddrPrivateKeyMap[currNodeChannelPubKey] = reqData.ChannelAddressPrivateKey

	//for rsmc
	if _, err := getAddressFromPubKey(reqData.CurrTempAddressPubKey); err != nil {
		err = errors.New("fail to get the signer's curr temp address pub key")
		log.Println(err)
		return nil, "", err
	}
	//for rsmc
	if tool.CheckIsString(&reqData.CurrTempAddressPrivateKey) == false {
		err = errors.New("fail to get the signer's curr temp address private key")
		log.Println(err)
		return nil, "", err
	}
	if _, err := tool.GetPubKeyFromWifAndCheck(reqData.CurrTempAddressPrivateKey, reqData.CurrTempAddressPubKey); err != nil {
		return nil, "", errors.New(reqData.CurrTempAddressPrivateKey + " and " + reqData.CurrTempAddressPubKey + " not the pair key")
	}
	tempAddrPrivateKeyMap[reqData.CurrTempAddressPubKey] = reqData.CurrTempAddressPrivateKey

	//for br
	creatorLastTempAddressPrivateKey := aliceDataJson.Get("lastTempAddressPrivateKey").String()
	if tool.CheckIsString(&creatorLastTempAddressPrivateKey) == false {
		err = errors.New("fail to get the starter's last temp address  private key")
		log.Println(err)
		return nil, targetUser, err
	}

	// get the funding transaction
	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, signer.PeerId)
	if fundingTransaction == nil {
		return nil, "", errors.New("not found fundingTransaction")
	}

	retData = make(map[string]interface{})

	retData["channelId"] = channelInfo.ChannelId
	retData["commitmentTxHash"] = reqData.RequestCommitmentHash
	retData["lastTempAddressPrivateKey"] = reqData.LastTempAddressPrivateKey
	retData["currTempAddressPubKey"] = reqData.CurrTempAddressPubKey

	//region 1、签名对方传过来的rsmcHex
	rsmcTxId, signedRsmcHex, err := rpcClient.BtcSignRawTransaction(aliceDataJson.Get("rsmcHex").String(), reqData.ChannelAddressPrivateKey)
	if err != nil {
		return nil, targetUser, errors.New("fail to sign rsmc hex ")
	}
	testResult, err := rpcClient.TestMemPoolAccept(signedRsmcHex)
	if err != nil {
		return nil, targetUser, err
	}
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, targetUser, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}

	// region 根据alice的临时地址+bob的通道address,获取alice2+bob的多签地址，并得到AliceSignedRsmcHex签名后的交易的input，为创建alice的RD和bob的BR做准备
	aliceMultiAddr, err := rpcClient.CreateMultiSig(2, []string{aliceDataJson.Get("currTempAddressPubKey").String(), currNodeChannelPubKey})
	if err != nil {
		return nil, "", err
	}
	aliceRsmcMultiAddress := gjson.Get(aliceMultiAddr, "address").String()
	aliceRsmcRedeemScript := gjson.Get(aliceMultiAddr, "redeemScript").String()
	tempJson, err := rpcClient.GetAddressInfo(aliceRsmcMultiAddress)
	if err != nil {
		return nil, "", err
	}
	aliceRsmcMultiAddressScriptPubKey := gjson.Get(tempJson, "scriptPubKey").String()

	inputs, err := getInputsForNextTxByParseTxHashVout(signedRsmcHex, aliceRsmcMultiAddress, aliceRsmcMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	retData["signedRsmcHex"] = signedRsmcHex
	//endregion

	//endregion

	// region 2、签名对方传过来的toOtherHex
	_, signedToOtherHex, err := rpcClient.BtcSignRawTransaction(aliceDataJson.Get("toOtherHex").String(), reqData.ChannelAddressPrivateKey)
	if err != nil {
		return nil, targetUser, errors.New("fail to sign toOther hex ")
	}
	testResult, err = rpcClient.TestMemPoolAccept(signedToOtherHex)
	if err != nil {
		return nil, targetUser, err
	}
	if gjson.Parse(testResult).Array()[0].Get("allowed").Bool() == false {
		return nil, targetUser, errors.New(gjson.Parse(testResult).Array()[0].Get("reject-reason").String())
	}
	retData["signedToOtherHex"] = signedToOtherHex
	//endregion

	//获取bob最新的承诺交易
	latestCommitmentTxInfo, _ := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, signer.PeerId)
	isFirstRequest := false
	if latestCommitmentTxInfo != nil && latestCommitmentTxInfo.Id > 0 {
		if latestCommitmentTxInfo.TxType != dao.CommitmentTransactionType_Rsmc {
			return nil, "", errors.New("wrong commitment tx type " + strconv.Itoa(int(latestCommitmentTxInfo.TxType)))
		}

		if latestCommitmentTxInfo.CurrState != dao.TxInfoState_CreateAndSign && latestCommitmentTxInfo.CurrState != dao.TxInfoState_RsmcCreate {
			return nil, "", errors.New("wrong commitment tx state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		}

		if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign { //有上一次的承诺交易
			isFirstRequest = true
		}
	} else { // 因为没有充值，没有最初的承诺交易C1b
		isFirstRequest = true
	}

	var amountToOther = 0.0
	//如果是本轮的第一次请求交易
	if isFirstRequest {
		//region 3、根据对方传过来的上一个交易的临时rsmc私钥，签名最近的BR交易，保证对方确实放弃了上一个承诺交易
		err := signLastBR(tx, *channelInfo, signer.PeerId, aliceDataJson.Get("lastTempAddressPrivateKey").String())
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		//endregion

		//region 4、创建C2b
		commitmentTxRequest := &bean.CommitmentTx{}
		commitmentTxRequest.ChannelId = channelInfo.ChannelId
		commitmentTxRequest.Amount = aliceDataJson.Get("amount").Float()
		commitmentTxRequest.ChannelAddressPrivateKey = reqData.ChannelAddressPrivateKey
		commitmentTxRequest.CurrTempAddressPubKey = reqData.CurrTempAddressPubKey
		commitmentTxRequest.CurrTempAddressPrivateKey = reqData.CurrTempAddressPrivateKey
		newCommitmentTxInfo, err := createCommitmentTxHex(tx, false, commitmentTxRequest, channelInfo, latestCommitmentTxInfo, *signer)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		amountToOther = newCommitmentTxInfo.AmountToOther

		retData["rsmcHex"] = newCommitmentTxInfo.RSMCTxHex
		retData["toOtherHex"] = newCommitmentTxInfo.ToOtherTxHex
		//endregion

		// region 5、根据alice的Rsmc，创建对应的BR,为下一个交易做准备，create BR2b tx  for bob
		var myAddress = channelInfo.AddressB
		if signer.PeerId == channelInfo.PeerIdA {
			myAddress = channelInfo.AddressA
		}
		senderCommitmentTx := &dao.CommitmentTransaction{}
		senderCommitmentTx.Id = -1
		senderCommitmentTx.PropertyId = fundingTransaction.PropertyId
		senderCommitmentTx.RSMCMultiAddress = aliceRsmcMultiAddress
		senderCommitmentTx.RSMCRedeemScript = aliceRsmcRedeemScript
		senderCommitmentTx.RSMCTxHex = signedRsmcHex
		senderCommitmentTx.RSMCTxid = rsmcTxId
		senderCommitmentTx.AmountToRSMC = newCommitmentTxInfo.AmountToOther
		err = createCurrBR(tx, channelInfo, senderCommitmentTx, inputs, myAddress, fundingTransaction.FunderAddress, reqData.ChannelAddressPrivateKey, *signer)
		if err != nil {
			log.Println(err)
			return nil, "", err
		}
		//endregion

	} else {
		retData["rsmcHex"] = latestCommitmentTxInfo.RSMCTxHex
		retData["toOtherHex"] = latestCommitmentTxInfo.ToOtherTxHex
		amountToOther = latestCommitmentTxInfo.AmountToOther
	}

	//region 6、根据签名后的AliceRsmc创建alice的RD create RD tx for alice
	outputAddress := channelInfo.AddressA
	if signer.PeerId == channelInfo.PeerIdA {
		outputAddress = channelInfo.AddressB
	}
	_, senderRdhex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
		aliceRsmcMultiAddress,
		[]string{
			reqData.ChannelAddressPrivateKey,
		},
		inputs,
		outputAddress,
		channelInfo.FundingAddress,
		channelInfo.PropertyId,
		amountToOther,
		0,
		1000,
		&aliceRsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, targetUser, errors.New("fail to create rd")
	}
	retData["senderRdHex"] = senderRdhex
	//endregion create RD tx for alice

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	return retData, "", err
}

func (this *commitmentTxSignedManager) AfterAliceSignCommitmentTranctionAtBobSide(data string, user *bean.User) (retData map[string]interface{}, err error) {
	jsonObj := gjson.Parse(data)

	var channelId = jsonObj.Get("channelId").String()
	var signedRsmcHex = jsonObj.Get("signedRsmcHex").String()
	var signedToOtherHex = jsonObj.Get("signedToOtherHex").String()
	var rdHex = jsonObj.Get("rdHex").String()

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

	latestCcommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find sender's commitmentTxInfo")
		log.Println(err)
		return nil, err
	}

	if latestCcommitmentTxInfo.CurrState != dao.TxInfoState_RsmcCreate {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCcommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, err
	}

	myChannelAddress := channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdA {
		myChannelAddress = channelInfo.AddressA
	}

	err = signRdTx(tx, channelInfo, signedRsmcHex, rdHex, *latestCcommitmentTxInfo, myChannelAddress, user)
	if err != nil {
		return nil, err
	}

	//更新alice的当前承诺交易
	latestCcommitmentTxInfo.SignAt = time.Now()
	latestCcommitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	latestCcommitmentTxInfo.RSMCTxHex = signedRsmcHex
	latestCcommitmentTxInfo.ToOtherTxHex = signedToOtherHex
	bytes, err := json.Marshal(latestCcommitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	latestCcommitmentTxInfo.CurrHash = msgHash
	_ = tx.Update(latestCcommitmentTxInfo)
	_ = tx.Commit()

	retData = make(map[string]interface{})
	retData["channelId"] = channelId
	return retData, nil
}

func createAliceSideTxs(tx storm.Node, signData *bean.CommitmentTxSigned, dataFromCreator dao.CommitmentTxRequestInfo, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, signer *bean.User) (*dao.CommitmentTransaction, error) {
	owner := channelInfo.PeerIdA

	var isAliceSendToBob = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceSendToBob = false
	}

	var lastCommitmentTx = &dao.CommitmentTransaction{}
	err := tx.Select(
		q.Eq("ChannelId", signData.ChannelId),
		q.Eq("Owner", owner),
		q.Or(
			q.Eq("PeerIdA", signer.PeerId),
			q.Eq("PeerIdB", signer.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(lastCommitmentTx)
	if err != nil {
		lastCommitmentTx = nil
	}

	if lastCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, errors.New("latest commitment tx state is wrong")
	}

	if lastCommitmentTx != nil {
		count, _ := tx.Select(
			q.Eq("CommitmentTxId", lastCommitmentTx.Id)).
			Count(&dao.BreachRemedyTransaction{})
		if count > 0 {
			err = errors.New("already exist BreachRemedyTransaction ")
			return nil, err
		}

		// create BRa tx  for bob ，let the lastCommitmentTx abort,
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdB, channelInfo, lastCommitmentTx, signer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if breachRemedyTransaction.Amount > 0 {
			lastTempAddressPrivateKey := ""
			if isAliceSendToBob {
				lastTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCreator.LastTempAddressPubKey]
			} else {
				lastTempAddressPrivateKey = signData.LastTempAddressPrivateKey
			}
			if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
				err = errors.New("fail to get the lastTempAddressPrivateKey")
				log.Println(err)
				return nil, err
			}

			if _, err := tool.GetPubKeyFromWifAndCheck(lastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey); err != nil {
				return nil, errors.New(lastTempAddressPrivateKey + " is wrong private key for RSMCTempAddressPubKey")
			}

			inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTx.RSMCTxHex, lastCommitmentTx.RSMCMultiAddress, lastCommitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
				lastCommitmentTx.RSMCMultiAddress,
				[]string{
					lastTempAddressPrivateKey,
					tempAddrPrivateKeyMap[channelInfo.PubKeyB],
				},
				inputs,
				channelInfo.AddressB,
				fundingTransaction.FunderAddress,
				fundingTransaction.PropertyId,
				breachRemedyTransaction.Amount,
				0,
				0,
				&lastCommitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			breachRemedyTransaction.Txid = txid
			breachRemedyTransaction.BrTxHex = hex
			breachRemedyTransaction.SignAt = time.Now()
			breachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
			err = tx.Save(breachRemedyTransaction)
		}

		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(
			q.Eq("ChannelId", signData.ChannelId),
			q.Eq("Owner", owner),
			q.Eq("CommitmentTxId", lastCommitmentTx.Id),
			q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
			OrderBy("CreateAt").
			Reverse().
			First(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		lastCommitmentTx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentTx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = tx.Update(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	// create Cna tx
	var outputBean = commitmentOutputBean{}
	if isAliceSendToBob {
		outputBean.RsmcTempPubKey = dataFromCreator.CurrTempAddressPubKey
		//default alice transfer to bob ,then alice minus money
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentTx != nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	} else {
		outputBean.RsmcTempPubKey = signData.CurrTempAddressPubKey
		// if bob transfer to alice,then alice add money
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentTx != nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	}
	outputBean.OppositeSideChannelAddress = channelInfo.AddressB
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB

	commitmentTxInfo, err := createCommitmentTx(owner, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Rsmc

	usedTxidTemp := ""
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUseSingleInput(
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
		log.Println(usedTxid)
		usedTxidTemp = usedTxid
		commitmentTxInfo.RSMCTxid = txid
		commitmentTxInfo.RSMCTxHex = hex
	}

	//create to Bob tx
	if commitmentTxInfo.AmountToOther > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseRestInput(
			int(commitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			usedTxidTemp,
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
		commitmentTxInfo.ToOtherTxHex = hex
	}
	if lastCommitmentTx != nil {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentTx.Id
	}

	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	if lastCommitmentTx != nil {
		commitmentTxInfo.LastHash = lastCommitmentTx.CurrHash
	}
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = tx.Update(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create RDna tx
	rdTransaction, err := createRDTx(owner, channelInfo, commitmentTxInfo, channelInfo.AddressA, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if isAliceSendToBob {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCreator.CurrTempAddressPubKey]
	} else {
		currTempAddressPrivateKey = signData.CurrTempAddressPrivateKey
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHex, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
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
	return commitmentTxInfo, err
}

func createBobSideTxs(tx storm.Node, signData *bean.CommitmentTxSigned, dataFromCreator dao.CommitmentTxRequestInfo, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, signer *bean.User) (*dao.CommitmentTransaction, error) {
	owner := channelInfo.PeerIdB
	var isAliceSendToBob = true
	if signer.PeerId == channelInfo.PeerIdA {
		isAliceSendToBob = false
	}

	var lastCommitmentTx = &dao.CommitmentTransaction{}
	err := tx.Select(
		q.Eq("ChannelId", signData.ChannelId),
		q.Eq("Owner", owner),
		q.Or(
			q.Eq("PeerIdA", signer.PeerId),
			q.Eq("PeerIdB", signer.PeerId))).
		OrderBy("CreateAt").Reverse().
		First(lastCommitmentTx)
	if err != nil {
		lastCommitmentTx = nil
	}

	//In unilataral funding mode, only Alice is required to fund the channel.
	//So during funding procedure, on Bob side, he has no commitment transaction and revockable delivery transaction.
	if lastCommitmentTx != nil {

		if lastCommitmentTx.CurrState != dao.TxInfoState_CreateAndSign {
			return nil, errors.New("latest commitment tx state is wrong")
		}

		count, _ := tx.Select(
			q.Eq("CommitmentTxId", lastCommitmentTx.Id)).
			Count(&dao.BreachRemedyTransaction{})
		if count > 0 {
			err = errors.New("already exist BreachRemedyTransaction ")
			return nil, err
		}

		// create BR tx for alice
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdA, channelInfo, lastCommitmentTx, signer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if breachRemedyTransaction.Amount > 0 {
			lastTempAddressPrivateKey := ""
			if isAliceSendToBob {
				lastTempAddressPrivateKey = signData.LastTempAddressPrivateKey
			} else {
				lastTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCreator.LastTempAddressPubKey]
			}

			if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
				err = errors.New("fail to get the lastTempAddressPrivateKey")
				log.Println(err)
				return nil, err
			}

			if _, err := tool.GetPubKeyFromWifAndCheck(lastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey); err != nil {
				return nil, errors.New(lastTempAddressPrivateKey + " is wrong private key for RSMCTempAddressPubKey")
			}

			inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentTx.RSMCTxHex, lastCommitmentTx.RSMCMultiAddress, lastCommitmentTx.RSMCMultiAddressScriptPubKey)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
				lastCommitmentTx.RSMCMultiAddress,
				[]string{
					lastTempAddressPrivateKey,
					tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				},
				inputs,
				channelInfo.AddressA,
				fundingTransaction.FunderAddress,
				fundingTransaction.PropertyId,
				breachRemedyTransaction.Amount,
				0,
				0,
				&lastCommitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			breachRemedyTransaction.Txid = txid
			breachRemedyTransaction.BrTxHex = hex
			breachRemedyTransaction.SignAt = time.Now()
			breachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
			err = tx.Save(breachRemedyTransaction)
		}

		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(
			q.Eq("ChannelId", signData.ChannelId),
			q.Eq("Owner", owner),
			q.Eq("CommitmentTxId", lastCommitmentTx.Id),
			q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
			OrderBy("CreateAt").
			Reverse().
			First(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		lastCommitmentTx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentTx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = tx.Update(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	// create Cnb tx
	var outputBean = commitmentOutputBean{}
	if isAliceSendToBob {
		outputBean.RsmcTempPubKey = signData.CurrTempAddressPubKey
		//by default, alice transfers money to bob,then bob's balance increases.
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentTx != nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	} else {
		outputBean.RsmcTempPubKey = dataFromCreator.CurrTempAddressPubKey
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		if lastCommitmentTx != nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Add(decimal.NewFromFloat(dataFromCreator.Amount)).Float64()
		}
	}
	outputBean.OppositeSideChannelAddress = channelInfo.AddressA
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA

	commitmentTxInfo, err := createCommitmentTx(owner, channelInfo, fundingTransaction, outputBean, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Rsmc

	usedTxidTemp := ""
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionUseSingleInput(
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
		log.Println(usedTxid)
		usedTxidTemp = usedTxid
		commitmentTxInfo.RSMCTxid = txid
		commitmentTxInfo.RSMCTxHex = hex
	}

	//create to alice tx
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseRestInput(
			int(commitmentTxInfo.TxType),
			channelInfo.ChannelAddress,
			usedTxidTemp,
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
		commitmentTxInfo.ToOtherTxHex = hex
	}

	if lastCommitmentTx != nil {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentTx.Id
	}
	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	commitmentTxInfo.CurrHash = ""
	commitmentTxInfo.LastHash = ""
	if lastCommitmentTx != nil {
		commitmentTxInfo.LastHash = lastCommitmentTx.CurrHash
	}
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = tx.Update(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create RDb tx
	rdTransaction, err := createRDTx(owner, channelInfo, commitmentTxInfo, channelInfo.AddressB, signer)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if isAliceSendToBob {
		currTempAddressPrivateKey = signData.CurrTempAddressPrivateKey
	} else {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[dataFromCreator.CurrTempAddressPubKey]
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHex, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
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
	return commitmentTxInfo, err
}

//在Rsmc二次签名完成后，创建对应的BR
func createCurrBR(tx storm.Node, channelInfo *dao.ChannelInfo,
	commitmentTx *dao.CommitmentTransaction, inputs []rpc.TransactionInputItem,
	outputAddress, changeToAddress string, channelPrivateKey string, user bean.User) (err error) {
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}
	_ = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Eq("InputTxid", commitmentTx.RSMCTxid),
		q.Or(
			q.Eq("PeerIdA", user.PeerId),
			q.Eq("PeerIdB", user.PeerId))).
		First(breachRemedyTransaction)
	if breachRemedyTransaction.Id == 0 {
		breachRemedyTransaction, err = createBRTx(user.PeerId, channelInfo, commitmentTx, &user)
		if err != nil {
			log.Println(err)
			return err
		}
		if breachRemedyTransaction.Amount > 0 {
			txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionUseUnsendInput(
				commitmentTx.RSMCMultiAddress,
				[]string{
					channelPrivateKey,
				},
				inputs,
				outputAddress,
				changeToAddress,
				channelInfo.PropertyId,
				breachRemedyTransaction.Amount,
				0,
				0,
				&commitmentTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return err
			}
			breachRemedyTransaction.OutAddress = outputAddress
			breachRemedyTransaction.Txid = txid
			breachRemedyTransaction.BrTxHex = hex
			breachRemedyTransaction.CurrState = dao.TxInfoState_RsmcCreate
			_ = tx.Save(breachRemedyTransaction)
		}
	}
	return nil
}

//对上一个承诺交易的br进行签名
func signLastBR(tx storm.Node, channelInfo dao.ChannelInfo, userPeerId string, lastTempAddressPrivateKey string) (err error) {
	lastBreachRemedyTransaction := &dao.BreachRemedyTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelInfo.ChannelId),
		q.Or(
			q.Eq("PeerIdA", userPeerId),
			q.Eq("PeerIdB", userPeerId))).
		OrderBy("CreateAt").
		Reverse().First(lastBreachRemedyTransaction)
	if lastBreachRemedyTransaction != nil && lastBreachRemedyTransaction.Id > 0 {
		inputs, err := getInputsForNextTxByParseTxHashVout(lastBreachRemedyTransaction.RsmcTxHex, lastBreachRemedyTransaction.OutAddress, lastBreachRemedyTransaction.InputAddressScriptPubKey)
		if err != nil {
			log.Println(err)
			return errors.New("fail to sign breachRemedyTransaction")
		}
		_, signedBRHex, err := rpcClient.OmniSignRawTransactionForUnsend(lastBreachRemedyTransaction.RsmcTxHex, inputs, lastTempAddressPrivateKey)
		if err != nil {
			return errors.New("fail to sign breachRemedyTransaction")
		}
		result, err := rpcClient.TestMemPoolAccept(signedBRHex)
		if err != nil {
			return errors.New("fail to sign breachRemedyTransaction")
		}
		if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
			if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
				return errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
			}
		}
		lastBreachRemedyTransaction.SignAt = time.Now()
		lastBreachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
		_ = tx.Save(lastBreachRemedyTransaction)
	}
	return nil
}

func signRdTx(tx storm.Node, channelInfo *dao.ChannelInfo, signedRsmcHex string, rdHex string, latestCcommitmentTxInfo dao.CommitmentTransaction, outputAddress string, user *bean.User) (err error) {
	inputs, err := getInputsForNextTxByParseTxHashVout(signedRsmcHex, latestCcommitmentTxInfo.RSMCMultiAddress, latestCcommitmentTxInfo.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return err
	}
	_, signedRdHex, err := rpcClient.OmniSignRawTransactionForUnsend(rdHex, inputs, tempAddrPrivateKeyMap[latestCcommitmentTxInfo.RSMCTempAddressPubKey])
	if err != nil {
		return err
	}
	result, err := rpcClient.TestMemPoolAccept(signedRdHex)
	if err != nil {
		return err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}

	rdHexDecode, err := rpcClient.DecodeRawTransaction(signedRdHex)
	if err != nil {
		return err
	}
	if checkRdOutputAddress(rdHexDecode, outputAddress) == false {
		return errors.New("rdtx has wrong output address")
	}
	aliceRdTxid := gjson.Get(rdHexDecode, "txid").String()
	rdTransaction, err := createRDTx(user.PeerId, channelInfo, &latestCcommitmentTxInfo, outputAddress, user)
	if err != nil {
		log.Println(err)
		return err
	}
	rdTransaction.RDType = 0
	rdTransaction.TxHash = signedRdHex
	rdTransaction.Txid = aliceRdTxid
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
