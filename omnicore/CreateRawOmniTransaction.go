package omnicore

import (
	"encoding/json"
	"io"
	"log"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type Unspent struct {
	TxID     string `json:"txid"`
	VOut     uint32 `json:"vout"`
	Sequence uint32 `json:"sequence"`
}

type PrevTx struct {
	TxID         string `json:"txid"`
	VOut         uint32 `json:"vout"`
	ScriptPubKey string `json:"scriptPubKey"`
	Value        string `json:"value"`
}

func CheckUnspent(unspent_json_list string) (string, error) {

	return "", nil
}

/*
 * https://github.com/OmniLayer/omnicore/wiki/Use-the-raw-transaction-API-to-create-a-Simple-Send-transaction
 */

/*
 * step 3) Construct transaction base
 * omnicore-cli "createrawtransaction" '[{"txid":"c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de","vout":0},{"txid":"ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3","vout":2}]' '{}'
 *
 */
func CreateRawTransaction(unspent_json_list string, btc_version int32) (*wire.MsgTx, string, error) {
	// txFee := 10000
	// weird, there is no definition of TxVersion = 2, but version 2 works.
	//tx := wire.NewMsgTx(wire.TxVersion + 1)
	tx := wire.NewMsgTx(btc_version)

	decoder := json.NewDecoder(strings.NewReader(unspent_json_list))
	for {
		var unspentTx Unspent
		if err := decoder.Decode(&unspentTx); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		hash, err := chainhash.NewHashFromStr(unspentTx.TxID)
		if err != nil {
			log.Fatalf("could not get hash from transaction ID: %v", err)
		}
		outPoint := wire.NewOutPoint(hash, unspentTx.VOut)
		txIn := wire.NewTxIn(outPoint, nil, nil)

		txIn.Sequence = unspentTx.Sequence
		tx.AddTxIn(txIn)

		//fmt.Printf("%s \n ", txToHex(tx))

		//fmt.Printf("%s: %d \n ", unspentTx.TxID, unspentTx.VOut)

	}

	return tx, TxToHex(tx), nil
}
