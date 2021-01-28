package agent

import (
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
	"log"
)

func getLatestCommitmentTx(channelId string, user bean.User) *dao.CommitmentTransaction {
	commitmentTransaction := &dao.CommitmentTransaction{}
	user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).
		OrderBy("Id").Reverse().
		First(commitmentTransaction)
	return commitmentTransaction
}
func getCurrCommitmentTx(channelId string, user bean.User) *dao.CommitmentTransaction {
	commitmentTransaction := &dao.CommitmentTransaction{}
	err := user.Db.Select(
		q.Eq("ChannelId", channelId),
		q.Eq("CurrState", dao.TxInfoState_Create)).
		OrderBy("Id").Reverse().
		First(commitmentTransaction)
	if err != nil {
		log.Println(err)
		return nil
	}
	return commitmentTransaction
}
