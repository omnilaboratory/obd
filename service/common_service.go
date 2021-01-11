package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/omnicore"
	"github.com/omnilaboratory/obd/tool"
	"github.com/shopspring/decimal"
	"log"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/tidwall/gjson"
)

func findUserIsOnline(nodePeerId, userPeerId string) error {
	if tool.CheckIsString(&userPeerId) {
		value, exists := OnlineUserMap[userPeerId]
		if exists && value != nil {
			return nil
		}
		if nodePeerId != P2PLocalNodeId {
			if conn2tracker.GetUserState(nodePeerId, userPeerId) > 0 {
				return nil
			}
		}
	}
	return errors.New(fmt.Sprintf(enum.Tips_user_notExistOrOnline, userPeerId))
}

func checkBtcFundFinish(tx storm.Node, channel dao.ChannelInfo, isFundOmni bool) error {
	if channel.CurrState > bean.ChannelState_WaitFundAsset {
		return nil
	}

	count, err := tx.Select(q.Eq("TemporaryChannelId", channel.TemporaryChannelId), q.Eq("IsFinish", true), q.Eq("Owner", channel.CreateBy)).Count(&dao.MinerFeeRedeemTransaction{})
	if err != nil {
		return err
	}
	//log.Println("listunspent", array)
	if count < config.BtcNeedFundTimes {
		return errors.New(enum.Tips_funding_notEnoughBtcFundingTime)
	}

	return nil
}

func getAddressFromPubKey(pubKey string) (address string, err error) {
	if tool.CheckIsString(&pubKey) == false {
		return "", errors.New(enum.Tips_common_empty + "pubKey")
	}
	address, err = tool.GetAddressFromPubKey(pubKey)
	if err != nil {
		return "", err
	}
	return address, nil
}

func createCommitmentTx(owner string, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, outputBean commitmentTxOutputBean, user *bean.User) (*dao.CommitmentTransaction, error) {
	commitmentTxInfo := &dao.CommitmentTransaction{}
	commitmentTxInfo.PeerIdA = channelInfo.PeerIdA
	commitmentTxInfo.PeerIdB = channelInfo.PeerIdB
	commitmentTxInfo.ChannelId = channelInfo.ChannelId
	commitmentTxInfo.PropertyId = fundingTransaction.PropertyId
	commitmentTxInfo.Owner = owner

	//input
	commitmentTxInfo.InputTxid = fundingTransaction.FundingTxid
	commitmentTxInfo.InputVout = fundingTransaction.FundingOutputIndex
	commitmentTxInfo.InputAmount = fundingTransaction.AmountA + fundingTransaction.AmountB

	//output to rsmc
	commitmentTxInfo.RSMCTempAddressPubKey = outputBean.RsmcTempPubKey
	multiAddr, err := omnicore.CreateMultiSig(2, []string{commitmentTxInfo.RSMCTempAddressPubKey, outputBean.OppositeSideChannelPubKey})
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.RSMCMultiAddress = gjson.Get(multiAddr, "address").String()
	commitmentTxInfo.RSMCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	commitmentTxInfo.RSMCMultiAddressScriptPubKey = gjson.Get(multiAddr, "scriptPubKey").String()

	if tool.CheckIsString(&outputBean.HtlcTempPubKey) {
		commitmentTxInfo.HTLCTempAddressPubKey = outputBean.HtlcTempPubKey
		multiAddr, err := omnicore.CreateMultiSig(2, []string{commitmentTxInfo.HTLCTempAddressPubKey, outputBean.OppositeSideChannelPubKey})
		if err != nil {
			return nil, err
		}
		commitmentTxInfo.HTLCMultiAddress = gjson.Get(multiAddr, "address").String()
		commitmentTxInfo.HTLCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
		commitmentTxInfo.HTLCMultiAddressScriptPubKey = gjson.Get(multiAddr, "scriptPubKey").String()
	}

	commitmentTxInfo.AmountToRSMC = outputBean.AmountToRsmc
	commitmentTxInfo.AmountToCounterparty = outputBean.AmountToCounterparty
	commitmentTxInfo.AmountToHtlc = outputBean.AmountToHtlc

	commitmentTxInfo.CreateBy = user.PeerId
	commitmentTxInfo.CreateAt = time.Now()
	commitmentTxInfo.LastEditTime = time.Now()

	return commitmentTxInfo, nil
}

