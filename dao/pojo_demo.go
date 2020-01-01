package dao

type DemoChannelInfo struct {
	Id      int    `storm:"id,increment" json:"id"`
	PeerIdA string `storm:"index"  json:"peer_id_a"`
	AmountA float64
	PeerIdB string `storm:"index" json:"peer_id_b"`
	AmountB float64
}
