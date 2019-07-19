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
	Interval       int //ms, intercal for boardcasting to all the connections.
	MaximumClients int //maximum clients for one server instance.
	PoolSize       int //go routine pool size
}

type DbManager struct {
	Db *bolt.DB // DB to store blocks 
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
