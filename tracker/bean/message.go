package bean

type RequestMessage struct {
	Type             MsgType `json:"type"`
	SenderNodePeerId string  `json:"sender_node_peer_id"`
	Data             string  `json:"data"`
}

type ReplyMessage struct {
	Type   MsgType     `json:"type"`
	Status bool        `json:"status"`
	From   string      `json:"from"`
	To     string      `json:"to"`
	Result interface{} `json:"result"`
}

//节点登录
type ObdNodeLoginRequest struct {
	NodeId string `json:"node_id"`
}

//节点的用户登录
type ObdNodeUserLoginRequest struct {
	UserId string `json:"user_id"`
}
