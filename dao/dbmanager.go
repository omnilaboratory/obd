package dao

import (
	"LightningOnOmni/config"
	"github.com/asdine/storm"
	"log"
)

type DbManager struct {
	Db *storm.DB //db
}

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
	}
	return DB_Manager.Db, nil
}
