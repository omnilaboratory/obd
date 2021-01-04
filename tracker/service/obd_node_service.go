package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gin-gonic/gin"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/omnilaboratory/obd/tracker/dao"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

//普通在线用户
var userOfOnlineMap map[string]dao.UserInfo
var obdOnlineNodesMap = make(map[string]*dao.ObdNodeInfo)

var db *storm.DB

func Start(chainType string) {
	var err error
	db, err = dao.DBService.GetTrackerDB(chainType)
	if err != nil {
		log.Println(err)
	}
	userOfOnlineMap = make(map[string]dao.UserInfo)
}

type obdNodeAccountManager struct {
	mu sync.Mutex
}

var NodeAccountService obdNodeAccountManager

func (service *obdNodeAccountManager) login(obdClient *ObdNode, msgData string) (retData interface{}, err error) {
	reqData := &bean.ObdNodeLoginRequest{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		return nil, err
	}
	if tool.CheckIsString(&reqData.NodeId) == false {
		return nil, errors.New("error node_id")
	}

	info := &dao.ObdNodeInfo{}
	_ = db.Select(q.Eq("NodeId", reqData.NodeId)).First(info)
	info.LatestLoginAt = time.Now()
	info.LatestLoginIp = obdClient.Socket.RemoteAddr().String()
	info.P2PAddress = reqData.P2PAddress
	if info.Id == 0 {
		info.NodeId = reqData.NodeId
		info.IsOnline = true
		_ = db.Save(info)
	} else {
		info.IsOnline = true
		_ = db.Update(info)
	}

	obdClient.Id = reqData.NodeId
	obdClient.IsLogin = true

	loginLog := &dao.ObdNodeLoginLog{ObdId: reqData.NodeId, LoginIp: obdClient.Socket.RemoteAddr().String(), LoginTime: time.Now()}
	_ = db.Save(loginLog)

	split := strings.Split(reqData.P2PAddress, "/")
	p2PNodeId := split[len(split)-1]
	obdOnlineNodesMap[p2PNodeId] = info
	obdClient.ObdP2pNodeId = p2PNodeId

	retData = "login successfully"
	return retData, err
}

func (service *obdNodeAccountManager) logout(obdClient *ObdNode) (err error) {
	if obdClient.IsLogin == false {
		return nil
	}
	info := &dao.ObdNodeInfo{}
	err = db.Select(q.Eq("NodeId", obdClient.Id)).First(info)
	if err != nil {
		return err
	}

	info.LatestOfflineAt = time.Now()
	info.IsOnline = false
	_ = db.Update(info)
	_ = db.UpdateField(info, "IsOnline", info.IsOnline)

	for userId, item := range userOfOnlineMap {
		if item.ObdNodeId == obdClient.Id {
			delete(userOfOnlineMap, userId)
			userInfo := &dao.UserInfo{}
			err = db.Select(q.Eq("ObdNodeId", item.ObdNodeId), q.Eq("UserId", userId)).First(userInfo)
			if err != nil {
				continue
			}
			userInfo.OfflineAt = time.Now()
			_ = db.Update(userInfo)
			userInfo.IsOnline = false
			_ = db.UpdateField(userInfo, "IsOnline", userInfo.IsOnline)
		}
	}

	obdClient.IsLogin = false
	return err
}

func (service *obdNodeAccountManager) userLogin(obdClient *ObdNode, msgData string) (retData interface{}, err error) {
	if obdClient.IsLogin == false {
		return nil, errors.New("obd need to login first")
	}
	reqData := &bean.ObdNodeUserLoginRequest{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		return nil, err
	}
	if tool.CheckIsString(&reqData.UserId) == false {
		return nil, errors.New("error node_id")
	}
	return service.updateUserInfo(obdClient.ObdP2pNodeId, obdClient.Id, reqData.UserId)
}

func (service *obdNodeAccountManager) updateUserInfo(obdP2pNodeId, obdClientId, userId string) (retData interface{}, err error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	info := &dao.UserInfo{}
	_ = db.Select(q.Eq("ObdNodeId", obdClientId), q.Eq("UserId", userId)).First(info)
	info.ObdP2pNodeId = obdP2pNodeId
	if info.Id == 0 {
		info.UserId = userId
		info.ObdNodeId = obdClientId
		info.IsOnline = true
		_ = db.Save(info)
	} else {
		if info.IsOnline == false {
			info.IsOnline = true
			_ = db.Update(info)
		}
	}
	userOfOnlineMap[info.UserId] = *info
	retData = "login successfully"
	return retData, err
}

