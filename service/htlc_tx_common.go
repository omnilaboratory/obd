package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/tool"
	"encoding/json"
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"log"
	"time"
)

func getTwoChannelOfSingleHop(htlcRAndHInfo dao.HtlcRAndHInfo, channelAliceInfos []dao.ChannelInfo, channelCarlInfos []dao.ChannelInfo) (string, *dao.ChannelInfo, *dao.ChannelInfo) {
	for _, aliceChannel := range channelAliceInfos {
		if aliceChannel.PeerIdA == htlcRAndHInfo.SenderPeerId {
			bobPeerId := aliceChannel.PeerIdB
			carlChannel, err := getCarlChannelHasInterNodeBob(htlcRAndHInfo, aliceChannel, channelCarlInfos, aliceChannel.PeerIdA, bobPeerId)
			if err == nil {
				return bobPeerId, &aliceChannel, carlChannel
			}
		} else {
			bobPeerId := aliceChannel.PeerIdA
			carlChannel, err := getCarlChannelHasInterNodeBob(htlcRAndHInfo, aliceChannel, channelCarlInfos, aliceChannel.PeerIdB, bobPeerId)
			if err == nil {
				return bobPeerId, &aliceChannel, carlChannel
			}
		}
	}
	return "", nil, nil
}

func getCarlChannelHasInterNodeBob(htlcRAndHInfo dao.HtlcRAndHInfo, aliceChannel dao.ChannelInfo, channelCarlInfos []dao.ChannelInfo, alicePeerId, bobPeerId string) (*dao.ChannelInfo, error) {
	//whether bob is online
	if err := FindUserIsOnline(bobPeerId); err != nil {
		return nil, err
	}

	//alice and bob's channel, whether alice has enough money
	aliceCommitmentTxInfo, err := getLatestCommitmentTx(aliceChannel.ChannelId, alicePeerId)
	if err != nil {
		return nil, err
	}
	if aliceCommitmentTxInfo.AmountToRSMC < (htlcRAndHInfo.Amount + tool.GetHtlcFee()) {
		return nil, errors.New("channel not have enough money")
	}

	//bob and carl's channel,whether bob has enough money
	for _, carlChannel := range channelCarlInfos {
		if (carlChannel.PeerIdA == bobPeerId && carlChannel.PeerIdB == htlcRAndHInfo.RecipientPeerId) ||
			(carlChannel.PeerIdB == bobPeerId && carlChannel.PeerIdA == htlcRAndHInfo.RecipientPeerId) {
			commitmentTxInfo, err := getLatestCommitmentTx(carlChannel.ChannelId, bobPeerId)
			if err != nil {
				continue
			}
			if commitmentTxInfo.AmountToRSMC < htlcRAndHInfo.Amount {
				continue
			}
			return &carlChannel, nil
		}
	}
	return nil, errors.New("not found the channel")
}

func getAllChannels(peerId string) (channelInfos []dao.ChannelInfo) {
	channelInfos = make([]dao.ChannelInfo, 0)
	_ = db.Select(q.Or(q.Eq("PeerIdA", peerId), q.Eq("PeerIdB", peerId)), q.Eq("CurrState", dao.ChannelState_Accept)).Find(&channelInfos)
	return channelInfos
}

