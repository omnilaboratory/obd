package bean

type BtcSendRequest struct {
	FromAddress           string  `json:"from_address"`
	FromAddressPrivateKey string  `json:"from_address_private_key"`
	ToAddress             string  `json:"to_address"`
	Amount                float64 `json:"amount"`
	MinerFee              float64 `json:"miner_fee"`
}
type OmniSendRequest struct {
	BtcSendRequest
	PropertyId int64 `json:"property_id"`
}
