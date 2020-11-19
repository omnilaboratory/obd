package omnicore

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

/*
 * Script opcodes, defined in omnicore script/script.h
 */
const (
	// push value
	OP_0 = 0x00
	//OP_FALSE = OP_0
	OP_FALSE     = 0x00
	OP_PUSHDATA1 = 0x4c
	OP_PUSHDATA2 = 0x4d
	OP_PUSHDATA4 = 0x4e
	OP_1NEGATE   = 0x4f
	OP_RESERVED  = 0x50
	OP_1         = 0x51
	OP_TRUE      = 0x51
	OP_2         = 0x52
	OP_3         = 0x53
	OP_4         = 0x54
	OP_5         = 0x55
	OP_6         = 0x56
	OP_7         = 0x57
	OP_8         = 0x58
	OP_9         = 0x59
	OP_10        = 0x5a
	OP_11        = 0x5b
	OP_12        = 0x5c
	OP_13        = 0x5d
	OP_14        = 0x5e
	OP_15        = 0x5f
	OP_16        = 0x60

	// control
	OP_NOP      = 0x61
	OP_VER      = 0x62
	OP_IF       = 0x63
	OP_NOTIF    = 0x64
	OP_VERIF    = 0x65
	OP_VERNOTIF = 0x66
	OP_ELSE     = 0x67
	OP_ENDIF    = 0x68
	OP_VERIFY   = 0x69
	OP_RETURN   = 0x6a

	// stack ops
	OP_TOALTSTACK   = 0x6b
	OP_FROMALTSTACK = 0x6c
	OP_2DROP        = 0x6d
	OP_2DUP         = 0x6e
	OP_3DUP         = 0x6f
	OP_2OVER        = 0x70
	OP_2ROT         = 0x71
	OP_2SWAP        = 0x72
	OP_IFDUP        = 0x73
	OP_DEPTH        = 0x74
	OP_DROP         = 0x75
	OP_DUP          = 0x76
	OP_NIP          = 0x77
	OP_OVER         = 0x78
	OP_PICK         = 0x79
	OP_ROLL         = 0x7a
	OP_ROT          = 0x7b
	OP_SWAP         = 0x7c
	OP_TUCK         = 0x7d

	// splice ops
	OP_CAT    = 0x7e
	OP_SUBSTR = 0x7f
	OP_LEFT   = 0x80
	OP_RIGHT  = 0x81
	OP_SIZE   = 0x82

	// bit logic
	OP_INVERT      = 0x83
	OP_AND         = 0x84
	OP_OR          = 0x85
	OP_XOR         = 0x86
	OP_EQUAL       = 0x87
	OP_EQUALVERIFY = 0x88
	OP_RESERVED1   = 0x89
	OP_RESERVED2   = 0x8a

	// numeric
	OP_1ADD      = 0x8b
	OP_1SUB      = 0x8c
	OP_2MUL      = 0x8d
	OP_2DIV      = 0x8e
	OP_NEGATE    = 0x8f
	OP_ABS       = 0x90
	OP_NOT       = 0x91
	OP_0NOTEQUAL = 0x92

	OP_ADD    = 0x93
	OP_SUB    = 0x94
	OP_MUL    = 0x95
	OP_DIV    = 0x96
	OP_MOD    = 0x97
	OP_LSHIFT = 0x98
	OP_RSHIFT = 0x99

	OP_BOOLAND            = 0x9a
	OP_BOOLOR             = 0x9b
	OP_NUMEQUAL           = 0x9c
	OP_NUMEQUALVERIFY     = 0x9d
	OP_NUMNOTEQUAL        = 0x9e
	OP_LESSTHAN           = 0x9f
	OP_GREATERTHAN        = 0xa0
	OP_LESSTHANOREQUAL    = 0xa1
	OP_GREATERTHANOREQUAL = 0xa2
	OP_MIN                = 0xa3
	OP_MAX                = 0xa4

	OP_WITHIN = 0xa5

	// crypto
	OP_RIPEMD160           = 0xa6
	OP_SHA1                = 0xa7
	OP_SHA256              = 0xa8
	OP_HASH160             = 0xa9
	OP_HASH256             = 0xaa
	OP_CODESEPARATOR       = 0xab
	OP_CHECKSIG            = 0xac
	OP_CHECKSIGVERIFY      = 0xad
	OP_CHECKMULTISIG       = 0xae
	OP_CHECKMULTISIGVERIFY = 0xaf

	// expansion
	OP_NOP1                = 0xb0
	OP_CHECKLOCKTIMEVERIFY = 0xb1
	OP_NOP2                = OP_CHECKLOCKTIMEVERIFY
	OP_CHECKSEQUENCEVERIFY = 0xb2
	OP_NOP3                = OP_CHECKSEQUENCEVERIFY
	OP_NOP4                = 0xb3
	OP_NOP5                = 0xb4
	OP_NOP6                = 0xb5
	OP_NOP7                = 0xb6
	OP_NOP8                = 0xb7
	OP_NOP9                = 0xb8
	OP_NOP10               = 0xb9

	// template matching params
	OP_SMALLINTEGER = 0xfa
	OP_PUBKEYS      = 0xfb
	OP_PUBKEYHASH   = 0xfd
	OP_PUBKEY       = 0xfe

	OP_INVALIDOPCODE = 0xff
)

