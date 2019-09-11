package service

import (
	"LightningOnOmni/dao"
	"github.com/asdine/storm/q"
	"log"
	"time"
)

type scheduleManager struct {
}

var ScheduleService = scheduleManager{}

func (service *scheduleManager) StartSchedule() {
	go func() {

		ticker10m := time.NewTicker(10 * time.Minute)
		defer ticker10m.Stop()

		ticker := time.NewTicker(10 * time.Second)
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
		_, err := rpcClient.SendRawTransaction(node.TransactionHex)
		if err == nil {
			node.IsEnable = false
			node.FinishAt = time.Now()
			db.Save(&node)
		}
	}
}
