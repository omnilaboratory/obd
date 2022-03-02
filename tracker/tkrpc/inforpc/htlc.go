package inforpc

import (
	"errors"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/config"
	"github.com/omnilaboratory/obd/tracker/service"
	"github.com/omnilaboratory/obd/tracker/tkrpc"
	"github.com/shopspring/decimal"
	"log"
	"strings"
	"sync"
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

func (manager *htlcManager) getPath( msgData string,pathRequest *tkrpc.HtlcGetPathReq ) (*tkrpc.HtlcGetPathRes ,error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	log.Println("getPath", msgData)
	if tool.CheckIsString(&pathRequest.RealPayerPeerId) == false {
		return nil, errors.New("wrong realPayerPeerId")
	}
	if tool.CheckIsString(&pathRequest.PayeePeerId) == false {
		return nil, errors.New("wrong SendeePeerId")
	}
	if pathRequest.Amount < tool.GetOmniDustBtc() {
		return nil, errors.New("wrong amount")
	}
	manager.createChannelNetwork(pathRequest.RealPayerPeerId, pathRequest.PayeePeerId, pathRequest.PropertyId, pathRequest.Amount, nil, true)
	resultIndex := manager.getPathIndex()

	res:=new(tkrpc.HtlcGetPathRes)
	res.SenderPeerId = pathRequest.RealPayerPeerId
	res.H = pathRequest.H
	res.Amount = pathRequest.Amount
	if resultIndex != -1 {
		splitArr := strings.Split(manager.openList[resultIndex].ChannelIds, ",")
		path := ""
		for i := len(splitArr) - 1; i > -1; i-- {
			path += splitArr[i] + ","
		}
		path = strings.TrimSuffix(path, ",")
		res.Path = path
	}
	log.Println("return path info", res)
	return res, nil
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
		lockResult := true
		for _, item := range channelIdArr {
			channelInfo:=new(tkrpc.ChannelInfo)
			err:=Orm.First(channelInfo,tkrpc.ChannelInfo{ChannelId: item,CurrState: tkrpc.ChannelState_CanUse}).Error
			if err == nil {
				if len(channelInfo.ObdNodeIda) > 0 && service.SendChannelLockInfoToObd(channelInfo.ChannelId, channelInfo.PeerIda, channelInfo.ObdNodeIda,false) == false {
					lockResult = false
					break
				}

				if len(channelInfo.ObdNodeIdb) > 0 && service.SendChannelLockInfoToObd(channelInfo.ChannelId, channelInfo.PeerIdb, channelInfo.ObdNodeIdb,false) == false {
					lockResult = false
					break
				}
			}
		}
		if lockResult == false {
			manager.openList = append(manager.openList[:resultIndex], manager.openList[resultIndex+1:]...)
			return manager.getPathIndex()
		} else {
			for _, item := range channelIdArr {
				channelInfo:=new(tkrpc.ChannelInfo)
				err:=Orm.First(channelInfo,tkrpc.ChannelInfo{ChannelId: item,CurrState: tkrpc.ChannelState_CanUse}).Error
				if err == nil {
					channelInfo.CurrState =tkrpc.ChannelState_LockByTracker
					Orm.Save(channelInfo)
				}
			}
			htlcPath := tkrpc.LockHtlcPath{Path: strings.Join(channelIdArr,","), CurrState: 0}
			Orm.Save(&htlcPath)
		}
	}
	return resultIndex
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

	var nodes []tkrpc.ChannelInfo
	err:=Orm.Where(tkrpc.ChannelInfo{PropertyId:propertyId,CurrState: tkrpc.ChannelState_CanUse}).Or(tkrpc.ChannelInfo{PeerIdb:currPayeePeerId,PeerIda: currPayeePeerId}).Find(&nodes).Order("id desc")
	if err == nil {
		for _, item := range nodes {

			// check userOnline
			if getUserState( item.PeerIda) == false {
				continue
			}
			if getUserState( item.PeerIdb) == false {
				continue
			}

			interSender := item.PeerIda
			leftAmount := item.AmountA
			if item.PeerIda == currPayeePeerId {
				interSender = item.PeerIdb
				leftAmount = item.AmountB
			}

			if leftAmount >= amount {
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

