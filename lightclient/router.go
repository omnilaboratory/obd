package lightclient

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/tool"
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
	bean.CurrObdNodeInfo.WebsocketLink = "ws://" + config.P2P_hostIp + ":" + strconv.Itoa(config.ServerPort) + "/ws" + config.ChainNodeType
	router.GET("/ws"+config.ChainNodeType, wsClientConnect)

	return router

}

func wsClientConnect(c *gin.Context) {
	session := c.GetHeader("session")
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

	if session == tool.GetGRpcSession() {
		client.IsGRpcRequest = true
	}

	go client.Write()
	go client.Read()
	globalWsClientManager.Connected <- client
}
