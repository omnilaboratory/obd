package bean

// 原有 type -100049: user wanna close htlc tx when tx is on getH state
type HtlcRequestCloseCurrTx struct {
	ChannelId                            string `json:"channel_id"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"` //	开通通道用到的私钥
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressIndex             int    `json:"curr_rsmc_temp_address_index"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`
	typeLengthValue
}

// 消息 -100049: user wanna close htlc tx when tx is on getH state
type HtlcCloseRequestCurrTx struct {
	ChannelId                            string `json:"channel_id"`
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressIndex             int    `json:"curr_rsmc_temp_address_index"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	typeLengthValue
}

// 返回值 -100049的返回值 (obd to alice) obd推给Alice (需要Alice签名的数据)
type NeedAliceSignRsmcDataForC4a struct {
	ChannelId              string               `json:"channel_id"` //the global channel id.
	C4aRsmcRawData         NeedClientSignTxData `json:"c4a_rsmc_raw_data"`
	C4aCounterpartyRawData NeedClientSignTxData `json:"c4a_counterparty_raw_data"`
}

//消息 -100110(alice to obd) alice发给obd (Alice签名C4a的结果)
type AliceSignedRsmcDataForC4a struct {
	ChannelId                    string `json:"channel_id"`
	RsmcPartialSignedHex         string `json:"rsmc_signed_hex"`
	CounterpartyPartialSignedHex string `json:"counterparty_signed_hex"`
	typeLengthValue
}

// 消息 p2p 49 p2p (obd of alice to obd of bob) 发送给bob的obd的数据 alice请求rsmc转账
type AliceRequestCloseHtlcCurrTxOfP2p struct {
	ChannelId                            string               `json:"channel_id"` //the global channel id.
	CommitmentTxHash                     string               `json:"commitment_tx_hash"`
	CurrTempAddressPubKey                string               `json:"curr_temp_address_pub_key"`
	LastRsmcTempAddressPrivateKey        string               `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string               `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string               `json:"last_htlc_temp_address_for_htnx_private_key"`
	RsmcPartialSignedData                NeedClientSignTxData `json:"rsmc_partial_signed_data"`
	CounterpartyPartialSignedData        NeedClientSignTxData `json:"counterparty_partial_signed_data"`
	CloserNodeAddress                    string               `json:"closer_node_address"`
	CloserPeerId                         string               `json:"closer_peer_id"`
	typeLengthValue
}

//消息 110049
type AliceRequestCloseHtlcCurrTxOfP2pToBobClient struct {
	ChannelId                        string               `json:"channel_id"` //the global channel id.
	C4aRsmcPartialSignedData         NeedClientSignTxData `json:"c4a_rsmc_partial_signed_data"`
	C4aCounterpartyPartialSignedData NeedClientSignTxData `json:"c4a_counterparty_partial_signed_data"`
	MsgHash                          string               `json:"msg_hash"`
}

//type -50: receiver sign the close request
type HtlcSignCloseCurrTx struct {
	MsgHash                              string `json:"msg_hash"`
	ChannelAddressPrivateKey             string `json:"channel_address_private_key"` //	开通通道用到的私钥
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressIndex             int    `json:"curr_rsmc_temp_address_index"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	CurrRsmcTempAddressPrivateKey        string `json:"curr_rsmc_temp_address_private_key"`
	typeLengthValue
}

//消息 100050 bob对关闭htlc交易请求的签收
type HtlcBobSignCloseCurrTx struct {
	MsgHash                              string `json:"msg_hash"`
	C4aRsmcCompleteSignedHex             string `json:"c4a_rsmc_complete_signed_hex"`
	C4aCounterpartyCompleteSignedHex     string `json:"c4a_counterparty_complete_signed_hex"`
	LastRsmcTempAddressPrivateKey        string `json:"last_rsmc_temp_address_private_key"`
	LastHtlcTempAddressPrivateKey        string `json:"last_htlc_temp_address_private_key"`
	LastHtlcTempAddressForHtnxPrivateKey string `json:"last_htlc_temp_address_for_htnx_private_key"`
	CurrRsmcTempAddressIndex             int    `json:"curr_rsmc_temp_address_index"`
	CurrRsmcTempAddressPubKey            string `json:"curr_rsmc_temp_address_pub_key"`
	typeLengthValue
}

