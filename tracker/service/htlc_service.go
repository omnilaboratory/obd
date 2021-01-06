package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/gin-gonic/gin"
	cbean "github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/omnilaboratory/obd/tracker/config"
	"github.com/omnilaboratory/obd/tracker/dao"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type graphEdge struct {
	ParentNodeIndex int
	PathPeerIds     string
	ChannelIds      string
	PathIndexArr    []int
	ChannelId       string
	Level           uint16
	CurrNodePeerId  string
	IsRoot          bool
	IsTarget        bool
}

type htlcManager struct {
	mu       sync.Mutex
	openList []*graphEdge
}

var HtlcService htlcManager

func (manager *htlcManager) getPath(obdClient *ObdNode, msgData string) (path interface{}, err error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if tool.CheckIsString(&msgData) == false {
		return "", errors.New("wrong inputData")
	}
	pathRequest := &bean.HtlcPathRequest{}
	err = json.Unmarshal([]byte(msgData), pathRequest)
	if err != nil {
		return "", err
	}

	if tool.CheckIsString(&pathRequest.RealPayerPeerId) == false {
		return "", errors.New("wrong realPayerPeerId")
	}
	if tool.CheckIsString(&pathRequest.PayeePeerId) == false {
		return "", errors.New("wrong SendeePeerId")
	}
	if pathRequest.Amount < tool.GetOmniDustBtc() {
		return "", errors.New("wrong amount")
	}

	manager.createChannelNetwork(pathRequest.RealPayerPeerId, pathRequest.PayeePeerId, pathRequest.PropertyId, pathRequest.Amount, nil, true)
	resultIndex := manager.getPathIndex()

	retNode := make(map[string]interface{})
	retNode["senderPeerId"] = pathRequest.RealPayerPeerId
	retNode["h"] = pathRequest.H
	retNode["amount"] = pathRequest.Amount
	retNode["path"] = ""
	if resultIndex != -1 {
		splitArr := strings.Split(manager.openList[resultIndex].ChannelIds, ",")
		path := ""
		for i := len(splitArr) - 1; i > -1; i-- {
			path += splitArr[i] + ","
		}
		path = strings.TrimSuffix(path, ",")
		retNode["path"] = path
	}
	return retNode, nil
}

func (manager *htlcManager) getPathIndex() int {
	resultIndex := -1
	minLength := 99
	for index, node := range manager.openList {
		if node.IsTarget {
			log.Println(node.ChannelIds)
			tempLength := len(strings.Split(node.ChannelIds, ","))
			if tempLength < minLength {
				resultIndex = index
				minLength = tempLength
			}
		}
	}
	if resultIndex != -1 {
		channelIdArr := strings.Split(manager.openList[resultIndex].ChannelIds, ",")
		locaResult := true
		for _, item := range channelIdArr {
			channelInfo := &dao.ChannelInfo{}
			err := db.Select(q.Eq("ChannelId", item), q.Eq("CurrState", cbean.ChannelState_CanUse)).First(channelInfo)
			if err == nil {
				if len(channelInfo.ObdNodeIdA) > 0 && sendChannelLockInfoToObd(channelInfo.ChannelId, channelInfo.PeerIdA, channelInfo.ObdNodeIdA) == false {
					locaResult = false
					break
				}

				if len(channelInfo.ObdNodeIdB) > 0 && sendChannelLockInfoToObd(channelInfo.ChannelId, channelInfo.PeerIdB, channelInfo.ObdNodeIdB) == false {
					locaResult = false
					break
				}
			}
		}
		if locaResult == false {
			manager.openList = append(manager.openList[:resultIndex], manager.openList[resultIndex+1:]...)
			resultIndex = manager.getPathIndex()
		}
	}
	return resultIndex
}

func (manager *htlcManager) updateHtlcInfo(obdClient *ObdNode, msgData string) (err error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if tool.CheckIsString(&msgData) == false {
		return errors.New("wrong inputData")
	}
	reqData := &bean.UpdateHtlcTxStateRequest{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		return err
	}

	if tool.CheckIsString(&reqData.Path) == false {
		return errors.New("path")
	}
	if tool.CheckIsString(&reqData.H) == false {
		return errors.New("h")
	}
	if tool.CheckIsString(&reqData.CurrChannelId) == false {
		return errors.New("currChannelId")
	}

	htlcTxInfo := &dao.HtlcTxInfo{}
	_ = db.Select(q.Eq("Path", reqData.Path), q.Eq("H", reqData.H)).First(htlcTxInfo)
	if htlcTxInfo.Id == 0 {
		htlcTxInfo.Path = reqData.Path
		htlcTxInfo.H = reqData.H
		htlcTxInfo.DirectionFlag = bean.HtlcTxState_PayMoney
		htlcTxInfo.CurrChannelId = reqData.CurrChannelId
		htlcTxInfo.CreateAt = time.Now()
		_ = db.Save(htlcTxInfo)
	} else {
		if tool.CheckIsString(&htlcTxInfo.R) == false && tool.CheckIsString(&reqData.R) {
			htlcTxInfo.R = reqData.R
		}
		htlcTxInfo.DirectionFlag = reqData.DirectionFlag
		htlcTxInfo.CurrChannelId = reqData.CurrChannelId
		_ = db.Update(htlcTxInfo)
	}
	return nil
}

