package enum

type SendTargetType int

const (
	SendTargetType_SendToNone     SendTargetType = -1
	SendTargetType_SendToAll      SendTargetType = 0
	SendTargetType_SendToSomeone  SendTargetType = 1
	SendTargetType_SendToExceptMe SendTargetType = 2
)

type MsgType int

const (
	MsgType_Tracker_Connect_301           MsgType = 301
	MsgType_Tracker_HeartBeat_302         MsgType = 302
	MsgType_Tracker_NodeLogin_303         MsgType = 303
	MsgType_Tracker_UserLogin_304         MsgType = 304
	MsgType_Tracker_UserLogout_305        MsgType = 305
	MsgType_Tracker_UpdateChannelInfo_350 MsgType = 350
	MsgType_Tracker_GetHtlcPath_351       MsgType = 351
	MsgType_Tracker_UpdateHtlcTxState_352 MsgType = 352
	MsgType_Tracker_UpdateUserInfo_353    MsgType = 353
)

// Transaction related messages, login is required: [-100000,-102000]
// Common messages, login is not required: [-102000,-103000]
// User center messages, login is required: [-103000,-104000]
const (
	MsgType_Error_0 MsgType = 0

	// region
	// Common messages, login is not required [-102000,-103000]
	MsgType_UserLogin_2001            MsgType = -102001
	MsgType_UserLogout_2002           MsgType = -102002
	MsgType_p2p_ConnectPeer_2003      MsgType = -102003
	MsgType_GetMnemonic_2004          MsgType = -102004
	MsgType_GetObdNodeInfo_2005       MsgType = -102005
	MsgType_GetMiniBtcFundAmount_2006 MsgType = -102006
	MsgType_HeartBeat_2007            MsgType = -102007
	MsgType_User_End_2099             MsgType = -102099

	MsgType_Core_GetNewAddress_2101                    MsgType = -102101
	MsgType_Core_GetMiningInfo_2102                    MsgType = -102102
	MsgType_Core_GetNetworkInfo_2103                   MsgType = -102103
	MsgType_Core_SignMessageWithPrivKey_2104           MsgType = -102104
	MsgType_Core_VerifyMessage_2105                    MsgType = -102105
	MsgType_Core_DumpPrivKey_2106                      MsgType = -102106
	MsgType_Core_ListUnspent_2107                      MsgType = -102107
	MsgType_Core_BalanceByAddress_2108                 MsgType = -102108
	MsgType_Core_FundingBTC_2109                       MsgType = -102109
	MsgType_Core_BtcCreateMultiSig_2110                MsgType = -102110
	MsgType_Core_Btc_ImportPrivKey_2111                MsgType = -102111
	MsgType_Core_Omni_GetBalance_2112                  MsgType = -102112
	MsgType_Core_Omni_CreateNewTokenFixed_2113         MsgType = -102113
	MsgType_Core_Omni_CreateNewTokenManaged_2114       MsgType = -102114
	MsgType_Core_Omni_GrantNewUnitsOfManagedToken_2115 MsgType = -102115
	MsgType_Core_Omni_RevokeUnitsOfManagedToken_2116   MsgType = -102116
	MsgType_Core_Omni_ListProperties_2117              MsgType = -102117
	MsgType_Core_Omni_GetTransaction_2118              MsgType = -102118
	MsgType_Core_Omni_GetProperty_2119                 MsgType = -102119
	MsgType_Core_Omni_FundingAsset_2120                MsgType = -102120
	MsgType_Core_Omni_Send_2121                        MsgType = -102121
	MsgType_Core_GetTransactionByTxid_2122             MsgType = -102122
	MsgType_Core_SignRawTransaction_2123               MsgType = -102123
	MsgType_Core_Omni_End_2199                         MsgType = -102199
	MsgType_Common_End_2999                            MsgType = -102999
	// endregion

	//region
	//User center messages, login is required: [-103000,-104000]

	//Create new address via mnemonic words
	MsgType_Mnemonic_CreateAddress_3000 MsgType = -103000
	//get address info via mnemonic words and index
	MsgType_Mnemonic_GetAddressByIndex_3001 MsgType = -103001

	//Omni asset funding list
	MsgType_FundingCreate_Asset_AllItem_3100 MsgType = -103100
	//Get omni asset funding info via ID
	MsgType_FundingCreate_Asset_ItemById_3101 MsgType = -103101
	//Get omni asset funding info via channel ID
	MsgType_FundingCreate_Asset_ItemByChannelId_3102 MsgType = -103102
	//Rounds of funding
	MsgType_FundingCreate_Asset_Count_3103 MsgType = -103103
	//BTC funding list
	MsgType_FundingCreate_Btc_AllItem_3104                      MsgType = -103104
	MsgType_FundingCreate_Btc_ItemById_3105                     MsgType = -103105
	MsgType_FundingCreate_Btc_ItemByTempChannelId_3106          MsgType = -103106
	MsgType_FundingCreate_Btc_RDAllItem_3107                    MsgType = -103107
	MsgType_FundingCreate_Btc_ItemRDById_3108                   MsgType = -103108
	MsgType_FundingCreate_Btc_ItemRDByTempChannelId_3109        MsgType = -103109
	MsgType_FundingCreate_Btc_ItemRDByTempChannelIdAndTxId_3110 MsgType = -103110
	MsgType_FundingCreate_Btc_ItemByChannelId_3111              MsgType = -103111

	MsgType_ChannelOpen_AllItem_3150         MsgType = -103150
	MsgType_ChannelOpen_ItemByTempId_3151    MsgType = -103151
	MsgType_ChannelOpen_Count_3152           MsgType = -103152
	MsgType_ChannelOpen_DelItemByTempId_3153 MsgType = -103153
	MsgType_GetChannelInfoByChannelId_3154   MsgType = -103154
	MsgType_GetChannelInfoByDbId_3155        MsgType = -103155
	MsgType_CheckChannelAddessExist_3156     MsgType = -103156

	MsgType_CommitmentTx_ItemsByChanId_3200              MsgType = -103200
	MsgType_CommitmentTx_ItemById_3201                   MsgType = -103201
	MsgType_CommitmentTx_Count_3202                      MsgType = -103202
	MsgType_CommitmentTx_LatestCommitmentTxByChanId_3203 MsgType = -103203
	MsgType_CommitmentTx_LatestRDByChanId_3204           MsgType = -103204
	MsgType_CommitmentTx_LatestBRByChanId_3205           MsgType = -103205
	MsgType_CommitmentTx_SendSomeCommitmentById_3206     MsgType = -103206
	MsgType_CommitmentTx_AllRDByChanId_3207              MsgType = -103207
	MsgType_CommitmentTx_AllBRByChanId_3208              MsgType = -103208

	MsgType_Htlc_GetLatestHT1aOrHE1b_3250             MsgType = -103250
	MsgType_Htlc_GetHT1aOrHE1bBySomeCommitmentId_3251 MsgType = -103251
	//endregion

	// region
	// Transaction related messages, login is required: [-100000,-102000]

	MsgType_SendChannelOpen_32 MsgType = -100032
	MsgType_ChannelOpen_32     MsgType = -32
	MsgType_RecvChannelOpen_32 MsgType = -110032

	MsgType_SendChannelAccept_33 MsgType = -100033
	MsgType_ChannelAccept_33     MsgType = -33
	MsgType_RecvChannelAccept_33 MsgType = -110033

	MsgType_FundingCreate_SendBtcFundingCreated_340       MsgType = -100340
	MsgType_FundingCreate_BtcFundingMinerRDTxToClient_341 MsgType = -100341
	MsgType_FundingCreate_BtcFundingCreated_340           MsgType = -340
	MsgType_FundingCreate_RecvBtcFundingCreated_340       MsgType = -110340

	MsgType_FundingSign_SendBtcSign_350 MsgType = -100350
	MsgType_FundingSign_BtcSign_350     MsgType = -350
	MsgType_FundingSign_RecvBtcSign_350 MsgType = -110350

	MsgType_FundingCreate_SendAssetFundingCreated_34 MsgType = -100034
	// 客户端签名：alice签名C1a
	MsgType_ClientSign_Duplex_AssetFunding_ChannelAddressSignC1a_1034 MsgType = -101034
	//MsgType_ClientSign_EchoAlice_AssetFunding_RdAndBr_1134 MsgType = -101134
	//MsgType_ClientSign_Duplex_AssetFunding_RdAndBr_1234 MsgType = -101234
	MsgType_FundingCreate_AssetFundingCreated_34     MsgType = -34
	MsgType_FundingCreate_RecvAssetFundingCreated_34 MsgType = -110034

	MsgType_FundingSign_SendAssetFundingSigned_35        MsgType = -100035
	MsgType_ClientSign_EchoBob_AssetFunding_RdAndBr_1035 MsgType = -101035
	MsgType_ClientSign_Duplex_AssetFunding_RdAndBr_1135  MsgType = -101135
	MsgType_FundingSign_AssetFundingSigned_35            MsgType = -35
	MsgType_FundingSign_RecvAssetFundingSigned_35        MsgType = -110035

	MsgType_CommitmentTx_SendCommitmentTransactionCreated_351                    MsgType = -100351
	MsgType_CommitmentTx_CommitmentTransactionCreated_351                        MsgType = -351
	MsgType_CommitmentTx_RecvCommitmentTransactionCreated_351                    MsgType = -110351
	MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352 MsgType = -100352
	MsgType_CommitmentTxSigned_ToAliceSign_353                                   MsgType = -353
	MsgType_CommitmentTxSigned_SecondToBobSign_354                               MsgType = -354
	MsgType_CommitmentTxSigned_RecvRevokeAndAcknowledgeCommitmentTransaction_352 MsgType = -110352

	MsgType_SendCloseChannelRequest_38 MsgType = -100038
	MsgType_CloseChannelRequest_38     MsgType = -38
	MsgType_RecvCloseChannelRequest_38 MsgType = -110038

	MsgType_SendCloseChannelSign_39 MsgType = -100039
	MsgType_CloseChannelSign_39     MsgType = -39
	MsgType_RecvCloseChannelSign_39 MsgType = -110039

	// path finding
	MsgType_HTLC_FindPath_401 MsgType = -100401
	MsgType_HTLC_Invoice_402  MsgType = -100402

	MsgType_HTLC_SendAddHTLC_40       MsgType = -100040
	MsgType_HTLC_AddHTLC_40           MsgType = -40
	MsgType_HTLC_RecvAddHTLC_40       MsgType = -110040
	MsgType_HTLC_SendAddHTLCSigned_41 MsgType = -100041
	MsgType_HTLC_PayerSignC3b_42      MsgType = -42
	MsgType_HTLC_PayeeCreateHTRD1a_43 MsgType = -43
	MsgType_HTLC_PayerSignHTRD1a_44   MsgType = -44
	MsgType_HTLC_RecvAddHTLCSigned_41 MsgType = -110041

	MsgType_HTLC_SendVerifyR_45     MsgType = -100045
	MsgType_HTLC_VerifyR_45         MsgType = -45
	MsgType_HTLC_RecvVerifyR_45     MsgType = -110045
	MsgType_HTLC_SendSignVerifyR_46 MsgType = -100046
	MsgType_HTLC_SendHerdHex_47     MsgType = -47
	MsgType_HTLC_SignHedHex_48      MsgType = -48
	MsgType_HTLC_RecvSignVerifyR_46 MsgType = -110046

	MsgType_HTLC_SendRequestCloseCurrTx_49 MsgType = -100049
	MsgType_HTLC_RequestCloseCurrTx_49     MsgType = -49
	MsgType_HTLC_RecvRequestCloseCurrTx_49 MsgType = -110049
	MsgType_HTLC_SendCloseSigned_50        MsgType = -100050
	MsgType_HTLC_CloseHtlcRequestSignBR_51 MsgType = -51
	MsgType_HTLC_CloseHtlcUpdateCnb_52     MsgType = -52
	MsgType_HTLC_RecvCloseSigned_50        MsgType = -110050

	//https://github.com/LightningOnOmnilayer/Omni-BOLT-spec/blob/master/OmniBOLT-05-Atomic-Swap-among-Channels.md
	MsgType_Atomic_SendSwap_80       MsgType = -100080
	MsgType_Atomic_Swap_80           MsgType = -80
	MsgType_Atomic_RecvSwap_80       MsgType = -110080
	MsgType_Atomic_SendSwapAccept_81 MsgType = -100081
	MsgType_Atomic_SwapAccept_81     MsgType = -81
	MsgType_Atomic_RecvSwapAccept_81 MsgType = -110081
	//endregion
)