func createRDTx(owner string, channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTransaction, toAddress string, user *bean.User) (*dao.RevocableDeliveryTransaction, error) {
	rda := &dao.RevocableDeliveryTransaction{}

	rda.CommitmentTxId = commitmentTxInfo.Id
	rda.PeerIdA = channelInfo.PeerIdA
	rda.PeerIdB = channelInfo.PeerIdB
	rda.ChannelId = channelInfo.ChannelId
	rda.PropertyId = commitmentTxInfo.PropertyId
	rda.Owner = owner

	//input
	rda.InputTxHex = commitmentTxInfo.RSMCTxHex
	rda.InputTxid = commitmentTxInfo.RSMCTxid
	rda.InputVout = 0
	rda.InputAmount = commitmentTxInfo.AmountToRSMC
	//output
	rda.OutputAddress = toAddress
	rda.Sequence = 1000
	rda.Amount = commitmentTxInfo.AmountToRSMC

	rda.CreateBy = user.PeerId
	rda.CreateAt = time.Now()
	rda.LastEditTime = time.Now()

	return rda, nil
}

func createBRTxObj(owner string, channelInfo *dao.ChannelInfo, brType dao.BRType, commitmentTxInfo *dao.CommitmentTransaction, user *bean.User) (*dao.BreachRemedyTransaction, error) {
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}
	breachRemedyTransaction.CommitmentTxId = commitmentTxInfo.Id
	breachRemedyTransaction.PeerIdA = channelInfo.PeerIdA
	breachRemedyTransaction.PeerIdB = channelInfo.PeerIdB
	breachRemedyTransaction.ChannelId = channelInfo.ChannelId
	breachRemedyTransaction.PropertyId = commitmentTxInfo.PropertyId
	breachRemedyTransaction.Owner = owner
	breachRemedyTransaction.Type = brType

	//input
	breachRemedyTransaction.TempPubKey = commitmentTxInfo.RSMCTempAddressPubKey
	breachRemedyTransaction.InputAddress = commitmentTxInfo.RSMCMultiAddress
	breachRemedyTransaction.InputAddressScriptPubKey = commitmentTxInfo.RSMCMultiAddressScriptPubKey
	breachRemedyTransaction.InputRedeemScript = commitmentTxInfo.RSMCRedeemScript
	breachRemedyTransaction.InputTxHex = commitmentTxInfo.RSMCTxHex
	breachRemedyTransaction.InputTxid = commitmentTxInfo.RSMCTxid
	breachRemedyTransaction.InputVout = 0
	breachRemedyTransaction.InputAmount = commitmentTxInfo.AmountToRSMC
	//output
	breachRemedyTransaction.Amount = commitmentTxInfo.AmountToRSMC

	breachRemedyTransaction.CreateBy = user.PeerId
	breachRemedyTransaction.CreateAt = time.Now()
	breachRemedyTransaction.LastEditTime = time.Now()

	return breachRemedyTransaction, nil
}