func (service *obdNodeAccountManager) updateUsers(obdClient *ObdNode, msgData string) (err error) {
	if tool.CheckIsString(&msgData) == false {
		return errors.New("wrong inputData")
	}

	infos := make([]bean.ObdNodeUserLoginRequest, 0)
	err = json.Unmarshal([]byte(msgData), &infos)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, item := range infos {
		userInfo := &dao.UserInfo{}
		_ = db.Select(q.Eq("ObdNodeId", obdClient.Id), q.Eq("UserId", item.UserId)).First(userInfo)
		if userInfo.Id == 0 {
			userInfo.UserId = item.UserId
			userInfo.ObdNodeId = obdClient.Id
			userInfo.IsOnline = true
			_ = db.Save(userInfo)
		} else {
			if userInfo.IsOnline == false {
				userInfo.IsOnline = true
				_ = db.Update(userInfo)
			}
		}
		userOfOnlineMap[userInfo.UserId] = *userInfo
	}
	return err
}

func (service *obdNodeAccountManager) userLogout(obdClient *ObdNode, msgData string) (err error) {
	if obdClient.IsLogin == false {
		return errors.New("obd need to login first")
	}
	reqData := &bean.ObdNodeUserLoginRequest{}
	err = json.Unmarshal([]byte(msgData), reqData)
	if err != nil {
		return err
	}
	if tool.CheckIsString(&reqData.UserId) == false {
		return errors.New("user_id is wrong")
	}
	info := &dao.UserInfo{}
	err = db.Select(q.Eq("ObdNodeId", obdClient.Id), q.Eq("UserId", reqData.UserId)).First(info)
	if err != nil {
		return err
	}

	info.OfflineAt = time.Now()
	err = db.Update(info)
	info.IsOnline = false
	err = db.UpdateField(info, "IsOnline", info.IsOnline)

	delete(userOfOnlineMap, info.UserId)
	return err
}

func (service *obdNodeAccountManager) GetNodeInfoByP2pAddress(context *gin.Context) {
	p2pAddress := context.Query("p2pAddress")
	if tool.CheckIsString(&p2pAddress) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error p2pAddress",
		})
		return
	}

	info := &dao.ObdNodeInfo{}
	err := db.Select(q.Eq("P2PAddress", p2pAddress), q.Eq("IsOnline", true)).First(info)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error p2pAddress",
		})
		return
	}

	retData := make(map[string]interface{})
	retData["id"] = info.Id
	context.JSON(http.StatusOK, gin.H{
		"data": retData,
	})
}

func (service *obdNodeAccountManager) GetUserState(context *gin.Context) {
	reqData := &bean.ObdNodeUserLoginRequest{}
	reqData.UserId = context.Query("userId")
	if tool.CheckIsString(&reqData.UserId) == false {
		context.JSON(http.StatusInternalServerError, gin.H{
			"msg": "error userId",
		})
		return
	}

	retData := make(map[string]interface{})
	retData["state"] = 0
	if _, ok := userOfOnlineMap[reqData.UserId]; ok == true {
		retData["state"] = 1
	} else {
		for _, item := range userOnlineOfOtherObdMap {
			if strings.Contains(item, reqData.UserId) {
				retData["state"] = 1
				break
			}
		}
	}

	context.JSON(http.StatusOK, gin.H{
		"msg":  "GetUserState",
		"data": retData,
	})
}

func (service *obdNodeAccountManager) GetAllUsers(context *gin.Context) {
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

	totalCount, _ := db.Count(&dao.UserInfo{})
	totalPage := totalCount / pageSize
	if totalCount%pageSize != 0 {
		totalPage += 1
	}
	if pageNum > totalPage {
		pageNum = totalPage
	}

	var infos []dao.UserInfo
	pageNum -= 1
	_ = db.Select(q.True()).OrderBy("IsOnline").OrderBy("Id").Reverse().Skip(pageNum * pageSize).Limit(pageSize).Find(&infos)
	context.JSON(http.StatusOK, gin.H{
		"data":       infos,
		"totalCount": totalCount,
		"totalPage":  totalPage,
		"pageNum":    pageNum + 1,
		"pageSize":   pageSize,
	})
}
func (service *obdNodeAccountManager) GetAllObdNodes(context *gin.Context) {
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

	totalCount, _ := db.Count(&dao.ObdNodeInfo{})
	totalPage := totalCount / pageSize
	if totalCount%pageSize != 0 {
		totalPage += 1
	}
	if pageNum > totalPage {
		pageNum = totalPage
	}

	var infos []dao.ObdNodeInfo
	pageNum -= 1
	_ = db.Select(q.True()).OrderBy("IsOnline").OrderBy("Id").Reverse().Skip(pageNum * pageSize).Limit(pageSize).Find(&infos)
	context.JSON(http.StatusOK, gin.H{
		"data":       infos,
		"totalCount": totalCount,
		"totalPage":  totalPage,
		"pageNum":    pageNum + 1,
		"pageSize":   pageSize,
	})
}
