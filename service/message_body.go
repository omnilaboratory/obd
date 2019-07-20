package service

type MessageBody struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Data      string `json:"data,omitempty"`
}