func (manager *htlcManager) GetHtlcCurrState(context *gin.Context) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	reqData := &bean.GetHtlcTxStateRequest{}
	reqData.Path = context.Query("path")
	if tool.CheckIsString(&reqData.Path) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error path",
		})
		return
	}
	reqData.H = context.Query("h")
	if tool.CheckIsString(&reqData.H) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error h",
		})
		return
	}

	htlcTxInfo := &dao.HtlcTxInfo{}
	_ = db.Select(q.Eq("Path", reqData.Path), q.Eq("H", reqData.H)).First(htlcTxInfo)
	retData := make(map[string]interface{})
	retData["flag"] = 0
	if htlcTxInfo.Id > 0 {
		retData["flag"] = htlcTxInfo.DirectionFlag
	}
	context.JSON(http.StatusOK, gin.H{
		"msg":  "htlcState",
		"data": retData,
	})
}

func (manager *htlcManager) createChannelNetwork(realPayerPeerId, currPayeePeerId string, propertyId int64, amount float64, currNode *graphEdge, isBegin bool) {
	if isBegin {
		manager.openList = make([]*graphEdge, 0)
		realPayeeEdge := &graphEdge{
			ParentNodeIndex: -1,
			CurrNodePeerId:  currPayeePeerId,
			PathPeerIds:     "",
			ChannelIds:      "",
			PathIndexArr:    make([]int, 0),
			Level:           0,
			IsRoot:          true,
			IsTarget:        false,
		}
		manager.openList = append(manager.openList, realPayeeEdge)
	}

	if currNode == nil {
		currNode = manager.openList[0]
	}
	//如果当前用户已经在备选的用户列表中出现了，就return
	if strings.Contains(currNode.PathPeerIds, currNode.CurrNodePeerId) {
		return
	}

	if currNode.Level > 0 {
		amount, _ = decimal.NewFromFloat(amount).Mul(decimal.NewFromFloat(1 + cfg.HtlcFeeRate*float64(currNode.Level))).Round(8).Float64()
	}

	currNodeIndex := 0
	for index, tempNode := range manager.openList {
		if tempNode == currNode {
			currNodeIndex = index
			break
		}
	}

	pathPeerIds := currNode.PathPeerIds + "," + currNode.CurrNodePeerId
	pathIndexArr := currNode.PathIndexArr
	if currNode.PathPeerIds == "" {
		pathPeerIds = currNode.CurrNodePeerId
		pathIndexArr = make([]int, 0)
	}
	pathIndexArr = append(pathIndexArr, currNodeIndex)

	var nodes []dao.ChannelInfo
	err := errors.New("no channel")

	err = db.Select(
		q.Eq("PropertyId", propertyId),
		q.Eq("CurrState", cbean.ChannelState_CanUse),
		q.Or(
			q.Eq("PeerIdB", currPayeePeerId),
			q.Eq("PeerIdA", currPayeePeerId))).OrderBy("Id").Reverse().
		Find(&nodes)

	if err == nil {
		for _, item := range nodes {

			// check userOnline
			if getUserState(item.ObdNodeIdA, item.PeerIdA) == false {
				continue
			}
			if getUserState(item.ObdNodeIdB, item.PeerIdB) == false {
				continue
			}

			interSender := item.PeerIdA
			leftAmount := item.AmountA
			if item.PeerIdA == currPayeePeerId {
				interSender = item.PeerIdB
				leftAmount = item.AmountB
			}

			if leftAmount >= amount {
				if _, ok := userOfOnlineMap[interSender]; ok == false {
					continue
				}
				channelIds := item.ChannelId
				if tool.CheckIsString(&currNode.ChannelIds) {
					channelIds = currNode.ChannelIds + "," + item.ChannelId
				}

				newEdge := graphEdge{
					ParentNodeIndex: currNodeIndex,
					PathPeerIds:     pathPeerIds,
					PathIndexArr:    pathIndexArr,
					Level:           currNode.Level + 1,
					IsRoot:          false,
					IsTarget:        false,
					CurrNodePeerId:  interSender,
					ChannelId:       item.ChannelId,
					ChannelIds:      channelIds,
				}
				manager.openList = append(manager.openList, &newEdge)

				if interSender == realPayerPeerId {
					newEdge.IsTarget = true
					newEdge.PathPeerIds += "," + newEdge.CurrNodePeerId
				} else {
					if newEdge.Level < 6 {
						manager.createChannelNetwork(realPayerPeerId, interSender, propertyId, amount, &newEdge, false)
					}
				}
			}
		}
	}
}
