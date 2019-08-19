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
	MsgType_UserLogin  MsgType = 1
	MsgType_UserLogout MsgType = 2

	MsgType_ChannelOpen                 MsgType = -32
	MsgType_ChannelOpen_ItemByTempId    MsgType = -3201
	MsgType_ChannelOpen_AllItem         MsgType = -3202
	MsgType_ChannelOpen_Count           MsgType = -3203
	MsgType_ChannelOpen_DelItemByTempId MsgType = -3204

	MsgType_ChannelAccept MsgType = -33

	MsgType_FundingCreate_Edit   MsgType = -34
	MsgType_FundingCreate_Item   MsgType = -3401
	MsgType_FundingCreate_Count  MsgType = -3402
	MsgType_FundingCreate_Del    MsgType = -3403
	MsgType_FundingCreate_DelAll MsgType = -3404

	MsgType_FundingSign_Edit   MsgType = -35
	MsgType_FundingSign_Item   MsgType = -3501
	MsgType_FundingSign_Count  MsgType = -3502
	MsgType_FundingSign_Del    MsgType = -3503
	MsgType_FundingSign_DelAll MsgType = -3504

	MsgType_CommitmentTx_Edit         MsgType = -351
	MsgType_CommitmentTx_ItemByChanId MsgType = -35101
	MsgType_CommitmentTx_Count        MsgType = -35102
	MsgType_CommitmentTx_Del          MsgType = -35103

	MsgType_CommitmentTxSigned MsgType = -352
	MsgType_GetBalanceRequest  MsgType = -353
	MsgType_GetBalanceRespond  MsgType = -354
)
