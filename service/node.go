package service

import (
	"LightningOnOmni/dao"
	"fmt"
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
	db, e := dao.DBService.GetDB()
	if e != nil {
		return e
	}
	return db.Save(node)
}
func (service *NodeService) Get(id interface{}) (data Node, err error) {
	db, e := dao.DBService.GetDB()
	var node Node
	if e != nil {
		return node, e
	}
	count, _ := db.Count(&node)
	fmt.Println(count)
	err = db.One("ID", id, &node)
	return node, err
}
