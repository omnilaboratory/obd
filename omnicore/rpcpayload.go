package omnicore

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"unsafe"

	"github.com/btcsuite/btcd/wire"
)

/*
* parsing.cpp, static bool isBigEndian()
 */
func IsLittleEndian() bool {
	n := 0x1234
	return *(*byte)(unsafe.Pointer(&n)) == 0x34
}

/**
 * Swaps byte order of 16 bit wide numbers on little-endian systems.
 * parsing.cpp, void SwapByteOrder16(uint16_t& us)
 */
func SwapByteOrder16(us uint16) uint16 {
	if IsLittleEndian() {
		us = (us >> 8) |
			(us << 8)
	}
	return us
}

/**
 * Swaps byte order of 32 bit wide numbers on little-endian systems.
 */
func SwapByteOrder32(ui uint32) uint32 {
	if IsLittleEndian() {
		ui = (ui >> 24) |
			((ui << 8) & 0x00FF0000) |
			((ui >> 8) & 0x0000FF00) |
			(ui << 24)
	}
	return ui
}

/**
 * Swaps byte order of 64 bit wide numbers on little-endian systems.
 */
func SwapByteOrder64(ull uint64) uint64 {
	if IsLittleEndian() {
		ull = (ull >> 56) |
			((ull << 40) & 0x00FF000000000000) |
			((ull << 24) & 0x0000FF0000000000) |
			((ull << 8) & 0x000000FF00000000) |
			((ull >> 8) & 0x00000000FF000000) |
			((ull >> 24) & 0x0000000000FF0000) |
			((ull >> 40) & 0x000000000000FF00) |
			(ull << 56)
	}

	return ull
}

func Uint16ToBytes(n uint16) []byte {
	return []byte{
		byte(n),
		byte(n >> 8),
	}
}
func Uint32ToBytes(n uint32) []byte {
	return []byte{
		byte(n),
		byte(n >> 8),
		byte(n >> 16),
		byte(n >> 24),
	}
}
func Uint64ToBytes(n uint64) []byte {
	return []byte{
		byte(n),
		byte(n >> 8),
		byte(n >> 16),
		byte(n >> 24),
		byte(n >> 32),
		byte(n >> 40),
		byte(n >> 48),
		byte(n >> 56),
	}
}

// createPayload.cpp, std::vector<unsigned char> CreatePayload_SimpleSend(uint32_t propertyId, uint64_t amount)
// This is a straight forward translation in golang.
func OmniCreatePayloadSimpleSend(property_id uint32, amount uint64) []byte {

	var messageType uint16 = 0
	var messageVer uint16 = 0
	messageType = SwapByteOrder16(messageType)
	messageVer = SwapByteOrder16(messageVer)
	property_id = SwapByteOrder32(property_id)
	amount = SwapByteOrder64(amount)

	len := 4
	s := make([][]byte, len)

	s[0] = Uint16ToBytes(messageType)
	s[1] = Uint16ToBytes(messageVer)
	s[2] = Uint32ToBytes(property_id)
	s[3] = Uint64ToBytes(amount)

	sep := []byte("")
	return bytes.Join(s, sep)

}

/*
 * rpcvalue.cpp, uint32_t ParsePropertyId(const UniValue& value)
 */
func ParsePropertyId(property_id string) (uint32, error) {
	id, err := strconv.ParseInt(property_id, 10, 64)
	if err == nil {
		fmt.Printf("i64 %v \n", id)
	}

	if id < 1 || int64(4294967295) < id {
		fmt.Println("Property identifier is out of range")
		return 0, errors.New("Property identifier is out of range")
	}

	return uint32(id), nil

}

func HexStr(byte_array []byte) string {
	return hex.EncodeToString(byte_array)
}

/*
 * step 2) Construct payload
 * omnicore-cli "omni_createpayload_simplesend" 2 "0.1"
 *
 * rpcpayload.cpp, static UniValue omni_createpayload_simplesend(const JSONRPCRequest& request)
 *
 */
func Omni_createpayload_simplesend(property_id_str string, amount_str string, divisible bool) ([]byte, string) {
	property_id, _ := ParsePropertyId(property_id_str)
	amount := StrToInt64(amount_str, divisible)

	//because amount must be less than the biggest unsigned 64-bit integer, and bigger than 0
	//so that here we transform amount to be uint64
	//TO DO: add check here.
	amount_uint64 := uint64(amount)

	payload := OmniCreatePayloadSimpleSend(property_id, amount_uint64)

	return payload, HexStr(payload)

}

func TxToHex(tx *wire.MsgTx) string {
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	tx.Serialize(buf)
	return hex.EncodeToString(buf.Bytes())
}
