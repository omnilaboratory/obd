package rpc

import (
	"context"
	"errors"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/proxy/pb"
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

	sendMsgToObd(request, "", "", enum.MsgType_HTLC_Invoice_402)

	data := <-addInvoiceChan

	if data.Status == false {
		return nil, errors.New(data.Result.(string))
	}
	log.Println(data.Result)
	resp := &pb.AddInvoiceResponse{PaymentRequest: data.Result.(string)}
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

	sendMsgToObd(request, "", "", enum.MsgType_HTLC_ParseInvoice_403)

	data := <-onceRequestChan

	if data.Status == false {
		return nil, errors.New(data.Result.(string))
	}
	log.Println(data.Result)
	dataResult := data.Result.(map[string]interface{})
	resp := &pb.ParseInvoiceResponse{
		Memo:                dataResult["description"].(string),
		PropertyId:          int64(dataResult["property_id"].(float64)),
		Value:               dataResult["amount"].(float64),
		CltvExpiry:          dataResult["expiry_time"].(string),
		H:                   dataResult["h"].(string),
		Private:             dataResult["is_private"].(bool),
		RecipientNodePeerId: dataResult["recipient_node_peer_id"].(string),
		RecipientUserPeerId: dataResult["recipient_user_peer_id"].(string),
	}
	return resp, nil
}

func (s *RpcServer) SendPayment(ctx context.Context, in *pb.SendPaymentRequest) (*pb.PaymentResp, error) {
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

	sendMsgToObd(request, "", "", enum.MsgType_HTLC_FindPath_401)

	for {
		data := <-payInvoiceChan
		if data.Status == false {
			return nil, errors.New(data.Result.(string))
		}
		if data.Type == enum.MsgType_HTLC_FinishTransferH_43 {
			log.Println(data.Result)
			dataResult := data.Result.(map[string]interface{})
			resp := &pb.PaymentResp{
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
	}
}
