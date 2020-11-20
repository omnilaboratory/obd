package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/asdine/storm/q"
)

type commitmentTxManager struct {
	operationFlag sync.Mutex
}

var tempRsmcCreateP2pData map[string]bean.AliceRequestToCreateCommitmentTxOfP2p
var tempP2pData_352 map[string]bean.PayeeSignCommitmentTxOfP2p

var CommitmentTxService commitmentTxManager

// step 1 协议号：100351  当发起转账人alice申请发起转账
func (this *commitmentTxManager) CommitmentTransactionCreated(msg bean.RequestMessage, creator *bean.User) (retData interface{}, needSign bool, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		return nil, false, errors.New(enum.Tips_common_empty + "msg.Data")
	}
	now := time.Now()
	log.Println("begin rsmc step 1 ", now)
	reqData := &bean.RequestCreateCommitmentTx{}
	err = json.Unmarshal([]byte(msg.Data), reqData)
	if err != nil {
		return nil, false, err
	}
	if tool.CheckIsString(&reqData.ChannelId) == false {
		return nil, false, errors.New(enum.Tips_common_empty + " channel_id")
	}

	if tool.CheckIsString(&reqData.LastTempAddressPrivateKey) == false {
		return nil, false, errors.New(enum.Tips_common_empty + "last_temp_address_private_key")
	}

	if reqData.Amount <= 0 {
		return nil, false, errors.New(enum.Tips_common_wrong + "payment amount")
	}

	tx, err := creator.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, reqData.ChannelId, creator.PeerId)
	if channelInfo == nil {
		err = errors.New(enum.Tips_funding_notFoundChannelByChannelId + reqData.ChannelId)
		log.Println(err)
		return nil, false, err
	}

	if channelInfo.CurrState == dao.ChannelState_NewTx {
		return nil, false, errors.New(enum.Tips_common_newTxMsg)
	}

	fundingTransaction := getFundingTransactionByChannelId(tx, channelInfo.ChannelId, creator.PeerId)
	duration := time.Now().Sub(fundingTransaction.CreateAt)
	if duration > time.Minute*30 {
		pass, err := checkChannelOmniAssetAmount(*channelInfo)
		if err != nil {
			return nil, false, err
		}
		if pass == false {
			err = errors.New(enum.Tips_rsmc_broadcastedChannel)
			log.Println(err)
			return nil, false, err
		}
	}

	if channelInfo.CurrState < dao.ChannelState_NewTx {
		return nil, false, errors.New("do not finish funding")
	}

	targetUser := channelInfo.PeerIdB
	if creator.PeerId == channelInfo.PeerIdB {
		targetUser = channelInfo.PeerIdA
	}

	if targetUser != msg.RecipientUserPeerId {
		return nil, false, errors.New(enum.Tips_rsmc_notTargetUser + msg.RecipientUserPeerId)
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, creator.PeerId)
	if err != nil {
		return nil, false, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
	}

	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_Init {
		tx.DeleteStruct(latestCommitmentTxInfo)
		latestCommitmentTxInfo, err = getLatestCommitmentTxUseDbTx(tx, reqData.ChannelId, creator.PeerId)
		if err != nil {
			return nil, false, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
		}
	}

	if latestCommitmentTxInfo.TxType != dao.CommitmentTransactionType_Rsmc {
		return nil, false, errors.New(enum.Tips_rsmc_errorCommitmentTxType + strconv.Itoa(int(latestCommitmentTxInfo.TxType)))
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_CreateAndSign &&
		latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
		return nil, false, errors.New(enum.Tips_rsmc_errorCommitmentTxState + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
	}

	//region check input data 检测输入数据
	//如果是第一次发起请求
	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign {
		balance := latestCommitmentTxInfo.AmountToRSMC
		if balance < reqData.Amount {
			return nil, false, errors.New(enum.Tips_rsmc_notEnoughBalance)
		}

		if _, err = tool.GetPubKeyFromWifAndCheck(reqData.LastTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey); err != nil {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_wrongPrivateKeyForLast, reqData.LastTempAddressPrivateKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
		}
	} else {
		if reqData.CurrTempAddressPubKey != latestCommitmentTxInfo.RSMCTempAddressPubKey {
			return nil, false, errors.New(fmt.Sprintf(enum.Tips_rsmc_notSameValueWhenCreate, reqData.CurrTempAddressPubKey, latestCommitmentTxInfo.RSMCTempAddressPubKey))
		}
		lastCommitmentTx := &dao.CommitmentTransaction{}
		_ = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, lastCommitmentTx)
		if _, err = tool.GetPubKeyFromWifAndCheck(reqData.LastTempAddressPrivateKey, lastCommitmentTx.RSMCTempAddressPubKey); err != nil {
			return nil, false, err
		}
	}

	if _, err = getAddressFromPubKey(reqData.CurrTempAddressPubKey); err != nil {
		return nil, false, errors.New(enum.Tips_common_wrong + "curr_temp_address_pub_key")
	}
	//endregion

	retSignData := bean.NeedAliceSignRsmcDataForC2a{}

	p2pData := &bean.AliceRequestToCreateCommitmentTxOfP2p{}
	p2pData.ChannelId = channelInfo.ChannelId
	p2pData.Amount = reqData.Amount
	p2pData.LastTempAddressPrivateKey = reqData.LastTempAddressPrivateKey
	p2pData.CurrTempAddressPubKey = reqData.CurrTempAddressPubKey

	needSign = false

	if latestCommitmentTxInfo.CurrState == dao.TxInfoState_CreateAndSign {
		//创建c2a omni的交易不能一个输入，多个输出，所以就是两个交易
		newCommitmentTxInfo, rawTx, err := createCommitmentTxHex(tx, true, reqData, channelInfo, latestCommitmentTxInfo, *creator)
		if err != nil {
			return nil, false, err
		}
		newCommitmentTxInfo.CurrState = dao.TxInfoState_Init
		_ = tx.UpdateField(newCommitmentTxInfo, "CurrState", dao.TxInfoState_Init)

		p2pData.CommitmentTxHash = newCommitmentTxInfo.CurrHash
		p2pData.RsmcRawData = rawTx.RsmcRawTxData
		p2pData.CounterpartyRawData = rawTx.ToCounterpartyRawTxData

		needSign = true

	} else {
		p2pData.CommitmentTxHash = latestCommitmentTxInfo.CurrHash
		if len(latestCommitmentTxInfo.RSMCTxid) == 0 {
			rawTx := &dao.CommitmentTxRawTx{}
			tx.Select(q.Eq("CommitmentTxId", latestCommitmentTxInfo.Id)).First(rawTx)
			if rawTx.Id == 0 {
				return nil, false, errors.New("not found rawTx")
			}
			p2pData.RsmcRawData = rawTx.RsmcRawTxData
			p2pData.CounterpartyRawData = rawTx.ToCounterpartyRawTxData
			needSign = true
		}
	}

	p2pData.PayerNodeAddress = msg.SenderNodePeerId
	p2pData.PayerPeerId = msg.SenderUserPeerId

	_ = tx.Commit()

	if needSign {
		if tempRsmcCreateP2pData == nil {
			tempRsmcCreateP2pData = make(map[string]bean.AliceRequestToCreateCommitmentTxOfP2p)
		}
		tempRsmcCreateP2pData[creator.PeerId+"_"+p2pData.ChannelId] = *p2pData

		retSignData.ChannelId = channelInfo.ChannelId
		retSignData.RsmcRawData = p2pData.RsmcRawData
		retSignData.CounterpartyRawData = p2pData.CounterpartyRawData

		return retSignData, true, nil
	}

	return p2pData, false, err
}

