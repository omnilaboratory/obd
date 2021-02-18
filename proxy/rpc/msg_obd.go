package rpc

type loginInfo struct {
	Mnemonic   string `json:"mnemonic"`
	LoginToken string `json:"login_token"`
}

type updateLoginToken struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
