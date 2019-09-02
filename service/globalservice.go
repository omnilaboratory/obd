package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
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
	TempAddress string
	AmountM     float64
	ToAddressB  string
	AmountB     float64
}

func init() {
	var err error
	db, err = dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
	}
	rpcClient = rpc.NewClient()
}

func createCommitmentATx(creatorSide int, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, outputBean commitmentOutputBean, user *bean.User) (*dao.CommitmentTxInfo, error) {
	commitmentTxInfo := &dao.CommitmentTxInfo{}
	commitmentTxInfo.PeerIdA = channelInfo.PeerIdA
	commitmentTxInfo.PeerIdB = channelInfo.PeerIdB
	commitmentTxInfo.ChannelId = channelInfo.ChannelId
	commitmentTxInfo.PropertyId = fundingTransaction.PropertyId
	commitmentTxInfo.CreatorSide = creatorSide

	//input
	commitmentTxInfo.InputTxid = fundingTransaction.FundingTxid
	commitmentTxInfo.InputVout = fundingTransaction.FundingOutputIndex
	commitmentTxInfo.InputAmount = fundingTransaction.AmountA + fundingTransaction.AmountB

	//output
	commitmentTxInfo.PubKey2 = outputBean.TempAddress
	commitmentTxInfo.PubKeyB = outputBean.ToAddressB
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{commitmentTxInfo.PubKey2, commitmentTxInfo.PubKeyB})
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.MultiAddress = gjson.Get(multiAddr, "address").String()
	commitmentTxInfo.RedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	commitmentTxInfo.AmountM = outputBean.AmountM
	commitmentTxInfo.AmountB = outputBean.AmountB

	commitmentTxInfo.CreateBy = user.PeerId
	commitmentTxInfo.CreateAt = time.Now()
	commitmentTxInfo.LastEditTime = time.Now()

	return commitmentTxInfo, nil
}
func createRDaTx(creatorSide int, channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTxInfo, toAddress string, user *bean.User) (*dao.RevocableDeliveryTransaction, error) {
	rda := &dao.RevocableDeliveryTransaction{}

	rda.CommitmentTxId = commitmentTxInfo.Id
	rda.PeerIdA = channelInfo.PeerIdA
	rda.PeerIdB = channelInfo.PeerIdB
	rda.ChannelId = channelInfo.ChannelId
	rda.PropertyId = commitmentTxInfo.PropertyId
	rda.CreatorSide = creatorSide

	//input
	rda.InputTxid = commitmentTxInfo.Txid
	rda.InputVout = 0
	rda.InputAmount = commitmentTxInfo.AmountM
	//output
	rda.PubKeyA = toAddress
	rda.Sequnence = 1000
	rda.Amount = commitmentTxInfo.AmountM

	rda.CreateBy = user.PeerId
	rda.CreateAt = time.Now()
	rda.LastEditTime = time.Now()

	return rda, nil
}
func createBRTx(creatorSide int, channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTxInfo, user *bean.User) (*dao.BreachRemedyTransaction, error) {
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}

	breachRemedyTransaction.CommitmentTxId = commitmentTxInfo.Id
	breachRemedyTransaction.PeerIdA = channelInfo.PeerIdA
	breachRemedyTransaction.PeerIdB = channelInfo.PeerIdB
	breachRemedyTransaction.ChannelId = channelInfo.ChannelId
	breachRemedyTransaction.PropertyId = commitmentTxInfo.PropertyId
	breachRemedyTransaction.CreatorSide = creatorSide

	//input
	breachRemedyTransaction.InputTxid = commitmentTxInfo.Txid
	breachRemedyTransaction.InputVout = 0
	breachRemedyTransaction.InputAmount = commitmentTxInfo.AmountM
	//output
	breachRemedyTransaction.PubKeyB = channelInfo.PubKeyB
	breachRemedyTransaction.Amount = commitmentTxInfo.AmountM

	breachRemedyTransaction.CreateBy = user.PeerId
	breachRemedyTransaction.CreateAt = time.Now()
	breachRemedyTransaction.LastEditTime = time.Now()

	return breachRemedyTransaction, nil
}
