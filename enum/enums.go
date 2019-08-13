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

const (
	MsgType_UserLogin          = 1
	MsgType_UserLogout         = 2
	MsgType_ChannelOpen        = -32
	MsgType_ChannelAccept      = -33
	MsgType_FundingCreated     = -34
	MsgType_FundingSigned      = -35
	MsgType_CommitmentTx       = -351
	MsgType_CommitmentTxSigned = -352
	MsgType_GetBalanceRequest  = -353
	MsgType_GetBalanceRespond  = -354
)
