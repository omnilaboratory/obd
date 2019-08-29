package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"github.com/asdine/storm"
	"github.com/tidwall/gjson"
	"log"
	"time"
)

var db *storm.DB
var rpcClient *rpc.Client

func init() {
	var err error
	db, err = dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
	}
	rpcClient = rpc.NewClient()
}

func createCommitmentATx(channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, user *bean.User) (*dao.CommitmentTxInfo, error) {
	commitmentTxInfo := &dao.CommitmentTxInfo{}
	commitmentTxInfo.PeerIdA = channelInfo.PeerIdA
	commitmentTxInfo.PeerIdB = channelInfo.PeerIdB
	commitmentTxInfo.ChannelId = channelInfo.ChannelId
	commitmentTxInfo.PropertyId = fundingTransaction.PropertyId
	commitmentTxInfo.CreateSide = 0

	//input
	commitmentTxInfo.InputTxid = fundingTransaction.FundingTxid
	commitmentTxInfo.InputVout = fundingTransaction.FundingOutputIndex
	commitmentTxInfo.InputAmount = fundingTransaction.AmountA + fundingTransaction.AmountB

	//output
	commitmentTxInfo.PubKeyA2 = fundingTransaction.FunderPubKey2ForCommitment
	commitmentTxInfo.PubKeyB = channelInfo.PubKeyB
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{commitmentTxInfo.PubKeyA2, commitmentTxInfo.PubKeyB})
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.MultiAddress = gjson.Get(multiAddr, "address").String()
	commitmentTxInfo.RedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	commitmentTxInfo.AmountM = fundingTransaction.AmountA
	commitmentTxInfo.AmountB = fundingTransaction.AmountB

	commitmentTxInfo.CreateBy = user.PeerId
	commitmentTxInfo.CreateAt = time.Now()
	commitmentTxInfo.LastEditTime = time.Now()

	return commitmentTxInfo, nil
}
func createRDaTx(channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTxInfo, user *bean.User) (*dao.RevocableDeliveryTransaction, error) {
	rda := &dao.RevocableDeliveryTransaction{}
	rda.PeerIdA = channelInfo.PeerIdA
	rda.PeerIdB = channelInfo.PeerIdB
	rda.ChannelId = channelInfo.ChannelId
	rda.PropertyId = commitmentTxInfo.PropertyId
	rda.CreateSide = 0

	//input
	rda.InputTxid = commitmentTxInfo.Txid
	rda.InputVout = 0
	rda.InputAmount = commitmentTxInfo.AmountM
	//output
	rda.PubKeyA = channelInfo.PubKeyA
	rda.Sequnence = 1000
	rda.Amount = commitmentTxInfo.AmountM

	rda.CreateBy = user.PeerId
	rda.CreateAt = time.Now()
	rda.LastEditTime = time.Now()

	return rda, nil
}
func createBRaTx(channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTxInfo, user *bean.User) (*dao.BreachRemedyTransaction, error) {
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}
	breachRemedyTransaction.PeerIdA = channelInfo.PeerIdA
	breachRemedyTransaction.PeerIdB = channelInfo.PeerIdB
	breachRemedyTransaction.ChannelId = channelInfo.ChannelId
	breachRemedyTransaction.PropertyId = commitmentTxInfo.PropertyId
	breachRemedyTransaction.CreateSide = 0

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
