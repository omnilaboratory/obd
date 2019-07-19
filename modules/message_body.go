package modules

/*
this structure will use protoBuffer for consistent communication and better performance.

type MessageBody struct {
    Sender    string `json:"sender,omitempty"`
    Recipient string `json:"recipient,omitempty"`
    Content   string `json:"content,omitempty"`

    TokenName string `json:"token_name,omitempty"`
    Price     string `price:"price,omitempty"`
    Amount     string `price:"amount,omitempty"`
    AcceptedCurrency string `accepted_currency:"accepted_currency,omitempty"`
}
*/
type MessageBody struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Data      string `json:"data,omitempty"`
}
