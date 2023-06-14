package op

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleEnode(t *testing.T) {

	// This testing hex equals to "OP_RETURN 6f6d6e6900000000000000020000000000989680",
	// which is encoded from payload of property id 2 and amount 0.1
	opreturn_hex, _ := hex.DecodeString("6a146f6d6e6900000000000000020000000000989680")
	op := &OpreturnSimple{PropertyId: 2, Amount: 10000000}
	bs, err := op.Encode()
	if err != nil {
		t.Fatalf("Encode err %v", err)
	}
	if !bytes.Equal(bs, opreturn_hex) {
		t.Fatalf("mismatch encode %s %v %v", err, bs, opreturn_hex)
	}
	op1 := &OpreturnSimple{}
	op1.Decode(opreturn_hex)
	t.Log(op1)
}
func TestEnode(t *testing.T) {
	op := &Opreturn{Outputs: []*Output{
		&Output{Index: 2, Amount: 10.5 * 100000000},
		&Output{Index: 3, Amount: 0.5 * 100000000},
		&Output{Index: 5, Amount: 15 * 100000000}},
		PropertyId: 31,
	}
	w := bytes.NewBuffer([]byte{})
	err := op.Encode(w)
	require.Nil(t, err)
	t.Logf("%x", w.Bytes())

	pks,err:=getOpPKs(w.Bytes())
	t.Log(hex.EncodeToString( pks),err)

	//decode opreturn_payload  and json print
	op.PrintWithDecode(w.Bytes())


	require.Equal(t, "000000070000001f0302000000003e95ba80030000000002faf080050000000059682f00", fmt.Sprintf("%x", w.Bytes()))

	//opr := new(Opreturn)
	//err = opr.Decode(w.Bytes())
	//require.Nil(t, err)
	//jsonOut, _ := json.MarshalIndent(opr, "", "  ")
	//t.Log(string(jsonOut))
}