// step 2 协议号：100360 当alice完成C2a的rsmc部分签名操作
func (this *commitmentTxManager) OnAliceSignC2aRawTxAtAliceSide(msg bean.RequestMessage, user *bean.User) (toAlice, retData interface{}, err error) {
	if tool.CheckIsString(&msg.Data) == false {
		err = errors.New(enum.Tips_common_empty + "msg.data")
		log.Println(err)
		return nil, nil, err
	}
	signedDataForC2a := bean.AliceSignedRsmcDataForC2a{}
	_ = json.Unmarshal([]byte(msg.Data), &signedDataForC2a)

	if tool.CheckIsString(&signedDataForC2a.ChannelId) == false {
		err = errors.New(enum.Tips_common_empty + "channel_id")
		log.Println(err)
		return nil, nil, err
	}

	p2pData := tempRsmcCreateP2pData[user.PeerId+"_"+signedDataForC2a.ChannelId]
	if &p2pData == nil {
		return nil, nil, errors.New(enum.Tips_common_wrong + "channel_id")
	}

	if tool.CheckIsString(&signedDataForC2a.RsmcSignedHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedDataForC2a.RsmcSignedHex, 1); pass == false {
			err = errors.New(enum.Tips_common_wrong + "rsmc_signed_hex")
			log.Println(err)
			return nil, nil, err
		}
	}

	if tool.CheckIsString(&signedDataForC2a.CounterpartySignedHex) == false {
		err = errors.New(enum.Tips_common_empty + "counterparty_signed_hex")
		log.Println(err)
		return nil, nil, err
	}

	if pass, _ := rpcClient.CheckMultiSign(true, signedDataForC2a.CounterpartySignedHex, 1); pass == false {
		err = errors.New(enum.Tips_common_wrong + "counterparty_signed_hex")
		log.Println(err)
		return nil, nil, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer tx.Rollback()

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, signedDataForC2a.ChannelId, user.PeerId)
	if err != nil {
		return nil, nil, errors.New(enum.Tips_channel_notFoundLatestCommitmentTx)
	}

	if len(latestCommitmentTxInfo.RSMCTxHex) > 0 {
		result, err := rpcClient.TestMemPoolAccept(signedDataForC2a.RsmcSignedHex)
		if err != nil {
			return nil, nil, err
		}
		txid := gjson.Parse(result).Array()[0].Get("txid").Str
		//封装好的签名数据，给bob的客户端签名使用
		latestCommitmentTxInfo.RSMCTxHex = signedDataForC2a.RsmcSignedHex
		latestCommitmentTxInfo.RSMCTxid = txid
	}

	if len(latestCommitmentTxInfo.ToCounterpartyTxHex) > 0 {
		result, err := rpcClient.TestMemPoolAccept(signedDataForC2a.CounterpartySignedHex)
		if err != nil {
			return nil, nil, err
		}
		txid := gjson.Parse(result).Array()[0].Get("txid").Str
		//封装好的签名数据，给bob的客户端签名使用
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedDataForC2a.CounterpartySignedHex
		latestCommitmentTxInfo.ToCounterpartyTxid = txid
	}
	latestCommitmentTxInfo.CurrState = dao.TxInfoState_Create
	_ = tx.Update(latestCommitmentTxInfo)

	tx.Commit()

	p2pData.RsmcRawData.Hex = signedDataForC2a.RsmcSignedHex
	p2pData.CounterpartyRawData.Hex = signedDataForC2a.CounterpartySignedHex

	toAliceResult := bean.AliceSignedRsmcDataForC2aResult{}
	toAliceResult.ChannelId = p2pData.ChannelId
	toAliceResult.CurrTempAddressPubKey = p2pData.CurrTempAddressPubKey
	toAliceResult.CommitmentTxHash = p2pData.CommitmentTxHash
	toAliceResult.Amount = p2pData.Amount
	return toAliceResult, p2pData, nil
}

