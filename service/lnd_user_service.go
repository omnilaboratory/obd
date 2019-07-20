package service

import (
	"LightningOnOmni/config"
	"encoding/json"
	"github.com/boltdb/bolt"
)

type UserState int

const (
	ErrorState UserState = -1
	Offline    UserState = 0
	OnLine     UserState = 1
)

//type = 1
type User struct {
	Id    string    `json:"id"`
	Email string    `json:"email"`
	State UserState `json:"state"`
}

type UserService struct {
}

var User_service = UserService{}

func (service *UserService) UserLogin(user *User) error {
	//打开数据库
	db, e := DB_Manager.GetDB()
	if e != nil {
		return e
	}
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(config.Userbucket))
		user.State = OnLine
		jsonData, err := json.Marshal(user)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(user.Email), []byte(jsonData))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
func (service *UserService) UserLogout(user *User) error {
	//打开数据库
	db, e := DB_Manager.GetDB()
	if e != nil {
		return e
	}
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(config.Userbucket))
		user.State = OnLine
		jsonData, err := json.Marshal(user)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(user.Email), []byte(jsonData))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
