package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/gin-gonic/gin"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/omnilaboratory/obd/tracker/dao"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type channelManager struct {
	mu           sync.Mutex
	BtcChainType string
}

var ChannelService channelManager

func (manager *channelManager) updateChannelInfo(obdClient *ObdNode, msgData string) (err error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if tool.CheckIsString(&msgData) == false {
		return errors.New("wrong inputData")
	}
	channelInfos := make([]bean.ChannelInfoRequest, 0)
	err = json.Unmarshal([]byte(msgData), &channelInfos)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, item := range channelInfos {
		if tool.CheckIsString(&item.ChannelId) == false ||
			tool.CheckIsString(&item.PeerIdA) == false ||
			tool.CheckIsString(&item.PeerIdB) == false {
			continue
		}
		channelInfo := &dao.ChannelInfo{}
		_ = db.Select(q.Eq("ChannelId", item.ChannelId)).First(channelInfo)
		if channelInfo.Id == 0 {
			channelInfo.ChannelId = item.ChannelId
			channelInfo.PropertyId = item.PropertyId
			channelInfo.CurrState = item.CurrState
			channelInfo.PeerIdA = item.PeerIdA
			channelInfo.PeerIdB = item.PeerIdB
			channelInfo.AmountA = item.AmountA
			channelInfo.AmountB = item.AmountB
			if item.IsAlice {
				channelInfo.ObdNodeIdA = obdClient.Id
			} else {
				channelInfo.ObdNodeIdB = obdClient.Id
			}
			channelInfo.CreateAt = time.Now()
			_ = db.Save(channelInfo)
		} else {
			channelInfo.PropertyId = item.PropertyId
			channelInfo.CurrState = item.CurrState
			channelInfo.PeerIdA = item.PeerIdA
			channelInfo.PeerIdB = item.PeerIdB
			channelInfo.AmountA = item.AmountA
			channelInfo.AmountB = item.AmountB
			if item.IsAlice {
				channelInfo.ObdNodeIdA = obdClient.Id
			} else {
				channelInfo.ObdNodeIdB = obdClient.Id
			}
			_ = db.Update(channelInfo)
		}
	}
	return err
}

func (manager *channelManager) GetChannelState(context *gin.Context) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	reqData := &bean.GetChannelStateRequest{}
	reqData.ChannelId = context.Query("channelId")
	if tool.CheckIsString(&reqData.ChannelId) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error channelId",
		})
		return
	}

	info := &dao.ChannelInfo{}
	_ = db.Select(q.Eq("ChannelId", reqData.ChannelId)).First(info)
	retData := make(map[string]interface{})
	retData["state"] = 0
	if info.Id > 0 {
		retData["state"] = info.CurrState
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "channelState",
		"data": retData,
	})
}

func (manager *channelManager) GetChannels(context *gin.Context) {
	pageNumStr := context.Query("pageNum")
	pageNum, _ := strconv.Atoi(pageNumStr)
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSizeStr := context.Query("pageSize")
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize <= 0 || pageSize > 20 {
		pageSize = 10
	}

	totalCount, _ := db.Count(&dao.ChannelInfo{})

	totalPage := totalCount / pageSize
	if totalCount%pageSize != 0 {
		totalPage += 1
	}
	if pageNum > totalPage {
		pageNum = totalPage
	}

	pageNum -= 1
	var infos []dao.ChannelInfo
	_ = db.Select(q.True()).OrderBy("Id").Reverse().Skip(pageNum * pageSize).Limit(pageSize).Find(&infos)
	context.JSON(http.StatusOK, gin.H{
		"data":       infos,
		"totalCount": totalCount,
		"totalPage":  totalPage,
		"pageNum":    pageNum + 1,
		"pageSize":   pageSize,
	})
}