// step 6 协议号：352 响应来自p2p的352号消息 推送110352消息
func (this *commitmentTxManager) OnGetBobC2bPartialSignTxAtAliceSide(msg bean.RequestMessage, data string, user *bean.User) (retData interface{}, needNoticeAlice bool, err error) {
	dataFromP2p352 := bean.PayeeSignCommitmentTxOfP2p{}
	_ = json.Unmarshal([]byte(data), &dataFromP2p352)

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, dataFromP2p352.ChannelId, user.PeerId)
	if channelInfo == nil {
		return nil, false, errors.New("not found channelInfo at targetSide")
	}

	//如果bob不同意这次交易
	if dataFromP2p352.Approval == false {
		latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, dataFromP2p352.ChannelId, user.PeerId)
		if err != nil {
			err = errors.New("fail to find sender's commitmentTxInfo")
			log.Println(err)
			return nil, false, err
		}
		_ = tx.DeleteStruct(latestCommitmentTxInfo)
		_ = tx.Commit()
		return dataFromP2p352, false, nil
	}

	channelInfo.CurrState = dao.ChannelState_NewTx
	_ = tx.Update(channelInfo)

	_ = tx.Commit()

	if tempP2pData_352 == nil {
		tempP2pData_352 = make(map[string]bean.PayeeSignCommitmentTxOfP2p)
	}
	tempP2pData_352[user.PeerId+"_"+dataFromP2p352.ChannelId] = dataFromP2p352

	needAliceSignRmscTxForC2b := bean.NeedAliceSignRsmcTxForC2b{}
	needAliceSignRmscTxForC2b.ChannelId = dataFromP2p352.ChannelId
	needAliceSignRmscTxForC2b.C2bRsmcPartialData = dataFromP2p352.C2bRsmcTxData
	needAliceSignRmscTxForC2b.C2bCounterpartyPartialData = dataFromP2p352.C2bCounterpartyTxData
	needAliceSignRmscTxForC2b.C2aRdPartialData = dataFromP2p352.C2aRdTxData
	needAliceSignRmscTxForC2b.PayeeNodeAddress = msg.SenderNodePeerId
	needAliceSignRmscTxForC2b.PayeePeerId = msg.SenderUserPeerId
	return needAliceSignRmscTxForC2b, false, nil
}