func checkBtcTxHex(btcFeeTxHexDecode string, channelInfo *dao.ChannelInfo, peerId string) (fundingTxid string, amountA float64, fundingOutputIndex uint32, err error) {
	jsonFundingTxHexDecode := gjson.Parse(btcFeeTxHexDecode)
	fundingTxid = jsonFundingTxHexDecode.Get("txid").String()

	//vin
	if jsonFundingTxHexDecode.Get("vin").IsArray() == false {
		err = errors.New(enum.Tips_funding_noVin)
		log.Println(err)
		return "", 0, 0, err
	}

	vin1 := jsonFundingTxHexDecode.Get("vin").Array()[0]
	asm := vin1.Get("scriptSig").Get("asm").Str
	if len(asm) == 0 {
		return "", 0, 0, errors.New("wrong vin asm")
	}
	split := strings.Split(asm, " ")
	if split[0] == "0" {
		return "", 0, 0, errors.New(enum.Tips_funding_noVin)
	}

	inTxid := vin1.Get("txid").String()
	inputTx := conn2tracker.GetTransactionById(inTxid)
	if err != nil {
		err = errors.New(enum.Tips_funding_wrongBtcHexVin + err.Error())
		log.Println(err)
		return "", 0, 0, err
	}

	jsonInputTxDecode := gjson.Parse(inputTx)
	flag := false
	inputHexDecode, err := omnicore.DecodeBtcRawTransaction(jsonInputTxDecode.Get("hex").String())
	if err != nil {
		err = errors.New(enum.Tips_funding_wrongBtcHexVin + err.Error())
		log.Println(err)
		return "", 0, 0, err
	}

	funderAddress := channelInfo.AddressA
	if peerId == channelInfo.PeerIdB {
		funderAddress = channelInfo.AddressB
	}
	jsonInputHexDecode := gjson.Parse(inputHexDecode)
	if jsonInputHexDecode.Get("vout").IsArray() {
		for _, item := range jsonInputHexDecode.Get("vout").Array() {
			addresses := item.Get("scriptPubKey").Get("addresses").Array()
			for _, subItem := range addresses {
				if subItem.String() == funderAddress {
					flag = true
					break
				}
			}
			if flag {
				break
			}
		}
	}

	if flag == false {
		log.Println(inputHexDecode)
		err = errors.New(enum.Tips_funding_wrongFunderAddressFromBtcHex)
		log.Println(err)
		return "", 0, 0, err
	}

	//vout
	flag = false
	if jsonFundingTxHexDecode.Get("vout").IsArray() == false {
		err = errors.New(enum.Tips_funding_notFoundVout)
		log.Println(err)
		return "", 0, 0, err
	}
	for _, item := range jsonFundingTxHexDecode.Get("vout").Array() {
		addresses := item.Get("scriptPubKey").Get("addresses").Array()
		for _, subItem := range addresses {
			if subItem.String() == channelInfo.ChannelAddress {
				amountA = item.Get("value").Float()
				fundingOutputIndex = uint32(item.Get("n").Int())
				flag = true
				break
			}
		}
		if flag {
			break
		}
	}
	if flag == false {
		log.Println(jsonFundingTxHexDecode)
		err = errors.New(enum.Tips_funding_wrongChannelAddressFromBtcHex)
		log.Println(err)
		return "", 0, 0, err
	}
	return fundingTxid, amountA, fundingOutputIndex, err
}

func checkOmniTxHex(fundingTxHexDecode string, channelInfo *dao.ChannelInfo, user string) (fundingTxid string, amountA float64, propertyId int64, err error) {
	jsonOmniTxHexDecode := gjson.Parse(fundingTxHexDecode)
	fundingTxid = jsonOmniTxHexDecode.Get("txid").String()

	funderAddress := channelInfo.FundingAddress

	sendingAddress := jsonOmniTxHexDecode.Get("sendingaddress").String()
	if sendingAddress != funderAddress {
		err = errors.New(fmt.Sprintf(enum.Tips_funding_wrongFunderAddressFromAssetHex, sendingAddress, funderAddress))
		log.Println(err)
		return "", 0, 0, err
	}
	referenceAddress := jsonOmniTxHexDecode.Get("referenceaddress").String()
	if referenceAddress != channelInfo.ChannelAddress {
		err = errors.New(fmt.Sprintf(enum.Tips_funding_wrongChannelAddressFromAssetHex, referenceAddress, channelInfo.ChannelAddress))
		log.Println(err)
		return "", 0, 0, err
	}

	amountA = jsonOmniTxHexDecode.Get("amount").Float()
	propertyId = jsonOmniTxHexDecode.Get("propertyid").Int()

	return fundingTxid, amountA, propertyId, err
}

//从未广播的交易hash数据中解析出他的输出，以此作为下个交易的输入
func getInputsForNextTxByParseTxHashVout(hex string, toAddress, scriptPubKey, redeemScript string) (inputs []bean.TransactionInputItem, err error) {
	result, err := omnicore.DecodeBtcRawTransaction(hex)
	if err != nil {
		return nil, err
	}
	jsonHex := gjson.Parse(result)
	//log.Println(jsonHex)
	if jsonHex.Get("vout").IsArray() {
		inputs = make([]bean.TransactionInputItem, 0, 0)
		for _, item := range jsonHex.Get("vout").Array() {
			if item.Get("scriptPubKey").Get("addresses").Exists() {
				addresses := item.Get("scriptPubKey").Get("addresses").Array()
				for _, address := range addresses {
					if address.String() == toAddress {
						node := bean.TransactionInputItem{}
						node.Txid = jsonHex.Get("txid").String()
						node.ScriptPubKey = scriptPubKey
						node.RedeemScript = redeemScript
						node.Vout = uint32(item.Get("n").Uint())
						node.Amount = item.Get("value").Float()
						if node.Amount > 0 {
							inputs = append(inputs, node)
						}
					}
				}
			}
		}
		if len(inputs) > 0 {
			return inputs, nil
		}
	}
	return nil, errors.New(enum.Tips_common_failToParseInputsFromUnsendTx)
}

