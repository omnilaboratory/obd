package addrtool

import (
	"LightningOnOmni/tool"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	Index   int    `json:"index"`
	Address string `json:"address"`
	PubKey  string `json:"pub_key"`
	Wif     string `json:"wif"`
}

func CreateWallet(mnemonic string, index uint32) (wallet *Wallet, err error) {
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
	coinTypeExtKey, _ := purposeExtKey.NewChildKey(bip32.FirstHardenedChild + 0)
	//m/purpose'/cointype'/account'
	accountExtKey, _ := coinTypeExtKey.NewChildKey(bip32.FirstHardenedChild + 0)
	//m/purpose'/cointype'/account'/change
	changeExtKey, _ := accountExtKey.NewChildKey(0)
	//m/purpose(44)'/coinType(0)'/account(0)'/change(0)/index(0)
	addrIndexExtKey, _ := changeExtKey.NewChildKey(index)

	wallet = &Wallet{}
	hash160Bytes := btcutil.Hash160(addrIndexExtKey.PublicKey().Key)
	wallet.Address = base58.CheckEncode(hash160Bytes[:ripemd160.Size], 0)
	wallet.PubKey = hex.EncodeToString(addrIndexExtKey.PublicKey().Key)

	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), addrIndexExtKey.Key)
	wif, _ := btcutil.NewWIF(privKey, &chaincfg.MainNetParams, true)
	wallet.Wif = wif.String()
	return wallet, nil
}
