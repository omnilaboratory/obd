package bean

type ChannelState int

const (
	ChannelState_Create            ChannelState = 10
	ChannelState_WaitFundAsset     ChannelState = 11
	ChannelState_NewTx             ChannelState = 12
	ChannelState_CanUse            ChannelState = 20
	ChannelState_Close             ChannelState = 21
	ChannelState_HtlcTx            ChannelState = 22
	ChannelState_OpenChannelRefuse ChannelState = 30

	ProtocolIdForUserState         = "tracker/userState/1.0.1"
	ProtocolIdForChannelInfoChange = "tracker/channelInfo/1.0.1"
	ProtocolIdForLockChannel       = "tracker/lockChannel/1.0.1"
)

//更新通道
type ChannelInfoRequest struct {
	ChannelId  string       `json:"channel_id"`
	PropertyId int64        `json:"property_id"`
	CurrState  ChannelState `json:"curr_state"`
	IsAlice    bool         `json:"is_alice"` //是否是alice方的节点
	PeerIdA    string       `json:"peer_ida"`
	PeerIdB    string       `json:"peer_idb"`
	AmountA    float64      `json:"amount_a"`
	AmountB    float64      `json:"amount_b"`
}

//节点登录
type ObdNodeLoginRequest struct {
	NodeId     string `json:"node_id"`
	P2PAddress string `json:"p2p_address"`
}

//节点的用户登录
type ObdNodeUserLoginRequest struct {
	UserId    string `json:"user_id"`
	P2pNodeId string `json:"p2p_node_id"`
}

//节点的用户登录
type TrackerLockChannelRequest struct {
	UserId    string `json:"user_id"`
	ChannelId string `json:"channel_id"`
}