func htlcAliceAbortLastCommitmentTx(tx storm.Node, channelInfo dao.ChannelInfo, user bean.User, fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen) error {
	owner := channelInfo.PeerIdA
	// 现在是创建PeerIdA方向，当前的操作者user可能对应PeerIdA，也可能对应PeerIdB
	// 默认假设当前操作者正好是PeerIdA，数据（公钥，私钥）的使用当前操作用户的
	var bobIsInterNodeSoAliceSend2Bob = true
	// 如果不是，那PeerIdA对应的就是另一个方，数据（公钥，私钥）的使用就要使用另一方的数据了
	if user.PeerId == channelInfo.PeerIdB {
		bobIsInterNodeSoAliceSend2Bob = false
	}

	//针对的是Cna
	var lastCommitmentATx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentATx)
	if err != nil {
		return err
	}

	//	为上一次的Rsmc交易构建BR交易，Alice宣布上一次的交易作废。
	// （惩罚交易：如果Alice广播这次作废的交易，因为BR交易的存在（bob才能广播），就会失去自己的钱，这个时候，对应的RD交易还需要等待1000个区块高度才能广播）
	if lastCommitmentATx != nil {
		//如果已经创建过了，return
		count, _ := tx.Select(q.Eq("CommitmentTxId", lastCommitmentATx.Id)).Count(&dao.BreachRemedyTransaction{})
		if count > 0 {
			err = errors.New("already exist BreachRemedyTransaction ")
			return err
		}

		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentATx.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return err
		}

		// create BRa tx  for bob ，let the lastCommitmentATx abort,
		// 为bob创建上一次交易的BR对象
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdB, &channelInfo, lastCommitmentATx, &user)
		if err != nil {
			log.Println(err)
			return err
		}

		//如果金额大于0
		if breachRemedyTransaction.Amount > 0 {
			lastTempAddressPrivateKey := ""
			// 如果当前操作用户是PeerIdA方（概念中的Alice方），则取当前操作人传入的数据
			if bobIsInterNodeSoAliceSend2Bob {
				lastTempAddressPrivateKey = requestData.LastTempAddressPrivateKey
			} else {
				// 如果当前操作用户是PeerIdB方，而我们现在正在处理Alice方，所以我们要取另一方的数据
				lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentATx.RSMCTempAddressPubKey]
			}
			if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
				err = errors.New("fail to get the lastTempAddressPrivateKey")
				log.Println(err)
				return err
			}

			inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentATx.RSMCTxHash, lastCommitmentATx.RSMCMultiAddress, lastCommitmentATx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return err
			}

			txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
				lastCommitmentATx.RSMCMultiAddress,
				[]string{
					lastTempAddressPrivateKey,
					tempAddrPrivateKeyMap[channelInfo.PubKeyB],
				},
				inputs,
				channelInfo.AddressB,
				fundingTransaction.FunderAddress,
				fundingTransaction.PropertyId,
				breachRemedyTransaction.Amount,
				0,
				0,
				&lastCommitmentATx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return err
			}
			breachRemedyTransaction.Txid = txid
			breachRemedyTransaction.TransactionSignHex = hex
			breachRemedyTransaction.SignAt = time.Now()
			breachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
			err = tx.Save(breachRemedyTransaction)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		lastCommitmentATx.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentATx)
		if err != nil {
			log.Println(err)
			return err
		}
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

