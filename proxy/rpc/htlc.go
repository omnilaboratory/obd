package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/proxy/pb"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"log"
)

func (s *RpcServer) AddInvoice(ctx context.Context, in *pb.Invoice) (*pb.AddInvoiceResponse, error) {
	log.Println("AddInvoice")
	_, err := checkLogin()
	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&in.CltvExpiry) == false {
		return nil, errors.New("wrong cltv_expiry")
	}
	if in.Value < 0 {
		return nil, errors.New("wrong value")
	}

	if in.PropertyId < 0 {
		return nil, errors.New("wrong property_id")
	}

	request := InvoiceInfo{
		PropertyId:  in.PropertyId,
		Amount:      in.Value,
		Description: in.Memo,
		ExpiryTime:  in.CltvExpiry,
		IsPrivate:   in.Private,
	}

	infoBytes, _ := json.Marshal(request)
	requestMessage := bean.RequestMessage{
		Type:             enum.MsgType_HTLC_Invoice_402,
		SenderNodePeerId: obcClient.User.P2PLocalPeerId,
		SenderUserPeerId: obcClient.User.PeerId,
		Data:             string(infoBytes)}
	_, dataBytes, status := obcClient.HtlcHModule(requestMessage)

	data := string(dataBytes)
	if status == false {
		return nil, errors.New(data)
	}
	dataMap := make(map[string]interface{})
	_ = json.Unmarshal(dataBytes, &dataMap)

	resp := &pb.AddInvoiceResponse{PaymentRequest: data}
	return resp, nil
}

func (s *RpcServer) ParseInvoice(ctx context.Context, in *pb.ParseInvoiceRequest) (*pb.ParseInvoiceResponse, error) {
	log.Println("ParseInvoice")
	_, err := checkLogin()
	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&in.PaymentRequest) == false {
		return nil, errors.New("wrong invoice")
	}

	request := ParseInvoice{
		Invoice: in.PaymentRequest,
	}

	infoBytes, _ := json.Marshal(request)
	requestMessage := bean.RequestMessage{
		Type:             enum.MsgType_HTLC_ParseInvoice_403,
		SenderNodePeerId: obcClient.User.P2PLocalPeerId,
		SenderUserPeerId: obcClient.User.PeerId,
		Data:             string(infoBytes)}
	_, dataBytes, status := obcClient.HtlcHModule(requestMessage)

	data := string(dataBytes)
	if status == false {
		return nil, errors.New(data)
	}

	dataMap := make(map[string]interface{})
	_ = json.Unmarshal(dataBytes, &dataMap)

	resp := &pb.ParseInvoiceResponse{
		Memo:                dataMap["description"].(string),
		PropertyId:          int64(dataMap["property_id"].(float64)),
		Value:               dataMap["amount"].(float64),
		CltvExpiry:          dataMap["expiry_time"].(string),
		H:                   dataMap["h"].(string),
		Private:             dataMap["is_private"].(bool),
		RecipientNodePeerId: dataMap["recipient_node_peer_id"].(string),
		RecipientUserPeerId: dataMap["recipient_user_peer_id"].(string),
	}
	return resp, nil
}
func (s *RpcServer) ListInvoices(ctx context.Context, in *pb.ListInvoiceRequest) (*pb.ListInvoiceResponse, error) {
	log.Println("ParseInvoice")
	_, err := checkLogin()
	if err != nil {
		return nil, err
	}

	infoBytes, _ := json.Marshal(in)

	data, err := service.HtlcQueryTxManager.ListInvoices(string(infoBytes), *obcClient.User)
	if err != nil {
		return nil, err
	}

	resp := &pb.ListInvoiceResponse{
		FirstIndexOffset: uint64(data["first_index_offset"].(int64)),
		LastIndexOffset:  uint64(data["last_index_offset"].(int64)),
	}
	invoices := data["invoices"].([]dao.InvoiceInfo)
	for _, item := range invoices {
		invoice := &pb.Invoice{}
		invoice.PropertyId = item.Detail.PropertyId
		invoice.Value = item.Detail.Amount
		invoice.Private = item.Detail.IsPrivate
		invoice.CltvExpiry = item.Detail.ExpiryTime.String()
		invoice.Memo = item.Detail.Description
		invoice.PaymentRequest = item.Invoice
		resp.Invoices = append(resp.Invoices, invoice)
	}
	return resp, nil
}

func (s *RpcServer) SendPayment(ctx context.Context, in *pb.SendRequest) (*pb.SendResponse, error) {
	log.Println("SendPayment")
	_, err := checkLogin()
	if err != nil {
		return nil, err
	}

	if tool.CheckIsString(&in.PaymentRequest) == false {
		return nil, errors.New("wrong payment_request")
	}

	request := bean.HtlcRequestFindPath{
		Invoice: in.PaymentRequest,
	}
	infoBytes, _ := json.Marshal(request)
	requestMessage := bean.RequestMessage{
		Type:             enum.MsgType_HTLC_FindPath_401,
		SenderNodePeerId: obcClient.User.P2PLocalPeerId,
		SenderUserPeerId: obcClient.User.PeerId,
		Data:             string(infoBytes)}
	obcClient.HtlcHModule(requestMessage)

	if obcClient.GrpcChan == nil {
		obcClient.GrpcChan = make(chan []byte)
	}

	message := <-obcClient.GrpcChan

	close(obcClient.GrpcChan)
	obcClient.GrpcChan = nil

	replyMessage := bean.ReplyMessage{}
	_ = json.Unmarshal(message, &replyMessage)
	if replyMessage.Status == false {
		return nil, errors.New(replyMessage.Result.(string))
	}

	dataResult := replyMessage.Result.(map[string]interface{})
	resp := &pb.SendResponse{
		PaymentHash:     dataResult["htlc_routing_packet"].(string),
		PaymentPreimage: dataResult["htlc_h"].(string),
		AmountToRsmc:    dataResult["amount_to_rsmc"].(float64),
		AmountToHtlc:    dataResult["amount_to_htlc"].(float64),
	}
	if dataResult["amount_to_counterparty"] != nil {
		resp.AmountToCounterparty = dataResult["amount_to_counterparty"].(float64)
	}

	return resp, nil
}
