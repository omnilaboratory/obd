package service

import (
	"errors"
	"github.com/asdine/storm/q"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tyler-smith/go-bip39"
	"log"
	"time"
)

type UserManager struct {
}

var UserService = UserManager{}

func (service *UserManager) UserLogin(user *bean.User) error {
	if user == nil {
		return errors.New(enum.Tips_user_nilUser)
	}
	if tool.CheckIsString(&user.Mnemonic) == false || bip39.IsMnemonicValid(user.Mnemonic) == false {
		return errors.New(enum.Tips_common_wrong + "mnemonic")
	}

	changeExtKey, err := HDWalletService.CreateChangeExtKey(user.Mnemonic)
	if err != nil {
		return err
	}
	var node dao.User
	user.PeerId = tool.SignMsgWithSha256([]byte(user.Mnemonic))
	userDB, err := dao.DBService.GetUserDB(user.PeerId)
	if err != nil {
		return err
	}
	err = userDB.Select(q.Eq("PeerId", user.PeerId)).First(&node)
	if node.Id == 0 {
		node = dao.User{}
		node.PeerId = user.PeerId
		node.P2PLocalPeerId = user.P2PLocalPeerId
		node.P2PLocalAddress = user.P2PLocalAddress
		node.CurrState = bean.UserState_OnLine
		node.CreateAt = time.Now()
		node.LatestLoginTime = node.CreateAt
		node.CurrAddrIndex = 0
		err = userDB.Save(&node)
	} else {
		node.P2PLocalPeerId = user.P2PLocalPeerId
		node.P2PLocalAddress = user.P2PLocalAddress
		node.CurrState = bean.UserState_OnLine
		node.LatestLoginTime = time.Now()
		err = userDB.Update(&node)
	}

	noticeTrackerUserLogin(node)

	if err != nil {
		return err
	}

	loginLog := &dao.UserLoginLog{}
	loginLog.PeerId = user.PeerId
	loginLog.LoginAt = time.Now()
	_ = userDB.Save(loginLog)

	user.State = node.CurrState
	user.CurrAddrIndex = node.CurrAddrIndex
	user.ChangeExtKey = changeExtKey
	user.Db = userDB
	return nil
}

func (service *UserManager) UserLogout(user *bean.User) error {
	if user == nil {
		return errors.New(enum.Tips_user_nilUser)
	}
	var node dao.User
	err := user.Db.Select(q.Eq("PeerId", user.PeerId)).First(&node)
	if err != nil {
		return err
	}
	node.CurrAddrIndex = user.CurrAddrIndex
	node.CurrState = bean.UserState_Offline
	err = user.Db.Update(&node)
	if err != nil {
		log.Println(err)
	}
	_ = user.Db.UpdateField(&node, "CurrState", bean.UserState_Offline)
	loginLog := &dao.UserLoginLog{}
	_ = user.Db.Select(q.Eq("PeerId", user.PeerId)).OrderBy("LoginAt").Reverse().First(loginLog)
	if loginLog.Id > 0 {
		loginLog.LogoutAt = time.Now()
		_ = user.Db.Update(loginLog)
	}
	noticeTrackerUserLogout(node)
	return user.Db.Close()
}
