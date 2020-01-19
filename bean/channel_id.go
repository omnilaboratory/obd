package bean

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"golang.org/x/crypto/salsa20"
	"log"
	"sync"
)

type ChannelID [32]byte

func init() {
	if _, err := rand.Read(ChannelIdService.chanIDSeed[:]); err != nil {
		log.Println(err)
	}
}

type channelIdManager struct {
	nonceMtx    sync.RWMutex
	chanIDNonce uint64
	chanIDSeed  [32]byte
}

var ChannelIdService = channelIdManager{}

// NewChanIDFromOutPoint converts a target OutPoint into a ChannelID that is usable within the network. In order to convert the OutPoint into a ChannelID,
// we XOR the lower 2-bytes of the txid within the OutPoint with the big-endian
// serialization of the Index of the OutPoint, truncated to 2-bytes.
func (service *channelIdManager) NewChanIDFromOutPoint(op *OutPoint) string {
	// First we'll copy the txid of the outpoint into our channel ID slice.
	var cid ChannelID
	copy(cid[:], op.Hash[:])

	// With the txid copied over, we'll now XOR the lower 2-bytes of the partial channelID with big-endian serialization of output index.
	xorTxid(&cid, uint16(op.Index))
	temp := cid[:]
	return hex.EncodeToString(temp)
}

// xorTxid performs the transformation needed to transform an OutPoint into a ChannelID.
// To do this, we expect the cid parameter to contain the txid unaltered and the outputIndex to be the output index
func xorTxid(cid *ChannelID, outputIndex uint16) {
	var buf [32]byte
	binary.BigEndian.PutUint16(buf[30:], outputIndex)
	cid[30] = cid[30] ^ buf[30]
	cid[31] = cid[31] ^ buf[31]
}

// NextTemporaryChanID returns the next free pending channel ID to be used to identify a particular future channel funding workflow.
func (service *channelIdManager) NextTemporaryChanID() string {
	// Obtain a fresh nonce. We do this by encoding the current nonce counter, then incrementing it by one.
	service.nonceMtx.Lock()
	var nonce [8]byte
	binary.LittleEndian.PutUint64(nonce[:], service.chanIDNonce)
	service.chanIDNonce++
	service.nonceMtx.Unlock()

	// We'll generate the next temporary channelID by "encrypting" 32-bytes of zeroes which'll extract 32 random bytes from our stream cipher.
	var (
		nextChanID [32]byte
		zeroes     [32]byte
	)
	salsa20.XORKeyStream(nextChanID[:], zeroes[:], nonce[:], &service.chanIDSeed)
	temp := nextChanID[:]
	return hex.EncodeToString(temp)
}

func (service *channelIdManager) IsEmpty(channelId ChannelID) bool {
	for _, item := range channelId {
		if item != 0 {
			return false
		}
	}
	return true
}
