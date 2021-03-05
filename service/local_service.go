package service

import (
	"errors"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/happierall/l"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/conn"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/tool"
	"log"
	"strings"
	"time"
)

var obdGlobalDB *storm.DB

func Start() {
	var err error
	obdGlobalDB, err = dao.DBService.GetGlobalDB()
	if err != nil {
		log.Println(err)
		return
	}
	checkInitConfig()
}

func checkInitConfig() {
	localConfig := &dao.ObdConfig{}
	_ = obdGlobalDB.Select().First(localConfig)
	if localConfig.Id == 0 {
		localConfig.InitHashCode = tool.GenerateInitHashCode()
		localConfig.AdminLoginToken = strings.ToLower(localConfig.InitHashCode)[0:8]
		_ = obdGlobalDB.Save(localConfig)
	}
	config.Init_node_chain_hash = localConfig.InitHashCode
	l.Log("admin login token:", l.Colorize(localConfig.AdminLoginToken, l.Red))
	log.Println("admin login token:", localConfig.AdminLoginToken)
}

func CheckIsAdmin(loginToken string) bool {
	if len(loginToken) == 0 {
		return false
	}

	localConfig := &dao.ObdConfig{}
	_ = obdGlobalDB.Select().First(localConfig)
	if tool.SignMsgWithMd5([]byte(localConfig.AdminLoginToken)) == loginToken {
		return true
	}
	return false
}

func UpdateAdminLoginToken(oldToken, newToken string) error {
	newToken = strings.TrimLeft(newToken, " ")
	newToken = strings.TrimRight(newToken, " ")
	if len(newToken) < 6 {
		return errors.New("token length must greater 6")
	}
	localConfig := &dao.ObdConfig{}
	_ = obdGlobalDB.Select().First(localConfig)
	if oldToken != localConfig.AdminLoginToken {
		return errors.New("wrong old token")
	}
	localConfig.AdminLoginToken = newToken
	_ = obdGlobalDB.Update(localConfig)
	return nil
}

func addRDTxToWaitDB(lastRevocableDeliveryTx *dao.RevocableDeliveryTransaction) (err error) {
	if lastRevocableDeliveryTx == nil || tool.CheckIsString(&lastRevocableDeliveryTx.TxHex) == false {
		return errors.New(enum.Tips_common_empty + "tx hex")
	}
	node := &dao.RDTxWaitingSend{}
	count, err := obdGlobalDB.Select(
		q.Eq("TransactionHex", lastRevocableDeliveryTx.TxHex)).
		Count(node)
	if count > 0 {
		return errors.New(enum.Tips_common_savedBefore)
	}
	node.TransactionHex = lastRevocableDeliveryTx.TxHex
	node.Type = 0
	node.IsEnable = true
	node.CreateAt = time.Now()
	err = obdGlobalDB.Save(node)
	if err != nil {
		return err
	}
	return nil
}

func addHT1aTxToWaitDB(htnx *dao.HTLCTimeoutTxForAAndExecutionForB, htrd *dao.RevocableDeliveryTransaction) error {
	node := &dao.RDTxWaitingSend{}
	count, err := obdGlobalDB.Select(
		q.Eq("TransactionHex", htnx.RSMCTxHex)).
		Count(node)
	if err == nil {
		return err
	}
	if count > 0 {
		return errors.New("already save")
	}
	node.TransactionHex = htnx.RSMCTxHex
	node.Type = 1
	node.IsEnable = true
	node.CreateAt = time.Now()
	node.HtnxIdAndHtnxRdId = make([]int, 2)
	node.HtnxIdAndHtnxRdId[0] = htnx.Id
	node.HtnxIdAndHtnxRdId[1] = htrd.Id
	err = obdGlobalDB.Save(node)
	if err != nil {
		return err
	}
	return nil
}

func addHTRD1aTxToWaitDB(htnxIdAndHtnxRdId []int) error {
	htnxId := htnxIdAndHtnxRdId[0]
	htrdId := htnxIdAndHtnxRdId[1]
	htnx := dao.HTLCTimeoutTxForAAndExecutionForB{}
	err := obdGlobalDB.One("Id", htnxId, &htnx)
	if err != nil {
		return err
	}

	htrd := dao.RevocableDeliveryTransaction{}
	err = obdGlobalDB.One("Id", htrdId, &htrd)
	if err != nil {
		return err
	}

	node := &dao.RDTxWaitingSend{}
	count, err := obdGlobalDB.Select(
		q.Eq("TransactionHex", htrd.TxHex)).
		Count(node)
	if err == nil {
		return err
	}
	if count > 0 {
		return errors.New("already save")
	}

	node.TransactionHex = htrd.TxHex
	node.Type = 0
	node.IsEnable = true
	node.CreateAt = time.Now()
	err = obdGlobalDB.Save(node)
	if err != nil {
		return err
	}

	htnx.CurrState = dao.TxInfoState_SendHex
	htnx.SendAt = time.Now()
	_ = obdGlobalDB.Update(htnx)

	return nil
}

//htlc timeout Delivery 1b
func addHTDnxTxToWaitDB(txInfo *dao.HTLCTimeoutDeliveryTxB) (err error) {
	node := &dao.RDTxWaitingSend{}
	count, err := obdGlobalDB.Select(
		q.Eq("TransactionHex", txInfo.TxHex)).
		Count(node)
	if err == nil {
		return err
	}
	if count > 0 {
		return errors.New("already save")
	}
	node.TransactionHex = txInfo.TxHex
	node.Type = 2
	node.IsEnable = true
	node.CreateAt = time.Now()
	err = obdGlobalDB.Save(node)
	if err != nil {
		return err
	}
	return nil
}

func sendRdTx() {
	log.Println("sendRdTx")
	var nodes []dao.RDTxWaitingSend
	err := obdGlobalDB.Select(q.Eq("IsEnable", true)).Find(&nodes)
	if err != nil {
		return
	}

	for _, node := range nodes {
		if tool.CheckIsString(&node.TransactionHex) {
			_, err := conn2tracker.SendRawTransaction(node.TransactionHex)
			if err == nil {
				if node.Type == 1 {
					_ = addHTRD1aTxToWaitDB(node.HtnxIdAndHtnxRdId)
				}
				_ = obdGlobalDB.UpdateField(&node, "IsEnable", false)
				_ = obdGlobalDB.UpdateField(&node, "FinishAt", time.Now())
			} else {
				if strings.Contains(err.Error(), "Code: -25,Msg: Missing inputs") {
					_ = obdGlobalDB.UpdateField(&node, "IsEnable", false)
					_ = obdGlobalDB.UpdateField(&node, "FinishAt", time.Now())
				}
			}
		}
	}
}
