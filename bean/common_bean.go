package bean

type UserInfoToTracker struct {
	ObdId      string `json:"obd_id"`
	UserPeerId string `json:"user_peer_id"`
	P2pNodeId  string `json:"p2p_node_id"`
}
