package service

import (
	"github.com/asdine/storm/q"
	"log"
	"obd/dao"
	"obd/tool"
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
				//case t := <-ticker.C:
				//	log.Println("timer 10s ", t)
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
				if node.Type == 1 {
					addHTRD1aTxToWaitDB(node.HtnxIdAndHtnxRdId)
				}
				_ = db.UpdateField(&node, "IsEnable", false)
				_ = db.UpdateField(&node, "FinishAt", time.Now())
			}
		}
	}
}
