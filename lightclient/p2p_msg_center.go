package lightclient

import (
	"encoding/json"
	"errors"
	"github.com/omnilaboratory/obd/agent"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/bean/enum"
	"github.com/omnilaboratory/obd/dao"
	"github.com/omnilaboratory/obd/service"
	"github.com/tidwall/gjson"
	"log"
)

func routerOfP2PNode(msg bean.RequestMessage, data string, client *Client) (retData string, isGoOn bool, retErr error) {
	defaultErr := errors.New("fail to deal msg in the inter node")
	status := false
	msgType := msg.Type
	switch msgType {
	case enum.MsgType_ChannelOpen_32:
		err := service.ChannelService.BeforeBobOpenChannelAtBobSide(data, client.User)
		if err == nil {
			status = true
			// when bob get the request for open channel
			if client.User.IsAdmin {
				msg.Data = data
				acceptOpenChannelInfo, err := agent.BeforeBobAcceptOpenChannel(&msg, client.User)
				if err == nil {
					signOpenChannelMsg := bean.RequestMessage{}
					marshal, _ := json.Marshal(acceptOpenChannelInfo)
					signOpenChannelMsg.Type = enum.MsgType_SendChannelAccept_33
					signOpenChannelMsg.RecipientNodePeerId = msg.SenderNodePeerId
					signOpenChannelMsg.RecipientUserPeerId = msg.SenderUserPeerId
					signOpenChannelMsg.Data = string(marshal)
					signedData, err := service.ChannelService.BobAcceptChannel(signOpenChannelMsg, client.User)
					if err == nil {
						signedOpenChannelMsg := bean.RequestMessage{}
						signedOpenChannelMsg.Type = enum.MsgType_ChannelAccept_33
						signedOpenChannelMsg.RecipientNodePeerId = signOpenChannelMsg.RecipientNodePeerId
						signedOpenChannelMsg.RecipientUserPeerId = signOpenChannelMsg.RecipientUserPeerId
						signedOpenChannelMsg.SenderNodePeerId = client.User.P2PLocalPeerId
						signedOpenChannelMsg.SenderUserPeerId = client.User.PeerId
						marshal, _ = json.Marshal(signedData)
						signedOpenChannelMsg.Data = string(marshal)
						_ = client.sendDataToP2PUser(signedOpenChannelMsg, true, signedOpenChannelMsg.Data)
						return "", false, nil
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}
		} else {
			defaultErr = err
		}
	case enum.MsgType_ChannelAccept_33:
		node, err := service.ChannelService.AfterBobAcceptChannelAtAliceSide(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingCreate_BtcFundingCreated_340:
		_, err := service.FundingTransactionService.BeforeSignBtcFundingCreatedAtBobSide(data, client.User)
		if err == nil {
			status = true
			if client.User.IsAdmin {
				signedRedeemTx, err := agent.BobSignFundBtcRedeemTx(data, client.User)
				if err == nil {
					newMsg := bean.RequestMessage{}
					newMsg.Type = enum.MsgType_FundingSign_SendBtcSign_350
					newMsg.SenderUserPeerId = client.User.PeerId
					newMsg.SenderNodePeerId = client.User.P2PLocalPeerId
					newMsg.RecipientUserPeerId = msg.SenderUserPeerId
					newMsg.RecipientNodePeerId = msg.SenderNodePeerId
					marshal, _ := json.Marshal(signedRedeemTx)
					newMsg.Data = string(marshal)
					signed, _, err := service.FundingTransactionService.FundingBtcTxSigned(newMsg, client.User)
					if err == nil {
						newMsg.Type = enum.MsgType_FundingSign_BtcSign_350
						marshal, _ := json.Marshal(signed)
						newMsg.Data = string(marshal)
						_ = client.sendDataToP2PUser(newMsg, true, newMsg.Data)
						return "", false, nil
					}
				}
			}
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingSign_BtcSign_350:
		_, err := service.FundingTransactionService.AfterBobSignBtcFundingAtAliceSide(data, client.User)
		if err == nil {
			status = true
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingCreate_AssetFundingCreated_34:
		_, err := service.FundingTransactionService.BeforeSignAssetFundingCreateAtBobSide(data, client.User)
		if err == nil {
			status = true
			// todo sign c1a
			if client.User.IsAdmin {
				signC1a, err := agent.BobSignC1a(data, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signC1a)
					signed, err := service.FundingTransactionService.AssetFundingSigned(string(marshal), client.User)
					if err == nil {
						marshal, _ := json.Marshal(signed)
						//	todo sign rd and br
						signedRdAndBrData, err := agent.BobSignRdAndBrOfC1a(string(marshal), client.User)
						if err == nil {
							marshal, _ := json.Marshal(signedRdAndBrData)
							aliceData, _, err := service.FundingTransactionService.OnBobSignedRDAndBR(string(marshal), client.User)
							if err == nil {
								newMsg := bean.RequestMessage{}
								newMsg.Type = enum.MsgType_FundingSign_AssetFundingSigned_35
								newMsg.SenderUserPeerId = client.User.PeerId
								newMsg.SenderNodePeerId = client.User.P2PLocalPeerId
								newMsg.RecipientUserPeerId = msg.SenderUserPeerId
								newMsg.RecipientNodePeerId = msg.SenderNodePeerId
								marshal, _ := json.Marshal(aliceData)
								newMsg.Data = string(marshal)
								_ = client.sendDataToP2PUser(newMsg, true, newMsg.Data)
								return "", false, nil
							}
						}
					}
				}
			}
		} else {
			defaultErr = err
		}
	case enum.MsgType_FundingSign_AssetFundingSigned_35:
		node, err := service.FundingTransactionService.OnGetBobSignedMsgAndSendDataToAlice(data, client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			if client.User.IsAdmin {
				signedData, _ := agent.AliceSignRdOfC1a(node, client.User)
				marshal, _ := json.Marshal(signedData)
				outData, err := service.FundingTransactionService.OnAliceSignedRdAtAliceSide(string(marshal), client.User)
				if err != nil {
					log.Println(err)
				}
				msg.Type = enum.MsgType_ClientSign_AssetFunding_AliceSignRD_1134
				marshal, _ = json.Marshal(outData)
				client.SendToMyself(msg.Type, true, string(marshal))
				return "", false, nil
			} else {
				return string(retData), true, nil
			}
		} else {
			defaultErr = err
		}
	case enum.MsgType_CommitmentTx_CommitmentTransactionCreated_351:
		node, err := service.CommitmentTxSignedService.BeforeBobSignCommitmentTransactionAtBobSide(data, client.User)
		if err == nil {
			status = true
			if client.User.IsAdmin {
				secondSignC2a, err := agent.RsmcBobSecondSignC2a(node, client.User)
				msg.Type = enum.MsgType_CommitmentTxSigned_SendRevokeAndAcknowledgeCommitmentTransaction_352
				msg.RecipientNodePeerId = msg.SenderNodePeerId
				msg.RecipientUserPeerId = msg.SenderUserPeerId
				msg.SenderUserPeerId = client.User.PeerId
				msg.SenderNodePeerId = client.User.P2PLocalPeerId
				marshal, _ := json.Marshal(secondSignC2a)
				msg.Data = string(marshal)
				transaction, _, err := service.CommitmentTxSignedService.RevokeAndAcknowledgeCommitmentTransaction(msg, client.User)
				if err == nil {
					signedDataForC2b, err := agent.RsmcBobFirstSignC2b(transaction, client.User)
					if err == nil {
						marshal, _ := json.Marshal(signedDataForC2b)
						_, retData, err := service.CommitmentTxSignedService.OnBobSignC2bTransactionAtBobSide(string(marshal), client.User)
						if err == nil {
							msg.Type = enum.MsgType_CommitmentTxSigned_ToAliceSign_352
							marshal, _ := json.Marshal(retData)
							msg.Data = string(marshal)
							err = client.sendDataToP2PUser(msg, true, msg.Data)
							return "", false, nil
						} else {
							log.Println(err)
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_ToAliceSign_352:
		node, needNoticeAlice, err := service.CommitmentTxService.OnGetBobC2bPartialSignTxAtAliceSide(msg, data, client.User)
		if err == nil {
			status = true
			if client.User.IsAdmin {
				signedData, err := agent.RsmcAliceSignC2b(node, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedData)
					needSignData, err := service.CommitmentTxService.OnAliceSignedC2bTxAtAliceSide(string(marshal), client.User)
					if err == nil {
						signedRdTxForC2b, err := agent.RsmcAliceSignRdOfC2b(needSignData, client.User)
						if err == nil {
							marshal, _ := json.Marshal(signedRdTxForC2b)
							_, bobRetData, _, err := service.CommitmentTxService.OnAliceSignedC2b_RDTxAtAliceSide(string(marshal), client.User)
							if err == nil {
								msg.Type = enum.MsgType_CommitmentTxSigned_SecondToBobSign_353
								msg.RecipientNodePeerId = msg.SenderNodePeerId
								msg.RecipientUserPeerId = msg.SenderUserPeerId
								msg.SenderUserPeerId = client.User.PeerId
								msg.SenderNodePeerId = client.User.P2PLocalPeerId
								marshal, _ := json.Marshal(bobRetData)
								msg.Data = string(marshal)
								err = client.sendDataToP2PUser(msg, true, msg.Data)
								return "", false, nil
							}
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}

			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		} else {
			if needNoticeAlice {
				client.SendToMyself(enum.MsgType_CommitmentTxSigned_RecvRevokeAndAcknowledgeCommitmentTransaction_352, true, string(err.Error()))
			}
		}
		defaultErr = err
	case enum.MsgType_CommitmentTxSigned_SecondToBobSign_353:
		node, err := service.CommitmentTxSignedService.OnGetAliceSignC2bTransactionAtBobSide(data, client.User)
		if err == nil {
			status = true
			if client.User.IsAdmin {
				signedData, err := agent.RsmcBobSignRdOfC2b(node, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedData)
					service.CommitmentTxSignedService.BobSignC2bRdAtBobSide(string(marshal), client.User)
					return "", false, nil
				}
			}
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_CloseChannelRequest_38:
		node, err := service.ChannelService.BeforeBobSignCloseChannelAtBobSide(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_CloseChannelSign_39:
		node, err := service.ChannelService.AfterBobSignCloseChannelAtAliceSide(data, *client.User)
		if err == nil {
			status = true
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_AddHTLC_40:
		node, err := service.HtlcForwardTxService.BeforeBobSignAddHtlcRequestAtBobSide_40(data, *client.User)
		if client.User.IsAdmin {
			signedData, err := agent.HtlcBobSignAddHtlcRequestAtBobSide_40(*node, client.User)
			if err == nil {
				marshal, _ := json.Marshal(signedData)
				txOfC3b, err := service.HtlcForwardTxService.BobSignedAddHtlcAtBobSide(string(marshal), *client.User)
				if err == nil {
					signedHtlcTxOfC3b, err := agent.HtlcBobSignC3b(txOfC3b, client.User)
					if err == nil {
						bytes, _ := json.Marshal(signedHtlcTxOfC3b)
						msg.Data = string(bytes)
						toAlice, _, err := service.HtlcForwardTxService.OnBobSignedC3bAtBobSide(msg, *client.User)
						if err == nil {
							marshal, _ := json.Marshal(toAlice)
							msg.Type = enum.MsgType_HTLC_NeedPayerSignC3b_41
							msg.RecipientNodePeerId = msg.SenderNodePeerId
							msg.RecipientUserPeerId = msg.SenderUserPeerId
							msg.SenderNodePeerId = client.User.P2PLocalPeerId
							msg.SenderUserPeerId = client.User.PeerId
							msg.Data = string(marshal)
							client.sendDataToP2PUser(msg, true, msg.Data)
							return "", false, nil
						} else {
							log.Println(err)
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}
		}
		if err == nil {
			retData, _ := json.Marshal(node)
			status = true
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_NeedPayerSignC3b_41:
		node, _, err := service.HtlcForwardTxService.AfterBobSignAddHtlcAtAliceSide_41(data, *client.User)
		if err == nil {
			status = true
			if client.User.IsAdmin {
				signedC3b, err := agent.HtlcAliceSignC3b(node, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedC3b)
					msg.Data = string(marshal)
					needSignHtlcPartTx, err := service.HtlcForwardTxService.OnAliceSignC3bAtAliceSide(msg, *client.User)
					if err == nil {
						signedData, err := agent.HtlcAliceSignedC3bSubTxAtAliceSide(needSignHtlcPartTx, client.User)
						if err == nil {
							marshal, _ := json.Marshal(signedData)
							msg.Data = string(marshal)
							_, bob, err := service.HtlcForwardTxService.OnAliceSignedC3bSubTxAtAliceSide(msg, *client.User)
							if err == nil {
								msg.Type = enum.MsgType_HTLC_PayeeCreateHTRD1a_42
								msg.RecipientNodePeerId = msg.SenderNodePeerId
								msg.RecipientUserPeerId = msg.SenderUserPeerId
								msg.SenderNodePeerId = client.User.P2PLocalPeerId
								msg.SenderUserPeerId = client.User.PeerId
								marshal, _ := json.Marshal(bob)
								msg.Data = string(marshal)
								client.sendDataToP2PUser(msg, true, msg.Data)
								return "", false, nil
							}
						} else {
							log.Println(err)
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}

			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayeeCreateHTRD1a_42:
		node, err := service.HtlcForwardTxService.OnGetNeedBobSignC3bSubTxAtBobSide(data, *client.User)
		if err == nil {
			status = true
			if client.User.IsAdmin {
				signedData, err := agent.HtlcBobSignedC3bSubTxAtBobSide(node, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedData)
					msg.Data = string(marshal)
					needSignHtRd, err := service.HtlcForwardTxService.OnBobSignedC3bSubTxAtBobSide(msg, *client.User)
					if err == nil {
						signedData, err := agent.HtlcBobSignHtRdAtBobSide(needSignHtRd, client.User)
						if err == nil {
							marshal, _ := json.Marshal(signedData)
							msg.Data = string(marshal)
							toAlice, toBob, err := service.HtlcForwardTxService.OnBobSignHtRdAtBobSide_42(msg.Data, *client.User)
							if err == nil {
								msg.Type = enum.MsgType_HTLC_PayerSignHTRD1a_43
								msg.RecipientNodePeerId = msg.SenderNodePeerId
								msg.RecipientUserPeerId = msg.SenderUserPeerId
								msg.SenderNodePeerId = client.User.P2PLocalPeerId
								msg.SenderUserPeerId = client.User.PeerId
								marshal, _ := json.Marshal(toAlice)
								msg.Data = string(marshal)
								client.sendDataToP2PUser(msg, true, msg.Data)

								afterH(toBob, *client, msg)

								return "", false, nil
							} else {
								log.Println(err)
							}
						} else {
							log.Println(err)
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}

			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_PayerSignHTRD1a_43:
		node, err := service.HtlcForwardTxService.OnGetHtrdTxDataFromBobAtAliceSide_43(data, *client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_VerifyR_45:
		responseData, err := service.HtlcBackwardTxService.OnGetHeSubTxDataAtAliceObdAtAliceSide(data, *client.User)
		if err == nil {
			if client.User.IsAdmin {
				signedData, err := agent.HtlcAliceSignedHeRdAtAliceSide(responseData, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedData)
					msg.Data = string(marshal)
					toAlice, toBob, err := service.HtlcBackwardTxService.OnAliceSignedHeRdAtAliceSide(msg, *client.User)
					if err == nil {
						msg.Type = enum.MsgType_HTLC_SendHerdHex_46
						msg.RecipientNodePeerId = msg.SenderNodePeerId
						msg.RecipientUserPeerId = msg.SenderUserPeerId
						msg.SenderNodePeerId = client.User.P2PLocalPeerId
						msg.SenderUserPeerId = client.User.PeerId
						marshal, _ := json.Marshal(toBob)
						msg.Data = string(marshal)
						client.sendDataToP2PUser(msg, true, msg.Data)
						// TODO 启动寻找上一个通道 回传R
						log.Println(toAlice)
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}
			retData, _ := json.Marshal(responseData)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_SendHerdHex_46:
		responseData, err := service.HtlcBackwardTxService.OnGetHeRdDataAtBobObd(data, *client.User)
		if err == nil {
			// TODO 关闭htlc
			if client.User.IsAdmin {
				closeHtlc, err := agent.HtlcRequestCloseHtlc(responseData, client.User)
				if err == nil {
					msg.Type = enum.MsgType_HTLC_Close_SendRequestCloseCurrTx_49
					msg.RecipientNodePeerId = msg.SenderNodePeerId
					msg.RecipientUserPeerId = msg.SenderUserPeerId
					msg.SenderNodePeerId = client.User.P2PLocalPeerId
					msg.SenderUserPeerId = client.User.PeerId
					marshal, _ := json.Marshal(closeHtlc)
					msg.Data = string(marshal)
					htlc, _, err := service.HtlcCloseTxService.RequestCloseHtlc(msg, *client.User)
					if err == nil {
						cxa, err := agent.HtlcCloseAliceSignedCxa(htlc, client.User)
						if err == nil {
							msg.Type = enum.MsgType_HTLC_Close_ClientSign_Alice_C4a_110
							marshal, _ := json.Marshal(cxa)
							msg.Data = string(marshal)
							_, toBob, err := service.HtlcCloseTxService.OnAliceSignedCxa(msg, *client.User)
							if err == nil {
								msg.Type = enum.MsgType_HTLC_Close_RequestCloseCurrTx_49
								marshal, _ := json.Marshal(toBob)
								msg.Data = string(marshal)
								err = client.sendDataToP2PUser(msg, true, msg.Data)
								return "", false, nil
							} else {
								log.Println(err)
							}
						} else {
							log.Println(err)
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}
			retData, _ := json.Marshal(responseData)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_Close_RequestCloseCurrTx_49:
		responseData, err := service.HtlcCloseTxService.OnObdOfBobGet49PData(data, *client.User)
		if err == nil {
			if client.User.IsAdmin {
				signedData, err := agent.HtlcCloseBobSignCloseHtlcRequest(responseData, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedData)
					msg.Data = string(marshal)
					needSignData, err := service.HtlcCloseTxService.OnBobSignCloseHtlcRequest(msg, *client.User)
					if err == nil {
						signedCxb, err := agent.HtlcCloseBobSignedCxb(needSignData, client.User)
						if err == nil {
							msg.Type = enum.MsgType_HTLC_Close_ClientSign_Bob_C4b_111
							msg.RecipientNodePeerId = msg.SenderNodePeerId
							msg.RecipientUserPeerId = msg.SenderUserPeerId
							marshal, _ := json.Marshal(signedCxb)
							msg.Data = string(marshal)
							toAlice, _, err := service.HtlcCloseTxService.OnBobSignedCxb(msg, *client.User)
							if err == nil {
								msg.Type = enum.MsgType_HTLC_CloseHtlcRequestSignBR_50
								msg.SenderNodePeerId = client.User.P2PLocalPeerId
								msg.SenderUserPeerId = client.User.PeerId
								marshal, _ := json.Marshal(toAlice)
								msg.Data = string(marshal)
								client.sendDataToP2PUser(msg, true, msg.Data)
								return "", false, nil
							} else {
								log.Println(err)
							}
						} else {
							log.Println(err)
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}
			retData, _ := json.Marshal(responseData)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_CloseHtlcRequestSignBR_50:
		responseData, _, err := service.HtlcCloseTxService.OnObdOfAliceGet50PData(data, *client.User)
		if err == nil {
			if client.User.IsAdmin {
				signedCxb, err := agent.HtlcCloseAliceSignedCxb(responseData, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedCxb)
					msg.Data = string(marshal)
					needSignData, err := service.HtlcCloseTxService.OnAliceSignedCxb(msg, *client.User)
					if err == nil {
						cxbBubTx, err := agent.HtlcCloseAliceSignedCxbBubTx(needSignData, client.User)
						if err == nil {
							marshal, _ := json.Marshal(cxbBubTx)
							msg.Data = string(marshal)
							_, toBob, err := service.HtlcCloseTxService.OnAliceSignedCxbBubTx(msg, *client.User)
							if err == nil {
								msg.Type = enum.MsgType_HTLC_CloseHtlcUpdateCnb_51
								msg.RecipientNodePeerId = msg.SenderNodePeerId
								msg.RecipientUserPeerId = msg.SenderUserPeerId
								msg.SenderNodePeerId = client.User.P2PLocalPeerId
								msg.SenderUserPeerId = client.User.PeerId
								marshal, _ := json.Marshal(toBob)
								msg.Data = string(marshal)
								client.sendDataToP2PUser(msg, true, msg.Data)
								return "", false, nil
							} else {
								log.Println(err)
							}
						} else {
							log.Println(err)
						}
					} else {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			}

			retData, _ := json.Marshal(responseData)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_HTLC_CloseHtlcUpdateCnb_51:
		node, err := service.HtlcCloseTxService.OnObdOfBobGet51PData(data, *client.User)
		if err == nil {
			if client.User.IsAdmin {
				signedCxb, err := agent.HtlcCloseBobSignedCxbSubTx(node, client.User)
				if err == nil {
					marshal, _ := json.Marshal(signedCxb)
					msg.Data = string(marshal)
					_, _ = service.HtlcCloseTxService.OnBobSignedCxbSubTx(msg, *client.User)
					return "", false, nil
				} else {
					log.Println(err)
				}
			}
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_Atomic_Swap_80:
		node, err := service.AtomicSwapService.BeforeSignAtomicSwapAtBobSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	case enum.MsgType_Atomic_SwapAccept_81:
		node, err := service.AtomicSwapService.BeforeSignAtomicSwapAcceptedAtAliceSide(data, client.User)
		if err == nil {
			retData, _ := json.Marshal(node)
			return string(retData), true, nil
		}
		defaultErr = err
	default:
		status = true
	}
	if status {
		defaultErr = nil
	}
	return "", true, defaultErr
}

func p2pMiddleNodeTransferData(msg *bean.RequestMessage, itemClient Client, data string, retData string) string {
	if msg.Type == enum.MsgType_ChannelOpen_32 {
		msg.Type = enum.MsgType_RecvChannelOpen_32
	}

	if msg.Type == enum.MsgType_ChannelAccept_33 {
		msg.Type = enum.MsgType_RecvChannelAccept_33
	}

	if msg.Type == enum.MsgType_FundingCreate_AssetFundingCreated_34 {
		msg.Type = enum.MsgType_FundingCreate_RecvAssetFundingCreated_34
	}

	if msg.Type == enum.MsgType_FundingSign_AssetFundingSigned_35 {
		msg.Type = enum.MsgType_FundingSign_RecvAssetFundingSigned_35
	}

	if msg.Type == enum.MsgType_FundingCreate_BtcFundingCreated_340 {
		msg.Type = enum.MsgType_FundingCreate_RecvBtcFundingCreated_340
	}

	if msg.Type == enum.MsgType_FundingSign_BtcSign_350 {
		msg.Type = enum.MsgType_FundingSign_RecvBtcSign_350
	}

	if msg.Type == enum.MsgType_CommitmentTx_CommitmentTransactionCreated_351 {
		msg.Type = enum.MsgType_CommitmentTx_RecvCommitmentTransactionCreated_351
	}

	if msg.Type == enum.MsgType_CloseChannelRequest_38 {
		msg.Type = enum.MsgType_RecvCloseChannelRequest_38
	}

	if msg.Type == enum.MsgType_CloseChannelSign_39 {
		msg.Type = enum.MsgType_RecvCloseChannelSign_39
	}

	if msg.Type == enum.MsgType_HTLC_AddHTLC_40 {
		msg.Type = enum.MsgType_HTLC_RecvAddHTLC_40
	}

	if msg.Type == enum.MsgType_CommitmentTxSigned_ToAliceSign_352 {
		//发给alice
		msg.Type = enum.MsgType_CommitmentTxSigned_RecvRevokeAndAcknowledgeCommitmentTransaction_352
		payerData := gjson.Parse(retData).String()
		data = payerData
	}

	//当353处理完成，就改成110353 推送给bob的客户端
	if msg.Type == enum.MsgType_CommitmentTxSigned_SecondToBobSign_353 {
		msg.Type = enum.MsgType_ClientSign_BobC2b_Rd_353
	}

	if msg.Type == enum.MsgType_HTLC_NeedPayerSignC3b_41 {
		msg.Type = enum.MsgType_HTLC_RecvAddHTLCSigned_41
	}

	//当42处理完成
	if msg.Type == enum.MsgType_HTLC_PayeeCreateHTRD1a_42 {
		msg.Type = enum.MsgType_HTLC_BobSignC3bSubTx_42
	}

	if msg.Type == enum.MsgType_HTLC_PayerSignHTRD1a_43 {
		msg.Type = enum.MsgType_HTLC_FinishTransferH_43
	}

	if msg.Type == enum.MsgType_HTLC_VerifyR_45 {
		msg.Type = enum.MsgType_HTLC_RecvVerifyR_45
	}

	//当47处理完成，发送48号协议给收款方
	if msg.Type == enum.MsgType_HTLC_SendHerdHex_46 {
		msg.Type = enum.MsgType_HTLC_RecvSignVerifyR_46
	}

	if msg.Type == enum.MsgType_HTLC_Close_RequestCloseCurrTx_49 {
		msg.Type = enum.MsgType_HTLC_Close_RecvRequestCloseCurrTx_49
	}

	if msg.Type == enum.MsgType_HTLC_CloseHtlcRequestSignBR_50 {
		msg.Type = enum.MsgType_HTLC_RecvCloseSigned_50
	}

	if msg.Type == enum.MsgType_HTLC_CloseHtlcUpdateCnb_51 {
		msg.Type = enum.MsgType_HTLC_Close_ClientSign_Bob_C4bSub_51
	}

	if msg.Type == enum.MsgType_Atomic_Swap_80 {
		msg.Type = enum.MsgType_Atomic_RecvSwap_80
	}

	if msg.Type == enum.MsgType_Atomic_SwapAccept_81 {
		msg.Type = enum.MsgType_Atomic_RecvSwapAccept_81
	}

	return data
}

func afterH(toBob interface{}, client Client, msg bean.RequestMessage) {
	r := agent.ROwnerGetHtlcRFromLocal(toBob, client.User)
	if r != "" {
		msg.Type = enum.MsgType_HTLC_SendVerifyR_45
		c3b := toBob.(*dao.CommitmentTransaction)
		sendR := bean.HtlcBobSendR{ChannelId: c3b.ChannelId, R: r}
		marshal, _ := json.Marshal(sendR)
		msg.Data = string(marshal)
		retData, err := service.HtlcBackwardTxService.SendRToPreviousNodeAtBobSide(msg, *client.User)
		if err == nil {
			signedData, err := agent.HtlcBobSignedHeRdAtBobSide(retData, client.User)
			if err == nil {
				marshal, _ := json.Marshal(signedData)
				msg.Data = string(marshal)
				toAlice, err := service.HtlcBackwardTxService.OnBobSignedHeRdAtBobSide(msg, *client.User)
				if err == nil {
					marshal, _ := json.Marshal(toAlice)
					msg.Type = enum.MsgType_HTLC_VerifyR_45
					msg.Data = string(marshal)
					client.sendDataToP2PUser(msg, true, msg.Data)
				} else {
					log.Println(err)
				}
			} else {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
	} else {
		//trigger send 40
		channelId, amount, msg := agent.InterUserGetNextNode(toBob, client.User)
		if len(channelId) > 0 {
			currNodeTx := toBob.(dao.CommitmentTransaction)
			msg.Type = enum.MsgType_HTLC_SendAddHTLC_40
			createHtlcTxForC3a := bean.CreateHtlcTxForC3a{}
			createHtlcTxForC3a.Amount = amount
			createHtlcTxForC3a.AmountToPayee = currNodeTx.HtlcAmountToPayee
			createHtlcTxForC3a.IsPayInvoice = true
			createHtlcTxForC3a.CltvExpiry = currNodeTx.HtlcCltvExpiry - 1
			createHtlcTxForC3a.H = currNodeTx.HtlcH
			createHtlcTxForC3a.Memo = currNodeTx.HtlcMemo
			createHtlcTxForC3a.RoutingPacket = currNodeTx.HtlcRoutingPacket
			marshal, _ := json.Marshal(createHtlcTxForC3a)
			msg.Data = string(marshal)
			client.htlcHModule(*msg)
		}
	}
}
