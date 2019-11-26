package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"errors"
	"github.com/asdine/storm/q"
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

	if tool.VerifyEmailFormat(user.PeerId) == false {
		return errors.New("E-mail is not correct.")
	}

	if tool.CheckIsString(&user.Password) == false {
		return errors.New("Password is empty.")
	}

	// Check out if the user already exists.
	var node dao.User
	err := db.Select(q.Eq("PeerId", user.PeerId)).First(&node)
	if err == nil {
		return errors.New("The user already exists.")
	}

	// A new user, sign up.
	node.PeerId = user.PeerId
	node.Password = tool.SignMsgWithSha256([]byte(user.Password))
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
	if tool.VerifyEmailFormat(user.PeerId) == false {
		return errors.New("err peerId")
	}
	if tool.CheckIsString(&user.Password) == false {
		return errors.New("err Password")
	}
	var node dao.User
	err := db.Select(q.Eq("PeerId", user.PeerId), q.Eq("Password", tool.SignMsgWithSha256([]byte(user.Password)))).First(&node)
	if err != nil {
		return errors.New("not found user from db")
	}
	node.State = bean.UserState_OnLine
	err = db.Update(&node)
	if err != nil {
		return err
	}
	user.State = node.State
	user.Password = node.Password
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
	node.State = bean.UserState_Offline
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
