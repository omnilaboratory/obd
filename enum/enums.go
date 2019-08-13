package enum

type UserState int

const (
	ErrorState UserState = -1
	Offline    UserState = 0
	OnLine     UserState = 1
)
