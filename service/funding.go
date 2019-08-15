package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/dao"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
	"golang.org/x/crypto/salsa20"
	"log"
	"sync"
)

type FundingCreateManager struct {
	nonceMtx    sync.RWMutex
	chanIDNonce uint64
	chanIDSeed  [32]byte
}

var FundingCreateService FundingCreateManager

func init() {
	if _, err := rand.Read(FundingCreateService.chanIDSeed[:]); err != nil {
		log.Println(err)
	}
}

// NextTemporaryChanID returns the next free pending channel ID to be used to identify a particular future channel funding workflow.
func (service *FundingCreateManager) NextTemporaryChanID() [32]byte {
	// Obtain a fresh nonce. We do this by encoding the current nonce counter, then incrementing it by one.
	service.nonceMtx.Lock()
	var nonce [8]byte
	binary.LittleEndian.PutUint64(nonce[:], service.chanIDNonce)
	service.chanIDNonce++
	service.nonceMtx.Unlock()

	// We'll generate the next temporary channelID by "encrypting" 32-bytes of zeroes which'll extract 32 random bytes from our stream cipher.
	var (
		nextChanID [32]byte
		zeroes     [32]byte
	)
	salsa20.XORKeyStream(nextChanID[:], zeroes[:], nonce[:], &service.chanIDSeed)
	return nextChanID
}

func (service *FundingCreateManager) CreateFunding(jsonData string) (node *dao.FundingCreated, err error) {
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

func (service *FundingCreateManager) GetFundingTx(id int) (node *dao.FundingCreated, err error) {
	db, _ := dao.DB_Manager.GetDB()
	var data = &dao.FundingCreated{}
	err = db.One("Id", id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (service *FundingCreateManager) DeleteTable() (err error) {
	//db, _ := dao.DB_Manager.GetDB()
	//var data = &dao.FundingCreated{}
	//return db.Drop(data)
	return nil
}
func (service *FundingCreateManager) DeleteItem(id int) (err error) {
	db, _ := dao.DB_Manager.GetDB()
	var data = &dao.FundingCreated{}
	db.One("Id", id, data)
	err = db.DeleteStruct(data)
	return err
}
func (service *FundingCreateManager) TotalCount() (count int, err error) {
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

	vo.TemporaryChannelId = FundingCreateService.NextTemporaryChanID()

	db, _ := dao.DB_Manager.GetDB()
	node := &dao.FundingSigned{}
	//https://www.ctolib.com/storm.html
	err = db.Select(
		q.Eq("FundeePubKey", vo.FundeePubKey),
		q.And(
			q.Eq("FunderPubKey", vo.FunderPubKey),
		),
	).First(node)
	node.FundingSigned = *vo
	if err != nil {
		err = db.Save(node)
	} else {
		err = db.Update(node)
	}
	return node, err
}
func (service *FundingSignManager) Item(id int) (signed *dao.FundingSigned, err error) {
	node := &dao.FundingSigned{}
	db, _ := dao.DB_Manager.GetDB()
	err = db.One("Id", id, node)
	return node, err
}
func (service *FundingSignManager) Del(id int) (signed *dao.FundingSigned, err error) {
	db, _ := dao.DB_Manager.GetDB()
	node := &dao.FundingSigned{}
	err = db.One("Id", id, node)
	if err == nil {
		err = db.DeleteStruct(node)
	}
	return node, err
}
func (service *FundingSignManager) DelAll() (err error) {
	db, _ := dao.DB_Manager.GetDB()
	err = db.Drop(&dao.FundingSigned{})
	return err
}

func (service *FundingSignManager) TotalCount() (count int, err error) {
	db, _ := dao.DB_Manager.GetDB()
	return db.Count(&dao.FundingSigned{})
}
