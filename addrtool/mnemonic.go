package addrtool

import (
	"errors"
	"github.com/tyler-smith/go-bip39"
)

func Bip39GenMnemonic(size int) (mnemonic string, err error) {
	entropy, err := bip39.NewEntropy(size)
	if err != nil {
		return "", nil
	}
	return bip39.NewMnemonic(entropy)
}

func Bip39MnemonicToSeed(mnemonic string, password string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("mnemonic not valid")
	}
	return bip39.NewSeed(mnemonic, password), nil
}
