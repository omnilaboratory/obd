package service

import (
	"encoding/hex"
	"errors"
	"github.com/asdine/storm/q"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"log"
	"path"
	"strings"
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
	net := tool.GetCoreNet()
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
	user.CurrAddrIndex = user.CurrAddrIndex + 1
	addrIndexExtKey, _ := changeExtKey.NewChildKey(uint32(user.CurrAddrIndex))

	wallet = &Wallet{}
	wallet.Index = user.CurrAddrIndex

	err = getWalletObj(addrIndexExtKey, wallet)
	if err != nil {
		return nil, err
	}

	var node dao.User
	_ = user.Db.Select(q.Eq("PeerId", user.PeerId)).First(&node)
	if node.Id > 0 {
		node.CurrAddrIndex = user.CurrAddrIndex
		_ = user.Db.Update(&node)
	}

	return wallet, nil
}

func (service *hdWalletManager) CreateChangeExtKey(mnemonic, seedPwd string) (changeExtKey *bip32.Key, err error) {
	if tool.CheckIsString(&mnemonic) == false {
		return nil, errors.New("error mnemonic")
	}
	valid := bip39.IsMnemonicValid(mnemonic)
	if valid == false {
		return nil, errors.New("wrong mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, seedPwd)
	masterKey, _ := bip32.NewMasterKey(seed)
	//m/purpose'
	purposeExtKey, _ := masterKey.NewChildKey(bip32.FirstHardenedChild + 44)
	//m/purpose'/cointype'
	coinType := 0
	if strings.Contains(config.ChainNodeType, "main") == false {
		coinType = 1
	}

	coinTypeExtKey, err := purposeExtKey.NewChildKey(bip32.FirstHardenedChild + uint32(coinType))
	return coinTypeExtKey, err

	//m/purpose'/cointype'/account'
	accountExtKey, _ := coinTypeExtKey.NewChildKey(bip32.FirstHardenedChild + 0)
	//m/purpose'/cointype'/account'/change
	return accountExtKey.NewChildKey(0)
}

func (service *hdWalletManager) ChangePwd(currentPasswd,newPasswd string) (err error) {
	key,err:= service.LoadNodeMasterKey(currentPasswd)
	if err == nil {
			err=service.SaveNodeMasterKey(key,newPasswd)
	}
	return err
}
func (service *hdWalletManager) LoadNodeMasterKey(passwd string) (changeExtKey *bip32.Key, err error) {
	var keyfilePath=path.Join(config.DataDirectory,"nodekey")
	keybs,err:= decryptHexFile(keyfilePath,passwd)
	if err == nil {
		changeExtKey,err=bip32.Deserialize(keybs)
	}
	return changeExtKey,err
}
func (service *hdWalletManager) GetKeyAddress(key *bip32.Key,passwd string) ( string) {
	net := tool.GetCoreNet()
	hash160Bytes := btcutil.Hash160(key.PublicKey().Key)
	addr, err := btcutil.NewAddressPubKeyHash(hash160Bytes, net)
	if err != nil {
		panic(err)
	}
	return addr.String()
}
func (service *hdWalletManager) SaveNodeMasterKey(key *bip32.Key,passwd string) ( err error) {
	var keybs []byte
	keybs,err=key.Serialize()
	if err == nil {
		var keyfilePath=path.Join(config.DataDirectory,"nodekey")
		err= encryptHexFile(keyfilePath,keybs,passwd)
		if err == nil {
			log.Println("SaveNodeMasterKey",keyfilePath)

		}
	}
	return  err
}
func (service *hdWalletManager) SaveNodeMasterKeyFormMnemonic(mnemonic,passwd,seedPwd string) (changeExtKey *bip32.Key,err error) {
	key,err:=service.CreateChangeExtKey(mnemonic,seedPwd)
	if err == nil {
		err=service.SaveNodeMasterKey(key,passwd)
	}
	return key,err
}
func (service *hdWalletManager) GenSeed() (mnemonic string, err error) {
	return service.Bip39GenMnemonic(256)
}