// step 7 协议号：100362(to Obd) 响应Alice对C2b的Rsmc的签名，然后创建C2b的Br和Rd，再推送Rd和Br的Raw交易给alice签名
func (this *commitmentTxManager) OnAliceSignedC2bTxAtAliceSide(data string, user *bean.User) (retData interface{}, err error) {
	aliceSignedRmscTxForC2b := bean.AliceSignedRsmcTxForC2b{}
	_ = json.Unmarshal([]byte(data), &aliceSignedRmscTxForC2b)

	dataFromP2p352 := tempP2pData_352[user.PeerId+"_"+aliceSignedRmscTxForC2b.ChannelId]
	if len(dataFromP2p352.ChannelId) == 0 {
		return nil, errors.New(enum.Tips_common_empty + "channel_id")
	}

	if tool.CheckIsString(&dataFromP2p352.C2bRsmcTxData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, aliceSignedRmscTxForC2b.C2bRsmcSignedHex, 2); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c2b_rsmc_signed_hex")
			log.Println(err)
			return nil, err
		}
	}
	dataFromP2p352.C2bRsmcTxData.Hex = aliceSignedRmscTxForC2b.C2bRsmcSignedHex

	if tool.CheckIsString(&dataFromP2p352.C2bCounterpartyTxData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(true, aliceSignedRmscTxForC2b.C2bCounterpartySignedHex, 2); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c2b_counterparty_signed_hex")
			log.Println(err)
			return nil, err
		}
	}
	dataFromP2p352.C2bCounterpartyTxData.Hex = aliceSignedRmscTxForC2b.C2bCounterpartySignedHex

	if tool.CheckIsString(&dataFromP2p352.C2aRdTxData.Hex) {
		if pass, _ := rpcClient.CheckMultiSign(false, aliceSignedRmscTxForC2b.C2aRdSignedHex, 2); pass == false {
			err = errors.New(enum.Tips_common_wrong + "signed c2a_rd_signed_hex")
			log.Println(err)
			return nil, err
		}
	}
	dataFromP2p352.C2aRdTxData.Hex = aliceSignedRmscTxForC2b.C2aRdSignedHex

	tempP2pData_352[user.PeerId+"_"+dataFromP2p352.ChannelId] = dataFromP2p352

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	needAliceSignRdTxForC2b := bean.NeedAliceSignRdTxForC2b{}

	channelInfo := getChannelInfoByChannelId(tx, dataFromP2p352.ChannelId, user.PeerId)
	if channelInfo == nil {
		return nil, errors.New("not found channelInfo at targetSide")
	}
	needAliceSignRdTxForC2b.ChannelId = channelInfo.ChannelId

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelInfo.ChannelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find sender's commitmentTxInfo")
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.CurrHash != dataFromP2p352.CommitmentTxHash {
		err = errors.New("wrong request hash, Please notice payee,")
		log.Println(err)
		return nil, err
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
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

	//处理对方的数据
	//签名对方传过来的rsmcHex
	bobSignedRsmcHex := aliceSignedRmscTxForC2b.C2bRsmcSignedHex

	//region create RD tx for bob
	bobMultiAddr, err := rpcClient.CreateMultiSig(2, []string{dataFromP2p352.CurrTempAddressPubKey, myChannelPubKey})
	if err != nil {
		return nil, err
	}
	c2bRsmcMultiAddress := gjson.Get(bobMultiAddr, "address").String()
	c2bRsmcRedeemScript := gjson.Get(bobMultiAddr, "redeemScript").String()
	addressJson, err := rpcClient.GetAddressInfo(c2bRsmcMultiAddress)
	if err != nil {
		return nil, err
	}
	c2bRsmcMultiAddressScriptPubKey := gjson.Get(addressJson, "scriptPubKey").String()

	c2bRsmcOutputs, err := getInputsForNextTxByParseTxHashVout(
		bobSignedRsmcHex,
		c2bRsmcMultiAddress,
		c2bRsmcMultiAddressScriptPubKey,
		c2bRsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if tool.CheckIsString(&bobSignedRsmcHex) {
		c2bRdHexData, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
			c2bRsmcMultiAddress,
			c2bRsmcOutputs,
			partnerChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			latestCommitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&c2bRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, errors.New("fail to create rd")
		}
		c2bRdRawData := bean.NeedClientSignTxData{}
		c2bRdRawData.Hex = c2bRdHexData["hex"].(string)
		c2bRdRawData.Inputs = c2bRdHexData["inputs"]
		c2bRdRawData.IsMultisig = true
		c2bRdRawData.PubKeyA = dataFromP2p352.CurrTempAddressPubKey
		c2bRdRawData.PubKeyB = myChannelPubKey
		needAliceSignRdTxForC2b.C2bRdRawData = c2bRdRawData
		//endregion create RD tx for bob

		//region 根据对对方的Rsmc签名，生成惩罚对方，自己获益BR
		res, err := rpcClient.TestMemPoolAccept(dataFromP2p352.C2bRsmcTxData.Hex)
		bobRsmcTxid := gjson.Parse(res).Array()[0].Get("txid").Str

		bobCommitmentTx := &dao.CommitmentTransaction{}
		bobCommitmentTx.Id = latestCommitmentTxInfo.Id
		bobCommitmentTx.PropertyId = channelInfo.PropertyId
		bobCommitmentTx.RSMCTempAddressPubKey = dataFromP2p352.CurrTempAddressPubKey
		bobCommitmentTx.RSMCMultiAddress = c2bRsmcMultiAddress
		bobCommitmentTx.RSMCRedeemScript = c2bRsmcRedeemScript
		bobCommitmentTx.RSMCMultiAddressScriptPubKey = c2bRsmcMultiAddressScriptPubKey
		bobCommitmentTx.RSMCTxHex = bobSignedRsmcHex
		bobCommitmentTx.RSMCTxid = bobRsmcTxid
		bobCommitmentTx.AmountToRSMC = latestCommitmentTxInfo.AmountToCounterparty
		c2bBrHexData, err := createCurrCommitmentTxRawBR(tx, dao.BRType_Rmsc, channelInfo, bobCommitmentTx, c2bRsmcOutputs, myChannelAddress, *user)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		c2bBrRawData := bean.NeedClientSignRawBRTxData{}
		c2bBrRawData.Hex = c2bBrHexData["hex"].(string)
		c2bBrRawData.Inputs = c2bBrHexData["inputs"]
		c2bBrRawData.BrId = c2bBrHexData["br_id"].(int)
		c2bBrRawData.IsMultisig = true
		c2bBrRawData.PubKeyA = dataFromP2p352.CurrTempAddressPubKey
		c2bBrRawData.PubKeyB = myChannelPubKey
		needAliceSignRdTxForC2b.C2bBrRawData = c2bBrRawData
	}
	//endregion

	_ = tx.Commit()
	return needAliceSignRdTxForC2b, nil
}

