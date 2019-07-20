package service

import (
	"LightningOnOmni/config"
	"github.com/boltdb/bolt"
	"log"
)

var DB_Manager = DbManager{
	Db: nil,
}

func (manager DbManager) GetDB() (*bolt.DB, error) {
	if DB_Manager.Db == nil {
		db, e := bolt.Open(config.DBname, 0600, nil)
		if e != nil {
			log.Println("open db fail")
			return nil, e
		}
		DB_Manager.Db = db
	}

	DB_Manager.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.Userbucket))
		if b == nil {
			tx.CreateBucketIfNotExists([]byte(config.Userbucket))
		}
		return nil
	})
	return DB_Manager.Db, nil
}
