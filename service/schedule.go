package service

import (
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"github.com/asdine/storm/q"
	"log"
	"time"
)

type scheduleManager struct {
}

var ScheduleService = scheduleManager{}

func (service *scheduleManager) StartSchedule() {
	go func() {

		return

		ticker10m := time.NewTicker(10 * time.Minute)
		defer ticker10m.Stop()

		ticker := time.NewTicker(10 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker10m.C:
				log.Println("timer 10m", t)
				sendRdTx()
			case t := <-ticker.C:
				log.Println("timer 10s ", t)
			}
		}
	}()
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
				db.UpdateField(&node, "IsEnable", false)
				db.UpdateField(&node, "FinishAt", time.Now())
			}
		}
	}
}