// step 8 协议号：100363 Alice完成对C2b的RD的签名
func (this *commitmentTxManager) OnAliceSignedC2b_RDTxAtAliceSide(data string, user *bean.User) (aliceRetData, bobRetData interface{}, needNoticeAlice bool, err error) {
	aliceSignedRdTxForC2b := bean.AliceSignedRdTxForC2b{}
	_ = json.Unmarshal([]byte(data), &aliceSignedRdTxForC2b)

	dataFromP2p352 := tempP2pData_352[user.PeerId+"_"+aliceSignedRdTxForC2b.ChannelId]
	if len(dataFromP2p352.ChannelId) == 0 {
		return nil, nil, false, errors.New(enum.Tips_common_empty + "channel_id")
	}

	//region 检测传入数据
	var channelId = dataFromP2p352.ChannelId
	if tool.CheckIsString(&channelId) == false {
		err = errors.New("wrong channelId")
		log.Println(err)
		return nil, nil, false, err
	}

	if tool.CheckIsString(&dataFromP2p352.CommitmentTxHash) == false {
		err = errors.New("wrong commitmentTxHash")
		log.Println(err)
		return nil, nil, false, err
	}

	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, nil, true, err
	}
	defer tx.Rollback()

	channelInfo := getChannelInfoByChannelId(tx, channelId, user.PeerId)
	if channelInfo == nil {
		return nil, nil, true, errors.New("not found channelInfo at targetSide")
	}

	latestCommitmentTxInfo, err := getLatestCommitmentTxUseDbTx(tx, channelId, user.PeerId)
	if err != nil {
		err = errors.New("fail to find sender's commitmentTxInfo")
		log.Println(err)
		return nil, nil, true, err
	}

	if latestCommitmentTxInfo.CurrHash != dataFromP2p352.CommitmentTxHash {
		err = errors.New("wrong request hash, Please notice payee,")
		log.Println(err)
		return nil, nil, true, err
	}

	if latestCommitmentTxInfo.CurrState != dao.TxInfoState_Create {
		err = errors.New("wrong commitmentTxInfo state " + strconv.Itoa(int(latestCommitmentTxInfo.CurrState)))
		log.Println(err)
		return nil, nil, false, err
	}

	aliceData := make(map[string]interface{})
	aliceData["channel_id"] = dataFromP2p352.ChannelId
	aliceData["approval"] = dataFromP2p352.Approval

	var c2aRsmcTestResult string
	var c2aSignedRsmcHex = dataFromP2p352.C2aSignedRsmcHex
	if tool.CheckIsString(&c2aSignedRsmcHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, c2aSignedRsmcHex, 2); pass == false {
			return nil, nil, false, errors.New(enum.Tips_common_wrong + "c2a_signed_rsmc_hex")
		}
		c2aRsmcTestResult, err = rpcClient.TestMemPoolAccept(c2aSignedRsmcHex)
		if err != nil {
			err = errors.New("wrong signedRsmcHex")
			log.Println(err)
			return nil, nil, false, err
		}
	}

	var signedToCounterpartyHex = dataFromP2p352.C2aSignedToCounterpartyTxHex
	var toCounterpartyTestResult string
	if tool.CheckIsString(&signedToCounterpartyHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, signedToCounterpartyHex, 2); pass == false {
			return nil, nil, false, errors.New(enum.Tips_common_wrong + "c2a_signed_to_counterparty_tx_hex")
		}
		toCounterpartyTestResult, err = rpcClient.TestMemPoolAccept(signedToCounterpartyHex)
		if err != nil {
			err = errors.New("wrong signedToOtherHex")
			log.Println(err)
			return nil, nil, false, err
		}
	}

	var aliceRdHex = dataFromP2p352.C2aRdTxData.Hex
	if tool.CheckIsString(&aliceRdHex) {
		if pass, _ := rpcClient.CheckMultiSign(false, aliceRdHex, 2); pass == false {
			return nil, nil, false, errors.New(enum.Tips_common_wrong + "c2a_rd_signed_hex")
		}
	}

	var bobRsmcHex = dataFromP2p352.C2bRsmcTxData.Hex
	if tool.CheckIsString(&bobRsmcHex) {
		if pass, _ := rpcClient.CheckMultiSign(true, bobRsmcHex, 2); pass == false {
			return nil, nil, false, errors.New(enum.Tips_common_wrong + "c2b_rsmc_signed_hex")
		}
	}

	var bobCurrTempAddressPubKey = dataFromP2p352.CurrTempAddressPubKey
	if tool.CheckIsString(&bobCurrTempAddressPubKey) == false {
		err = errors.New("wrong curr_temp_address_pub_key")
		log.Println(err)
		return nil, nil, false, err
	}

	var c2bToCounterpartyTxHex = dataFromP2p352.C2bCounterpartyTxData.Hex
	if len(c2bToCounterpartyTxHex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, c2bToCounterpartyTxHex, 2); pass == false {
			return nil, nil, false, errors.New(enum.Tips_common_wrong + "c2b_counterparty_tx_data_hex")
		}
	}
	//endregion

	fundingTransaction := getFundingTransactionByChannelId(tx, channelId, user.PeerId)
	if fundingTransaction == nil {
		return nil, nil, true, errors.New("not found fundingTransaction at targetSide")
	}

	var myChannelPubKey = channelInfo.PubKeyA
	var myChannelAddress = channelInfo.AddressA
	var partnerChannelAddress = channelInfo.AddressB
	if user.PeerId == channelInfo.PeerIdB {
		myChannelAddress = channelInfo.AddressB
		myChannelPubKey = channelInfo.PubKeyB
		partnerChannelAddress = channelInfo.AddressA
	}

	//region 根据对方传过来的上一个交易的临时rsmc私钥，签名上一次的BR交易，保证对方确实放弃了上一个承诺交易
	var bobLastTempAddressPrivateKey = dataFromP2p352.LastTempAddressPrivateKey
	err = signLastBR(tx, dao.BRType_Rmsc, *channelInfo, user.PeerId, bobLastTempAddressPrivateKey, latestCommitmentTxInfo.LastCommitmentTxId)
	if err != nil {
		log.Println(err)
		return nil, nil, false, err
	}
	//endregion

	if tool.CheckIsString(&c2aSignedRsmcHex) {
		latestCommitmentTxInfo.RSMCTxHex = c2aSignedRsmcHex
		latestCommitmentTxInfo.RSMCTxid = gjson.Parse(c2aRsmcTestResult).Array()[0].Get("txid").Str

		// 保存Rd交易
		err = saveRdTx(tx, channelInfo, c2aSignedRsmcHex, aliceRdHex, latestCommitmentTxInfo, myChannelAddress, user)
		if err != nil {
			return nil, nil, true, err
		}
	}

	latestCommitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	latestCommitmentTxInfo.SignAt = time.Now()

	if tool.CheckIsString(&signedToCounterpartyHex) {
		latestCommitmentTxInfo.ToCounterpartyTxHex = signedToCounterpartyHex
		latestCommitmentTxInfo.ToCounterpartyTxid = gjson.Parse(toCounterpartyTestResult).Array()[0].Get("txid").Str
	}

	//重新生成交易id
	bytes, err := json.Marshal(latestCommitmentTxInfo)
	latestCommitmentTxInfo.CurrHash = tool.SignMsgWithSha256(bytes)
	_ = tx.Update(latestCommitmentTxInfo)

	lastCommitmentTxInfo := dao.CommitmentTransaction{}
	err = tx.One("Id", latestCommitmentTxInfo.LastCommitmentTxId, &lastCommitmentTxInfo)
	if err == nil {
		lastCommitmentTxInfo.CurrState = dao.TxInfoState_Abord
		_ = tx.Update(lastCommitmentTxInfo)
	}

	channelInfo.CurrState = dao.ChannelState_CanUse
	_ = tx.Update(channelInfo)

	//返回给alice的数据
	aliceData["latest_commitment_tx_info"] = latestCommitmentTxInfo

	//处理对方的数据
	bobData := bean.AliceSignedC2bTxDataP2p{}
	bobData.C2aCommitmentTxHash = dataFromP2p352.CommitmentTxHash

	//签名对方传过来的rsmcHex
	c2bSignedRsmcHex := dataFromP2p352.C2bRsmcTxData.Hex
	if len(c2bSignedRsmcHex) > 0 {
		if pass, _ := rpcClient.CheckMultiSign(true, c2bSignedRsmcHex, 2); pass == false {
			return nil, nil, false, errors.New(enum.Tips_common_wrong + "c2b_rsmc_tx_data_hex")
		}
	}

	err = checkBobRemcData(c2bSignedRsmcHex, latestCommitmentTxInfo)
	if err != nil {
		return nil, nil, false, err
	}
	bobData.C2bRsmcSignedHex = c2bSignedRsmcHex

	//region create RD tx for bob
	c2bMultiAddr, err := rpcClient.CreateMultiSig(2, []string{bobCurrTempAddressPubKey, myChannelPubKey})
	if err != nil {
		return nil, nil, false, err
	}
	c2bRsmcMultiAddress := gjson.Get(c2bMultiAddr, "address").String()
	c2bRsmcRedeemScript := gjson.Get(c2bMultiAddr, "redeemScript").String()
	addressJson, err := rpcClient.GetAddressInfo(c2bRsmcMultiAddress)
	if err != nil {
		return nil, nil, false, err
	}
	c2bRsmcMultiAddressScriptPubKey := gjson.Get(addressJson, "scriptPubKey").String()

	c2bRsmcOutputs, err := getInputsForNextTxByParseTxHashVout(
		c2bSignedRsmcHex,
		c2bRsmcMultiAddress,
		c2bRsmcMultiAddressScriptPubKey,
		c2bRsmcRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, nil, false, err
	}

	if len(c2bRsmcOutputs) > 0 {
		c2bRdHexData, err := rpcClient.OmniCreateRawTransactionUseUnsendInput(
			c2bRsmcMultiAddress,
			c2bRsmcOutputs,
			partnerChannelAddress,
			channelInfo.FundingAddress,
			channelInfo.PropertyId,
			latestCommitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			1000,
			&c2bRsmcRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, nil, false, errors.New("fail to create rd")
		}
		c2bRdRawData := bean.NeedClientSignTxData{}
		c2bRdRawData.Hex = aliceSignedRdTxForC2b.C2bRdSignedHex
		c2bRdRawData.Inputs = c2bRdHexData["inputs"]
		c2bRdRawData.IsMultisig = true
		c2bRdRawData.PubKeyA = dataFromP2p352.CurrTempAddressPubKey
		c2bRdRawData.PubKeyB = myChannelPubKey
		bobData.C2bRdPartialData = c2bRdRawData
		//endregion create RD tx for alice

		//region 根据对对方的Rsmc签名，生成惩罚对方，自己获益BR
		err = updateCurrCommitmentTxRawBR(tx, aliceSignedRdTxForC2b.C2bBrId, aliceSignedRdTxForC2b.C2bBrSignedHex, *user)
		if err != nil {
			log.Println(err)
			return nil, nil, false, err
		}
	}

	//endregion
	_ = tx.Commit()

	bobData.C2bCounterpartySignedHex = c2bToCounterpartyTxHex
	bobData.ChannelId = channelId
	return aliceData, bobData, true, nil
}

