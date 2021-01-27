package agent

import (
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/dao"
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
