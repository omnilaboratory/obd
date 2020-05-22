package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/rpc"
	"github.com/omnilaboratory/obd/tool"
	trackerBean "github.com/omnilaboratory/obd/tracker/bean"
	"github.com/shopspring/decimal"
	"log"
	"time"

	"github.com/asdine/storm"
	"github.com/tidwall/gjson"
)

var db *storm.DB
var P2PLocalPeerId string
var rpcClient *rpc.Client
var TrackerWsConn *websocket.Conn

//for store the privateKey
var tempAddrPrivateKeyMap = make(map[string]string)
var OnlineUserMap = make(map[string]bool)

func FindUserIsOnline(peerId string) error {
	if tool.CheckIsString(&peerId) {
		value, exists := OnlineUserMap[peerId]
		if exists && value == true {
			return nil
		}
	}
	return errors.New(peerId + " not exist or online")
}

type commitmentOutputBean struct {
	AmountToRsmc               float64
	AmountToCounterparty       float64
	AmountToHtlc               float64
	RsmcTempPubKey             string
	HtlcTempPubKey             string
	OppositeSideChannelPubKey  string
	OppositeSideChannelAddress string
}

func init() {
	var err error
	db, err = dao.DBService.GetGlobalDB()
	if err != nil {
		log.Println(err)
	}
	rpcClient = rpc.NewClient()
}

func checkBtcFundFinish(address string) error {
	result, err := rpcClient.ListUnspent(address)
	if err != nil {
		return err
	}
	array := gjson.Parse(result).Array()
	log.Println("listunspent", array)
	if len(array) < 3 {
		return errors.New("btc fund have been not finished")
	}

	pMoney := config.GetOmniDustBtc()
	out, _ := decimal.NewFromFloat(config.GetMinerFee()).Add(decimal.NewFromFloat(pMoney)).Mul(decimal.NewFromFloat(4.0)).Float64()
	count := 0
	for _, item := range array {
		amount := item.Get("amount").Float()
		if amount >= out {
			count++
		}
	}
	if count < 3 {
		return errors.New("btc amount error, must greater " + tool.FloatToString(out, 8))
	}
	return nil
}

func getAddressFromPubKey(pubKey string) (address string, err error) {
	if tool.CheckIsString(&pubKey) == false {
		return "", errors.New("empty pubKey")
	}
	address, err = tool.GetAddressFromPubKey(pubKey)
	if err != nil {
		return "", err
	}
	_, err = rpcClient.ValidateAddress(address)
	if err != nil {
		return "", err
	}
	return address, nil
}

func createCommitmentTx(owner string, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, outputBean commitmentOutputBean, user *bean.User) (*dao.CommitmentTransaction, error) {
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
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{commitmentTxInfo.RSMCTempAddressPubKey, outputBean.OppositeSideChannelPubKey})
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.RSMCMultiAddress = gjson.Get(multiAddr, "address").String()
	commitmentTxInfo.RSMCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	json, err := rpcClient.GetAddressInfo(commitmentTxInfo.RSMCMultiAddress)
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.RSMCMultiAddressScriptPubKey = gjson.Get(json, "scriptPubKey").String()

	if tool.CheckIsString(&outputBean.HtlcTempPubKey) {
		commitmentTxInfo.HTLCTempAddressPubKey = outputBean.HtlcTempPubKey
		multiAddr, err := rpcClient.CreateMultiSig(2, []string{commitmentTxInfo.HTLCTempAddressPubKey, outputBean.OppositeSideChannelPubKey})
		if err != nil {
			return nil, err
		}
		commitmentTxInfo.HTLCMultiAddress = gjson.Get(multiAddr, "address").String()
		commitmentTxInfo.HTLCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
		json, err := rpcClient.GetAddressInfo(commitmentTxInfo.HTLCMultiAddress)
		if err != nil {
			return nil, err
		}
		commitmentTxInfo.HTLCMultiAddressScriptPubKey = gjson.Get(json, "scriptPubKey").String()
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
		err = errors.New("wrong Tx input vin")
		log.Println(err)
		return "", 0, 0, err
	}
	inTxid := jsonFundingTxHexDecode.Get("vin").Array()[0].Get("txid").String()
	inputTx, err := rpcClient.GetTransactionById(inTxid)
	if err != nil {
		err = errors.New("wrong input: " + err.Error())
		log.Println(err)
		return "", 0, 0, err
	}

	jsonInputTxDecode := gjson.Parse(inputTx)
	flag := false
	inputHexDecode, err := rpcClient.DecodeRawTransaction(jsonInputTxDecode.Get("hex").String())
	if err != nil {
		err = errors.New("wrong input: " + err.Error())
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
		err = errors.New("wrong vin " + jsonFundingTxHexDecode.Get("vin").String())
		log.Println(err)
		return "", 0, 0, err
	}

	//vout
	flag = false
	if jsonFundingTxHexDecode.Get("vout").IsArray() == false {
		err = errors.New("wrong Tx vout")
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
		err = errors.New("wrong vout " + jsonFundingTxHexDecode.Get("vout").String())
		log.Println(err)
		return "", 0, 0, err
	}
	return fundingTxid, amountA, fundingOutputIndex, err
}