// 广播某一次承诺交易
func (this *commitmentTxManager) SendSomeCommitmentById(data string, user *bean.User) (retData interface{}, err error) {
	id, err := strconv.Atoi(data)
	if err != nil {
		return nil, err
	}
	tx, err := user.Db.Begin(true)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	commitmentTransaction := &dao.CommitmentTransaction{}
	err = tx.One("Id", id, commitmentTransaction)
	if err != nil || commitmentTransaction.Id == 0 {
		return nil, err
	}
	if commitmentTransaction.CurrState != dao.TxInfoState_CreateAndSign {
		return nil, errors.New("wrong commitment state")
	}

	//region 广播承诺交易 最近的rsmc的资产分配交易 因为是omni资产，承诺交易被拆分成了两个独立的交易
	if tool.CheckIsString(&commitmentTransaction.RSMCTxHex) {
		commitmentTxid, err := rpcClient.SendRawTransaction(commitmentTransaction.RSMCTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxid)
	}
	if tool.CheckIsString(&commitmentTransaction.ToCounterpartyTxHex) {
		commitmentTxidToBob, err := rpcClient.SendRawTransaction(commitmentTransaction.ToCounterpartyTxHex)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(commitmentTxidToBob)
	}
	//endregion

	//region 广播RD
	latestRevocableDeliveryTx := &dao.RevocableDeliveryTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", commitmentTransaction.ChannelId),
		q.Eq("CommitmentTxId", commitmentTransaction.Id),
		q.Eq("Owner", user.PeerId)).
		OrderBy("CreateAt").Reverse().
		First(latestRevocableDeliveryTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = rpcClient.SendRawTransaction(latestRevocableDeliveryTx.TxHex)
	if err != nil {
		log.Println(err)
		msg := err.Error()
		//如果omnicore返回的信息里面包含了non-BIP68-final (code 64)， 则说明因为需要等待1000个区块高度，广播是对的
		if strings.Contains(msg, "non-BIP68-final (code 64)") == false &&
			strings.Contains(msg, "Code: -25,Msg: Missing inputs") == false {
			return nil, err
		}
	}
	//endregion

	// region update state
	commitmentTransaction.CurrState = dao.TxInfoState_SendHex
	commitmentTransaction.SendAt = time.Now()
	err = tx.Update(commitmentTransaction)
	if err != nil {
		return nil, err
	}

	latestRevocableDeliveryTx.CurrState = dao.TxInfoState_SendHex
	latestRevocableDeliveryTx.SendAt = time.Now()
	err = tx.Update(latestRevocableDeliveryTx)
	if err != nil {
		return nil, err
	}

	err = addRDTxToWaitDB(latestRevocableDeliveryTx)
	if err != nil {
		return nil, err
	}
	//endregion

	return nil, nil
}
