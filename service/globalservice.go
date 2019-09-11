package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"errors"
	"log"
	"time"

	"github.com/asdine/storm"
	"github.com/tidwall/gjson"
)

var db *storm.DB
var rpcClient *rpc.Client

//for store the privateKey of -351
var tempAddrPrivateKeyMap = make(map[string]string)

type commitmentOutputBean struct {
	TempPubKey string
	AmountM    float64
	ToPubKey   string
	ToAddress  string
	AmountB    float64
}

func init() {
	var err error
	db, err = dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
	}
	rpcClient = rpc.NewClient()
}

func getAddressFromPubKey(pubKey string) (address string, err error) {
	if tool.CheckIsString(&pubKey) {
		return "", errors.New("empty pubKey")
	}
	address, err = tool.GetAddressFromPubKey(pubKey)
	if err != nil {
		return "", err
	}
	isValid, err := rpcClient.ValidateAddress(address)
	if err != nil {
		return "", err
	}
	if isValid == false {
		return "", errors.New("invalid fundingPubKey")
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

	//output
	commitmentTxInfo.TempAddressPubKey = outputBean.TempPubKey
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{commitmentTxInfo.TempAddressPubKey, outputBean.ToPubKey})
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.MultiAddress = gjson.Get(multiAddr, "address").String()
	commitmentTxInfo.RedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	json, err := rpcClient.GetAddressInfo(commitmentTxInfo.MultiAddress)
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.ScriptPubKey = gjson.Get(json, "scriptPubKey").String()

	commitmentTxInfo.AmountM = outputBean.AmountM
	commitmentTxInfo.AmountB = outputBean.AmountB

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
	rda.InputTxid = commitmentTxInfo.Txid
	rda.InputVout = 0
	rda.InputAmount = commitmentTxInfo.AmountM
	//output
	rda.OutputAddress = toAddress
	rda.Sequence = 1000
	rda.Amount = commitmentTxInfo.AmountM

	rda.CreateBy = user.PeerId
	rda.CreateAt = time.Now()
	rda.LastEditTime = time.Now()

	return rda, nil
}
func createBRTx(owner string, channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTransaction, user *bean.User) (*dao.BreachRemedyTransaction, error) {
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}
	breachRemedyTransaction.CommitmentTxId = commitmentTxInfo.Id
	breachRemedyTransaction.PeerIdA = channelInfo.PeerIdA
	breachRemedyTransaction.PeerIdB = channelInfo.PeerIdB
	breachRemedyTransaction.ChannelId = channelInfo.ChannelId
	breachRemedyTransaction.PropertyId = commitmentTxInfo.PropertyId
	breachRemedyTransaction.Owner = owner

	//input
	breachRemedyTransaction.InputTxid = commitmentTxInfo.Txid
	breachRemedyTransaction.InputVout = 0
	breachRemedyTransaction.InputAmount = commitmentTxInfo.AmountM
	//output
	breachRemedyTransaction.Amount = commitmentTxInfo.AmountM

	breachRemedyTransaction.CreateBy = user.PeerId
	breachRemedyTransaction.CreateAt = time.Now()
	breachRemedyTransaction.LastEditTime = time.Now()

	return breachRemedyTransaction, nil
}
