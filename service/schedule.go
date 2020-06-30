package service

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type scheduleManager struct{}

var ScheduleService = scheduleManager{}

func (service *scheduleManager) StartSchedule() {
	go func() {
		ticker10m := time.NewTicker(10 * time.Minute)
		defer ticker10m.Stop()

		//ticker := time.NewTicker(10 * time.Hour)
		//defer ticker.Stop()

		for {
			select {
			case t := <-ticker10m.C:
				log.Println("timer 10m", t)
				sendRdTx()
				checkBR()
			}
		}
	}()
}

//检查通道地址的金额是否变动了，根据交易的txid，广播br
func checkBR() {
	_dir := "dbdata" + config.ChainNode_Type
	files, _ := ioutil.ReadDir(_dir)
	dbNames := make([]string, 0)
	userPeerIds := make([]string, 0)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "user_") && strings.HasSuffix(f.Name(), ".db") {
			peerId := strings.TrimPrefix(f.Name(), "user_")
			peerId = strings.TrimSuffix(peerId, ".db")
			value, exists := OnlineUserMap[peerId]
			if exists && value != nil {
				userPeerIds = append(userPeerIds, peerId)
			} else {
				dbNames = append(dbNames, f.Name())
			}
		}
	}

	for _, peerId := range userPeerIds {
		user, _ := OnlineUserMap[peerId]
		if user != nil {
			checkRsmcAndSendBR(user.Db)
		}
	}

	for _, dbName := range dbNames {
		db, err := storm.Open(_dir + "/" + dbName)
		if err == nil {
			checkRsmcAndSendBR(db)
			_ = db.Close()
		}
	}
}

func checkRsmcAndSendBR(db storm.Node) {
	var channelInfos []dao.ChannelInfo
	err := db.All(&channelInfos)
	if err == nil {
		for _, channelInfo := range channelInfos {
			if len(channelInfo.ChannelId) > 0 {
				if channelInfo.CurrState == dao.ChannelState_CanUse || channelInfo.CurrState == dao.ChannelState_HtlcTx {
					result, err := rpcClient.OmniGetbalance(channelInfo.ChannelAddress, int(channelInfo.PropertyId))
					if err != nil {
						continue
					}
					balance := gjson.Get(result, "balance").Float()
					if balance < channelInfo.Amount {
						transactionsStr, err := rpcClient.OmniListTransactions(channelInfo.ChannelAddress, 100, 1)
						if err != nil {
							continue
						}
						transactions := gjson.Parse(transactionsStr).Array()
						for _, item := range transactions {
							txid := item.Get("txid").Str
							if tool.CheckIsString(&txid) == false {
								continue
							}
							breachRemedy := &dao.BreachRemedyTransaction{}
							_ = db.Select(q.Eq("CurrState", dao.TxInfoState_CreateAndSign), q.Eq("InputTxid", txid)).First(breachRemedy)
							if breachRemedy.Id > 0 {
								_, err = rpcClient.SendRawTransaction(breachRemedy.BrTxHex)
								if err != nil {
									log.Println("send br by timer")
									breachRemedy.CurrState = dao.TxInfoState_SendHex
									breachRemedy.SendAt = time.Now()
									_ = db.Update(breachRemedy)
								}
							}
						}
					}
				}
			}
		}
	}
}

func sendRdTx() {
	var nodes []dao.RDTxWaitingSend
	err := db.Select(q.Eq("IsEnable", true)).Find(&nodes)
	if err != nil {
		return
	}

	for _, node := range nodes {
		if tool.CheckIsString(&node.TransactionHex) {
			_, err := rpcClient.SendRawTransaction(node.TransactionHex)
			if err == nil {
				if node.Type == 1 {
					_ = addHTRD1aTxToWaitDB(node.HtnxIdAndHtnxRdId)
				}
				_ = db.UpdateField(&node, "IsEnable", false)
				_ = db.UpdateField(&node, "FinishAt", time.Now())
			}
		}
	}
}
