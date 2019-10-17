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

func getTwoChannelOfSingleHop(htlcCreateRandHInfo dao.HtlcCreateRandHInfo, channelAliceInfos []dao.ChannelInfo, channelCarlInfos []dao.ChannelInfo) (string, *dao.ChannelInfo, *dao.ChannelInfo) {
	for _, aliceChannel := range channelAliceInfos {
		if aliceChannel.PeerIdA == htlcCreateRandHInfo.SenderPeerId {
			bobPeerId := aliceChannel.PeerIdB
			carlChannel, err := getCarlChannelHasInterNodeBob(htlcCreateRandHInfo, aliceChannel, channelCarlInfos, aliceChannel.PeerIdA, bobPeerId)
			if err == nil {
				return bobPeerId, &aliceChannel, carlChannel
			}
		} else {
			bobPeerId := aliceChannel.PeerIdA
			carlChannel, err := getCarlChannelHasInterNodeBob(htlcCreateRandHInfo, aliceChannel, channelCarlInfos, aliceChannel.PeerIdB, bobPeerId)
			if err == nil {
				return bobPeerId, &aliceChannel, carlChannel
			}
		}
	}
	return "", nil, nil
}

func getCarlChannelHasInterNodeBob(htlcCreateRandHInfo dao.HtlcCreateRandHInfo, aliceChannel dao.ChannelInfo, channelCarlInfos []dao.ChannelInfo, alicePeerId, bobPeerId string) (*dao.ChannelInfo, error) {
	//whether bob is online
	if err := FindUserIsOnline(bobPeerId); err != nil {
		return nil, err
	}

	//alice and bob's channel, whether alice has enough money
	aliceCommitmentTxInfo, err := getLatestCommitmentTx(aliceChannel.ChannelId, alicePeerId)
	if err != nil {
		return nil, err
	}
	if aliceCommitmentTxInfo.AmountToRSMC < (htlcCreateRandHInfo.Amount + tool.GetHtlcFee()) {
		return nil, errors.New("channel not have enough money")
	}

	//bob and carl's channel,whether bob has enough money
	for _, carlChannel := range channelCarlInfos {
		if (carlChannel.PeerIdA == bobPeerId && carlChannel.PeerIdB == htlcCreateRandHInfo.RecipientPeerId) ||
			(carlChannel.PeerIdB == bobPeerId && carlChannel.PeerIdA == htlcCreateRandHInfo.RecipientPeerId) {
			commitmentTxInfo, err := getLatestCommitmentTx(carlChannel.ChannelId, bobPeerId)
			if err != nil {
				continue
			}
			if commitmentTxInfo.AmountToRSMC < htlcCreateRandHInfo.Amount {
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

func htlcAliceAbortLastCommitmentTx(tx storm.Node, channelInfo dao.ChannelInfo, user bean.User, fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen) error {
	owner := channelInfo.PeerIdA
	// 现在是创建PeerIdA方向，当前的操作者user可能对应PeerIdA，也可能对应PeerIdB
	// 默认假设当前操作者正好是PeerIdA，数据（公钥，私钥）的使用当前操作用户的
	var currOpUserIsPeerIdA = true
	// 如果不是，那PeerIdA对应的就是另一个方，数据（公钥，私钥）的使用就要使用另一方的数据了
	if user.PeerId == channelInfo.PeerIdB {
		currOpUserIsPeerIdA = false
	}

	//针对的是Cna
	var lastCommitmentATx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentATx)
	if err != nil {
		lastCommitmentATx = nil
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
			if currOpUserIsPeerIdA {
				lastTempAddressPrivateKey = htlcRequestOpen.LastTempAddressPrivateKey
			} else {
				// 如果当前操作用户是PeerIdB方，而我们现在正在处理概念中的Alice方，所以我们要取另一方的数据
				// 这些数据都是存在内存中，用完就删除
				lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentATx.RSMCTempAddressPubKey]
			}
			if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
				err = errors.New("fail to get the lastTempAddressPrivateKey")
				log.Println(err)
				return err
			}

			inputs, err := getInputsOfNextTxByParseTxHashVout(lastCommitmentATx.RSMCTxHash, lastCommitmentATx.RSMCMultiAddress, lastCommitmentATx.RSMCRedeemScript)
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
		}

		lastCommitmentATx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentATx)
		if err != nil {
			log.Println(err)
			return err
		}
		err = tx.Update(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

//概念的bob方结束上一次的Rsmc的交易
func htlcBobAbortLastCommitmentTx(tx storm.Node, channelInfo dao.ChannelInfo, user bean.User, fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen) error {
	owner := channelInfo.PeerIdB
	// 现在是创建PeerIdB方向，当前的操作者user可能对应PeerIdA，也可能对应PeerIdB
	// 默认假设当前操作者正好是PeerIdA，数据（公钥，私钥）的使用当前操作用户的
	var currOpUserIsPeerIdA = true
	// 如果不是，那PeerIdA对应的就是另一个方，数据（公钥，私钥）的使用就要使用另一方的数据了
	if user.PeerId == channelInfo.PeerIdB {
		currOpUserIsPeerIdA = false
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
		// 为概念的Alice方创建上一次交易的BR对象
		breachRemedyTransaction, err := createBRTx(channelInfo.PeerIdA, &channelInfo, lastCommitmentBTx, &user)
		if err != nil {
			log.Println(err)
			return err
		}

		//如果金额大于0
		if breachRemedyTransaction.Amount > 0 {
			lastTempAddressPrivateKey := ""
			if currOpUserIsPeerIdA {
				lastTempAddressPrivateKey = tempAddrPrivateKeyMap[lastCommitmentBTx.RSMCTempAddressPubKey]
			} else {
				lastTempAddressPrivateKey = htlcRequestOpen.LastTempAddressPrivateKey
			}
			if tool.CheckIsString(&lastTempAddressPrivateKey) == false {
				err = errors.New("fail to get the lastTempAddressPrivateKey")
				log.Println(err)
				return err
			}

			inputs, err := getInputsOfNextTxByParseTxHashVout(lastCommitmentBTx.RSMCTxHash, lastCommitmentBTx.RSMCMultiAddress, lastCommitmentBTx.RSMCRedeemScript)
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
				channelInfo.AddressB,
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
		}

		lastCommitmentBTx.CurrState = dao.TxInfoState_Abord
		lastRDTransaction.CurrState = dao.TxInfoState_Abord
		err = tx.Update(lastCommitmentBTx)
		if err != nil {
			log.Println(err)
			return err
		}
		err = tx.Update(lastRDTransaction)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// 创建PeerIdA方的htlc的承诺交易，rsmc的Rd
// 这里要做一个判断，作为这次交易的发起者，
// 如果PeerIdA是发起者，在这Cna的逻辑中创建HT1a和HED1a
// 如果PeerIdB是发起者，那么在Cna中就应该创建HTLC Time Delivery 1b(HED1b) 和HTLC Execution  1a(HE1b)
func htlcCreateAliceTxes(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen,
	singleHopTxBaseInfo dao.HtlcSingleHopTxBaseInfo, hAndRInfo dao.HtlcCreateRandHInfo) (*dao.CommitmentTransaction, error) {
	owner := channelInfo.PeerIdA
	currOperatorIsTxStarter := true
	if operator.PeerId == channelInfo.PeerIdB {
		currOperatorIsTxStarter = false
	}

	var lastCommitmentATx = &dao.CommitmentTransaction{}
	err := tx.Select(q.Eq("ChannelId", channelInfo.ChannelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(lastCommitmentATx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// create Cna tx
	commitmentTxInfo, err := htlcCreateCna(tx, channelInfo, operator, fundingTransaction, htlcRequestOpen, singleHopTxBaseInfo, hAndRInfo, currOperatorIsTxStarter, lastCommitmentATx, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create RDna tx
	_, err = htlcCreateRDOfRsmc(
		tx, channelInfo, operator, fundingTransaction, htlcRequestOpen,
		singleHopTxBaseInfo, currOperatorIsTxStarter, commitmentTxInfo, owner)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if currOperatorIsTxStarter { // 如果当前操作人是PeerIdA 创建HT1a
		// create ht1a
		htlcTimeoutTxA, err := createHtlcTimeoutTx(tx, owner, channelInfo, fundingTransaction, *commitmentTxInfo, htlcRequestOpen, operator)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(htlcTimeoutTxA)
		// 继续创建htrd
		rdTransaction, err := htlcCreateRD(tx, channelInfo, operator, fundingTransaction, htlcRequestOpen, singleHopTxBaseInfo, currOperatorIsTxStarter, htlcTimeoutTxA, owner)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println(rdTransaction)

	} else { // 如果当前操作人是PeerIdA 创建HED

	}

	return commitmentTxInfo, nil
}

func createHtlcTimeoutTx(tx storm.Node, owner string, channelInfo dao.ChannelInfo, fundingTransaction dao.FundingTransaction, commitmentTxInfo dao.CommitmentTransaction, htlcRequestOpen bean.HtlcRequestOpen, operator bean.User) (htlcTimeoutTxA *dao.HTLCTimeoutTxA, err error) {
	outputBean := commitmentOutputBean{}
	outputBean.AmountToRsmc = commitmentTxInfo.AmountToHtlc
	outputBean.RsmcTempPubKey = htlcRequestOpen.CurrHtlcTempAddressForHt1aPubKey
	outputBean.ToChannelAddress = channelInfo.PubKeyB
	htlcTimeoutTxA, err = createHtlcTimeoutTxObj(owner, channelInfo, commitmentTxInfo, outputBean, operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	inputs, err := getInputsOfNextTxByParseTxHashVout(commitmentTxInfo.HtlcTxHash, commitmentTxInfo.HTLCMultiAddress, commitmentTxInfo.HTLCMultiAddressScriptPubKey)
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
		htlcTimeoutTxA.RSMCMultiAddress,
		fundingTransaction.FunderAddress,
		fundingTransaction.PropertyId,
		htlcTimeoutTxA.RSMCOutAmount,
		0,
		htlcTimeoutTxA.Timeout,
		&commitmentTxInfo.HTLCRedeemScript)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htlcTimeoutTxA.RSMCTxid = txid
	htlcTimeoutTxA.RSMCTxHash = hex
	htlcTimeoutTxA.SignAt = time.Now()
	htlcTimeoutTxA.CurrState = dao.TxInfoState_CreateAndSign
	err = tx.Save(htlcTimeoutTxA)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return htlcTimeoutTxA, nil
}

func createHtlcTimeoutTxObj(owner string, channelInfo dao.ChannelInfo, commitmentTxInfo dao.CommitmentTransaction, outputBean commitmentOutputBean, user bean.User) (*dao.HTLCTimeoutTxA, error) {
	htlcTimeoutTxA := &dao.HTLCTimeoutTxA{}
	htlcTimeoutTxA.ChannelId = channelInfo.ChannelId
	htlcTimeoutTxA.PropertyId = commitmentTxInfo.PropertyId
	htlcTimeoutTxA.Owner = owner
	//input
	htlcTimeoutTxA.InputHex = commitmentTxInfo.HtlcTxHash

	//output to rsmc
	htlcTimeoutTxA.RSMCTempAddressPubKey = outputBean.RsmcTempPubKey
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{htlcTimeoutTxA.RSMCTempAddressPubKey, outputBean.ToChannelPubKey})
	if err != nil {
		return nil, err
	}
	htlcTimeoutTxA.RSMCMultiAddress = gjson.Get(multiAddr, "address").String()
	htlcTimeoutTxA.RSMCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	jsonData, err := rpcClient.GetAddressInfo(htlcTimeoutTxA.RSMCMultiAddress)
	if err != nil {
		return nil, err
	}
	htlcTimeoutTxA.RSMCMultiAddressScriptPubKey = gjson.Get(jsonData, "scriptPubKey").String()
	htlcTimeoutTxA.RSMCOutAmount = outputBean.AmountToRsmc
	htlcTimeoutTxA.CreateBy = user.PeerId
	htlcTimeoutTxA.CreateAt = time.Now()
	return htlcTimeoutTxA, nil
}

func htlcCreateCna(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen,
	singleHopTxBaseInfo dao.HtlcSingleHopTxBaseInfo, hAndRInfo dao.HtlcCreateRandHInfo,
	currOperatorIsTxStarter bool, lastCommitmentATx *dao.CommitmentTransaction, owner string) (*dao.CommitmentTransaction, error) {
	var outputBean = commitmentOutputBean{}
	if currOperatorIsTxStarter {
		outputBean.RsmcTempPubKey = htlcRequestOpen.CurrRsmcTempAddressPubKey
		outputBean.AmountToRsmc, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToRSMC).Sub(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
		outputBean.AmountToOther = lastCommitmentATx.AmountToOther
		outputBean.HtlcTempPubKey = htlcRequestOpen.CurrHtlcTempAddressPubKey
		outputBean.AmountToHtlc = hAndRInfo.Amount
	} else {
		outputBean.RsmcTempPubKey = singleHopTxBaseInfo.BobCurrRsmcTempPubKey
		outputBean.AmountToRsmc = lastCommitmentATx.AmountToRSMC
		outputBean.AmountToOther, _ = decimal.NewFromFloat(lastCommitmentATx.AmountToOther).Add(decimal.NewFromFloat(hAndRInfo.Amount)).Float64()
		outputBean.HtlcTempPubKey = singleHopTxBaseInfo.BobCurrHtlcTempForHt1bPubKey
		outputBean.AmountToHtlc = hAndRInfo.Amount
	}
	outputBean.ToChannelAddress = channelInfo.AddressB
	outputBean.ToChannelPubKey = channelInfo.PubKeyB

	commitmentTxInfo, err := createCommitmentTx(owner, &channelInfo, &fundingTransaction, outputBean, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commitmentTxInfo.TxType = 1

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

	commitmentTxInfo.LastCommitmentTxId = lastCommitmentATx.Id

	commitmentTxInfo.SignAt = time.Now()
	commitmentTxInfo.CurrState = dao.TxInfoState_CreateAndSign
	commitmentTxInfo.LastHash = ""
	commitmentTxInfo.CurrHash = ""
	if lastCommitmentATx != nil {
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

func htlcCreateRDOfRsmc(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen,
	singleHopTxBaseInfo dao.HtlcSingleHopTxBaseInfo, currOperatorIsTxStarter bool,
	commitmentTxInfo *dao.CommitmentTransaction, owner string) (*dao.RevocableDeliveryTransaction, error) {

	rdTransaction, err := createRDTx(owner, &channelInfo, commitmentTxInfo, channelInfo.AddressA, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if currOperatorIsTxStarter {
		currTempAddressPrivateKey = htlcRequestOpen.CurrRsmcTempAddressPrivateKey
	} else {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[singleHopTxBaseInfo.BobCurrRsmcTempPubKey]
	}

	inputs, err := getInputsOfNextTxByParseTxHashVout(commitmentTxInfo.RSMCTxHash, commitmentTxInfo.RSMCMultiAddress, commitmentTxInfo.RSMCMultiAddressScriptPubKey)
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

func htlcCreateRD(tx storm.Node, channelInfo dao.ChannelInfo, operator bean.User,
	fundingTransaction dao.FundingTransaction, htlcRequestOpen bean.HtlcRequestOpen,
	singleHopTxBaseInfo dao.HtlcSingleHopTxBaseInfo, currOperatorIsTxStarter bool,
	htlcTimeoutTxA *dao.HTLCTimeoutTxA, owner string) (*dao.RevocableDeliveryTransaction, error) {

	rdTransaction, err := createHtlcRDTxObj(owner, &channelInfo, htlcTimeoutTxA, channelInfo.AddressA, &operator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	currTempAddressPrivateKey := ""
	if currOperatorIsTxStarter {
		currTempAddressPrivateKey = htlcRequestOpen.CurrRsmcTempAddressPrivateKey
	} else {
		currTempAddressPrivateKey = tempAddrPrivateKeyMap[singleHopTxBaseInfo.BobCurrRsmcTempPubKey]
	}

	inputs, err := getInputsOfNextTxByParseTxHashVout(htlcTimeoutTxA.RSMCTxHash, htlcTimeoutTxA.RSMCMultiAddress, htlcTimeoutTxA.RSMCMultiAddressScriptPubKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txid, hex, err := rpcClient.OmniCreateAndSignRawTransactionForUnsendInputTx(
		htlcTimeoutTxA.RSMCMultiAddress,
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
		&htlcTimeoutTxA.RSMCRedeemScript)
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

func createHtlcRDTxObj(owner string, channelInfo *dao.ChannelInfo, htlcTimeoutTxA *dao.HTLCTimeoutTxA, toAddress string, user *bean.User) (*dao.RevocableDeliveryTransaction, error) {
	rda := &dao.RevocableDeliveryTransaction{}
	rda.CommitmentTxId = htlcTimeoutTxA.Id
	rda.PeerIdA = channelInfo.PeerIdA
	rda.PeerIdB = channelInfo.PeerIdB
	rda.ChannelId = channelInfo.ChannelId
	rda.PropertyId = htlcTimeoutTxA.PropertyId
	rda.Owner = owner

	//input
	rda.InputTxid = htlcTimeoutTxA.RSMCTxid
	rda.InputVout = 0
	rda.InputAmount = htlcTimeoutTxA.RSMCOutAmount
	//output
	rda.OutputAddress = toAddress
	rda.Sequence = 1000
	rda.Amount = htlcTimeoutTxA.RSMCOutAmount

	rda.CreateBy = user.PeerId
	rda.CreateAt = time.Now()
	rda.LastEditTime = time.Now()

	return rda, nil
}
