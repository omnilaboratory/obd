package dao

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"time"
)

type User struct {
	Id int `storm:"id,increment" `
	bean.User
}

type ChannelState int

const (
	ChannelState_Create        ChannelState = 10
	ChannelState_Accept        ChannelState = 20
	ChannelState_Close         ChannelState = 21
	ChannelState_Defuse        ChannelState = 30
	ChannelState_FundingDefuse ChannelState = 31
)

type ChannelInfo struct {
	bean.OpenChannelInfo
	Id            int            `storm:"id,increment" json:"id"`
	PeerIdA       string         `json:"peer_id_a"`
	PeerIdB       string         `json:"peer_id_b"`
	PubKeyA       string         `json:"pub_key_a"`
	PubKeyB       string         `json:"pub_key_b"`
	ChannelPubKey string         `json:"channel_pub_key"`
	RedeemScript  string         `json:"redeem_script"`
	ChannelId     bean.ChannelID `json:"channel_id"`
	CurrState     ChannelState   `json:"curr_state"`
	CreateAt      time.Time      `json:"create_at"`
	AcceptAt      time.Time      `json:"accept_at"`
}

type CloseChannel struct {
	bean.CloseChannel
	Id       int       `storm:"id,increment" json:"id"`
	CreateAt time.Time `json:"create_at"`
}

type FundingTransactionState int

const (
	FundingTransactionState_Create FundingTransactionState = 10
	FundingTransactionState_Accept FundingTransactionState = 20
	FundingTransactionState_Defuse FundingTransactionState = 30
)

type FundingTransaction struct {
	Id                         int                     `storm:"id,increment" `
	PeerIdA                    string                  `json:"peer_id_a"`
	PeerIdB                    string                  `json:"peer_id_b"`
	TemporaryChannelId         chainhash.Hash          `json:"temporary_channel_id"`
	ChannelId                  bean.ChannelID          `json:"channel_id"`
	PropertyId                 int64                   `json:"property_id"`
	FunderPubKey               string                  `json:"funder_pub_key"`
	AmountA                    float64                 `json:"amount_a"`
	FunderPubKey2ForCommitment string                  `json:"funder_pub_key_2_for_commitment"`
	FundingTxid                string                  `json:"funding_txid"`
	FundingOutputIndex         uint32                  `json:"funding_output_index"`
	FunderSignature            string                  `json:"funder_signature"`
	FundeePubKey               string                  `json:"fundee_pub_key"`
	AmountB                    float64                 `json:"amount_b"`
	FundeeSignature            string                  `json:"fundee_signature"`
	ChannelPubKey              string                  `json:"channel_pub_key"`
	RedeemScript               string                  `json:"redeem_script"`
	CreateBy                   string                  `json:"create_by"`
	CreateAt                   time.Time               `json:"create_at"`
	FundeeSignAt               time.Time               `json:"fundee_sign_at"`
	TxId                       string                  `json:"tx_id"`
	CurrState                  FundingTransactionState `json:"curr_state"`
}

type TxInfoState int

const (
	TxInfoState_OtherSign  TxInfoState = 10
	TxInfoState_MyselfSign TxInfoState = 20
	TxInfoState_Abord      TxInfoState = 30
)

//CommitmentTransaction
type CommitmentTxInfo struct {
	Id                 int            `storm:"id,increment" json:"id" `
	LastCommitmentTxId int            `json:"last_commitment_tx_id"`
	PeerIdA            string         `json:"peer_id_a"`
	PeerIdB            string         `json:"peer_id_b"`
	ChannelId          bean.ChannelID `json:"channel_id"`
	PropertyId         int64          `json:"property_id"`
	CreatorSide        int            `json:"creator_side"`  // 0 alice 1 bob
	InputTxid          string         `json:"input_txid"`    //input txid  from channeladdr: alice&bob multtaddr, so need  sign of alice2 and bob
	InputVout          uint32         `json:"input_vout"`    // input vout
	InputAmount        float64        `json:"input_amount"`  //input amount
	PubKey2            string         `json:"pub_key2"`      //output alice2
	PubKeyB            string         `json:"pub_key_b"`     //output bob
	MultiAddress       string         `json:"multi_address"` //output alice2&bob multiaddr
	RedeemScript       string         `json:"redeem_script"`
	AmountM            float64        `json:"amount_m"` // amount to multiaddr
	AmountB            float64        `json:"amount_b"` //amount to bob
	TxHexFirstSign     string         `json:"tx_hex_first_sign"`
	TxHexEndSign       string         `json:"tx_hex_end_sign"`
	Txid               string         `json:"txid"`
	CurrState          TxInfoState    `json:"curr_state"`
	CreateBy           string         `json:"create_by"`
	CreateAt           time.Time      `json:"create_at"`
	FirstSignAt        time.Time      `json:"first_sign_at"`
	EndSignAt          time.Time      `json:"end_sign_at"`
	LastEditTime       time.Time      `json:"last_edit_time"`
}

