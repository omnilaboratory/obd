package service

import (
	"LightningOnOmni/config"
	"github.com/boltdb/bolt"
)

type UserService struct {
}

var User_service = UserService{}

func (service *UserService) UserLogin(user *User) error {
	//打开数据库
	db, e := DB_Manager.GetDB()
	if e != nil {
		return e
	}
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(config.Userbucket))
		if bucket == nil {
			b, err := tx.CreateBucket([]byte(config.Userbucket))
			if err != nil {
				return e
			}
			bucket = b
		}
		return nil
	})
	return nil
}
