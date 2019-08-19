package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/dao"
	"errors"
)

type UserManager struct {
}

var UserService = UserManager{}

func (service *UserManager) UserLogin(user *bean.User) error {
	if user != nil {
		errors.New("user is nil")
	}
	//打开数据库
	db, e := dao.DBService.GetDB()
	if e != nil {
		return e
	}
	user.State = enum.UserState_OnLine
	var node dao.User

	e = db.One("Email", user.Email, &node)
	node.Email = user.Email
	node.Password = user.Password
	node.State = user.State

	if node.Id == 0 {
		return db.Save(node)
	} else {
		return db.Update(node)
	}
}
func (service *UserManager) UserLogout(user *bean.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	//打开数据库
	db, e := dao.DBService.GetDB()
	if e != nil {
		return e
	}

	var node dao.User

	e = db.One("Email", user.Email, &node)
	if node.Id == 0 {
		return errors.New("user not found")
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
	e = db.One("Email", email, &node)
	if node.Id == 0 {
		return nil, errors.New("user not exist")
	}
	return &node, nil
}
