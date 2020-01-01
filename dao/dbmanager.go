package dao

import (
	"LightningOnOmni/config"
	"github.com/asdine/storm"
	"log"
)

//storm  doc  https://github.com/asdine/storm#getting-started

type dbManager struct {
	Db *storm.DB //db
}

var DBService = dbManager{
	Db: nil,
}

func (manager dbManager) GetDB() (*storm.DB, error) {
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
