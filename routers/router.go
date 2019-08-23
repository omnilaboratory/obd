package routers

import (
	"LightningOnOmni/grpcpack"
	pb "LightningOnOmni/grpcpack/pb"
	"LightningOnOmni/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"time"
)

func InitRouter(conn *grpc.ClientConn) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	go GlobalWsClientManager.Start()
	router.GET("/ws", wsClientConnect)

	return router

	routerForRpc(conn, router)

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

	return router
}

func routerForRpc(conn *grpc.ClientConn, router *gin.Engine) {

	if conn == nil {
		return
	}
	client := pb.NewBtcServiceClient(conn)

	var grpcService = grpcpack.GetGrpcService()
	grpcService.SetClient(client)

	apiRpc := router.Group("/api/rpc/btc")
	{

		//curl -X POST -H "Content-Type:application / json" -d "{\"id\":1,\"method\":\"Test.Greet\",\"params\":[{\"name\" :\"world\"}]}" http://localhost:60020/api/rpc/btc/
		apiRpc.Any("/", handlerRpcReq)

		//curl -H "Content-Type:application/x-www-form-urlencoded" -d "label=admin" -X POST http://localhost:60020/api/rpc/btc/getnewaddress
		//curl http://localhost:60020/api/rpc/btc/getnewaddress -H "content-type: application/json"  -d "{\"label\":{\"test\":\"abc\"}}"
		apiRpc.POST("/getnewaddress", grpcService.GetNewAddress)
		apiRpc.POST("/createmultisig", grpcService.CreateMultiSig)

		//curl http://localhost:60020/api/rpc/btc/getnewaddress/254698748@qq.com -v
		//apiRpc.GET("/getnewaddress/:label", grpcservice.GetNewAddress)

		apiRpc.GET("/getblockcount", grpcService.GetBlockCount)
		apiRpc.GET("/getmininginfo", grpcService.GetMiningInfo)
	}
}

func handlerRpcReq(c *gin.Context) {
	bytes, _ := c.GetRawData()
	log.Println(string(bytes))
	parse := gjson.Parse(string(bytes))
	log.Println(parse)
	c.JSON(http.StatusOK, gin.H{
		"result": parse.Raw,
	})
}

func wsClientConnect(c *gin.Context) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if error != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}

	uuidStr, _ := uuid.NewV4()
	client := &Client{
		Id:          uuidStr.String(),
		Socket:      conn,
		SendChannel: make(chan []byte)}

	GlobalWsClientManager.Register <- client
	go client.Write()
	client.Read()
}
