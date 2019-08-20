package routers

import (
	pb "LightningOnOmni/grpcpack/pb"
	"LightningOnOmni/service"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

func InitRouter(conn *grpc.ClientConn) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	go GlobalWsClientManager.Start()
	router.GET("/ws", wsClientConnect)

	apiv1 := router.Group("/api/v1")
	{
		apiv1.GET("/tags", func(context *gin.Context) {
			context.JSON(http.StatusOK, gin.H{
				"msg": "test",
			})
		})
		apiv1.GET("/saveNode", func(context *gin.Context) {
			nodeService := service.NodeService{}
			node := service.Node{Name: "name", Date: time.Now()}
			nodeService.Save(&node)

			context.JSON(http.StatusOK, gin.H{
				"msg": "test",
			})
		})
		apiv1.GET("/getNode", service.HttpService.GetNodeData)
		apiv1.GET("/test", service.HttpService.TestBd)

		apiv1.GET("/userLogin", service.HttpService.UserLogin)
		apiv1.GET("/userLogout", service.HttpService.UserLogout)
		apiv1.GET("/userInfo", service.HttpService.UserInfo)
	}

	//test grpc
	routerForRpc(conn, router)
	return router
}

func routerForRpc(conn *grpc.ClientConn, router *gin.Engine) {
	client := pb.NewBtcServiceClient(conn)
	apiRpc := router.Group("/api/rpc")
	{
		apiRpc.GET("/btc/newaddress/:label", func(c *gin.Context) {
			label := c.Param("label")
			// Contact the server and print out its response.
			req := &pb.AddressRequest{Label: label}
			res, err := client.GetNewAddress(c, req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			jsonStr, _ := json.Marshal(res)
			c.JSON(http.StatusOK, gin.H{
				"result": string(jsonStr),
			})
		})
		apiRpc.GET("/btc/blockcount", func(c *gin.Context) {
			// Contact the server and print out its response.
			res, err := client.GetBlockCount(c, &pb.EmptyRequest{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			jsonStr, _ := json.Marshal(res)
			c.JSON(http.StatusOK, gin.H{
				"result": string(jsonStr),
			})
		})
	}
}

func wsClientConnect(c *gin.Context) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if error != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}

	uuid_str, _ := uuid.NewV4()
	client := &Client{
		Id:          uuid_str.String(),
		Socket:      conn,
		SendChannel: make(chan []byte)}

	GlobalWsClientManager.Register <- client
	go client.Write()
	client.Read()
}
