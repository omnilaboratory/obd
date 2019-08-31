package service

import (
	"LightningOnOmni/config"
	"LightningOnOmni/rpc"
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

func (service *ScheduleManager) StartSchudule() {
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
	client := rpc.NewClient()
	result, err := client.GetNetworkInfo()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(result)
}