/**
 * script/standard.h
 * Default setting for nMaxDatacarrierBytes. 80 bytes of data, +1 for OP_RETURN,
 * +2 for the pushdata opcodes.
 */
const MAX_OP_RETURN_RELAY uint = 83

/*
 * defined in standard.cpp
 */
const nMaxDatacarrierBytes uint = MAX_OP_RETURN_RELAY

/**
 * Embeds a payload in an OP_RETURN output, prefixed with a transaction marker.
 *
 * The request is rejected, if the size of the payload with marker is larger than
 * the allowed data carrier size ("-datacarriersize=n").
 *
 * encoding.cpp bool OmniCore_Encode_ClassC(const std::vector<unsigned char>& vchPayload,
 *	std::vector<std::pair <CScript, int64_t> >& vecOutputs)
 *
 * This is the golang version.
 */
func OmniCore_Encode_ClassC(payload_bytes []byte) []byte {

	vchOmBytes := GetOmMarker()
	s := make([][]byte, 2)

	s[0] = vchOmBytes
	s[1] = payload_bytes

	sep := []byte("")
	omni_payload := bytes.Join(s, sep)

	if (uint(len(omni_payload))) > nMaxDatacarrierBytes {
		fmt.Println("omni_payload is too long, only 83 bytes are allowed")
		return nil
	}
	//add op code data

	//CScript script;
	//script << OP_RETURN << vchData;

	//add op code, which is OP_RETURN in this case, and data .
	op_return_data := make([][]byte, 2)
	op_return_data[0] = []byte{OP_RETURN}
	op_return_data[1] = addOpCodeData(omni_payload)
	op_return := bytes.Join(op_return_data, sep)

	return op_return
}

// integer to byte array
/*
func UIntToBytes(integer uint, b int) ([]byte,error) {
	switch b {
	case 8:
		tmp := uint8(integer)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(),nil
	case 16:
		tmp := uint16(integer)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(),nil
	case 32:
		tmp := Uint32(integer)
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, &tmp)
		return bytesBuffer.Bytes(),nil
	}
	return nil,fmt.Errorf("UIntToBytes b param is invaild, must be 8, 16 or 32")
*/
/*
 * omnicore.cpp, const std::vector<unsigned char> GetOmMarker()
 * static unsigned char pch[] = {0x6f, 0x6d, 0x6e, 0x69}; // Hex-encoded: "omni"
 * return std::vector<unsigned char>(pch, pch + sizeof(pch) / sizeof(pch[0]));
 *
 * This is the golang version.
 */
func GetOmMarker() []byte {
	//var a = []byte("omni")
	//var a = [...]byte{0x6f, 0x6d, 0x6e, 0x69}
	return []byte{0x6f, 0x6d, 0x6e, 0x69}
}

/*
 * Embeds a payload with class C (op-return) encoding.
 */
func AddOpReturn(tx *wire.MsgTx, data []byte) *wire.MsgTx {

	outputs := OmniCore_Encode_ClassC(data)

	new_output := wire.NewTxOut(0, outputs)
	tx.AddTxOut(new_output)
	return tx
}

//omnicore-cli "omni_createrawtx_opreturn" "0100000002de95b97cf4c67ec01485fd698ec154a325ff69dd3e58435d7024bae7f69534c20000000000ffffffffb3b60aaa69b860c9bf31e742e3b37e75a2a553fd0bebf8aaf7da0e9bb07316ee0200000000ffffffff0000000000" "00000000000000020000000000989680"
func Omni_createrawtx_opreturn(base_tx *wire.MsgTx, payload_bytes []byte, payload_str string) (*wire.MsgTx, error) {

	// extend the transaction
	//tx = OmniTxBuilder(tx)
	//        .addOpReturn(payload)
	//        .build();

	tx := AddOpReturn(base_tx, payload_bytes)
	return tx, nil
}

/*
 * script.h overloaded operator <<. Here we translate it to golang func, but for OP_RETURN only
 * https://wiki.bitcoinsv.io/index.php/Pushdata_Opcodes
 */
func addOpCodeData(data []byte) []byte {
	var data_length []byte
	length := len(data)

	if length < OP_PUSHDATA1 {
		//insert(end(), (unsigned char)b.size());
		len_8 := uint8(length)
		data_length = []byte{len_8}

	} else if length <= 0xff {

		//insert(end(), OP_PUSHDATA1);
		//insert(end(), (unsigned char)b.size());
		len_8 := uint8(length)
		data_length = []byte{OP_PUSHDATA1, len_8}

	} else {
		fmt.Println("Only support Opcodes 1-75 (0x01 - 0x4B)  and OP_PUSHDATA1 (76 or 0x4C)")
		return nil
	}

	s := make([][]byte, 2)
	s[0] = data_length
	s[1] = data
	opcode_data := bytes.Join(s, []byte(""))
	return opcode_data
}
