package dao

import (
	"github.com/asdine/storm"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
	"log"
)

type dbManager struct {
	Db *storm.DB //db
}

var DBService dbManager

func (manager dbManager) GetTrackerDB(chainType string) (*storm.DB, error) {
	if DBService.Db == nil {
		_dir := "dbdata" + chainType
		_ = tool.PathExistsAndCreate(_dir)
		db, e := storm.Open(_dir + "/" + config.TrackerDbName)
		if e != nil {
			log.Println("open db fail")
			return nil, e
		}
		DBService.Db = db
	}
	return DBService.Db, nil
}
