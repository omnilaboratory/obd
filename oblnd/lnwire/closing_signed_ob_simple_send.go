package lnwire

import (
	"bytes"
	"io"

	"github.com/btcsuite/btcd/btcutil"
)

// ClosingSigned is sent by both parties to a channel once the channel is clear
// of HTLCs, and is primarily concerned with negotiating fees for the close
// transaction. Each party provides a signature for a transaction with a fee
// that they believe is fair. The process terminates when both sides agree on
// the same fee, or when one side force closes the channel.
//
// NOTE: The responder is able to send a signature without any additional
// messages as all transactions are assembled observing BIP 69 which defines a
// canonical ordering for input/outputs. Therefore, both sides are able to
// arrive at an identical closure transaction as they know the order of the
// inputs/outputs.
type ObClosingSignedSimpleSend struct {
	// ChannelID serves to identify which channel is to be closed.
	ChannelID ChannelID

	// FeeSatoshis is the total fee in satoshis that the party to the
	// channel would like to propose for the close transaction.
	FeeSatoshis btcutil.Amount

	// Signature is for the proposed channel close transaction.
	//ob simple send  have multi sig
	Signatures []Sig

	// ExtraData is the set of data that was appended to this message to
	// fill out the full maximum transport message size. These fields can
	// be used to specify optional data such as custom TLV fields.
	ExtraData ExtraOpaqueData
	SigEnd    bool
}

// NewClosingSigned creates a new empty ClosingSigned message.
func NewObClosingSignedSimpleSend(cid ChannelID, fs btcutil.Amount,
	sig []Sig) *ObClosingSignedSimpleSend {

	return &ObClosingSignedSimpleSend{
		ChannelID:   cid,
		FeeSatoshis: fs,
		Signatures:  sig,
	}
}

// A compile time check to ensure ClosingSigned implements the lnwire.Message
// interface.
var _ Message = (*ObClosingSignedSimpleSend)(nil)

// Decode deserializes a serialized ClosingSigned message stored in the passed
// io.Reader observing the specified protocol version.
//
// This is part of the lnwire.Message interface.
func (c *ObClosingSignedSimpleSend) Decode(r io.Reader, pver uint32) error {
	return ReadElements(
		r, &c.ChannelID, &c.FeeSatoshis, &c.Signatures, &c.SigEnd, &c.ExtraData,
	)
}

// Encode serializes the target ClosingSigned into the passed io.Writer
// observing the protocol version specified.
//
// This is part of the lnwire.Message interface.
func (c *ObClosingSignedSimpleSend) Encode(w *bytes.Buffer, pver uint32) error {
	if err := WriteChannelID(w, c.ChannelID); err != nil {
		return err
	}

	if err := WriteSatoshi(w, c.FeeSatoshis); err != nil {
		return err
	}

	if err := WriteSigs(w, c.Signatures); err != nil {
		return err
	}
	if err := WriteBool(w, c.SigEnd); err != nil {
		return err
	}
	return WriteBytes(w, c.ExtraData)
}

// MsgType returns the integer uniquely identifying this message type on the
// wire.
//
// This is part of the lnwire.Message interface.
func (c *ObClosingSignedSimpleSend) MsgType() MessageType {
	return MsgObClosingSignedSimpleSend
}
