package lightclient

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"net/http"
	"strconv"
)

func InitRouter(conn *grpc.ClientConn) *gin.Engine {
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
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if error != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}

	uuidStr := uuid.NewV4()
	client := &Client{
		Id:          uuidStr.String(),
		Socket:      conn,
		SendChannel: make(chan []byte)}

	globalWsClientManager.Connected <- client
	go client.Write()
	client.Read()
}
