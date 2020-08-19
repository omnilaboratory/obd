package enum

const (
	Tips_common_wrong = "错误的 "

	Tips_channel_notFoundChannelInCreate               = "未能通过临时通道id找到处于创建状态的通道: "
	Tips_channel_notThePeerIdB                         = "你不是通道创建的接收方 "
	Tips_channel_changePubkeyForChannel                = "因为生成的通道多钱地址已经用过了，请换一个pubkey"
	Tips_channel_wrongLatestCommitmentTxState          = "最新承诺交易的状态不满足要求"
	Tips_channel_notFoundLatestCommitmentTx            = "没有找到最新的承诺交易"
	Tips_channel_LatestCommitmentTxNotInReadySendState = "当前承诺交易的状态不是可以广播的状态"

	Tips_funding_notFoundChannelByTempId    = "未能通过临时通道id找到对应的通道: "
	Tips_funding_notFoundChannelByChannelId = "未能通过通道id找到对应的通道: "
	Tips_funding_notFundBtcState            = "当前通道不是处于充值btc的阶段 "
	Tips_funding_notFundAssetState          = "当前通道不是处于充值资产的阶段 "
	Tips_funding_failDecodeRawTransaction   = "DecodeRawTransaction失败 "
	Tips_funding_enoughBtcFundingTime       = "btc的充值次数够3次了"
	Tips_funding_btcAmountMustGreater       = "btc的数量太少了，必须大于 "
	Tips_funding_btcTxBeenSend              = "这个交易已经被广播过了"
	Tips_funding_fundTxIsRunning            = "最新的btc充值正在执行，请稍后"
	Tips_funding_notFoundFundTx             = "未能找到btc的充值交易"
)
