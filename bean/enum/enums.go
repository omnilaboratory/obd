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
	MsgType_UserLogin     MsgType = 1
	MsgType_UserLogout    MsgType = 2
	MsgType_ChannelOpen   MsgType = -32
	MsgType_ChannelAccept MsgType = -33

	MsgType_FundingCreated         MsgType = -34
	MsgType_GetFundingCreated      MsgType = -3401
	MsgType_DelTableFundingCreated MsgType = -3402
	MsgType_DelItemFundingCreated  MsgType = -3403
	MsgType_CountFundingCreated    MsgType = -3404

	MsgType_FundingSigned      MsgType = -35
	MsgType_CommitmentTx       MsgType = -351
	MsgType_CommitmentTxSigned MsgType = -352
	MsgType_GetBalanceRequest  MsgType = -353
	MsgType_GetBalanceRespond  MsgType = -354
)
