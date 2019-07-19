package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/websocket"
	"log"
	"net/http"
)

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Hash  string `json:"hash"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpooint)
}

func reader(conn *websocket.Conn) {

	defer func() {
		conn.Close()
		fmt.Println("socket closed after reading...")
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			break
		}

		log.Println(string(p))
		var user User
		json.Unmarshal(p, &user)

		log.Println(user.Id)
		log.Println(user.Email)
		log.Println(user.Name)
		hash := sha256.New()
		bytes, err := json.Marshal(user)
		hash.Write(bytes)
		hashInBytes := hash.Sum(nil)
		hashValue := hex.EncodeToString(hashInBytes)
		user.Hash = hashValue

		bytes, err = json.Marshal(user)
		log.Println(user.Hash)

		for node, v := range connections {
			log.Println(v)
			if node != conn {
				if err := node.WriteMessage(messageType, bytes); err != nil {
					log.Println(err)
					break
				}
			}
		}
	}
}

type connection struct {
	// websocket 连接器
	ws *websocket.Conn

	// 发送信息的缓冲 channel
	send chan []byte
}

type hub struct {
	// 注册了的连接器
	connections map[*connection]bool

	// 从连接器中发入的信息
	broadcast chan []byte

	// 从连接器中注册请求
	register chan *connection

	// 从连接器中注销请求
	unregister chan *connection
}

var connections map[*websocket.Conn]bool = make(map[*websocket.Conn]bool)

func wsEndpooint(writer http.ResponseWriter, request *http.Request) {
	//upgrader.CheckOrigin = func(r *http.Request) bool {	return true }
	//conn, err := upgrader.Upgrade(writer, request, nil)
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Client Successfully Connected")
	connections[conn] = true
	reader(conn)
}

func homePage(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Home Page")
}

func main() {
	fmt.Println("go web sockets")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":60020", nil))
}
