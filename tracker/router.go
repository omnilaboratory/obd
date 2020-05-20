package tracker

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/tracker/service"
	"github.com/satori/go.uuid"
	"net/http"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	go service.ObdNodeManager.TrackerStart()
	router.GET("/ws", wsClientConnect)
	return router
}

func wsClientConnect(c *gin.Context) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	uuidStr := uuid.NewV4()
	newClient := &service.ObdNode{
		Id:          uuidStr.String(),
		Socket:      conn,
		SendChannel: make(chan []byte)}
	service.ObdNodeManager.Connected <- newClient
	go newClient.Write()
	newClient.Read()
}
