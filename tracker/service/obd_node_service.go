package service

import (
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/bean"
	"github.com/omnilaboratory/obd/tracker/dao"
	"log"
	"sync"
	"time"
)

var db *storm.DB

func init() {
	var err error
	db, err = dao.DBService.GetTrackerDB()
	if err != nil {
		log.Println(err)
	}
}

type obdNodeAccountManager struct {
	mu sync.Mutex
}

var nodeAccountService obdNodeAccountManager

func (this *obdNodeAccountManager) login(obdClient *ObdNode, msgData string) (retData interface{}, err error) {
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
	loginLog := &dao.ObdNodeLoginLog{LoginIp: obdClient.Socket.RemoteAddr().String(), LoginTime: time.Now()}
	_ = db.Save(loginLog)

	retData = "login successfully"
	return retData, err
}
func (this *obdNodeAccountManager) logout(obdClient *ObdNode) (err error) {
	if obdClient.IsLogin == false {
		return nil
	}
	info := &dao.ObdNodeInfo{}
	err = db.Select(q.Eq("NodeId", obdClient.Id)).First(info)
	if err != nil {
		return err
	}

	info.OfflineAt = time.Now()
	info.IsOnline = false
	err = db.Update(info)
	err = db.UpdateField(info, "IsOnline", info.IsOnline)

	obdClient.IsLogin = false
	return err
}

func (this *obdNodeAccountManager) userLogin(obdClient *ObdNode, msgData string) (retData interface{}, err error) {
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
	info := &dao.UserInfo{}
	_ = db.Select(q.Eq("ObdNodeId", obdClient.Id), q.Eq("UserId", reqData.UserId)).First(info)
	if info.Id == 0 {
		info.UserId = reqData.UserId
		info.ObdNodeId = obdClient.Id
		info.IsOnline = true
		_ = db.Save(info)
	} else {
		if info.IsOnline == false {
			info.IsOnline = true
			_ = db.Update(info)
		}
	}
	retData = "login successfully"
	return retData, err
}

func (this *obdNodeAccountManager) userLogout(obdClient *ObdNode, msgData string) (err error) {
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
	return err
}
