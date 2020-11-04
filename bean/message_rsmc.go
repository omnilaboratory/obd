package bean

//消息 -100351(alice to obd) Alice发给obd (commitment_tx) 申请转账
type RequestToCreateCommitmentTx struct {
	ChannelId                 string  `json:"channel_id"` //the global channel id.
	Amount                    float64 `json:"amount"`     //amount of the payment
	LastTempAddressPrivateKey string  `json:"last_temp_address_private_key"`
	CurrTempAddressIndex      int     `json:"curr_temp_address_index"`
	CurrTempAddressPubKey     string  `json:"curr_temp_address_pub_key"`
	typeLengthValue
}

// 返回值 -100351的返回值 (obd to alice) obd推给Alice (需要Alice签名的数据)
type NeedAliceSignRsmcDataForC2a struct {
	ChannelId           string                  `json:"channel_id"` //the global channel id.
	RsmcRawData         NeedClientSignRawTxData `json:"rsmc_raw_data"`
	CounterpartyRawData NeedClientSignRawTxData `json:"counterparty_raw_data"`
}

//消息 -100360(alice to obd) alice发给obd (Alice签名C2a的结果)
type AliceSignedRsmcDataForC2a struct {
	ChannelId             string `json:"channel_id"`
	RsmcSignedHex         string `json:"rsmc_signed_hex"`
	CounterpartySignedHex string `json:"counterparty_signed_hex"`
	typeLengthValue
}

//返回值 -100360(alice to obd) alice发给obd (Alice签名C2a的结果)
type AliceSignedRsmcDataForC2aResult struct {
	ChannelId             string  `json:"channel_id"` //the global channel id.
	CommitmentTxHash      string  `json:"commitment_tx_hash"`
	Amount                float64 `json:"amount"` //amount of the payment
	CurrTempAddressPubKey string  `json:"curr_temp_address_pub_key"`
}

// 消息 p2p 351 p2p (obd of alice to obd of bob) 发送给bob的obd的数据 alice请求rsmc转账
type AliceRequestToCreateCommitmentTxOfP2p struct {
	AliceSignedRsmcDataForC2aResult
	LastTempAddressPrivateKey string                  `json:"last_temp_address_private_key"`
	RsmcRawData               NeedClientSignRawTxData `json:"rsmc_raw_data"`
	CounterpartyRawData       NeedClientSignRawTxData `json:"counterparty_raw_data"`
	PayerNodeAddress          string                  `json:"payer_node_address"`
	PayerPeerId               string                  `json:"payer_peer_id"`
	typeLengthValue
}

// 返回值 -110351  (obd to bob) obd推给bob  bob所在的obd发送给bob的需要bob签名的数据：C2a的rsmc和toB
type PayerRequestCommitmentTxToBobClient struct {
	AliceRequestToCreateCommitmentTxOfP2p
	MsgHash string `json:"msg_hash"`
}

// p2p消息 -100352 (bob to obd) (commitment_tx_signed)  bob完成对C2a rsmc和toB的签名，并对这次rsmc转账签收
type PayeeSendSignCommitmentTx struct {
	ChannelId                 string `json:"channel_id"`
	MsgHash                   string `json:"msg_hash"`
	C2aRsmcSignedHex          string `json:"c2a_rsmc_signed_hex"`
	C2aCounterpartySignedHex  string `json:"c2a_counterparty_signed_hex"`
	LastTempAddressPrivateKey string `json:"last_temp_address_private_key"` // bob2's private key
	CurrTempAddressIndex      int    `json:"curr_temp_address_index"`
	CurrTempAddressPubKey     string `json:"curr_temp_address_pub_key"` // bob3 or alice3
	Approval                  bool   `json:"approval"`                  // true agree false disagree
	typeLengthValue
}

// 返回值 -100352 (obd to bob) 推送给bob，需要bob对C2b的交易进行签名
type NeedBobSignRawDataForC2b struct {
	ChannelId              string                    `json:"channel_id"` //the global channel id.
	C2bRsmcRawData         NeedClientSignRawTxData   `json:"c2b_rsmc_raw_data"`
	C2bCounterpartyRawData NeedClientSignRawTxData   `json:"c2b_counterparty_raw_data"`
	C2aRdRawData           NeedClientSignRawTxData   `json:"c2a_rd_raw_data"`
	C2aBrRawData           NeedClientSignRawBRTxData `json:"c2a_br_raw_data"`
}

// 消息 100361 （bob to obd） bob对C2b的签名结果
type BobSignedRsmcDataForC2b struct {
	ChannelId                string `json:"channel_id"` //the global channel id.
	C2bRsmcSignedHex         string `json:"c2b_rsmc_signed_hex"`
	C2bCounterpartySignedHex string `json:"c2b_counterparty_signed_hex"`
	C2aRdSignedHex           string `json:"c2a_rd_signed_hex"`
	C2aBrSignedHex           string `json:"c2a_br_signed_hex"`
	C2aBrId                  int64  `json:"c2a_br_id"`
	typeLengthValue
}

