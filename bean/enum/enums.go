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

	MsgType_Core_GetNewAddress                         MsgType = 1001
	MsgType_Core_GetMiningInfo                         MsgType = 1002
	MsgType_Core_GetNetworkInfo                        MsgType = 1003
	MsgType_Core_SignMessageWithPrivKey                MsgType = 1004
	MsgType_Core_VerifyMessage                         MsgType = 1005
	MsgType_Core_DumpPrivKey                           MsgType = 1006
	MsgType_Core_ListUnspent                           MsgType = 1007
	MsgType_Core_BalanceByAddress                      MsgType = 1008
	MsgType_Core_BtcCreateAndSignRawTransaction_1009   MsgType = 1009
	MsgType_Core_Omni_CreateAndSignRawTransaction_2001 MsgType = 2001

	MsgType_ChannelOpen                 MsgType = -32
	MsgType_ChannelOpen_ItemByTempId    MsgType = -3201
	MsgType_ChannelOpen_AllItem         MsgType = -3202
	MsgType_ChannelOpen_Count           MsgType = -3203
	MsgType_ChannelOpen_DelItemByTempId MsgType = -3204
	MsgType_ForceCloseChannel           MsgType = -3205

	MsgType_ChannelAccept MsgType = -33

	MsgType_FundingCreate_OmniCreate   MsgType = -34
	MsgType_FundingCreate_BtcCreate    MsgType = -3400
	MsgType_FundingCreate_ItemByTempId MsgType = -3401
	MsgType_FundingCreate_ItemById     MsgType = -3402
	MsgType_FundingCreate_ALlItem      MsgType = -3403
	MsgType_FundingCreate_Count        MsgType = -3404
	MsgType_FundingCreate_DelById      MsgType = -3405

	MsgType_FundingSign_OmniSign MsgType = -35
	MsgType_FundingSign_BtcSign  MsgType = -3500

	MsgType_CommitmentTx_Create                     MsgType = -351
	MsgType_CommitmentTx_ItemsByChanId              MsgType = -35101
	MsgType_CommitmentTx_ItemById                   MsgType = -35102
	MsgType_CommitmentTx_Count                      MsgType = -35103
	MsgType_CommitmentTx_LatestCommitmentTxByChanId MsgType = -35104
	MsgType_CommitmentTx_LatestRDByChanId           MsgType = -35105
	MsgType_CommitmentTx_LatestBRByChanId           MsgType = -35106
	MsgType_SendBreachRemedyTransaction             MsgType = -35107
	MsgType_CommitmentTx_AllRDByChanId              MsgType = -35108
	MsgType_CommitmentTx_AllBRByChanId              MsgType = -35109
	MsgType_CommitmentTx_GetBroadcastCommitmentTx   MsgType = -35110
	MsgType_CommitmentTx_GetBroadcastRDTx           MsgType = -35111
	MsgType_CommitmentTx_GetBroadcastBRTx           MsgType = -35112

	MsgType_CommitmentTxSigned_Sign         MsgType = -352
	MsgType_CommitmentTxSigned_ItemByChanId MsgType = -35201
	MsgType_CommitmentTxSigned_ItemById     MsgType = -35202
	MsgType_CommitmentTxSigned_Count        MsgType = -35203

	MsgType_GetBalanceRequest MsgType = -353
	MsgType_GetBalanceRespond MsgType = -354

	MsgType_CloseChannelRequest MsgType = -38
	MsgType_CloseChannelSign    MsgType = -39
)
