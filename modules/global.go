package modules

import (
	"github.com/boltdb/bolt"
	"log"
)

var Global_manager = ClientManager{
	Broadcast:   make(chan []byte),
	Register:    make(chan *Client),
	Unregister:  make(chan *Client),
	Clients_map: make(map[*Client]bool),
}

type GlobleParams struct {
	Interval       int //ms,间隔多少毫秒下发一次
	MaximumClients int //单机最多客户数
	PoolSize       int //go routine pool size
}

type DbManager struct {
	Db *bolt.DB //存放区块的数据库
}

var DB_Manager = DbManager{
	Db: nil,
}

func (manager DbManager) GetDB() (*bolt.DB, error) {
	if DB_Manager.Db == nil {
		db, e := bolt.Open(DBname, 0644, nil)
		if e != nil {
			log.Println("open db fail")
			return nil, e
		}
		DB_Manager.Db = db
	}
	return DB_Manager.Db, nil
}

var Global_params = GlobleParams{
	Interval:       1000,
	MaximumClients: 1024 * 1024,
	PoolSize:       4 * 1024,
}
