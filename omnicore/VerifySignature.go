package omnicore

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
)

func VerifySignature(pubkey_hex string, signature_hex string) bool {
	// Decode hex-encoded serialized public key.
	pubKeyBytes, _ := hex.DecodeString(pubkey_hex)
	pubKey, _ := btcec.ParsePubKey(pubKeyBytes, btcec.S256())

	sigBytes, _ := hex.DecodeString(signature_hex)
	signature, _ := btcec.ParseSignature(sigBytes, btcec.S256())

	result := signature.Verify(signature.Serialize(), pubKey)
	return result
}
