package service

import (
	"LightningOnOmni/config/chainhash"
)

//type: -34 (funding_created)
type Funding_created struct {
	Id                   int            `storm:"id,increment" `
	Temporary_channel_id chainhash.Hash `json:"temporary_channel_id"`
	Funder_pubKey        chainhash.Hash `json:"funder_pub_key"`
	Property_id          int            `json:"property_id"`
	Max_assets           float64        `json:"max_assets"`
	Amount_a             float64        `json:"amount_a"`
}

type Funding_Service struct {
}

var FundingService Funding_Service

func (service *Funding_Service) CreateFunding() error {
	db, e := DB_Manager.GetDB()
	if e != nil {
		return e
	}
	tempId, _ := Channel_Service.getTemporayChaneelId()
	hashes, _ := chainhash.NewHashFromStr("abc")
	node := &Funding_created{
		Temporary_channel_id: *tempId,
		Funder_pubKey:        *hashes,
		Property_id:          31,
		Max_assets:           1000,
		Amount_a:             20,
	}
	return db.Save(node)
}

//type: -35 (funding_signed)
type Funding_signed struct {
	//the same as the temporary_channel_id in the open_channel message
	Temporary_channel_id chainhash.ChainHash
	//the omni address of funder Alice
	Funder_pubKey chainhash.Hash
	// the id of the Omni asset
	Asset_id int
	//amount of the asset on Alice side
	Amount_a float64
	//the omni address of fundee Bob
	Fundee_pubKey chainhash.Hash
	//amount of the asset on Bob side
	Amount_b float64
	//signature of fundee Bob
	Fundee_signature chainhash.Signauture
	//redeem script used to generate P2SH address
	RedeemScript string
	//hash of redeemScript
	P2sh_address chainhash.Hash
	//final global channel id generated
	Channel_id chainhash.Hash
}
