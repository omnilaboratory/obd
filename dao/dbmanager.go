package dao

import (
	"LightningOnOmni/config"
	"LightningOnOmni/tool"
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

func (manager dbManager) GetUserDB(peerId string) (*storm.DB, error) {

	_dir := "dbdata"
	_ = tool.PathExistsAndCreate(_dir)

	db, e := storm.Open(_dir + "/user_" + peerId + ".db")
	if e != nil {
		log.Println("open db fail")
		return nil, e
	}
	return db, nil
}