//响应 100050的结果
type NeedBobSignRawDataForC4b struct {
	ChannelId              string               `json:"channel_id"`
	C4bRsmcRawData         NeedClientSignTxData `json:"c4b_rsmc_raw_data"`
	C4bCounterpartyRawData NeedClientSignTxData `json:"c4b_counterparty_raw_data"`
	C4aRdRawData           NeedClientSignTxData `json:"c4a_rd_raw_data"`
	C4aBrRawData           NeedClientSignTxData `json:"c4a_br_raw_data"`
}

// 消息 1000111对100050的签名结果
type BobSignedRsmcDataForC4b struct {
	ChannelId                       string `json:"channel_id"`
	C4bRsmcPartialSignedHex         string `json:"c4b_rsmc_signed_hex"`
	C4bCounterpartyPartialSignedHex string `json:"c4b_counterparty_signed_hex"`
	C4aRdPartialSignedHex           string `json:"c4a_rd_signed_hex"`
	C4aBrPartialSignedHex           string `json:"c4a_br_signed_hex"`
	typeLengthValue
}

// 消息 100361 （bob to obd） bob对C2b的签名结果
type BobSignedRsmcDataForC4bResult struct {
	ChannelId        string `json:"channel_id"`
	CommitmentTxHash string `json:"commitment_tx_hash"`
}

// 原有 p2p消息 50号协议
type HtlcCloseCloseeSignedInfoToCloser struct {
	ChannelId                                  string `json:"channel_id"`
	CloseeLastRsmcTempAddressPrivateKey        string `json:"closee_last_rsmc_temp_address_private_key"`
	CloseeLastHtlcTempAddressPrivateKey        string `json:"closee_last_htlc_temp_address_private_key"`
	CloseeLastHtlcTempAddressForHtnxPrivateKey string `json:"closee_last_htlc_temp_address_for_htnx_private_key"`
	CloseeCurrRsmcTempAddressPubKey            string `json:"closee_curr_rsmc_temp_address_pub_key"`
	CloseeRsmcHex                              string `json:"closee_rsmc_hex"`
	CloseeToCounterpartyTxHex                  string `json:"closee_to_counterparty_tx_hex"`
	CloserCommitmentTxHash                     string `json:"closer_commitment_tx_hash"`
	CloserSignedRsmcHex                        string `json:"closer_signed_rsmc_hex"`
	CloserRsmcRdHex                            string `json:"closer_rsmc_rd_hex"`
	CloserSignedToCounterpartyTxHex            string `json:"closer_signed_to_counterparty_tx_hex"`
}

// p2p消息 50号协议传递的消息
type CloseeSignCloseHtlcTxOfP2p struct {
	ChannelId                                  string               `json:"channel_id"`
	CloseeLastRsmcTempAddressPrivateKey        string               `json:"closee_last_rsmc_temp_address_private_key"`
	CloseeLastHtlcTempAddressPrivateKey        string               `json:"closee_last_htlc_temp_address_private_key"`
	CloseeLastHtlcTempAddressForHtnxPrivateKey string               `json:"closee_last_htlc_temp_address_for_htnx_private_key"`
	CloseeCurrRsmcTempAddressPubKey            string               `json:"closee_curr_rsmc_temp_address_pub_key"`
	CloserCommitmentTxHash                     string               `json:"closer_commitment_tx_hash"`
	C4aRsmcCompleteSignedHex                   string               `json:"c4a_rsmc_complete_signed_hex"`
	C4aCounterpartyCompleteSignedHex           string               `json:"c4a_counterparty_complete_signed_hex"`
	C4aRdPartialSignedData                     NeedClientSignTxData `json:"c4a_rd_partial_signed_data"`
	C4bRsmcPartialSignedData                   NeedClientSignTxData `json:"c4b_rsmc_partial_signed_data"`
	C4bCounterpartyPartialSignedData           NeedClientSignTxData `json:"c4b_counterparty_partial_signed_data"`
	PayeeNodeAddress                           string               `json:"payee_node_address"`
	PayeePeerId                                string               `json:"payee_peer_id"`
}

