package service

import (
	"github.com/boltdb/bolt"
	"lnd-server/modules"
)

type UserService struct {
}

var User_service = UserService{}

func (service *UserService) UserLogin(user *modules.User) error {

	//打开数据库
	db, e := modules.DB_Manager.GetDB()
	if e != nil {
		return e
	}
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(modules.Userbucket))
		if bucket == nil {
			b, err := tx.CreateBucket([]byte(modules.Userbucket))
			if err != nil {
				return e
			}
			bucket = b
		}
		return nil
	})
	return nil
}