func getLatestCommitmentTxUseDbTx(tx storm.Node, channelId string, owner string) (commitmentTxInfo *dao.CommitmentTransaction, err error) {
	commitmentTxInfo = &dao.CommitmentTransaction{}
	err = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", owner)).
		OrderBy("CreateAt").Reverse().First(commitmentTxInfo)
	return commitmentTxInfo, err
}

//根据通道id获取通道信息
func getChannelInfoByChannelId(tx storm.Node, channelId string, userPeerId string) (channelInfo *dao.ChannelInfo) {
	channelInfo = &dao.ChannelInfo{}
	err := tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", userPeerId),
			q.Eq("PeerIdB", userPeerId)),
		q.Or(
			q.Eq("CurrState", bean.ChannelState_CanUse),
			q.Eq("CurrState", bean.ChannelState_LockByTracker),
			q.Eq("CurrState", bean.ChannelState_NewTx))).
		First(channelInfo)
	if err != nil {
		return nil
	}
	return channelInfo
}

func getLockChannelForHtlc(tx storm.Node, channelId string, userPeerId string) (channelInfo *dao.ChannelInfo) {
	channelInfo = &dao.ChannelInfo{}
	_ = tx.Select(
		q.Eq("ChannelId", channelId),
		q.Or(
			q.Eq("PeerIdA", userPeerId),
			q.Eq("PeerIdB", userPeerId)),
		q.Or(
			q.Eq("CurrState", bean.ChannelState_CanUse),
			q.Eq("CurrState", bean.ChannelState_LockByTracker))).
		First(channelInfo)
	return channelInfo
}

//根据通道id获取通道信息
func getFundingTransactionByChannelId(dbTx storm.Node, channelId string, userPeerId string) (fundingTransaction *dao.FundingTransaction) {
	fundingTransaction = &dao.FundingTransaction{}
	err := dbTx.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("CurrState", dao.FundingTransactionState_Accept),
		q.Or(
			q.Eq("PeerIdA", userPeerId),
			q.Eq("PeerIdB", userPeerId))).
		OrderBy("CreateAt").
		Reverse().First(fundingTransaction)
	if err != nil {
		return nil
	}
	return fundingTransaction
}

