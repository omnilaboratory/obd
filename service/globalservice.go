package service

import (
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"github.com/asdine/storm"
	"log"
)

var db *storm.DB
var rpcClient *rpc.Client

func init() {
	var err error
	db, err = dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
	}
	rpcClient = rpc.NewClient()
}
