package lightclient

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"strconv"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	go globalWsClientManager.Start()
	bean.CurrObdNodeInfo.WebsocketLink = "ws://" + config.P2P_hostIp + ":" + strconv.Itoa(config.ServerPort) + "/ws" + config.ChainNode_Type
	router.GET("/ws"+config.ChainNode_Type, wsClientConnect)

	return router

}

func wsClientConnect(c *gin.Context) {
	wsConn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		http.NotFound(c.Writer, c.Request)
		return
	}

	client := &Client{
		Id:          uuid.NewV4().String(),
		Socket:      wsConn,
		SendChannel: make(chan []byte)}

	go client.Write()
	globalWsClientManager.Connected <- client
	client.Read()
}
