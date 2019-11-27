package addrtool

import (
	"crypto/sha256"
	"errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

func Bip39GenMnemonic() (string, error) {
	size := 256
	entropy, _ := bip39.NewEntropy(size)
	return bip39.NewMnemonic(entropy)
}

func Bip39MnemonicToSeed(mnemonic string, password string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("mnemonic not valid")
	}
	return bip39.NewSeed(mnemonic, password), nil
}

//func DcrSeedToMnemonic(seed []byte) string {
//	var buf bytes.Buffer
//	for i, b := range seed {
//		if i != 0 {
//			buf.WriteRune(' ')
//		}
//		buf.WriteString(byteToMnemonic(b, i))
//	}
//	checksum := checksumByte(seed)
//	buf.WriteRune(' ')
//	buf.WriteString(byteToMnemonic(checksum, len(seed)))
//	return buf.String()
//}

// DecodeMnemonics returns the decoded value that is encoded by words.  Any
// words that are whitespace are empty are skipped.
//func DcrMnemonicToSeed(words []string) ([]byte, error) {
//	decoded := make([]byte, len(words))
//	idx := 0
//	for _, w := range words {
//		w = strings.TrimSpace(w)
//		if w == "" {
//			continue
//		}
//		b, ok := wordIndexes[strings.ToLower(w)]
//		if !ok {
//			return nil, errors.New("word %v is not in the PGP word list")
//		}
//		if int(b%2) != idx%2 {
//			return nil, errors.New("word %v is not valid at position %v, ")
//		}
//		decoded[idx] = byte(b / 2)
//		idx++
//	}
//	return decoded[:idx], nil
//}

func checksumByte(data []byte) byte {
	intermediateHash := sha256.Sum256(data)
	return sha256.Sum256(intermediateHash[:])[0]
}

// byteToMnemonic returns the PGP word list encoding of b when found at index.
//func byteToMnemonic(b byte, index int) string {
//	bb := uint16(b) * 2
//	if index%2 != 0 {
//		bb++
//	}
//	return wordList[bb]
//}

type Key struct {
	private  bool
	k        *hdkeychain.ExtendedKey
	mnemonic string
}

type CoinType uint32

const (
	BTC CoinType = 0
)

func (k *Key) deriveKey(coinTyp CoinType, account, index uint32) (*Key, error) {
	// m/49'
	purpose, err := k.k.Child(44)
	if err != nil {
		return nil, err
	}

	// m/49'/1'
	coinType, err := purpose.Child(uint32(coinTyp))
	if err != nil {
		return nil, err
	}

	// m/49'/1'/0'
	acctX, err := coinType.Child(account)
	if err != nil {
		return nil, err
	}
	// Derive the extended key for the account 0 external chain.  This
	// gives the path:
	//   m/0H/0
	// 0 is external, 1 is internal address (used for change, wallet software)
	acctXExt, err := acctX.Child(0)
	if err != nil {
		return nil, err
	}
	// Derive the Indexth extended key for the account X external chain.
	// m/49'/1'/0'/0
	acctXExternalX, err := acctXExt.Child(index)
	if err != nil {
		return nil, err
	}
	return &Key{private: k.private, k: acctXExternalX}, nil
}

// creates the Wallet Import Format string encoding of a WIF structure.
func (k *Key) WIF(coinTyp CoinType, account, index uint32) (string, error) {
	acctXExternalX, err := k.deriveKey(coinTyp, account, index)
	if err != nil {
		return "", err
	}
	compress := false //?
	priv, err := acctXExternalX.k.ECPrivKey()
	if err != nil {
		return "", err
	}
	wf, err := btcutil.NewWIF(priv, &chaincfg.MainNetParams, compress)
	if err != nil {
		return "", err
	}
	return wf.String(), nil
}
