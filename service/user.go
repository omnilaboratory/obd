package service

import (
	"errors"
	"github.com/asdine/storm/q"
	"github.com/tyler-smith/go-bip39"
	"log"
	"obd/bean"
	"obd/dao"
	"obd/tool"
	"time"
)

type UserManager struct {
}

var UserService = UserManager{}

// UserSignUp
func (service *UserManager) UserSignUp(user *bean.User) error {
	// Check data if correct.
	if user == nil {
		return errors.New("user is nil")
	}

	if tool.CheckIsString(&user.PeerId) == false {
		return errors.New("Peer ID  is not correct.")
	}

	// Check out if the user already exists.
	var node dao.User
	err := db.Select(q.Eq("PeerId", user.PeerId)).First(&node)
	if err == nil {
		return errors.New("The user already exists.")
	}

	// A new user, sign up.
	node.PeerId = user.PeerId
	node.CreateAt = time.Now()

	err = db.Save(&node)
	if err != nil {
		return err
	}

	return nil
}

func (service *UserManager) UserLogin(user *bean.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if tool.CheckIsString(&user.Mnemonic) == false || bip39.IsMnemonicValid(user.Mnemonic) == false {
		return errors.New("err Mnemonic")
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
		node.P2PLocalPeerId = user.P2PLocalPeerId
		node.P2PLocalAddress = user.P2PLocalAddress
		node.PeerId = user.PeerId
		node.State = bean.UserState_OnLine
		node.CreateAt = time.Now()
		node.LatestLoginTime = node.CreateAt
		node.CurrAddrIndex = -1
		err = userDB.Save(&node)
	} else {
		node.P2PLocalPeerId = user.P2PLocalPeerId
		node.P2PLocalAddress = user.P2PLocalAddress
		node.State = bean.UserState_OnLine
		node.LatestLoginTime = time.Now()
		err = userDB.Update(&node)
	}
	if err != nil {
		return err
	}

	loginLog := &dao.UserLoginLog{}
	loginLog.PeerId = user.PeerId
	loginLog.LoginAt = time.Now()
	_ = userDB.Save(loginLog)

	user.State = node.State
	user.CurrAddrIndex = node.CurrAddrIndex
	user.ChangeExtKey = changeExtKey
	user.Db = userDB
	return nil
}

func (service *UserManager) UserLogout(user *bean.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	var node dao.User
	err := user.Db.Select(q.Eq("PeerId", user.PeerId)).First(&node)
	if err != nil {
		return err
	}
	node.CurrAddrIndex = user.CurrAddrIndex
	node.State = bean.UserState_Offline
	err = user.Db.Update(&node)
	if err != nil {
		log.Println(err)
	}
	loginLog := &dao.UserLoginLog{}
	_ = user.Db.Select(q.Eq("PeerId", user.PeerId)).OrderBy("LoginAt").Reverse().First(loginLog)
	if loginLog.Id > 0 {
		loginLog.LogoutAt = time.Now()
		_ = user.Db.Update(loginLog)
	}

	return user.Db.Close()
}

func (service *UserManager) UserInfo(email string) (user *dao.User, e error) {

	db, e := dao.DBService.GetGlobalDB()
	if e != nil {
		return nil, errors.New("db is not exist")
	}

	var node dao.User
	e = db.One("PeerId", email, &node)
	if node.Id == 0 {
		return nil, errors.New("user not exist")
	}
	return &node, nil
}