func checkOmniTxHex(fundingTxHexDecode string, channelInfo *dao.ChannelInfo, user string) (fundingTxid string, amountA float64, propertyId int64, err error) {
	jsonOmniTxHexDecode := gjson.Parse(fundingTxHexDecode)
	fundingTxid = jsonOmniTxHexDecode.Get("txid").String()

	funderAddress := channelInfo.AddressA
	if user == channelInfo.PeerIdB {
		funderAddress = channelInfo.AddressB
	}

	sendingAddress := jsonOmniTxHexDecode.Get("sendingaddress").String()
	if sendingAddress != funderAddress {
		err = errors.New("wrong Tx input")
		log.Println(err)
		return "", 0, 0, err
	}
	referenceAddress := jsonOmniTxHexDecode.Get("referenceaddress").String()
	if referenceAddress != channelInfo.ChannelAddress {
		err = errors.New("wrong Tx output")
		log.Println(err)
		return "", 0, 0, err
	}

	amountA = jsonOmniTxHexDecode.Get("amount").Float()
	propertyId = jsonOmniTxHexDecode.Get("propertyid").Int()

	return fundingTxid, amountA, propertyId, err
}

//从未广播的交易hash数据中解析出他的输出，以此作为下个交易的输入
func getInputsForNextTxByParseTxHashVout(hex string, toAddress, scriptPubKey string) (inputs []rpc.TransactionInputItem, err error) {
	result, err := rpcClient.DecodeRawTransaction(hex)
	if err != nil {
		return nil, err
	}
	jsonHex := gjson.Parse(result)
	log.Println(jsonHex)
	if jsonHex.Get("vout").IsArray() {
		inputs = make([]rpc.TransactionInputItem, 0, 0)
		for _, item := range jsonHex.Get("vout").Array() {
			if item.Get("scriptPubKey").Get("addresses").Exists() {
				addresses := item.Get("scriptPubKey").Get("addresses").Array()
				for _, address := range addresses {
					if address.String() == toAddress {
						node := rpc.TransactionInputItem{}
						node.Txid = jsonHex.Get("txid").String()
						node.ScriptPubKey = scriptPubKey
						node.Vout = uint32(item.Get("n").Uint())
						node.Amount = item.Get("value").Float()
						if node.Amount > 0 {
							inputs = append(inputs, node)
						}
					}
				}
			}
		}
		return inputs, nil
	}
	return nil, errors.New("no inputs")
}