// 消息 100361 （bob to obd） bob对C2b的签名结果
type BobSignedRsmcDataForC2bResult struct {
	ChannelId        string `json:"channel_id"` //the global channel id.
	CommitmentTxHash string `json:"commitment_tx_hash"`
	Approval         bool   `json:"approval"`
}

//p2p消息 -100361 -> 352 (obd of bob to obd of alice) bob的obd把c2b的相关信息通过p2p推送过alice的obd
type PayeeSignCommitmentTxOfP2p struct {
	BobSignedRsmcDataForC2bResult
	C2bRsmcTxData                NeedClientSignRawTxData `json:"c2b_rsmc_tx_data"`
	C2bCounterpartyTxData        NeedClientSignRawTxData `json:"c2b_counterparty_tx_data"`
	LastTempAddressPrivateKey    string                  `json:"last_temp_address_private_key"`
	CurrTempAddressPubKey        string                  `json:"curr_temp_address_pub_key"`
	C2aSignedRsmcHex             string                  `json:"c2a_signed_rsmc_hex"`
	C2aSignedToCounterpartyTxHex string                  `json:"c2a_signed_to_counterparty_tx_hex"`
	C2aRdTxData                  NeedClientSignRawTxData `json:"c2a_rd_tx_data"`
	typeLengthValue
}

// 返回值 110352（obd to Alice）352给Alice的返回值 obd推送给alice，为c2b的Rsmc和toBob，以及c2a的Rd签名
type NeedAliceSignRmscTxForC2b struct {
	ChannelId                  string                  `json:"channel_id"` //the global channel id.
	C2bRsmcPartialData         NeedClientSignRawTxData `json:"c2b_rsmc_partial_data"`
	C2bCounterpartyPartialData NeedClientSignRawTxData `json:"c2b_counterparty_partial_data"`
	C2aRdPartialData           NeedClientSignRawTxData `json:"c2a_rd_partial_data"`
}

// 消息 100362（alice to obd） Alice完成对C2b的相关信息签名
type AliceSignedRmscTxForC2b struct {
	ChannelId                string `json:"channel_id"` //the global channel id.
	C2bRsmcSignedHex         string `json:"c2b_rsmc_signed_hex"`
	C2bCounterpartySignedHex string `json:"c2b_counterparty_signed_hex"`
	C2aRdSignedHex           string `json:"c2a_rd_signed_hex"`
	typeLengthValue
}

// 返回值 100362的返回值（obd to Alice） obd推送给alice，为c2b的Rd和BR签名
type NeedAliceSignRdTxForC2b struct {
	ChannelId    string                    `json:"channel_id"` //the global channel id.
	C2bRdRawData NeedClientSignRawTxData   `json:"c2b_rd_raw_data"`
	C2bBrRawData NeedClientSignRawBRTxData `json:"c2b_br_raw_data"`
}

// 消息 100363（alice to obd） Alice完成对C2b的Rd和BR的相关信息签名
type AliceSignedRdTxForC2b struct {
	ChannelId      string `json:"channel_id"`
	C2bRdSignedHex string `json:"c2b_rd_signed_hex"`
	C2bBrSignedHex string `json:"c2b_br_signed_hex"`
	C2bBrId        int64  `json:"c2b_br_id"`
	typeLengthValue
}

// p2p 消息 353（obd of alice to obd of bob） Alice完成c2b的Rd和Br的签名，把签名结果发给bob所在的obd
type AliceSignedC2bTxDataP2p struct {
	ChannelId                string                  `json:"channel_id"`
	C2aCommitmentTxHash      string                  `json:"c2a_commitment_tx_hash"`
	C2bRsmcSignedHex         string                  `json:"c2b_rsmc_signed_hex"`
	C2bCounterpartySignedHex string                  `json:"c2b_counterparty_signed_hex"`
	C2bRdPartialData         NeedClientSignRawTxData `json:"c2b_rd_partial_data"`
}

// 返回值 110353（obd to bob） 353给bob的返回值 把需要签名的rd交易推给bob
type NeedBobSignRdTxForC2b struct {
	ChannelId        string                  `json:"channel_id"`
	C2bRdPartialData NeedClientSignRawTxData `json:"c2b_rd_partial_data"`
}

// 消息 100364（to obd） bob签名rd完成后的结果
type BobSignedRdTxForC2b struct {
	ChannelId      string `json:"channel_id"`
	C2bRdSignedHex string `json:"c2b_rd_signed_hex"`
	typeLengthValue
}
