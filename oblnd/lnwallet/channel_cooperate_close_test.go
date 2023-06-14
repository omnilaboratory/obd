package lnwallet

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/lightningnetwork/lnd/channeldb"
	"github.com/lightningnetwork/lnd/lnwallet/chainfee"
	"testing"
)

func Test_OBAssetCoopClose(t *testing.T) {
	EnableTestLog()
	testObAsset_CoopClose(t, &coopCloseTestCase{
		chanType: channeldb.SingleFunderTweaklessBit,
	})
}
func testObAsset_CoopClose(t *testing.T, testCase *coopCloseTestCase) {

	// Create a test channel which will be used for the duration of this
	// unittest. The channel will be funded evenly with Alice having 5 BTC,
	// and Bob having 5 BTC.
	aliceChannel, bobChannel, cleanUp, err := CreateTestChannels(
		testCase.chanType,
	)
	if err != nil {
		t.Fatalf("unable to create test channels: %v", err)
	}
	defer cleanUp()

	aliceDeliveryScript := bobsPrivKey[:]
	bobDeliveryScript := testHdSeed[:]

	aliceFeeRate := chainfee.SatPerKWeight(
		aliceChannel.channelState.LocalCommitment.FeePerKw,
	)
	bobFeeRate := chainfee.SatPerKWeight(
		bobChannel.channelState.LocalCommitment.FeePerKw,
	)

	// We'll start with both Alice and Bob creating a new close proposal
	// with the same fee.
	aliceFee := aliceChannel.CalcFee(aliceFeeRate)
	aliceSig, _, _, _, err := aliceChannel.Ob_AssetCreateCloseProposal(
		aliceFee, aliceDeliveryScript, bobDeliveryScript, nil,
	)
	if err != nil {
		t.Fatalf("unable to create alice coop close proposal: %v", err)
	}

	bobFee := bobChannel.CalcFee(bobFeeRate)
	bobSig, _, _, _, err := bobChannel.Ob_AssetCreateCloseProposal(
		bobFee, bobDeliveryScript, aliceDeliveryScript, nil,
	)
	if err != nil {
		t.Fatalf("unable to create bob coop close proposal: %v", err)
	}

	// With the proposals created, both sides should be able to properly
	// process the other party's signature. This indicates that the
	// transaction is well formed, and the signatures verify.
	aliceCloseTx1, aliceCloseTx2, bobTxBalance, err := bobChannel.OB_AssetCompleteCooperativeClose(
		bobSig, aliceSig, bobDeliveryScript, aliceDeliveryScript,
		bobFee,
	)
	if err != nil {
		t.Fatalf("unable to complete alice cooperative close: %v", err)
	}
	bobCloseSha := aliceCloseTx1.TxHash()

	bobCloseTx1, bobCloseTx2, aliceTxBalance, err := aliceChannel.OB_AssetCompleteCooperativeClose(
		aliceSig, bobSig, aliceDeliveryScript, bobDeliveryScript,
		aliceFee,
	)
	if err != nil {
		t.Fatalf("unable to complete bob cooperative close: %v", err)
	}
	aliceCloseSha := bobCloseTx1.TxHash()

	if bobCloseSha != aliceCloseSha {
		t.Fatalf("alice and bob close transactions don't match: %v", err)
	}
	if aliceCloseTx2.TxHash() != bobCloseTx2.TxHash() {
		t.Fatalf("alice and bob close transactions don't match: %v", err)
	}

	t.Logf("%v %v", spew.Sdump(aliceCloseTx1), spew.Sdump(aliceCloseTx2))

	return
	// Finally, make sure the final balances are correct from both's
	// perspective.
	aliceBalance := aliceChannel.channelState.LocalCommitment.
		LocalBtcBalance.ToSatoshis()

	// The commit balance have had the initiator's (Alice) commitfee and
	// any anchors subtracted, so add that back to the final expected
	// balance. Alice also pays the coop close fee, so that must be
	// subtracted.
	commitFee := aliceChannel.channelState.LocalCommitment.CommitFee
	expBalanceAlice := aliceBalance + commitFee +
		testCase.anchorAmt - bobFee
	if aliceTxBalance != expBalanceAlice {
		t.Fatalf("expected balance %v got %v", expBalanceAlice,
			aliceTxBalance)
	}

	// Bob is not the initiator, so his final balance should simply be
	// equal to the latest commitment balance.
	expBalanceBob := bobChannel.channelState.LocalCommitment.
		LocalBtcBalance.ToSatoshis()
	if bobTxBalance != expBalanceBob {
		t.Fatalf("expected bob's balance to be %v got %v",
			expBalanceBob, bobTxBalance)
	}
}

