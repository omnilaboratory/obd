package enum

type UserState int

const (
	UserState_ErrorState UserState = -1
	UserState_Offline    UserState = 0
	UserState_OnLine     UserState = 1
)

type SendTargetType int

const (
	SendTargetType_SendToNone     SendTargetType = -1
	SendTargetType_SendToAll      SendTargetType = 0
	SendTargetType_SendToSomeone  SendTargetType = 1
	SendTargetType_SendToExceptMe SendTargetType = 2
)

type MsgType int

const (
	MsgType_Error      MsgType = 0
	MsgType_UserLogin  MsgType = 1
	MsgType_UserLogout MsgType = 2

	MsgType_Core_GetNewAddress_1001                    MsgType = 1001
	MsgType_Core_GetMiningInfo_1002                    MsgType = 1002
	MsgType_Core_GetNetworkInfo_1003                   MsgType = 1003
	MsgType_Core_SignMessageWithPrivKey_1004           MsgType = 1004
	MsgType_Core_VerifyMessage_1005                    MsgType = 1005
	MsgType_Core_DumpPrivKey_1006                      MsgType = 1006
	MsgType_Core_ListUnspent_1007                      MsgType = 1007
	MsgType_Core_BalanceByAddress_1008                 MsgType = 1008
	MsgType_Core_BtcCreateAndSignRawTransaction_1009   MsgType = 1009
	MsgType_Core_Omni_CreateAndSignRawTransaction_2001 MsgType = 2001

	MsgType_ChannelOpen_N32                   MsgType = -32
	MsgType_ChannelOpen_ItemByTempId_N3201    MsgType = -3201
	MsgType_ChannelOpen_AllItem_N3202         MsgType = -3202
	MsgType_ChannelOpen_Count_N3203           MsgType = -3203
	MsgType_ChannelOpen_DelItemByTempId_N3204 MsgType = -3204
	MsgType_ForceCloseChannel_N3205           MsgType = -3205

	MsgType_ChannelAccept_N33 MsgType = -33

	MsgType_FundingCreate_OmniCreate_N34     MsgType = -34
	MsgType_FundingCreate_BtcCreate_N3400    MsgType = -3400
	MsgType_FundingCreate_ItemByTempId_N3401 MsgType = -3401
	MsgType_FundingCreate_ItemById_N3402     MsgType = -3402
	MsgType_FundingCreate_ALlItem_N3403      MsgType = -3403
	MsgType_FundingCreate_Count_N3404        MsgType = -3404
	MsgType_FundingCreate_DelById_N3405      MsgType = -3405

	MsgType_FundingSign_OmniSign_N35  MsgType = -35
	MsgType_FundingSign_BtcSign_N3500 MsgType = -3500

	MsgType_CommitmentTx_Create_N351                       MsgType = -351
	MsgType_CommitmentTx_ItemsByChanId_N35101              MsgType = -35101
	MsgType_CommitmentTx_ItemById_N35102                   MsgType = -35102
	MsgType_CommitmentTx_Count_N35103                      MsgType = -35103
	MsgType_CommitmentTx_LatestCommitmentTxByChanId_N35104 MsgType = -35104
	MsgType_CommitmentTx_LatestRDByChanId_N35105           MsgType = -35105
	MsgType_CommitmentTx_LatestBRByChanId_N35106           MsgType = -35106
	MsgType_SendBreachRemedyTransaction_N35107             MsgType = -35107
	MsgType_CommitmentTx_AllRDByChanId_N35108              MsgType = -35108
	MsgType_CommitmentTx_AllBRByChanId_N35109              MsgType = -35109
	MsgType_CommitmentTx_GetBroadcastCommitmentTx_N35110   MsgType = -35110
	MsgType_CommitmentTx_GetBroadcastRDTx_N35111           MsgType = -35111
	MsgType_CommitmentTx_GetBroadcastBRTx_N35112           MsgType = -35112

	MsgType_CommitmentTxSigned_Sign_N352           MsgType = -352
	MsgType_CommitmentTxSigned_ItemByChanId_N35201 MsgType = -35201
	MsgType_CommitmentTxSigned_ItemById_N35202     MsgType = -35202
	MsgType_CommitmentTxSigned_Count_N35203        MsgType = -35203

	MsgType_GetBalanceRequest_N353 MsgType = -353
	MsgType_GetBalanceRespond_N354 MsgType = -354

	MsgType_CloseChannelRequest_N38 MsgType = -38
	MsgType_CloseChannelSign_N39    MsgType = -39

	MsgType_HTLC_RequestH_N40               MsgType = -40
	MsgType_HTLC_CreatedRAndHInfoList_N4001 MsgType = -4001
	MsgType_HTLC_CreatedRAndHInfoItem_N4002 MsgType = -4002
	MsgType_HTLC_RespondH_N41               MsgType = -41
	MsgType_HTLC_SignedRAndHInfoList_N4101  MsgType = -4101
	MsgType_HTLC_SignedRAndHInfoItem_N4102  MsgType = -4102
	MsgType_HTLC_FindPathAndSendH_N42       MsgType = -42
	MsgType_HTLC_SignGetH_N43               MsgType = -43
	MsgType_HTLC_CreateCommitmentTx_N44     MsgType = -44
	MsgType_HTLC_SendH_N45                  MsgType = -45
)
