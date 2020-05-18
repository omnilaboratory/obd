package service

import (
	"fmt"
	"github.com/omnilaboratory/obd/dao"
	"time"
)

type Node struct {
	ID   int `storm:"id,increment" `
	Name string
	Date time.Time
}
type NodeService struct {
}

func (service *NodeService) Save(node *Node) error {
	db, e := dao.DBService.GetGlobalDB()
	if e != nil {
		return e
	}
	return db.Save(node)
}
func (service *NodeService) Get(id interface{}) (data Node, err error) {
	db, e := dao.DBService.GetGlobalDB()
	var node Node
	if e != nil {
		return node, e
	}
	count, _ := db.Count(&node)
	fmt.Println(count)
	err = db.One("ID", id, &node)
	return node, err
}
