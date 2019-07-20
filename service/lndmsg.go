package service

type Message struct {
	Type      int    `json:"type"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Data      string `json:"data"`
}

//type = 1
type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

type chainHash string
type point string

//type = -32
type OpenChannelData struct {
	Chain_hash                    chainHash `json:"chain_hash"`
	Temporary_channel_id          []byte    `json:"temporary_channel_id"`
	funding_satoshis              uint64    `json:"funding_satoshis"`
	push_msat                     uint64    `json:"push_msat"`
	dust_limit_satoshis           uint64    `json:"dust_limit_satoshis"`
	max_htlc_value_in_flight_msat uint64    `json:"max_htlc_value_in_flight_msat"`
	channel_reserve_satoshis      uint64    `json:"channel_reserve_satoshis"`
	htlc_minimum_msat             uint64    `json:"htlc_minimum_msat"`
	feerate_per_kw                uint32    `json:"feerate_per_kw"`
	to_self_delay                 uint16    `json:"to_self_delay"`
	max_accepted_htlcs            uint16    `json:"max_accepted_htlcs"`
	funding_pubkey                point     `json:"funding_pubkey"`
	revocation_basepoint          point     `json:"revocation_basepoint"`
	payment_basepoint             point     `json:"payment_basepoint"`
	delayed_payment_basepoint     point     `json:"delayed_payment_basepoint"`
	htlc_basepoint                point     `json:"htlc_basepoint"`
}
