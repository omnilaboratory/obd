package bean

import (
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/bean/enum"
)

type RequestMessage struct {
	Type            enum.MsgType `json:"type"`
	SenderPeerId    string       `json:"sender_peer_id"`
	RecipientPeerId string       `json:"recipient_peer_id"`
	Data            string       `json:"data"`
	PubKey          string       `json:"pub_key"`
	Signature       string       `json:"signature"`
}
type ReplyMessage struct {
	Type   enum.MsgType `json:"type"`
	Status bool         `json:"status"`
	From   string       `json:"from"`
	To     string       `json:"to"`
	Result interface{}  `json:"result"`
}

//type = 1
type User struct {
	PeerId   string         `json:"peer_id"`
	Password string         `json:"password"`
	State    enum.UserState `json:"state"`
}

//https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/OmniBOLT-03-RSMC-and-OmniLayer-Transactions.md
//type = -32
type OpenChannelInfo struct {
	ChainHash                chainhash.ChainHash `json:"chain_hash"`
	TemporaryChannelId       chainhash.Hash      `json:"temporary_channel_id"`
	FundingSatoshis          uint64              `json:"funding_satoshis"`
	PushMsat                 uint64              `json:"push_msat"`
	DustLimitSatoshis        uint64              `json:"dust_limit_satoshis"`
	MaxHtlcValueInFlightMsat uint64              `json:"max_htlc_value_in_flight_msat"`
	ChannelReserveSatoshis   uint64              `json:"channel_reserve_satoshis"`
	HtlcMinimumMsat          uint64              `json:"htlc_minimum_msat"`
	FeeRatePerKw             uint32              `json:"fee_rate_per_kw"`
	ToSelfDelay              uint16              `json:"to_self_delay"`
	MaxAcceptedHtlcs         uint16              `json:"max_accepted_htlcs"`
	FundingPubKey            string              `json:"funding_pubkey"`
	FundingAddress           string              `json:"funding_address"`
	RevocationBasePoint      chainhash.Point     `json:"revocation_base_point"`
	PaymentBasePoint         chainhash.Point     `json:"payment_base_point"`
	DelayedPaymentBasePoint  chainhash.Point     `json:"delayed_payment_base_point"`
	HtlcBasePoint            chainhash.Point     `json:"htlc_base_point"`
}

//type = -33
type AcceptChannelInfo struct {
	OpenChannelInfo
	Approval bool `json:"approval"`
}

//type: -38 (close_channel)
type CloseChannel struct {
	ChannelId ChannelID `json:"channel_id"`
}

//type: -39 (close_channel_sign)
type CloseChannelSign struct {
	ChannelId               ChannelID `json:"channel_id"`
	RequestCloseChannelHash string    `json:"request_close_channel_hash"`
	Approval                bool      `json:"approval"` // true agree false disagree
}

//type: -35107 (SendBreachRemedyTransaction)
type SendBreachRemedyTransaction struct {
	ChannelId                ChannelID `json:"channel_id"`
	ChannelAddressPrivateKey string    `json:"channel_address_private_key"` // openChannel address
}

//type: -34 (funding_created)
type FundingCreated struct {
	TemporaryChannelId       chainhash.Hash `json:"temporary_channel_id"`
	PropertyId               int64          `json:"property_id"`
	MaxAssets                float64        `json:"max_assets"`
	AmountA                  float64        `json:"amount_a"`
	FundingTxHex             string         `json:"funding_tx_hex"`
	TempAddressPubKey        string         `json:"temp_address_pub_key"`
	TempAddressPrivateKey    string         `json:"temp_address_private_key"`
	ChannelAddressPrivateKey string         `json:"channel_address_private_key"`
}

//type: -3400 (FundingBtcCreated)
type FundingBtcCreated struct {
	TemporaryChannelId       chainhash.Hash `json:"temporary_channel_id"`
	Amount                   float64        `json:"amount_a"`
	FundingTxHex             string         `json:"funding_tx_hex"`
	ChannelAddressPrivateKey string         `json:"channel_address_private_key"`
}

