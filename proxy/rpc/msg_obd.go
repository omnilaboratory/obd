package rpc

type loginInfo struct {
	Mnemonic   string `json:"mnemonic"`
	LoginToken string `json:"login_token"`
}

type updateLoginToken struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type InvoiceInfo struct {
	ExpiryTime  string  `json:"expiry_time"`
	PropertyId  int64   `json:"property_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	IsPrivate   bool    `json:"is_private"`
}
type ParseInvoice struct {
	Invoice string `json:"invoice"`
}
