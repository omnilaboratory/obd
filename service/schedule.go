package service

import (
	"LightningOnOmni/config"
	"LightningOnOmni/dao"
	"github.com/asdine/storm/q"
	"log"
	"time"
)

type ScheduleManager struct {
	delay time.Duration
}

var ScheduleService = ScheduleManager{
	config.Schedule_Delay1,
}
var ticker = &time.Ticker{}

func (service *ScheduleManager) StartSchedule() {
	go func() {
		ticker = time.NewTicker(service.delay)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				log.Println(t)
				task()
			}
		}
	}()
}

func task() {
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