func getLatestCommitmentTx(channelId string, owner string) (commitmentTxInfo *dao.CommitmentTransaction, err error) {
	commitmentTxInfo = &dao.CommitmentTransaction{}
	err = db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("Owner", owner)).
		OrderBy("CreateAt").Reverse().First(commitmentTxInfo)
	return commitmentTxInfo, err
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
		q.Eq("CurrState", dao.ChannelState_CanUse)).
		First(channelInfo)
	if err != nil {
		return nil
	}
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

	aliceRdTxid := checkHexOutputAddressFromOmniDecode(signedRdHex, inputs, outputAddress)
	if aliceRdTxid == "" {
		return errors.New("rdtx has wrong output address")
	}
	rdTransaction, err := createRDTx(user.PeerId, channelInfo, &latestCcommitmentTxInfo, outputAddress, user)
	if err != nil {
		log.Println(err)
		return err
	}
	rdTransaction.RDType = 0
	rdTransaction.TxHex = signedRdHex
	rdTransaction.Txid = aliceRdTxid
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func signHTD1bTx(tx storm.Node, signedHtlcHex string, htd1bHex string, latestCcommitmentTxInfo dao.CommitmentTransaction, outputAddress string, user *bean.User) (err error) {
	inputs, err := getInputsForNextTxByParseTxHashVout(signedHtlcHex, latestCcommitmentTxInfo.HTLCMultiAddress, latestCcommitmentTxInfo.HTLCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return err
	}
	signedHtd1bTxid, signedHtd1bHex, err := rpcClient.OmniSignRawTransactionForUnsend(htd1bHex, inputs, tempAddrPrivateKeyMap[latestCcommitmentTxInfo.HTLCTempAddressPubKey])
	if err != nil {
		return err
	}
	result, err := rpcClient.TestMemPoolAccept(signedHtd1bHex)
	if err != nil {
		return err
	}
	if gjson.Parse(result).Array()[0].Get("allowed").Bool() == false {
		if gjson.Parse(result).Array()[0].Get("reject-reason").String() != "missing-inputs" {
			return errors.New(gjson.Parse(result).Array()[0].Get("reject-reason").String())
		}
	}

	owner := latestCcommitmentTxInfo.PeerIdA
	if user.PeerId == latestCcommitmentTxInfo.PeerIdA {
		owner = latestCcommitmentTxInfo.PeerIdB
	}

	htlcTimeOut := getHtlcTimeout(latestCcommitmentTxInfo.HtlcChannelPath, latestCcommitmentTxInfo.ChannelId)
	htlcTimeoutDeliveryTx := &dao.HTLCTimeoutDeliveryTxB{}
	htlcTimeoutDeliveryTx.ChannelId = latestCcommitmentTxInfo.ChannelId
	htlcTimeoutDeliveryTx.CommitmentTxId = latestCcommitmentTxInfo.Id
	htlcTimeoutDeliveryTx.PropertyId = latestCcommitmentTxInfo.PropertyId
	htlcTimeoutDeliveryTx.OutputAddress = outputAddress
	htlcTimeoutDeliveryTx.InputHex = latestCcommitmentTxInfo.HtlcTxHex
	htlcTimeoutDeliveryTx.OutAmount = latestCcommitmentTxInfo.AmountToHtlc
	htlcTimeoutDeliveryTx.Owner = owner
	htlcTimeoutDeliveryTx.CurrState = dao.TxInfoState_CreateAndSign
	htlcTimeoutDeliveryTx.CreateBy = user.PeerId
	htlcTimeoutDeliveryTx.Timeout = htlcTimeOut
	htlcTimeoutDeliveryTx.CreateAt = time.Now()

	htlcTimeoutDeliveryTx.Txid = signedHtd1bTxid
	htlcTimeoutDeliveryTx.TxHex = signedHtd1bHex
	err = tx.Save(htlcTimeoutDeliveryTx)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func createMultiSig(pubkey1 string, pubkey2 string) (multiAddress, redeemScript, scriptPubKey string, err error) {
	aliceRsmcMultiAddr, err := rpcClient.CreateMultiSig(2, []string{pubkey1, pubkey2})
	if err != nil {
		return "", "", "", err
	}
	multiAddress = gjson.Get(aliceRsmcMultiAddr, "address").String()
	redeemScript = gjson.Get(aliceRsmcMultiAddr, "redeemScript").String()
	tempJson, err := rpcClient.GetAddressInfo(multiAddress)
	if err != nil {
		return "", "", "", err
	}
	scriptPubKey = gjson.Get(tempJson, "scriptPubKey").String()
	return multiAddress, redeemScript, scriptPubKey, nil
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
		outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(reqData.Amount)).Float64()
		outputBean.OppositeSideChannelAddress = channelInfo.AddressB
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
	} else {
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(reqData.Amount)).Float64()
		outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(reqData.Amount)).Float64()
		outputBean.OppositeSideChannelAddress = channelInfo.AddressA
		outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
	}

	if lastCommitmentTx != nil && lastCommitmentTx.Id > 0 {
		if isSender {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(reqData.Amount)).Float64()
			outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToCounterparty).Add(decimal.NewFromFloat(reqData.Amount)).Float64()
		} else {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(reqData.Amount)).Float64()
			outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToCounterparty).Sub(decimal.NewFromFloat(reqData.Amount)).Float64()
		}
	}

	if lastCommitmentTx.TxType == dao.CommitmentTransactionType_Htlc {
		if lastCommitmentTx.HtlcSender == currUser.PeerId {
			outputBean.AmountToCounterparty, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToCounterparty).Add(decimal.NewFromFloat(lastCommitmentTx.AmountToHtlc)).Float64()
		} else {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Add(decimal.NewFromFloat(lastCommitmentTx.AmountToHtlc)).Float64()
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
	if commitmentTxInfo.AmountToCounterparty > 0 {
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
			commitmentTxInfo.AmountToCounterparty,
			0,
			0, &channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		commitmentTxInfo.ToCounterpartyTxid = txid
		commitmentTxInfo.ToCounterpartyTxHex = hex
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

func sendMsgToTracker(msgType enum.MsgType, data interface{}) {
	message := trackerBean.RequestMessage{}
	message.Type = msgType

	dataBytes, _ := json.Marshal(data)
	dataStr := string(dataBytes)
	parse := gjson.Parse(dataStr)
	result := parse.Value()
	if parse.Exists() == false {
		result = dataStr
	}
	message.Data = result

	bytes, _ := json.Marshal(message)
	err := TrackerWsConn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		log.Println("write:", err)
		return
	}
}
