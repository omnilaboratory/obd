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
	log.Println("-------------------  UserLogin user.PeerId " + user.PeerId)
	userDB, err := dao.DBService.GetUserDB(user.PeerId)
	if err != nil {
		return err
	}
	err = userDB.Select(q.Eq("PeerId", user.PeerId)).First(&node)
	if node.Id == 0 {
		node = dao.User{}
		node.PeerId = user.PeerId
		node.State = bean.UserState_OnLine
		node.CreateAt = time.Now()
		node.LatestLoginTime = node.CreateAt
		node.CurrAddrIndex = -1
		err = userDB.Save(&node)
	} else {
		node.State = bean.UserState_OnLine
		node.LatestLoginTime = time.Now()
		err = userDB.Update(&node)
	}
	if err != nil {
		return err
	}
	user.Db = userDB

	user.State = node.State
	user.CurrAddrIndex = node.CurrAddrIndex
	user.ChangeExtKey = changeExtKey

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
	node.State = bean.UserState_Offline
	err = user.Db.Update(node)
	if err != nil {
		log.Println(err)
	}
	user.Db.Close()

	return nil
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
