package omnicore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

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
 * omnicore-cli "omni_createrawtx_reference" "0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff010000000000000000166a146f6d6e690000000000000002000000000098968000000000"
 * 											"1Njbpr7EkLA1R8ag8bjRN7oks7nv5wUn3o"
 */
func Omni_createrawtx_reference(base_tx *wire.MsgTx, receiver_address string, defaultNet *chaincfg.Params) (*wire.MsgTx, error) {

	destinationAddress, err := btcutil.DecodeAddress(receiver_address, defaultNet)
	if err != nil {
		return nil, errors.New("Fail decoding receiver address")
	}

	destinationPkScript, _ := txscript.PayToAddrScript(destinationAddress)

	// TO DO: add logic to calc dust
	val := OmniGetDustThreshold(destinationPkScript)
	// if val < dust {
	//	val = dust
	//}
	redeemTxOut := wire.NewTxOut(val, destinationPkScript)

	base_tx.AddTxOut(redeemTxOut)
	return base_tx, nil
}

/*
 * defined in amount.h
 */
/** No amount larger than this (in satoshi) is valid.
 *
 * Note that this constant is *not* the total money supply, which in Bitcoin
 * currently happens to be less than 21,000,000 BTC for various reasons, but
 * rather a sanity check. As this sanity check is used by consensus-critical
 * validation code, the exact value of the MAX_MONEY constant is consensus
 * critical; in unusual circumstances like a(nother) overflow bug that allowed
 * for the creation of coins out of thin air modification could lead to a fork.
 * */
const COIN = 100000000
const MAX_MONEY = 21000000 * COIN

func MoneyRange(money int64) bool {
	if money >= 0 && money <= MAX_MONEY {
		return true
	} else {
		return false
	}
}

func AmountFromValue(value string) int64 {
	// amount in satoshi = bitcoin_amount * COIN
	amount := StrToInt64(value, true)
	if MoneyRange(amount) {
		return amount
	} else {
		fmt.Println("AmountFromValue(value string): value out of range")
		return 0
	}
}

/*
 * script.cpp, int64_t OmniGetDustThreshold(const CScript& scriptPubKey)
 *
 * Determines the minimum output amount to be spent by an output, based on the
 * scriptPubKey size in relation to the minimum relay fee.
 *
 * @param scriptPubKey[in]  The scriptPubKey
 * @return The dust threshold value
 *
 */
func OmniGetDustThreshold(scriptPubKey []byte) int64 {
	//TO DO, add logic to calculate dust threshhold.
	if len(scriptPubKey) > 0 {
		return 546
	} else {
		return COIN
	}

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
 * This example specifies a transaction fee of 0.0006 BTC.
 * The golang implementation does not accept number expressed in exponential from.
 */
func Omni_createrawtx_change(base_tx *wire.MsgTx, prev_txs_json_list string, destination_address string, miner_fee_in_btc string, defaultNet *chaincfg.Params) (*wire.MsgTx, error) {

	destination, err := btcutil.DecodeAddress(destination_address, defaultNet)
	destinationPkScript, _ := txscript.PayToAddrScript(destination)
	if err != nil {
		return nil, errors.New("Omni_createrawtx_change(...): Fail decoding destination address, which accepts the change.")
	}

	txFee_in_satoshi := AmountFromValue(miner_fee_in_btc)

	var total_val_in int64 = 0

	decoder := json.NewDecoder(strings.NewReader(prev_txs_json_list))
	for {
		var prev_tx PrevTx
		if err := decoder.Decode(&prev_tx); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		//prev_txid := prev_tx.TxID
		//prev_vout := prev_tx.VOut
		//prev_scriptPubKey := prev_tx.ScriptPubKey
		//prev_value := prev_tx.Value

		total_val_in = total_val_in + AmountFromValue(prev_tx.Value)

	}

	/*
	 * createTx.cpp, TxBuilder& TxBuilder::addChange(...) adds an output for change
	 *
	 * int64_t txChange = view.GetValueIn(tx) - tx.GetValueOut() - txFee;
	 * int64_t minValue = OmniGetDustThreshold(scriptPubKey);
	 *
	 */

	// because in Omni_createrawtx_reference(...) there is a dust output.
	txChange := total_val_in - OmniGetDustThreshold(destinationPkScript) - txFee_in_satoshi
	minValue := OmniGetDustThreshold(destinationPkScript)

	if txChange < minValue {
		return base_tx, nil
	}
	redeemTxOut := wire.NewTxOut(txChange, destinationPkScript)

	/*
	*  std::vector<CTxOut>::iterator it = transaction.vout.end();
	*  if (position < transaction.vout.size()) {
	*  it = transaction.vout.begin() + position;
	*  }
	 */

	TxOutArray := base_tx.TxOut
	redeemTxOutArray := []*wire.TxOut{redeemTxOut}
	TxOutArray = append(redeemTxOutArray, TxOutArray...)

	base_tx.TxOut = TxOutArray
	//base_tx.AddTxOut(redeemTxOut)

	return base_tx, nil
}
