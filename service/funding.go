package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/dao"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"github.com/tidwall/gjson"
	"golang.org/x/crypto/salsa20"
	"log"
	"sync"
)

type FundingManager struct {
	nonceMtx    sync.RWMutex
	chanIDNonce uint64
	chanIDSeed  [32]byte
}

var FundingCreateService FundingManager

func init() {
	if _, err := rand.Read(FundingCreateService.chanIDSeed[:]); err != nil {
		log.Println(err)
	}
}

// NextTemporaryChanID returns the next free pending channel ID to be used to identify a particular future channel funding workflow.
func (service *FundingManager) NextTemporaryChanID() [32]byte {
	// Obtain a fresh nonce. We do this by encoding the current nonce counter, then incrementing it by one.
	service.nonceMtx.Lock()
	var nonce [8]byte
	binary.LittleEndian.PutUint64(nonce[:], service.chanIDNonce)
	service.chanIDNonce++
	service.nonceMtx.Unlock()

	// We'll generate the next pending channelID by "encrypting" 32-bytes of zeroes which'll extract 32 random bytes from our stream cipher.
	var (
		nextChanID [32]byte
		zeroes     [32]byte
	)
	salsa20.XORKeyStream(nextChanID[:], zeroes[:], nonce[:], &service.chanIDSeed)
	return nextChanID
}

func (service *FundingManager) CreateFunding(jsonData string) (node *dao.FundingCreated, err error) {
	node = &dao.FundingCreated{}

	tempId := service.NextTemporaryChanID()
	node.TemporaryChannelIdStr = string(tempId[:])
	node.FunderPubKeyStr = gjson.Get(jsonData, "funderPubKey").String()
	hashes, _ := chainhash.NewHashFromStr(node.FunderPubKeyStr)
	data := &bean.FundingCreated{
		TemporaryChannelId: tempId,
		FunderPubKey:       *hashes,
		PropertyId:         gjson.Get(jsonData, "propertyId").Int(),
		MaxAssets:          gjson.Get(jsonData, "maxAssets").Float(),
		AmountA:            gjson.Get(jsonData, "amountA").Float(),
	}
	node.FundingCreated = *data

	db, _ := dao.DB_Manager.GetDB()
	err = db.Save(node)
	return node, err
}

func (service *FundingManager) GetFundingTx(id int) (node *dao.FundingCreated, err error) {
	db, _ := dao.DB_Manager.GetDB()
	var data = &dao.FundingCreated{}
	err = db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *FundingManager) DeleteTable() (err error) {
	//db, _ := dao.DB_Manager.GetDB()
	//var data = &dao.FundingCreated{}
	//return db.Drop(data)
	return nil
}
func (service *FundingManager) DeleteItem(id int) (err error) {
	db, _ := dao.DB_Manager.GetDB()
	var data = &dao.FundingCreated{}
	db.One("Id", id, data)
	err = db.DeleteStruct(data)
	return err
}
func (service *FundingManager) TotalCount() (count int, err error) {
	db, _ := dao.DB_Manager.GetDB()
	var data = &dao.FundingCreated{}
	return db.Count(data)
}

type FundingSignManager struct{}

var FundingSignService FundingSignManager

func (service *FundingSignManager) Edit(jsonData string) (signed *dao.FundingSigned, err error) {
	vo := &bean.FundingSigned{}
	err = json.Unmarshal([]byte(jsonData), vo)
	if err != nil {
		return nil, err
	}
	return signed, nil
}
