package dao

import (
	"LightningOnOmni/config"
	"github.com/asdine/storm"
	"log"
)

type DbManager struct {
	Db *storm.DB //db
}

var DBService = DbManager{
	Db: nil,
}

func (manager DbManager) GetDB() (*storm.DB, error) {
	if DBService.Db == nil {
		db, e := storm.Open(config.DBname)
		if e != nil {
			log.Println("open db fail")
			return nil, e
		}
		DBService.Db = db
	}
	return DBService.Db, nil
}