func saveRdTx(tx storm.Node, channelInfo *dao.ChannelInfo, signedRsmcHex string, signedRdHex string, latestCommitmentTxInfo *dao.CommitmentTransaction, outputAddress string, user *bean.User) (err error) {
	inputs, err := getInputsForNextTxByParseTxHashVout(signedRsmcHex, latestCommitmentTxInfo.RSMCMultiAddress, latestCommitmentTxInfo.RSMCMultiAddressScriptPubKey, latestCommitmentTxInfo.RSMCRedeemScript)
	if err != nil || len(inputs) == 0 {
		log.Println(err)
		return err
	}

	aliceRdTxid := checkHexOutputAddressFromOmniDecode(signedRdHex, outputAddress)
	if aliceRdTxid == "" {
		return errors.New(enum.Tips_common_wrongAddressOfRD)
	}
	rdTransaction, err := createRDTx(user.PeerId, channelInfo, latestCommitmentTxInfo, outputAddress, user)
	if err != nil {
		log.Println(err)
		return err
	}
	rdTransaction.RDType = 0
	rdTransaction.TxHex = signedRdHex
	rdTransaction.Txid = aliceRdTxid
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	rdTransaction.SignAt = time.Now()
	err = tx.Save(rdTransaction)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func saveHTD1bTx(tx storm.Node, signedHtlcHex string, signedHtd1bHex string, latestCommitmentTxInfo dao.CommitmentTransaction, outputAddress string, user *bean.User) (err error) {
	inputs, err := getInputsForNextTxByParseTxHashVout(signedHtlcHex, latestCommitmentTxInfo.HTLCMultiAddress, latestCommitmentTxInfo.HTLCMultiAddressScriptPubKey, latestCommitmentTxInfo.HTLCRedeemScript)
	if err != nil || len(inputs) == 0 {
		log.Println(err)
		return err
	}

	owner := latestCommitmentTxInfo.PeerIdA
	if user.PeerId == latestCommitmentTxInfo.PeerIdA {
		owner = latestCommitmentTxInfo.PeerIdB
	}

	htlcTimeOut := latestCommitmentTxInfo.HtlcCltvExpiry
	htlcTimeoutDeliveryTx := &dao.HTLCTimeoutDeliveryTxB{}
	htlcTimeoutDeliveryTx.ChannelId = latestCommitmentTxInfo.ChannelId
	htlcTimeoutDeliveryTx.CommitmentTxId = latestCommitmentTxInfo.Id
	htlcTimeoutDeliveryTx.PropertyId = latestCommitmentTxInfo.PropertyId
	htlcTimeoutDeliveryTx.OutputAddress = outputAddress
	htlcTimeoutDeliveryTx.InputTxid = latestCommitmentTxInfo.HTLCTxid
	htlcTimeoutDeliveryTx.InputHex = latestCommitmentTxInfo.HtlcTxHex
	htlcTimeoutDeliveryTx.OutAmount = latestCommitmentTxInfo.AmountToHtlc
	htlcTimeoutDeliveryTx.Owner = owner
	htlcTimeoutDeliveryTx.CurrState = dao.TxInfoState_CreateAndSign
	htlcTimeoutDeliveryTx.CreateBy = user.PeerId
	htlcTimeoutDeliveryTx.Timeout = htlcTimeOut
	htlcTimeoutDeliveryTx.CreateAt = time.Now()

	htlcTimeoutDeliveryTx.Txid = omnicore.GetTxId(signedHtd1bHex)
	htlcTimeoutDeliveryTx.TxHex = signedHtd1bHex
	err = tx.Save(htlcTimeoutDeliveryTx)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func createMultiSig(pubkey1 string, pubkey2 string) (multiAddress, redeemScript, scriptPubKey string, err error) {
	aliceRsmcMultiAddr, err := omnicore.CreateMultiSig(2, []string{pubkey1, pubkey2})
	if err != nil {
		return "", "", "", err
	}
	multiAddress = gjson.Get(aliceRsmcMultiAddr, "address").String()
	redeemScript = gjson.Get(aliceRsmcMultiAddr, "redeemScript").String()
	scriptPubKey = gjson.Get(aliceRsmcMultiAddr, "scriptPubKey").String()
	return multiAddress, redeemScript, scriptPubKey, nil
}

func createCommitmentTxHex(dbTx storm.Node, isSender bool, reqData *bean.RequestCreateCommitmentTx, channelInfo *dao.ChannelInfo, lastCommitmentTx *dao.CommitmentTransaction, currUser bean.User) (commitmentTxInfo *dao.CommitmentTransaction, rawTx dao.CommitmentTxRawTx, err error) {
	//1、转账给bob的交易：输入：通道其中一个input，输出：给bob
	//2、转账后的余额的交易：输入：通道总的一个input,输出：一个多签地址，这个钱又需要后续的RD才能赎回
	// create Cna tx
	fundingTransaction := getFundingTransactionByChannelId(dbTx, channelInfo.ChannelId, currUser.PeerId)
	if fundingTransaction == nil {
		err = errors.New(enum.Tips_funding_notFoundFundAssetTx)
		return nil, rawTx, err
	}

	var outputBean = commitmentTxOutputBean{}
	outputBean.RsmcTempPubKey = reqData.CurrTempAddressPubKey
	if currUser.PeerId == channelInfo.PeerIdA {
		//default alice transfer to bob ,then alice minus money
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(reqData.Amount)).Round(8).Float64()
		outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(reqData.Amount)).Round(8).Float64()
		outputBean.OppositeSideChannelAddress = channelInfo.AddressB
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
	} else {
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(reqData.Amount)).Round(8).Float64()
		outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(reqData.Amount)).Round(8).Float64()
		outputBean.OppositeSideChannelAddress = channelInfo.AddressA
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
	}

	if lastCommitmentTx != nil && lastCommitmentTx.Id > 0 {
		if isSender {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(reqData.Amount)).Round(8).Float64()
			outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToCounterparty).Add(decimal.NewFromFloat(reqData.Amount)).Round(8).Float64()
		} else {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(reqData.Amount)).Round(8).Float64()
			outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToCounterparty).Sub(decimal.NewFromFloat(reqData.Amount)).Round(8).Float64()
		}
	}

	if lastCommitmentTx.TxType == dao.CommitmentTransactionType_Htlc {
		if lastCommitmentTx.CurrState == dao.TxInfoState_Htlc_GetH {
			if lastCommitmentTx.HtlcSender == currUser.PeerId {
				outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentTx.AmountToHtlc)).Round(8).Float64()
			} else {
				outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToCounterparty).Add(decimal.NewFromFloat(lastCommitmentTx.AmountToHtlc)).Round(8).Float64()
			}
			//	TxInfoState_Htlc_GetR State
		} else {
			if lastCommitmentTx.HtlcSender == currUser.PeerId {
				outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToCounterparty).Add(decimal.NewFromFloat(lastCommitmentTx.AmountToHtlc)).Round(8).Float64()
			} else {
				outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentTx.AmountToHtlc)).Round(8).Float64()
			}
		}
	}
	commitmentTxInfo, err = createCommitmentTx(currUser.PeerId, channelInfo, fundingTransaction, outputBean, &currUser)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}

	listUnspent, err := GetAddressListUnspent(dbTx, *channelInfo)
	if err != nil {
		return nil, rawTx, err
	}

	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Rsmc
	commitmentTxInfo.RSMCTempAddressIndex = reqData.CurrTempAddressIndex
	rawTx = dao.CommitmentTxRawTx{}
	usedTxidTemp := ""
	if commitmentTxInfo.AmountToRSMC > 0 {
		rsmcTxData, usedTxid, err := omnicore.OmniCreateRawTransactionUseSingleInput(
			listUnspent,
			channelInfo.ChannelAddress,
			commitmentTxInfo.RSMCMultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, rawTx, err
		}

		usedTxidTemp = usedTxid
		commitmentTxInfo.RsmcInputTxid = usedTxid
		commitmentTxInfo.RSMCTxHex = rsmcTxData["hex"].(string)

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = commitmentTxInfo.RSMCTxHex
		signHexData.Inputs = rsmcTxData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = channelInfo.PubKeyA
		signHexData.PubKeyB = channelInfo.PubKeyB
		rawTx.RsmcRawTxData = signHexData
	}

	//create to Counterparty tx
	if commitmentTxInfo.AmountToCounterparty > 0 {
		toBobTxData, err := omnicore.OmniCreateRawTransactionUseRestInput(
			int(commitmentTxInfo.TxType),
			listUnspent,
			channelInfo.ChannelAddress,
			usedTxidTemp,
			outputBean.OppositeSideChannelAddress,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToCounterparty,
			getBtcMinerAmount(channelInfo.BtcAmount),
			&channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, rawTx, err
		}
		commitmentTxInfo.ToCounterpartyTxHex = toBobTxData["hex"].(string)

		signHexData := bean.NeedClientSignTxData{}
		signHexData.Hex = commitmentTxInfo.ToCounterpartyTxHex
		signHexData.Inputs = toBobTxData["inputs"]
		signHexData.IsMultisig = true
		signHexData.PubKeyA = channelInfo.PubKeyA
		signHexData.PubKeyB = channelInfo.PubKeyB
		rawTx.ToCounterpartyRawTxData = signHexData
	}
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	if lastCommitmentTx != nil && lastCommitmentTx.Id > 0 {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentTx.Id
		commitmentTxInfo.LastHash = lastCommitmentTx.CurrHash
	}
	commitmentTxInfo.CurrState = dao.TxInfoState_Create
	err = dbTx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}

	rawTx.CommitmentTxId = commitmentTxInfo.Id
	_ = dbTx.Save(rawTx)

	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = dbTx.Update(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, rawTx, err
	}
	return commitmentTxInfo, rawTx, nil
}

