package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"errors"
	"github.com/asdine/storm/q"
)

type UserManager struct {
}

var UserService = UserManager{}

func (service *UserManager) UserLogin(user *bean.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if tool.CheckIsString(&user.PeerId) == false {
		return errors.New("err peerId")
	}
	if tool.CheckIsString(&user.Password) == false {
		return errors.New("err Password")
	}
	var node dao.User
	err := db.Select(q.Eq("PeerId", user.PeerId), q.Eq("Password", user.Password)).First(&node)
	if err != nil {
		return errors.New("not found user from db")
	}
	node.State = user.State
	err = db.Update(node)
	if err != nil {
		return err
	}
	user.State = enum.UserState_OnLine
	return nil
}
func (service *UserManager) UserLogout(user *bean.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	var node dao.User
	err := db.Select(q.Eq("PeerId", user.PeerId), q.Eq("Password", user.Password)).First(&node)
	if err != nil {
		return err
	}
	node.State = enum.UserState_Offline
	return db.Update(node)
}

func (service *UserManager) UserInfo(email string) (user *dao.User, e error) {

	db, e := dao.DBService.GetDB()
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
