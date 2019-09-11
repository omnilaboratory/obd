package service

import (
	"LightningOnOmni/dao"
	"github.com/asdine/storm/q"
	"log"
	"testing"
	"time"
)

func TestTask(t *testing.T) {
	log.Println("aaa")
	node := &dao.RDTxWaitingSend{}
	node.TransactionHex = "111"
	node.IsEnable = true
	node.CreateAt = time.Now()
	db.Save(node)

	var nodes []dao.RDTxWaitingSend
	err := db.Select().Find(&nodes)
	if err != nil {
		return
	}

	for _, item := range nodes {
		item.IsEnable = true
		item.TransactionHex = "33333"
		item.FinishAt = time.Now()
		err := db.Save(&item)
		log.Println(err)
	}
	var nodes2 []dao.RDTxWaitingSend

	db.Select(q.Eq("IsEnable", true)).Find(&nodes2)
	if err != nil {
		return
	}

}