func GetBtcMinerFundMiniAmount() float64 {
	out, _ := decimal.NewFromFloat(omnicore.GetMinerFee()).Add(decimal.NewFromFloat(2 * tool.GetOmniDustBtc())).Mul(decimal.NewFromFloat(4.0)).Round(8).Float64()
	return out
}

func getBtcMinerAmount(total float64) float64 {
	return tool.GetBtcMinerAmount(total)
}

func checkChannelOmniAssetAmount(channelInfo dao.ChannelInfo) (bool, error) {
	balance := conn2tracker.GetOmniBalance(channelInfo.ChannelAddress, int(channelInfo.PropertyId))
	if balance == channelInfo.Amount {
		return true, nil
	}
	return false, nil
}

func GetAddressListUnspent(tx storm.Node, channelInfo dao.ChannelInfo) (listUnspentResult string, err error) {
	listUnspentArr := []dao.ChannelAddressListUnspent{}
	tx.Select(q.Eq("ChannelId", channelInfo.ChannelId)).Find(&listUnspentArr)
	if len(listUnspentArr) < 4 {
		btcRequestArr := []dao.FundingBtcRequest{}
		_ = tx.Select(q.Eq("TemporaryChannelId", channelInfo.TemporaryChannelId), q.Eq("IsFinish", true)).Find(&btcRequestArr)
		if len(btcRequestArr) != 3 {
			return "", errors.New("error btc fund")
		}
		for _, item := range btcRequestArr {
			outputFromHex := getOutputFromHex(item.TxHash, channelInfo.ChannelAddress)
			outputFromHex.ChannelId = channelInfo.ChannelId
			tx.Save(outputFromHex)
		}
		omniRequestArr := []dao.FundingTransaction{}
		_ = tx.Select(q.Eq("TemporaryChannelId", channelInfo.TemporaryChannelId)).Find(&omniRequestArr)
		if len(omniRequestArr) != 1 {
			return "", errors.New("error omni fund")
		}
		for _, item := range omniRequestArr {
			outputFromHex := getOutputFromHex(item.FundingTxHex, channelInfo.ChannelAddress)
			outputFromHex.ChannelId = channelInfo.ChannelId
			tx.Save(outputFromHex)
		}
		tx.Select(q.Eq("ChannelId", channelInfo.ChannelId)).Find(&listUnspentArr)
	}
	marshal, _ := json.Marshal(&listUnspentArr)
	return string(marshal), nil
}