// 110050的推送信息
type NeedAliceSignRsmcTxForC4b struct {
	ChannelId                        string               `json:"channel_id"` //the global channel id.
	C4aRdPartialSignedData           NeedClientSignTxData `json:"c4a_rd_partial_signed_data"`
	C4bRsmcPartialSignedData         NeedClientSignTxData `json:"c4b_rsmc_partial_signed_data"`
	C4bCounterpartyPartialSignedData NeedClientSignTxData `json:"c4b_counterparty_partial_signed_data"`
}

// 消息 100112（alice to obd） Alice完成对C2b的相关信息签名
type AliceSignedRsmcTxForC4b struct {
	ChannelId                        string `json:"channel_id"`
	C4aRdCompleteSignedHex           string `json:"c4a_rd_complete_signed_hex"`
	C4bRsmcCompleteSignedHex         string `json:"c4b_rsmc_complete_signed_hex"`
	C4bCounterpartyCompleteSignedHex string `json:"c4b_counterparty_complete_signed_hex"`
	typeLengthValue
}

// 返回值 100112的返回值（obd to Alice） obd推送给alice，为c4b的Rd和BR签名
type NeedAliceSignRdTxForC4b struct {
	ChannelId        string                    `json:"channel_id"` //the global channel id.
	C4bRdRawData     NeedClientSignTxData      `json:"c4b_rd_raw_data"`
	C4bBrRawData     NeedClientSignRawBRTxData `json:"c4b_br_raw_data"`
	PayeeNodeAddress string                    `json:"payee_node_address"`
	PayeePeerId      string                    `json:"payee_peer_id"`
}

// 消息 100113（alice to obd） Alice完成对C4b的Rd和BR的相关信息签名
type AliceSignedRdTxForC4b struct {
	ChannelId             string `json:"channel_id"`
	C2bRdPartialSignedHex string `json:"c_2_b_rd_partial_signed_hex"`
	C2bBrPartialSignedHex string `json:"c_2_b_br_partial_signed_hex"`
	C2bBrId               int64  `json:"c_2_b_br_id"`
	typeLengthValue
}

// p2p 消息 51（obd of alice to obd of bob） Alice完成c2b的Rd和Br的签名，把签名结果发给bob所在的obd
type AliceSignedC4bTxDataP2p struct {
	ChannelId                        string               `json:"channel_id"`
	C4aCommitmentTxHash              string               `json:"c4a_commitment_tx_hash"`
	C4bRsmcCompleteSignedHex         string               `json:"c4b_rsmc_complete_signed_hex"`
	C4bCounterpartyCompleteSignedHex string               `json:"c4b_counterparty_complete_signed_hex"`
	C4bRdPartialSignedData           NeedClientSignTxData `json:"c4b_rd_partial_signed_data"`
}

// 返回值 110353（obd to bob） 353给bob的返回值 把需要签名的rd交易推给bob
type NeedBobSignRdTxForC4b struct {
	ChannelId              string               `json:"channel_id"`
	C4bRdPartialSignedData NeedClientSignTxData `json:"c4b_rd_partial_signed_data"`
}

// 消息 100364（to obd） bob签名rd完成后的结果
type BobSignedRdTxForC4b struct {
	ChannelId                  string `json:"channel_id"`
	C4bRdRsmcCompleteSignedHex string `json:"c4b_rd_rsmc_complete_signed_hex"`
	typeLengthValue
}
