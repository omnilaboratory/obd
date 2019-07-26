package service

import (
	"LightningOnOmni/config"
	"github.com/asdine/storm"
	"github.com/boltdb/bolt"
	"log"
)

var DB_Manager = DbManager{
	Db: nil,
}

func (manager DbManager) GetDB() (*storm.DB, error) {
	if DB_Manager.Db == nil {
		db, e := storm.Open(config.DBname)
		if e != nil {
			log.Println("open db fail")
			return nil, e
		}
		DB_Manager.Db = db
		db.Init(config.Userbucket)
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
