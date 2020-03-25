package service

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"obd/bean"
	"obd/config"
	"obd/dao"
	"obd/tool"
)

type Wallet struct {
	Index   int    `json:"index"`
	Address string `json:"address"`
	PubKey  string `json:"pub_key"`
	//PrivateKey    string `json:"private_key"`
	Wif string `json:"wif"`
}

type hdWalletManager struct {
}

var HDWalletService hdWalletManager

func (service *hdWalletManager) GetAddressByIndex(user *bean.User, index uint32) (wallet *Wallet, err error) {

	if user == nil || user.ChangeExtKey == nil {
		return nil, errors.New("error mnemonic")
	}

	changeExtKey := user.ChangeExtKey
	//m/purpose(44)'/coinType(0)'/account(0)'/change(0)/index(0)
	addrIndexExtKey, _ := changeExtKey.NewChildKey(index)

	wallet = &Wallet{}
	wallet.Index = int(index)
	err = getWalletObj(addrIndexExtKey, wallet)
	return wallet, err
}

func getWalletObj(addrIndexExtKey *bip32.Key, wallet *Wallet) (err error) {
	net := &chaincfg.MainNetParams
	if config.ChainNode_Type == "test" {
		net = &chaincfg.TestNet3Params
	}
	hash160Bytes := btcutil.Hash160(addrIndexExtKey.PublicKey().Key)
	addr, err := btcutil.NewAddressPubKeyHash(hash160Bytes, net)
	if err != nil {
		return err
	}
	wallet.Address = addr.String()

	wallet.PubKey = hex.EncodeToString(addrIndexExtKey.PublicKey().Key)

	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), addrIndexExtKey.Key)

	//wallet.PrivateKey = hex.EncodeToString(privKey.Serialize())
	wif, _ := btcutil.NewWIF(privKey, net, true)
	//wallet.Wifobj = wif
	//wallet.PrivateKeyObj = privKey
	wallet.Wif = wif.String()

	return nil
}
func (service *hdWalletManager) CreateNewAddress(user *bean.User) (wallet *Wallet, err error) {
	if user == nil || user.ChangeExtKey == nil {
		return nil, errors.New("error mnemonic")
	}

	changeExtKey := user.ChangeExtKey
	//m/purpose(44)'/coinType(0)'/account(0)'/change(0)/index(0)
	var userDb dao.User
	err = db.One("PeerId", user.PeerId, &userDb)
	if err != nil {
		return nil, err
	}
	user.CurrAddrIndex = user.CurrAddrIndex + 1
	userDb.CurrAddrIndex = user.CurrAddrIndex
	addrIndexExtKey, _ := changeExtKey.NewChildKey(uint32(user.CurrAddrIndex))

	wallet = &Wallet{}
	wallet.Index = userDb.CurrAddrIndex

	err = getWalletObj(addrIndexExtKey, wallet)
	if err != nil {
		return nil, err
	}
	_ = db.Update(&userDb)
	return wallet, nil
}

func (service *hdWalletManager) CreateChangeExtKey(mnemonic string) (changeExtKey *bip32.Key, err error) {
	if tool.CheckIsString(&mnemonic) == false {
		return nil, errors.New("error mnemonic")
	}
	valid := bip39.IsMnemonicValid(mnemonic)
	if valid == false {
		return nil, errors.New("wrong mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, "")
	masterKey, _ := bip32.NewMasterKey(seed)
	//m/purpose'
	purposeExtKey, _ := masterKey.NewChildKey(bip32.FirstHardenedChild + 44)
	//m/purpose'/cointype'
	coinType := 0
	if config.ChainNode_Type == "test" {
		coinType = 1
	}
	coinTypeExtKey, err := purposeExtKey.NewChildKey(bip32.FirstHardenedChild + uint32(coinType))
	//m/purpose'/cointype'/account'
	accountExtKey, _ := coinTypeExtKey.NewChildKey(bip32.FirstHardenedChild + 0)
	//m/purpose'/cointype'/account'/change
	return accountExtKey.NewChildKey(0)
}
