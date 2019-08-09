package rpc

import (
	"fmt"
	"testing"
)

func TestClient_GetBlockCount(t *testing.T) {
	client := NewClient()
	//hex := "02000000000101d5df58bb45caac351120a9d228e580a7ffa9f928d78925a3ccbd746f256cb0fe0000000023220020b92c9c68bdc00c5f838e64fc436b994546c6e27911881a3384d0c54f4acc91f10a0000000160decb1d0000000017a914c893cec2a62aca000b8cf963a9530f24a4a4e23987040047304402207183958394599d0023995852d89d8cfd57687feb3638f92082d7fcc911e78ff4022003f50ae3514ca0387d9dae533398526a8c92e847e7f241f98af762eaad234c0901473044022079c25009db65d6e82bd1fb5271164ab41701359fe44e42d9ac93e40bc3026b460220143c20659211159cfec5e11ac95c49a3488baac43d57de1e6c735e7708e4ddb901475221035792effbf392a1f59ebc1a6df0eec31ca83f434706d07b5797f8cb7df8c27def210345f04d044f19006bd60620b86354296ad8a526027c00117b0f4324b0d59db1a152ae00000000"
	//result, _ := client.Omni_getbalance("n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn",31)
	//result, err := client.Omni_gettransaction("1f35a5a0d166fef6b2e24b321efa1108a01e6ea1205397bc96613fb73b902d18")
	//result, err := client.Omni_listtransactions()
	var keys []string
	result, err := client.BtcCreateAndSignRawTransaction("n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn", keys, "n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA", 0.0001, 0.00001, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
}

func TestClient_DecodeRawTransaction(t *testing.T) {
	client := NewClient()
	//hex := "0100000001be898a63ce2f0bf3297189fb9103aa63f0eafd74a6cb1974f2bbd59b1ed13140520000006a47304402206b5eb7fdcc6df35ee0cd5186b6d9e6f29bcaa8a85af1f9a4726c976ee746f00b0220764d3d9f7c86d61afd773ac31f63256fe8b766819253f55f837451b7c93cad3d01210343c3b527ebf385b15cedfa7e9b840b32353482bfc29ddc931a155ff41db0123effffffff02306f0100000000001976a914d8b7fc6790003813df46aacd8bdc84d1f672147b88ac10270000000000001976a914fd1dc7fb3727d29b43200480580049c3bf8a041b88ac00000000"
	hex := "0100000001946f7d2d3b7f61e87c20a96fd01f9c6ef50fc534588c9d5dd73d0c7241deca36010000006a47304402204107627dcece427aa8c565e76c9d18af455833fd66cbbc8fbbf96bd68f828f10022043c84bcbf09d543b18b66745611280d2bee1ed398e4aab2d5f0fcffd6d99e56b01210343c3b527ebf385b15cedfa7e9b840b32353482bfc29ddc931a155ff41db0123effffffff0210270000000000001976a914fd1dc7fb3727d29b43200480580049c3bf8a041b88ac38440100000000001976a914d8b7fc6790003813df46aacd8bdc84d1f672147b88ac00000000"
	result, err := client.SendRawTransaction(hex)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
}

func TestClient_Validateaddress(t *testing.T) {
	client := NewClient()
	//r, err := client.Omni_getbalance(" n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA", 121)
	r, err := client.Validateaddress(" n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r)
}
func TestClient_OmniRawTransaction(t *testing.T) {
	client := NewClient()
	txid, err := client.OmniRawTransaction("n1Grf4JGHUC2CdHHoDRYb7jbVKU2Fv8Tsn", nil, "n4bJvpVHks3Fz9wWB9f445LGV5xTS6LGpA", 121, 7, 0.00001, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(txid)
}

func TestClient_GetTransactionById(t *testing.T) {
	client := NewClient()
	result, err := client.GetTransactionById("434b1d74135ec0bf01c0d086792afdcee8c9440ad0aa10dc1882a901ca2b71e4")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
}
