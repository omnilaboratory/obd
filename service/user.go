package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/enum"
	"LightningOnOmni/dao"
	"errors"
)

type UserService struct {
}

var User_service = UserService{}

func (service *UserService) UserLogin(user *bean.User) error {
	if user != nil {
		errors.New("user is nil")
	}
	//打开数据库
	db, e := dao.DB_Manager.GetDB()
	if e != nil {
		return e
	}
	user.State = enum.UserState_OnLine
	var node bean.User

	e = db.One("Email", user.Email, &node)
	if node.Id == 0 {
		return db.Save(user)
	} else {
		return db.Update(user)
	}
}
func (service *UserService) UserLogout(user *bean.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	//打开数据库
	db, e := dao.DB_Manager.GetDB()
	if e != nil {
		return e
	}

	var node bean.User

	e = db.One("Email", user.Email, &node)
	if node.Id == 0 {
		return errors.New("user not found")
	}

	user.State = enum.UserState_Offline
	return db.Update(user)
}

func (service *UserService) UserInfo(email string) (user *bean.User, e error) {

	db, e := dao.DB_Manager.GetDB()
	if e != nil {
		return nil, errors.New("db is not exist")
	}

	var node bean.User
	e = db.One("Email", email, &node)
	if node.Id == 0 {
		return nil, errors.New("user not exist")
	}
	return &node, nil
}
