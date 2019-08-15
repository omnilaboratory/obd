package bean

import "encoding/binary"

type ChannelID [32]byte

// NewChanIDFromOutPoint converts a target OutPoint into a ChannelID that is usable within the network. In order to convert the OutPoint into a ChannelID,
// we XOR the lower 2-bytes of the txid within the OutPoint with the big-endian
// serialization of the Index of the OutPoint, truncated to 2-bytes.
func NewChanIDFromOutPoint(op *OutPoint) ChannelID {
	// First we'll copy the txid of the outpoint into our channel ID slice.
	var cid ChannelID
	copy(cid[:], op.Hash[:])

	// With the txid copied over, we'll now XOR the lower 2-bytes of the partial channelID with big-endian serialization of output index.
	xorTxid(&cid, uint16(op.Index))

	return cid
}

// xorTxid performs the transformation needed to transform an OutPoint into a ChannelID.
// To do this, we expect the cid parameter to contain the txid unaltered and the outputIndex to be the output index
func xorTxid(cid *ChannelID, outputIndex uint16) {
	var buf [32]byte
	binary.BigEndian.PutUint16(buf[30:], outputIndex)
	cid[30] = cid[30] ^ buf[30]
	cid[31] = cid[31] ^ buf[31]
}