// close channel , alice or bob wait 1000 sequence to drawback the balance
type RevocableDeliveryTransaction struct {
	Id             int            `storm:"id,increment" json:"id" `
	CommitmentTxId int            `json:"commitment_tx_id"`
	PeerIdA        string         `json:"peer_id_a"`
	PeerIdB        string         `json:"peer_id_b"`
	ChannelId      bean.ChannelID `json:"channel_id"`
	PropertyId     int64          `json:"property_id"`
	CreatorSide    int            `json:"creator_side"` // 0 alice 1 bob
	InputTxid      string         `json:"input_txid"`   //input txid  from commitTx alice2&bob multtaddr, so need  sign of alice2 and bob
	InputVout      uint32         `json:"input_vout"`   // input vout
	InputAmount    float64        `json:"input_amount"` //input amount
	PubKeyA        string         `json:"pub_key_a"`    //output alice
	Sequnence      int            `json:"sequnence"`
	Amount         float64        `json:"amount"` // output alice amount
	TxHexFirstSign string         `json:"tx_hex_first_sign"`
	TxHexEndSign   string         `json:"tx_hex_end_sign"`
	Txid           string         `json:"txid"`
	CurrState      TxInfoState    `json:"curr_state"`
	CreateBy       string         `json:"create_by"`
	CreateAt       time.Time      `json:"create_at"`
	FirstSignAt    time.Time      `json:"first_sign_at"`
	EndSignAt      time.Time      `json:"end_sign_at"`
	LastEditTime   time.Time      `json:"last_edit_time"`
}

// to punish alice do not admit the newest commitment tx
type BreachRemedyTransaction struct {
	Id             int            `storm:"id,increment" json:"id" `
	CommitmentTxId int            `json:"commitment_tx_id"` // parent commitmentTx id
	PeerIdA        string         `json:"peer_id_a"`
	PeerIdB        string         `json:"peer_id_b"`
	ChannelId      bean.ChannelID `json:"channel_id"`
	PropertyId     int64          `json:"property_id"`
	CreatorSide    int            `json:"creator_side"`      // 0 alice 1 bob
	InputTxid      string         `json:"input_txid"`        //input txid  from commitTx alice2&bob multtAddr, so need  sign of alice2 and bob
	InputVout      uint32         `json:"input_vout"`        // input vout
	InputAmount    float64        `json:"input_amount"`      //input amount
	PubKeyB        string         `json:"pub_key_b"`         //output bob
	Amount         float64        `json:"amount"`            // output boob amount
	TxHexFirstSign string         `json:"tx_hex_first_sign"` // first alice2 sign
	TxHexEndSign   string         `json:"tx_hex_end_sign"`   // end bob sign
	Txid           string         `json:"txid"`
	CurrState      TxInfoState    `json:"curr_state"`
	CreateBy       string         `json:"create_by"`
	CreateAt       time.Time      `json:"create_at"`
	FirstSignAt    time.Time      `json:"first_sign_at"`
	EndSignAt      time.Time      `json:"end_sign_at"`
	LastEditTime   time.Time      `json:"last_edit_time"`
}

type GetBalanceRequest struct {
	Id int `storm:"id,increment" `
	bean.GetBalanceRequest
	CreateAt time.Time `json:"create_at"`
}

type GetBalanceRespond struct {
	Id int `storm:"id,increment" `
	bean.GetBalanceRespond
	CreateAt time.Time `json:"create_at"`
}
