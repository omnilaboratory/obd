package service

import (
	"github.com/asdine/storm"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type scheduleManager struct{}

var ScheduleService = scheduleManager{}

func (service *scheduleManager) StartSchedule() {
	go func() {
		ticker8m := time.NewTicker(8 * time.Minute)
		defer ticker8m.Stop()

		for {
			select {
			case t := <-ticker8m.C:
				log.Println("timer 8m", t)
				go sendRdTx()
				go checkBR()
			}
		}
	}()
}

//检查通道地址的金额是否变动了，根据交易的txid，广播br
func checkBR() {
	log.Println("checkBR")
	_dir := config.DataDirectory + "/" + config.ChainNodeType
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
				if channelInfo.CurrState == bean.ChannelState_CanUse || channelInfo.CurrState == bean.ChannelState_HtlcTx {
					_, _ = dealBrTx(&channelInfo, db)
				}
			}
		}
	}
}
