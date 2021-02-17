package rpc

type loginInfo struct {
	Mnemonic   string `json:"mnemonic"`
	LoginToken string `json:"login_token"`
}

type updateLoginToken struct {
	OldLoginToken string `json:"old_login_token"`
	NewLoginToken string `json:"new_login_token"`
}
