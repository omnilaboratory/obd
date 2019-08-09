package config

type Message struct {
	Type      int    `json:"type"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Data      string `json:"data"`
}