func CheckExist(msgType MsgType) bool {
	switch msgType {
	case MsgType_Error_0:
		return true
	case MsgType_UserLogin_2001:
		return true
	case MsgType_UserLogout_2002:
		return true
	case MsgType_p2p_ConnectPeer_2003:
		return true
	case MsgType_GetMnemonic_2004:
		return true
	case MsgType_GetObdNodeInfo_2005:
		return true
	case MsgType_GetMiniBtcFundAmount_2006:
		return true
	case MsgType_HeartBeat_2007:
		return true
	case MsgType_Core_GetNewAddress_2101:
		return true
	case MsgType_Core_GetMiningInfo_2102:
		return true
	case MsgType_Core_GetNetworkInfo_2103:
		return true
	case MsgType_Core_SignMessageWithPrivKey_2104:
		return true
	case MsgType_Core_VerifyMessage_2105:
		return true
	case MsgType_Core_DumpPrivKey_2106:
		return true
	case MsgType_Core_ListUnspent_2107:
		return true
	case MsgType_Core_BalanceByAddress_2108:
		return true
	case MsgType_Core_FundingBTC_2109:
		return true
	case MsgType_Core_BtcCreateMultiSig_2110:
		return true
	case MsgType_Core_Btc_ImportPrivKey_2111:
		return true
	case MsgType_Core_Omni_GetBalance_2112:
		return true
	case MsgType_Core_Omni_CreateNewTokenFixed_2113:
		return true
	case MsgType_Core_Omni_CreateNewTokenManaged_2114:
		return true
	case MsgType_Core_Omni_GrantNewUnitsOfManagedToken_2115:
		return true
	case MsgType_Core_Omni_RevokeUnitsOfManagedToken_2116:
		return true
	case MsgType_Core_Omni_ListProperties_2117:
		return true
	case MsgType_Core_Omni_GetTransaction_2118:
		return true
	case MsgType_Core_Omni_GetProperty_2119:
		return true
	case MsgType_Core_Omni_FundingAsset_2120:
		return true
	case MsgType_Core_Omni_Send_2121:
		return true
	case MsgType_Core_GetTransactionByTxid_2122:
		return true
	case MsgType_Core_SignRawTransaction_2123:
		return true
	case MsgType_Mnemonic_CreateAddress_3000:
		return true
	case MsgType_Mnemonic_GetAddressByIndex_3001:
		return true
	case MsgType_SendChannelOpen_32:
		return true
	case MsgType_ChannelOpen_AllItem_3150:
		return true
	case MsgType_ChannelOpen_ItemByTempId_3151:
		return true
	case MsgType_ChannelOpen_Count_3152:
		return true
	case MsgType_ChannelOpen_DelItemByTempId_3153:
		return true
	case MsgType_GetChannelInfoByChannelId_3154:
		return true
	case MsgType_GetChannelInfoByDbId_3155:
		return true
	case MsgType_CheckChannelAddessExist_3156:
		return true
	case MsgType_SendChannelAccept_33:
		return true
	case MsgType_FundingCreate_SendAssetFundingCreated_34:
		return true
	case MsgType_ClientSign_Duplex_AssetFunding_ChannelAddressSignC1a_1034:
		return true
	case MsgType_FundingCreate_Asset_AllItem_3100:
		return true
	case MsgType_FundingCreate_Asset_ItemById_3101:
		return true
	case MsgType_FundingCreate_Asset_ItemByChannelId_3102:
		return true
	case MsgType_FundingCreate_Asset_Count_3103:
		return true
	case MsgType_FundingCreate_SendBtcFundingCreated_340:
		return true
	case MsgType_FundingCreate_BtcFundingMinerRDTxToClient_341:
		return true
	case MsgType_FundingCreate_Btc_AllItem_3104:
		return true
	case MsgType_FundingCreate_Btc_ItemById_3105:
		return true
	case MsgType_FundingCreate_Btc_ItemByTempChannelId_3106:
		return true
	case MsgType_FundingCreate_Btc_RDAllItem_3107:
		return true
	case MsgType_FundingCreate_Btc_ItemRDById_3108:
		return true
	case MsgType_FundingCreate_Btc_ItemRDByTempChannelId_3109:
		return true
	case MsgType_FundingCreate_Btc_ItemRDByTempChannelIdAndTxId_3110:
		return true
	case MsgType_FundingCreate_Btc_ItemByChannelId_3111:
		return true
	case MsgType_FundingSign_SendAssetFundingSigned_35:
		return true
	case MsgType_FundingSign_SendBtcSign_350:
		return true
	case MsgType_CommitmentTx_SendCommitmentTransactionCreated_351:
		return true
	case MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352:
		return true
	case MsgType_CommitmentTx_ItemsByChanId_3200:
		return true
	case MsgType_CommitmentTx_ItemById_3201:
		return true
	case MsgType_CommitmentTx_Count_3202:
		return true
	case MsgType_CommitmentTx_LatestCommitmentTxByChanId_3203:
		return true
	case MsgType_CommitmentTx_LatestRDByChanId_3204:
		return true
	case MsgType_CommitmentTx_LatestBRByChanId_3205:
		return true
	case MsgType_CommitmentTx_SendSomeCommitmentById_3206:
		return true
	case MsgType_CommitmentTx_AllRDByChanId_3207:
		return true
	case MsgType_CommitmentTx_AllBRByChanId_3208:
		return true
	case MsgType_Htlc_GetLatestHT1aOrHE1b_3250:
		return true
	case MsgType_Htlc_GetHT1aOrHE1bBySomeCommitmentId_3251:
		return true
	case MsgType_SendCloseChannelRequest_38:
		return true
	case MsgType_SendCloseChannelSign_39:
		return true
	case MsgType_HTLC_FindPath_401:
		return true
	case MsgType_HTLC_Invoice_402:
		return true
	case MsgType_HTLC_SendAddHTLC_40:
		return true
	case MsgType_HTLC_SendAddHTLCSigned_41:
		return true
	case MsgType_HTLC_SendVerifyR_45:
		return true
	case MsgType_HTLC_SendSignVerifyR_46:
		return true
	case MsgType_HTLC_SendRequestCloseCurrTx_49:
		return true
	case MsgType_HTLC_SendCloseSigned_50:
		return true
	case MsgType_Atomic_SendSwap_80:
		return true
	case MsgType_Atomic_SendSwapAccept_81:
		return true
	}
	return false
}
