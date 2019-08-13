package routers

import (
	"LightningOnOmni/service"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"time"
)

func InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	go service.GlobalWsClientManager.Start()
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
		apiv1.GET("/getNode", getNodeData)
		apiv1.GET("/test", testBd)

		apiv1.GET("/userLogin", userLogin)
		apiv1.GET("/userLogout", userLogout)
		apiv1.GET("/userInfo", userInfo)
	}
	return router
}

func testBd(context *gin.Context) {
	node, e := service.TheFundingService.CreateFunding("")
	if e != nil {
		context.JSON(http.StatusOK, gin.H{
			"msg":  "userInfo",
			"data": e.Error(),
		})
		return
	}
	bytes, _ := json.Marshal(node)
	context.JSON(http.StatusOK, gin.H{
		"msg":  "test CreateFunding",
		"data": string(bytes),
	})
}

func userInfo(context *gin.Context) {
	user, e := service.User_service.UserInfo(context.Query("email"))
	if e != nil {
		context.JSON(http.StatusOK, gin.H{
			"msg":  "userInfo",
			"data": e.Error(),
		})
		return
	}
	bytes, _ := json.Marshal(user)
	context.JSON(http.StatusOK, gin.H{
		"msg":  "userInfo",
		"data": string(bytes),
	})
}

func userLogin(context *gin.Context) {
	user := service.User{}
	user.Email = context.Query("email")
	service.User_service.UserLogin(&user)
	bytes, _ := json.Marshal(user)
	context.JSON(http.StatusOK, gin.H{
		"msg":  "userLogin",
		"data": string(bytes),
	})
}
func userLogout(context *gin.Context) {
	user := service.User{}
	user.Email = context.Query("email")
	logout := service.User_service.UserLogout(&user)
	if logout != nil {
		context.JSON(http.StatusOK, gin.H{
			"msg":  "userLogout",
			"data": logout.Error(),
		})
	} else {
		bytes, _ := json.Marshal(user)
		context.JSON(http.StatusOK, gin.H{
			"msg":  "userLogout",
			"data": string(bytes),
		})
	}

}

func getNodeData(context *gin.Context) {
	nodeService := service.NodeService{}
	id, _ := strconv.Atoi(context.Query("id"))
	data, _ := nodeService.Get(id)
	bytes, _ := json.Marshal(data)

	context.JSON(http.StatusOK, gin.H{
		"msg":  "getNodeData",
		"data": string(bytes),
	})

}

func wsClientConnect(c *gin.Context) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if error != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}

	uuid_str, _ := uuid.NewV4()
	client := &service.Client{
		Id:           uuid_str.String(),
		Socket:       conn,
		Send_channel: make(chan []byte)}

	service.GlobalWsClientManager.Register <- client
	go client.Write()
	client.Read()
}

func test(writer http.ResponseWriter, request *http.Request) {
	bytes, err := json.Marshal(&service.User{Id: 1, Email: "123@qq.com"})
	if err != nil {
		fmt.Fprintf(writer, "wrong data")
		return
	}
	fmt.Fprintf(writer, string(bytes))
}
