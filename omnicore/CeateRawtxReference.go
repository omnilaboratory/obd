package omnicore

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

/*
 * https://github.com/OmniLayer/omnicore/wiki/Use-the-raw-transaction-API-to-create-a-Simple-Send-transaction
 */

/*
 * step 5) Attach reference/receiver output
 */
func Omni_createrawtx_reference(base_tx *wire.MsgTx, receiver_address string, defaultNet *chaincfg.Params) (*wire.MsgTx, error) {

	destinationAddress, err := btcutil.DecodeAddress(receiver_address, defaultNet)
	if err != nil {
		return nil, errors.New("Fail decoding receiver address")
	}

	// TO DO: add logic to calc dust
	// val := value
	// if val < dust {
	//	val = dust
	//}
	destinationPkScript, _ := txscript.PayToAddrScript(destinationAddress)
	redeemTxOut := wire.NewTxOut(0, destinationPkScript)

	base_tx.AddTxOut(redeemTxOut)
	return base_tx, nil
}

/*
 * step 6) Specify miner fee and attach change output (as needed)
 *
 * omnicore-cli "omni_createrawtx_change" "0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff020000000000000000166a146f6d6e690000000000000002000000000098968022020000000000001976a914ee692ea81da1b12d3dd8f53fd504865c9d843f5288ac00000000"
 *	'[{"txid":"c23495f6e7ba24705d43583edd69ff25a354c18e69fd8514c07ec6f47cb995de","vout":0,"scriptPubKey":"76a914c6734676a08e3c6438bd95fa62c57939c988a17b88ac","value":0.001},
 *    {"txid":"ee1673b09b0edaf7aaf8eb0bfd53a5a2757eb3e342e731bfc960b869aa0ab6b3","vout":2,"scriptPubKey":"76a914c6734676a08e3c6438bd95fa62c57939c988a17b88ac","value":0.0083566}]
 * 	'
 *   "1K6JtSvrHtyFmxdtGZyZEF7ydytTGqasNc"
 *   0.0006
 *
 * This example specifies a transaction fee of 0.0006 BTC
 */
func Omni_createrawtx_change(base_tx *wire.MsgTx, prev_txs_json_list string, destination_address string, miner_fee_in_btc int64, defaultNet *chaincfg.Params) (*wire.MsgTx, error) {
	/*
		destination, err := btcutil.DecodeAddress(destination_address, defaultNet)
		destinationPkScript, _ := txscript.PayToAddrScript(destination)
		if err != nil {
			return nil, errors.New("Fail decoding destination address, which accepts the change.")
		}

		txFee := miner_fee_in_btc

		total_val_in := 0
		decoder := json.NewDecoder(strings.NewReader(prev_txs_json_list))
		for {
			var prev_txs PrevTx
			if err := decoder.Decode(&prev_txs); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			prev_txid := prev_txs.TxID
			prev_vout := prev_txs.VOut
			prev_scriptPubKey := prev_txs.ScriptPubKey
			prev_value := prev_txs.Value

			redeemTxOut := wire.NewTxOut(0, destinationPkScript)

			base_tx.AddTxOut(redeemTxOut)

		}
	*/
	return base_tx, nil
}