func getOutputFromHex(hex string, toAddress string) *dao.ChannelAddressListUnspent {
	result, err := omnicore.DecodeBtcRawTransaction(hex)
	if err != nil {
		return nil
	}
	jsonHex := gjson.Parse(result)
	if jsonHex.Get("vout").IsArray() {
		for _, item := range jsonHex.Get("vout").Array() {
			if item.Get("scriptPubKey").Get("addresses").Exists() {
				addresses := item.Get("scriptPubKey").Get("addresses").Array()
				for _, address := range addresses {
					if address.String() == toAddress {
						if item.Get("value").Float() > 0 {
							node := &dao.ChannelAddressListUnspent{}
							node.Txid = jsonHex.Get("txid").String()
							node.ScriptPubKey = item.Get("scriptPubKey").Get("hex").Str
							node.Vout = uint32(item.Get("n").Uint())
							node.Amount = item.Get("value").Float()
							return node
						}
					}
				}
			}
		}
	}
	return nil
}

func verifyCompleteSignHex(inputs interface{}, signedHex string) error {
	var items []bean.RawTxInputItem

	var inputArr []interface{}
	switch inputs.(type) {
	case []interface{}:
		inputArr = inputs.([]interface{})
	case []map[string]interface{}:
		for _, temp := range inputs.([]map[string]interface{}) {
			inputArr = append(inputArr, temp)
		}
	}
	for _, temp := range inputArr {
		item := temp.(map[string]interface{})
		inputItem := bean.RawTxInputItem{}
		inputItem.ScriptPubKey = item["scriptPubKey"].(string)
		inputItem.RedeemScript = item["redeemScript"].(string)
		items = append(items, inputItem)
	}
	if omnicore.VerifySignatureHex(items, signedHex) != nil {
		return errors.New(fmt.Sprintf(enum.Tips_common_failToSign, "signed_hex"))
	}
	return nil
}
