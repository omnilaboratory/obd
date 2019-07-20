package service

import (
	"github.com/boltdb/bolt"
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

var Global_params = GlobleParams{
	Interval:       1000,
	MaximumClients: 1024 * 1024,
	PoolSize:       4 * 1024,
}
