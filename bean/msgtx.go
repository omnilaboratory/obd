package bean

import "LightningOnOmni/bean/chainhash"

// OutPoint defines a bitcoin data type that is used to track previous transaction outputs.
type OutPoint struct {
	Hash  chainhash.Hash
	Index uint32
}