//概念的bob方结束上一次的Rsmc的交易
func htlcBobAbortLastCommitmentTx(tx storm.Node, channelInfo dao.ChannelInfo, user bean.User, fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen) error {
	owner := channelInfo.PeerIdB
	// 现在是创建PeerIdB方向，当前的操作者user可能对应PeerIdA，也可能对应PeerIdB
	// 默认假设当前操作者正好是PeerIdA，数据（公钥，私钥）的使用当前操作用户的
	var bobIsInterNodeSoAliceSend2Bob = true
	// 如果不是，那PeerIdA对应的就是另一个方，数据（公钥，私钥）的使用就要使用另一方的数据了
	if user.PeerId == channelInfo.PeerIdB {
		bobIsInterNodeSoAliceSend2Bob = false
	}

	//针对的是Cnb
	var lastCommitmentBTx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentBTx)
	if err != nil {
		lastCommitmentBTx = nil
	}
	//	为上一次的Rsmc交易构建BR交易，Bob宣布上一次的交易作废。
	// （惩罚交易：如果Bob广播这次作废的交易，因为BR交易的存在（alice才能广播），bob就会失去自己的钱，这个时候，对应的RD交易还需要等待1000个区块高度才能广播）
	if lastCommitmentBTx != nil {
		//如果已经创建过了，return
		count, _ := tx.Select(q.Eq("CommitmentTxId", lastCommitmentBTx.Id)).Count(&dao.BreachRemedyTransaction{})
		if count > 0 {
			err = errors.New("already exist BreachRemedyTransaction ")
			return err
		}

		// RD的存在意义是：告诉对方，我的钱是锁定在这个临时地址的，要取出来，是需要等待1000个10分钟的，
		// 当然最终自己也能取回钱，关键的地方是如果创建了新的后续的承诺交易，就会产生BR交易（惩罚交易），
		// 如果有了新的承诺交易，而某人想耍赖，不承认新的交易，而去广播之前的交易，因为等待的1000个10分钟，通过BR就能让违反规则的人血本无归
		// 如果没有新的交易，当前操作者也能取回自己的钱，虽然要等待1000个区块，但是也不会让自己的钱丢失
		lastRDTransaction := &dao.RevocableDeliveryTransaction{}
		err = tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CommitmentTxId", lastCommitmentBTx.Id), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return err
		}

		// create BRa tx  for bob ，let the lastCommitmentBTx abort,
		// 为了惩罚Bob，为Alice方创建上一次交易的BR交易
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdA, &channelInfo, lastCommitmentBTx, &user)
		if err != nil {
			log.Println(err)
			return err
		}

		//如果金额大于0
		if breachRemedyTransaction.Amount > 0 {
			lastTempAddressPrivateKey := ""
			if bobIsInterNodeSoAliceSend2Bob {
				lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentBTx.RSMCTempAddressPubKey]
			} else {
				lastTempAddressPrivateKey = requestData.LastTempAddressPrivateKey
			}
			if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
				err = errors.New("fail to get the lastTempAddressPrivateKey")
				log.Println(err)
				return err
			}

			inputs, err := getInputsForNextTxByParseTxHashVout(lastCommitmentBTx.RSMCTxHash, lastCommitmentBTx.RSMCMultiAddress, lastCommitmentBTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return err
			}

			txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
				lastCommitmentBTx.RSMCMultiAddress,
				[]string{
					lastTempAddressPrivateKey,
					tempAddrPrivateKeyMap[channelInfo.PubKeyB],
				},
				inputs,
				channelInfo.AddressA,
				fundingTransaction.FunderAddress,
				fundingTransaction.PropertyId,
				breachRemedyTransaction.Amount,
				0,
				0,
				&lastCommitmentBTx.RSMCRedeemScript)
			if err != nil {
				log.Println(err)
				return err
			}
			breachRemedyTransaction.Txid = txid
			breachRemedyTransaction.TransactionSignHex = hex
			breachRemedyTransaction.SignAt = time.Now()
			breachRemedyTransaction.CurrState = dao.TxInfoState_CreateAndSign
			err = tx.Save(breachRemedyTransaction)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		lastCommitmentBTx.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentBTx)
		if err != nil {
			log.Println(err)
			return err
		}
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func createHtlcTimeoutTxForAliceSide(tx storm.Node, owner string, channelInfo dao.ChannelInfo, fundingTransaction dao.FundingTransaction, commitmentTxInfo dao.CommitmentTransaction, requestData bean.HtlcRequestOpen, timeout int, operator bean.User) (htlcTimeoutTx *dao.HTLCTimeoutTxForAAndExecutionForB, err error) {
	outputBean := commitmentOutputBean{}
	outputBean.AmountToRsmc = commitmentTxInfo.AmountToHtlc
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
	outputBean.RsmcTempPubKey = requestData.CurrHtlcTempAddressForHt1aPubKey

	htlcTimeoutTx, err = createHtlcTimeoutTxObj(tx, owner, channelInfo, commitmentTxInfo, outputBean, timeout, operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.HtlcTxHash, commitmentTxInfo.HTLCMultiAddress, commitmentTxInfo.HTLCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.HTLCMultiAddress,
		[]string{
			requestData.CurrHtlcTempAddressPrivateKey,
			tempAddrPrivateKeyMap[channelInfo.PubKeyB],
		},
		inputs,
		htlcTimeoutTx.RSMCMultiAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		htlcTimeoutTx.RSMCOutAmount,
		0,
		htlcTimeoutTx.Timeout,
		&commitmentTxInfo.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcTimeoutTx.RSMCTxid = txid
	htlcTimeoutTx.RSMCTxHash = hex
	htlcTimeoutTx.SignAt = time.Now()
	htlcTimeoutTx.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(htlcTimeoutTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return htlcTimeoutTx, nil
}

func createHtlcTimeoutTxForBobSide(tx storm.Node, owner string, channelInfo dao.ChannelInfo, fundingTransaction dao.FundingTransaction, commitmentTxInfo dao.CommitmentTransaction, requestData bean.HtlcRequestOpen, timeout int, operator bean.User) (htlcTimeoutTx *dao.HTLCTimeoutTxForAAndExecutionForB, err error) {
	outputBean := commitmentOutputBean{}
	outputBean.AmountToRsmc = commitmentTxInfo.AmountToHtlc
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
	outputBean.RsmcTempPubKey = requestData.CurrHtlcTempAddressForHt1aPubKey
	htlcTimeoutTx, err = createHtlcTimeoutTxObj(tx, owner, channelInfo, commitmentTxInfo, outputBean, timeout, operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.HtlcTxHash, commitmentTxInfo.HTLCMultiAddress, commitmentTxInfo.HTLCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.HTLCMultiAddress,
		[]string{
			requestData.CurrHtlcTempAddressPrivateKey,
			tempAddrPrivateKeyMap[channelInfo.PubKeyA],
		},
		inputs,
		htlcTimeoutTx.RSMCMultiAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		htlcTimeoutTx.RSMCOutAmount,
		0,
		htlcTimeoutTx.Timeout,
		&commitmentTxInfo.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcTimeoutTx.RSMCTxid = txid
	htlcTimeoutTx.RSMCTxHash = hex
	htlcTimeoutTx.SignAt = time.Now()
	htlcTimeoutTx.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(htlcTimeoutTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return htlcTimeoutTx, nil
}

func createHtlcTimeoutDeliveryTx(tx storm.Node, owner string, outputAddress string, timeout int, channelInfo dao.ChannelInfo, fundingTransaction dao.FundingTransaction, commitmentTxInfo dao.CommitmentTransaction, htlcRequestOpen bean.HtlcRequestOpen, operator bean.User) (htlcTimeoutDeliveryTx *dao.HTLCTimeoutDeliveryTxB, err error) {
	htlcTimeoutDeliveryTx = &dao.HTLCTimeoutDeliveryTxB{}
	htlcTimeoutDeliveryTx.ChannelId = channelInfo.ChannelId
	htlcTimeoutDeliveryTx.CommitmentTxId = commitmentTxInfo.Id
	htlcTimeoutDeliveryTx.PropertyId = commitmentTxInfo.PropertyId
	htlcTimeoutDeliveryTx.OutputAddress = outputAddress
	htlcTimeoutDeliveryTx.InputHex = commitmentTxInfo.HtlcTxHash
	htlcTimeoutDeliveryTx.OutAmount = commitmentTxInfo.AmountToHtlc
	htlcTimeoutDeliveryTx.Owner = owner
	htlcTimeoutDeliveryTx.CurrState = dao.TxInfoState_CreateAndSign
	htlcTimeoutDeliveryTx.CreateBy = operator.PeerId
	htlcTimeoutDeliveryTx.Timeout = timeout
	htlcTimeoutDeliveryTx.CreateAt = time.Now()

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.HtlcTxHash, commitmentTxInfo.HTLCMultiAddress, commitmentTxInfo.HTLCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.HTLCMultiAddress,
		[]string{
			htlcRequestOpen.CurrHtlcTempAddressPrivateKey,
			tempAddrPrivateKeyMap[channelInfo.PubKeyB],
		},
		inputs,
		htlcTimeoutDeliveryTx.OutputAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		htlcTimeoutDeliveryTx.OutAmount,
		0,
		htlcTimeoutDeliveryTx.Timeout,
		&commitmentTxInfo.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcTimeoutDeliveryTx.Txid = txid
	htlcTimeoutDeliveryTx.TxHash = hex
	err = tx.Save(htlcTimeoutDeliveryTx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return htlcTimeoutDeliveryTx, nil
}

func createHtlcTimeoutTxObj(tx storm.Node, owner string, channelInfo dao.ChannelInfo, commitmentTxInfo dao.CommitmentTransaction, outputBean commitmentOutputBean, timeout int, user bean.User) (*dao.HTLCTimeoutTxForAAndExecutionForB, error) {
	htlcTimeoutTx := &dao.HTLCTimeoutTxForAAndExecutionForB{}
	htlcTimeoutTx.ChannelId = channelInfo.ChannelId
	htlcTimeoutTx.CommitmentTxId = commitmentTxInfo.Id
	htlcTimeoutTx.PropertyId = commitmentTxInfo.PropertyId
	htlcTimeoutTx.Owner = owner
	count, err := tx.Select(q.Eq("ChannelId", htlcTimeoutTx.ChannelId), q.Eq("CommitmentTxId", htlcTimeoutTx.CommitmentTxId), q.Eq("Owner", owner)).Count(htlcTimeoutTx)
	if err == nil {
		if count > 0 {
			return nil, errors.New("already exist")
		}
	}
	//input
	htlcTimeoutTx.InputHex = commitmentTxInfo.HtlcTxHash

	//output to rsmc
	htlcTimeoutTx.RSMCTempAddressPubKey = outputBean.RsmcTempPubKey
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{htlcTimeoutTx.RSMCTempAddressPubKey, outputBean.OppositeSideChannelPubKey})
	if err != nil {
		return nil, err
	}
	htlcTimeoutTx.RSMCMultiAddress = gjson.Get(multiAddr, "address").String()
	htlcTimeoutTx.RSMCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	jsonData, err := rpcClient.GetAddressInfo(htlcTimeoutTx.RSMCMultiAddress)
	if err != nil {
		return nil, err
	}
	htlcTimeoutTx.RSMCMultiAddressScriptPubKey = gjson.Get(jsonData, "scriptPubKey").String()
	htlcTimeoutTx.RSMCOutAmount = outputBean.AmountToRsmc
	htlcTimeoutTx.Timeout = timeout
	htlcTimeoutTx.CreateBy = user.PeerId
	htlcTimeoutTx.CreateAt = time.Now()
	return htlcTimeoutTx, nil
}

func htlcCreateCna(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen,
	htlcSingleHopPathInfo dao.HtlcSingleHopPathInfo, hAndRInfo dao.HtlcRAndHInfo,
	bobIsInterNodeSoAliceSend2Bob bool, lastCommitmentATx *dao.CommitmentTransaction, owner string) (*dao.CommitmentTransaction, error) {
	// htlc的资产分配方案
	var outputBean = commitmentOutputBean{}
	if bobIsInterNodeSoAliceSend2Bob { //Alice send money to bob
		//	alice借道bob，bob作为中间商，而当前的操作者是alice
		//	这个时候，我们在创建Cna，那么当前操作者Alice传进来的信息就是创建临时多签地址，转账等交易需要的信息了
		//	而bob作为中间商，他的余额应该是不变的，变化的是alice的余额，一部分被锁定在了tohtlc的临时多签地址里面了
		outputBean.RsmcTempPubKey = requestData.CurrRsmcTempAddressPubKey
		outputBean.HtlcTempPubKey = requestData.CurrHtlcTempAddressPubKey
		if lastCommitmentATx == nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
			outputBean.AmountToOther = fundingTransaction.AmountB
		} else {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToRSMC).Sub(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
			outputBean.AmountToOther = lastCommitmentATx.AmountToOther
		}
	} else { //	bob send money to alice
		//	bob借道Alice作为中间节点转账，也就是当前操作者实际是Bob，Alice是中间商，当前通道作为接收者，
		// 	而这个时候，我们正在创建Cna，为了Alice创建，那么，就需要Alice的信息了
		// 	Alice作为中间商，她的余额应该不变，变化的是给bob的钱，因为借道，所以bob钱就应该锁定
		outputBean.RsmcTempPubKey = htlcSingleHopPathInfo.BobCurrRsmcTempPubKey
		outputBean.HtlcTempPubKey = htlcSingleHopPathInfo.BobCurrHtlcTempForHt1bPubKey
		if lastCommitmentATx == nil {
			outputBean.AmountToRsmc = fundingTransaction.AmountA
			outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
		} else {
			outputBean.AmountToRsmc = lastCommitmentATx.AmountToRSMC
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToOther).Sub(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
		}
	}
	outputBean.AmountToHtlc = hAndRInfo.Amount
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyB
	outputBean.OppositeSideChannelAddress = channelInfo.AddressB

	commitmentTxInfo, err := createCommitmentTx(owner, &channelInfo, &fundingTransaction, outputBean, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Htlc

	allUsedTxidTemp := ""
	// rsmc
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			commitmentTxInfo.RSMCMultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += usedTxid
		commitmentTxInfo.RSMCTxid = txid
		commitmentTxInfo.RSMCTxHash = hex
	}

	//htlc
	if commitmentTxInfo.AmountToHtlc > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			commitmentTxInfo.HTLCMultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToHtlc,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, allUsedTxidTemp)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += "," + usedTxid
		commitmentTxInfo.HTLCTxid = txid
		commitmentTxInfo.HtlcTxHash = hex
	}

	//create to Bob tx
	if commitmentTxInfo.AmountToOther > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTxToBob(
			channelInfo.ChannelAddress,
			allUsedTxidTemp,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			channelInfo.AddressB,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToOther,
			0,
			0, &channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		commitmentTxInfo.ToOtherTxid = txid
		commitmentTxInfo.ToOtherTxHash = hex
	}

	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_Htlc_GetH
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	if lastCommitmentATx != nil {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentATx.Id
		commitmentTxInfo.LastHash = lastCommitmentATx.CurrHash
	}
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = tx.Update(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, nil
}

func htlcCreateCnb(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen,
	htlcSingleHopPathInfo dao.HtlcSingleHopPathInfo, hAndRInfo dao.HtlcRAndHInfo,
	bobIsInterNodeSoAliceSend2Bob bool, lastCommitmentTx *dao.CommitmentTransaction, owner string) (*dao.CommitmentTransaction, error) {
	// htlc的资产分配方案
	var outputBean = commitmentOutputBean{}
	if bobIsInterNodeSoAliceSend2Bob { //Alice send money to bob
		//	alice借道bob，bob作为中间商，而当前的操作者是alice
		//	这个时候，我们在创建Cna，那么当前操作者Alice传进来的信息就是创建临时多签地址，转账等交易需要的信息了
		//	而bob作为中间商，他的余额应该是不变的，变化的是alice的余额，一部分被锁定在了tohtlc的临时多签地址里面了
		outputBean.RsmcTempPubKey = htlcSingleHopPathInfo.BobCurrRsmcTempPubKey
		outputBean.HtlcTempPubKey = htlcSingleHopPathInfo.BobCurrHtlcTempForHt1bPubKey
		if lastCommitmentTx == nil {
			// 给bob，bob是收款方，bob本身的余额还是放到RSMC里面
			outputBean.AmountToRsmc = fundingTransaction.AmountB
			// 给Alice的，Alice转账给bob，Alice要锁定一部分资金
			outputBean.AmountToOther, _ = decimal.NewFromFloat(fundingTransaction.AmountA).Sub(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
		} else {
			outputBean.AmountToRsmc = lastCommitmentTx.AmountToRSMC
			outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToOther).Sub(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
		}
	} else { //	bob send money to alice bob是发送方 bob的钱就要减少，alice的钱不变
		// requestData 请求数据就是当前用户Bob的数据，在创建Cnb的时候，需要用bob的临时地址和Alice的通道地址的私钥完成交易的构建
		outputBean.RsmcTempPubKey = requestData.CurrRsmcTempAddressPubKey
		outputBean.HtlcTempPubKey = requestData.CurrHtlcTempAddressPubKey
		if lastCommitmentTx == nil {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(fundingTransaction.AmountB).Add(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
			outputBean.AmountToOther = fundingTransaction.AmountA
		} else {
			outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentTx.AmountToRSMC).Sub(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
			outputBean.AmountToOther = lastCommitmentTx.AmountToOther
		}
	}
	outputBean.AmountToHtlc = hAndRInfo.Amount
	outputBean.OppositeSideChannelPubKey = channelInfo.PubKeyA
	outputBean.OppositeSideChannelAddress = channelInfo.AddressA

	commitmentTxInfo, err := createCommitmentTx(owner, &channelInfo, &fundingTransaction, outputBean, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = dao.CommitmentTransactionType_Htlc

	allUsedTxidTemp := ""
	// rsmc
	if commitmentTxInfo.AmountToRSMC > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			commitmentTxInfo.RSMCMultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToRSMC,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, "")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += usedTxid
		commitmentTxInfo.RSMCTxid = txid
		commitmentTxInfo.RSMCTxHash = hex
	}

	//htlc
	if commitmentTxInfo.AmountToHtlc > 0 {
		txid, hex, usedTxid, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTx(
			channelInfo.ChannelAddress,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			commitmentTxInfo.HTLCMultiAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToHtlc,
			0,
			0, &channelInfo.ChannelAddressRedeemScript, allUsedTxidTemp)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		allUsedTxidTemp += "," + usedTxid
		commitmentTxInfo.HTLCTxid = txid
		commitmentTxInfo.HtlcTxHash = hex
	}

	//create to alice tx
	if commitmentTxInfo.AmountToOther > 0 {
		txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForCommitmentTxToBob(
			channelInfo.ChannelAddress,
			allUsedTxidTemp,
			[]string{
				tempAddrPrivateKeyMap[channelInfo.PubKeyA],
				tempAddrPrivateKeyMap[channelInfo.PubKeyB],
			},
			channelInfo.AddressA,
			fundingTransaction.FunderAddress,
			fundingTransaction.PropertyId,
			commitmentTxInfo.AmountToOther,
			0,
			0, &channelInfo.ChannelAddressRedeemScript)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		commitmentTxInfo.ToOtherTxid = txid
		commitmentTxInfo.ToOtherTxHash = hex
	}

	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_Htlc_GetH
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	if lastCommitmentTx != nil {
		commitmentTxInfo.LastCommitmentTxId = lastCommitmentTx.Id
		commitmentTxInfo.LastHash = lastCommitmentTx.CurrHash
	}
	err = tx.Save(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	bytes, err := json.Marshal(commitmentTxInfo)
	msgHash := tool.SignMsgWithSha256(bytes)
	commitmentTxInfo.CurrHash = msgHash
	err = tx.Update(commitmentTxInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return commitmentTxInfo, nil
}

func htlcCreateRDOfRsmc(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen,
	htlcSingleHopPathInfo dao.HtlcSingleHopPathInfo, bobIsInterNodeSoAliceSend2Bob bool,
	commitmentTxInfo *dao.CommitmentTransaction, owner string) (*dao.RevocableDeliveryTransaction, error) {

	rdTransaction, err := createRDTx(owner, &channelInfo, commitmentTxInfo, channelInfo.AddressA, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if bobIsInterNodeSoAliceSend2Bob {
		currTempAddressPrivateKey = htlcRequestOpen.CurrRsmcTempAddressPrivateKey
	} else {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[htlcSingleHopPathInfo.BobCurrRsmcTempPubKey]
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHash, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.RSMCMultiAddress,
		[]string{
			currTempAddressPrivateKey,
			tempAddrPrivateKeyMap[channelInfo.PubKeyB],
		},
		inputs,
		rdTransaction.OutputAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		rdTransaction.Amount,
		0,
		rdTransaction.Sequence,
		&commitmentTxInfo.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	rdTransaction.Txid = txid
	rdTransaction.TransactionSignHex = hex
	rdTransaction.SignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return rdTransaction, nil
}

func createHtlcRD(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, requestData bean.HtlcRequestOpen, bobIsInterNodeSoAliceSend2Bob bool,
	htlcTimeoutTx *dao.HTLCTimeoutTxForAAndExecutionForB, owner string) (*dao.RevocableDeliveryTransaction, error) {

	outAddress := channelInfo.AddressA
	if bobIsInterNodeSoAliceSend2Bob == false {
		outAddress = channelInfo.AddressB
	}

	count, _ := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("CommitmentTxId", htlcTimeoutTx.Id), q.Eq("Owner", owner), q.Eq("RDType", 1)).Count(&dao.RevocableDeliveryTransaction{})
	if count > 0 {
		return nil, errors.New("already create")
	}

	rdTransaction, err := createHtlcRDTxObj(owner, &channelInfo, htlcTimeoutTx, outAddress, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := requestData.CurrHtlcTempAddressForHt1aPrivateKey
	currChannelAddressPrivateKey := tempAddrPrivateKeyMap[channelInfo.PubKeyB]
	if bobIsInterNodeSoAliceSend2Bob == false {
		currChannelAddressPrivateKey = tempAddrPrivateKeyMap[channelInfo.PubKeyA]
	}

	inputs, err := getInputsForNextTxByParseTxHashVout(htlcTimeoutTx.RSMCTxHash, htlcTimeoutTx.RSMCMultiAddress, htlcTimeoutTx.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		htlcTimeoutTx.RSMCMultiAddress,
		[]string{
			currTempAddressPrivateKey,
			currChannelAddressPrivateKey,
		},
		inputs,
		rdTransaction.OutputAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		rdTransaction.Amount,
		0,
		rdTransaction.Sequence,
		&htlcTimeoutTx.RSMCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	rdTransaction.Txid = txid
	rdTransaction.TransactionSignHex = hex
	rdTransaction.SignAt = time.Now()
	rdTransaction.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(rdTransaction)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return rdTransaction, nil
}

func createHtlcRDTxObj(owner string, channelInfo *dao.ChannelInfo, htlcTimeoutTx *dao.HTLCTimeoutTxForAAndExecutionForB, toAddress string,
	user *bean.User) (*dao.RevocableDeliveryTransaction, error) {
	htrd := &dao.RevocableDeliveryTransaction{}
	htrd.ChannelId = channelInfo.ChannelId
	htrd.CommitmentTxId = htlcTimeoutTx.Id
	htrd.PeerIdA = channelInfo.PeerIdA
	htrd.PeerIdB = channelInfo.PeerIdB
	htrd.PropertyId = htlcTimeoutTx.PropertyId
	htrd.Owner = owner
	htrd.RDType = 1

	//input
	htrd.InputTxid = htlcTimeoutTx.RSMCTxid
	htrd.InputVout = 0
	htrd.InputAmount = htlcTimeoutTx.RSMCOutAmount
	//output
	htrd.OutputAddress = toAddress
	htrd.Sequence = 1000
	htrd.Amount = htlcTimeoutTx.RSMCOutAmount

	htrd.CreateBy = user.PeerId
	htrd.CreateAt = time.Now()
	htrd.LastEditTime = time.Now()

	return htrd, nil
}

//创建hed1a  此交易要修改创建时机，等到bob拿到R的时候，再来创建，这个时候就需要广播交易（关闭通道），那么在很多情况下，其实是不用创建的
func htlcCreateExecutionDeliveryA(tx storm.Node, channelInfo dao.ChannelInfo, fundingTransaction dao.FundingTransaction,
	commitmentTxInfo dao.CommitmentTransaction, htlcRequestOpen bean.HtlcRequestOpen, owner string, hAndRInfo dao.HtlcRAndHInfo, R string) (hednA *dao.HTLCExecutionDeliveryA, err error) {

	if R != hAndRInfo.R {
		return nil, errors.New("error R")
	}

	//	alice 借道 bob 给carl 转账，给bob创建这个交易
	hednA = &dao.HTLCExecutionDeliveryA{}
	hednA.Owner = owner
	hednA.OutputAddress = channelInfo.AddressB
	hednA.OutAmount = commitmentTxInfo.AmountToHtlc

	inputs, err := getInputsForNextTxByParseTxHashVout(commitmentTxInfo.HtlcTxHash, commitmentTxInfo.HTLCMultiAddress, commitmentTxInfo.HTLCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		commitmentTxInfo.HTLCMultiAddress,
		[]string{
			htlcRequestOpen.CurrHtlcTempAddressPrivateKey,
			tempAddrPrivateKeyMap[channelInfo.PubKeyB],
		},
		inputs,
		hednA.OutputAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		hednA.OutAmount,
		0,
		0,
		&commitmentTxInfo.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	hednA.Txid = txid
	hednA.TxHash = hex
	hednA.CreateAt = time.Now()
	hednA.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(hednA)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return hednA, nil
}

func getHtlcLatestCommitmentTx(channelId bean.ChannelID, owner string) (commitmentTxInfo *dao.CommitmentTransaction, err error) {
	commitmentTxInfo = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", channelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_Htlc_GetR)).OrderBy("CreateAt").Reverse().First(commitmentTxInfo)
	return commitmentTxInfo, err
}

func createHtlcBRTx(owner string, channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTransaction, user *bean.User) (*dao.HTLCBreachRemedyTransaction, error) {
	hbr := &dao.HTLCBreachRemedyTransaction{}
	hbr.CommitmentTxId = commitmentTxInfo.Id
	hbr.PeerIdA = channelInfo.PeerIdA
	hbr.PeerIdB = channelInfo.PeerIdB
	hbr.ChannelId = channelInfo.ChannelId
	hbr.PropertyId = commitmentTxInfo.PropertyId
	hbr.Owner = owner

	//input
	hbr.InputTxid = commitmentTxInfo.HTLCTxid
	hbr.InputVout = 0
	hbr.InputAmount = commitmentTxInfo.AmountToHtlc
	//output
	hbr.Amount = commitmentTxInfo.AmountToHtlc

	hbr.CreateBy = user.PeerId
	hbr.CreateAt = time.Now()
	hbr.LastEditTime = time.Now()

	return hbr, nil
}
func createHtlcTimeoutBRTx(owner string, channelInfo *dao.ChannelInfo, txInfo dao.HTLCTimeoutTxForAAndExecutionForB, user *bean.User) (*dao.HTLCTimeoutBreachRemedyTransaction, error) {
	htbr := &dao.HTLCTimeoutBreachRemedyTransaction{}
	htbr.CommitmentTxId = txInfo.Id
	htbr.ChannelId = channelInfo.ChannelId
	htbr.PropertyId = txInfo.PropertyId
	htbr.Owner = owner
	htbr.InputHash = txInfo.RSMCTxHash
	//output
	htbr.Amount = txInfo.RSMCOutAmount

	htbr.CreateBy = user.PeerId
	htbr.CreateAt = time.Now()
	htbr.LastEditTime = time.Now()
	return htbr, nil
}
