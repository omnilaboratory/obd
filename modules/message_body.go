package modules

/*
这个结构需要改成二进制的结构，传输和加解码效率比较好，推荐用google的protoBuff

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