//type: -3500 (FundingBtcSigned)
type FundingBtcSigned struct {
	TemporaryChannelId       chainhash.Hash `json:"temporary_channel_id"`
	FundingTxid              string         `json:"funding_txid"`
	ChannelAddressPrivateKey string         `json:"channel_address_private_key"`
	Approval                 bool           `json:"approval"`
}

//type: -35 (funding_signed)
type FundingSigned struct {
	ChannelId ChannelID `json:"channel_id"`
	//the omni address of funder Alice
	FunderPubKey string `json:"funder_pub_key"`
	// the id of the Omni asset
	PropertyId int `json:"property_id"`
	//amount of the asset on Alice side
	AmountA float64 `json:"amount_a"`
	//the omni address of fundee Bob
	FundeePubKey string `json:"fundee_pub_key"`
	//amount of the asset on Bob side
	AmountB float64 `json:"amount_b"`
	//signature of fundee Bob
	FundeeChannelAddressPrivateKey string `json:"fundee_channel_address_private_key"`
	//redeem script used to generate P2SH address
	RedeemScript string `json:"redeem_script"`
	//hash of redeemScript
	P2shAddress string `json:"p2sh_address"`
	Approval    bool   `json:"approval"`
}

//type: -351 (commitment_tx)
type CommitmentTx struct {
	ChannelId                 ChannelID `json:"channel_id"` //the global channel id.
	RequestCommitmentHash     string    `json:"request_commitment_hash"`
	PropertyId                int       `json:"property_id"` //the id of the Omni asset
	Amount                    float64   `json:"amount"`      //amount of the payment
	ChannelAddressPrivateKey  string    `json:"channel_address_private_key"`
	LastTempAddressPrivateKey string    `json:"last_temp_address_private_key"`
	CurrTempAddressPubKey     string    `json:"curr_temp_address_pub_key"`
	CurrTempAddressPrivateKey string    `json:"curr_temp_address_private_key"`
}

//type: -352 (commitment_tx_signed)
type CommitmentTxSigned struct {
	ChannelId                 ChannelID `json:"channel_id"`
	RequestCommitmentHash     string    `json:"request_commitment_hash"`
	ChannelAddressPrivateKey  string    `json:"channel_address_private_key"`   // bob private key
	LastTempAddressPrivateKey string    `json:"last_temp_address_private_key"` // bob2's private key
	CurrTempAddressPubKey     string    `json:"curr_temp_address_pub_key"`     // bob3 or alice3
	CurrTempAddressPrivateKey string    `json:"curr_temp_address_private_key"`
	Approval                  bool      `json:"approval"` // true agree false disagree
}

//type: -353 (get_balance_request)
type GetBalanceRequest struct {
	//the global channel id.
	ChannelId ChannelID `json:"channel_id"`
	//the p2sh address generated in funding_signed message.
	P2shAddress string `json:"p2sh_address"`
	// the channel owner, Alice or Bob, can query the balance.
	Who chainhash.Hash `json:"who"`
	//the signature of Alice or Bob
	Signature chainhash.Signature `json:"signature"`
}

//type: -354 (get_balance_respond)
type GetBalanceRespond struct {
	//the global channel id.
	ChannelId ChannelID `json:"channel_id"`
	//the asset id generated by Omnilayer protocol.
	PropertyId int `json:"property_id"`
	//the name of the asset.
	Name string `json:"name"`
	//balance in this channel
	Balance float64 `json:"balance"`
	//currently not in use
	Reserved float64 `json:"reserved"`
	//currently not in use
	Frozen float64 `json:"frozen"`
}

//type -40: alice tell carl ,she wanna transfer some money to Carl
type HtlcHRequest struct {
	PropertyId      int     `json:"property_id"`
	Amount          float64 `json:"amount"`
	RecipientPeerId string  `json:"recipient_peer_id"`
}

//type -41: carl tell alice the H,and he ca
type HtlcHRespond struct {
	RequestHash string  `json:"request_hash"`
	Approval    bool    `json:"approval"` // true agree false disagree
	PropertyId  int     `json:"property_id"`
	Amount      float64 `json:"amount"`
}

//type -43: alice request create htlc
type HtlcRequestCreate struct {
	H string `json:"h"`
}

//type -44: bob sign the request
type HtlcSignRequestCreate struct {
	Approval bool `json:"approval"` // true agree false disagree
}
