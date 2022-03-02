package inforpc

import (
	"context"
	"github.com/omnilaboratory/obd/tool"
	cfg "github.com/omnilaboratory/obd/tracker/config"
	"github.com/omnilaboratory/obd/tracker/tkrpc"
	"time"
)

var Orm=cfg.Orm
type ImpInfoServer struct {
	tkrpc.UnimplementedInfoTrackerServer
}
func (s *ImpInfoServer) HeartBeat(heartBeatServer tkrpc.InfoTracker_HeartBeatServer) error {
	var ninfo *tkrpc.NodeInfo
	defer func() {
		if ninfo != nil {
			ninfo.IsOnline = 2
			Orm.Save(ninfo)
		}
	}()
	for {
		req, err := heartBeatServer.Recv()
		if err != nil {
			break
		}
		if ninfo == nil {
			err = Orm.FirstOrCreate(ninfo, tkrpc.NodeInfo{NodeId: req.NodeId, P2PAddress: req.P2PAddress}).Error
			if err != nil {
				break
			}
		}
		ninfo.HearBeatCounter += 1
		ninfo.LastHearBeat = time.Now().Unix()
		ninfo.IsOnline = 1
		Orm.Save(ninfo)
	}
	return nil
}

func (s *ImpInfoServer) UpdateNodeInfo(ctx context.Context, req *tkrpc.UpdateNodeInfoReq) (*tkrpc.EmptyRes, error) {
	panic("implement me")
}

func (s *ImpInfoServer) UpdateUserInfo(ctx context.Context, req *tkrpc.UpdateUserInfoReq) (*tkrpc.EmptyRes, error) {
	uinfo := new(tkrpc.UserInfo)
	err := Orm.FirstOrCreate(uinfo, tkrpc.UserInfo{UserId: req.UserId, NodeId: req.NodeId}).Assign(tkrpc.UserInfo{IsOnline: req.IsOnline}).Error
	return &tkrpc.EmptyRes{}, err
}

func (s *ImpInfoServer) UpdateUserInfos(ctx context.Context, req *tkrpc.UpdateUserInfosReq) (*tkrpc.EmptyRes, error) {
	var err error
	for _, uinfo := range req.UpdateUserInfoReqs {
		err = Orm.FirstOrCreate(uinfo, tkrpc.UserInfo{UserId: uinfo.UserId, NodeId: uinfo.NodeId}).Assign(tkrpc.UserInfo{IsOnline: uinfo.IsOnline}).Error
		if err != nil {
			break
		}
	}
	return &tkrpc.EmptyRes{}, err
}

func (s *ImpInfoServer) GetUserInfo(ctx context.Context, req *tkrpc.UpdateUserInfoReq) (uinfo *tkrpc.UserInfo, err error) {
	err=Orm.First(uinfo,tkrpc.UserInfo{UserId: req.UserId,NodeId: req.NodeId}).Error
	return
}

func (s *ImpInfoServer) GetUserInfos(ctx context.Context, req *tkrpc.ListReq) (list *tkrpc.UserInfosRes, err error) {
	list=new(tkrpc.UserInfosRes)
	Orm.Model(tkrpc.UserInfo{}).Count(&list.Count)
	err=Orm.Limit(req.Limit()).Offset(req.Offset()).Order(req.SortStr()).Find(&list.Results).Error
	return
}

func (s *ImpInfoServer) GetNodes(ctx context.Context, req *tkrpc.ListReq) (list *tkrpc.NodeInfosRes, err error) {
	list=new(tkrpc.NodeInfosRes)
	Orm.Model(tkrpc.NodeInfo{}).Count(&list.Count)
	err=Orm.Limit(req.Limit()).Offset(req.Offset()).Order(req.SortStr()).Find(&list.Results).Error
	return
}

func (s *ImpInfoServer) UpdateChannelInfo(ctx context.Context, info *tkrpc.ChannelInfo) (*tkrpc.ChannelInfo, error) {
	panic("implement me")
}

func (s *ImpInfoServer) UpdateChannelInfos(ctx context.Context, info *tkrpc.UpdateChannelInfosReq) (*tkrpc.EmptyRes, error) {
	for _, item := range info.ChannelInfos {
		err:=item.ValidteClientData()
		if err != nil {
			return nil,err
		}
		channelInfo:=new(tkrpc.ChannelInfo)
		Orm.First(channelInfo,tkrpc.ChannelInfo{ChannelId: item.ChannelId})
		if channelInfo.Id==0{
			channelInfo.ChannelId = item.ChannelId
			channelInfo.PropertyId = item.PropertyId
			channelInfo.CurrState = item.CurrState
			channelInfo.PeerIda = item.PeerIda
			channelInfo.PeerIdb = item.PeerIdb
			channelInfo.AmountA = item.AmountA
			channelInfo.AmountB = item.AmountB
		}else{
			channelInfo.PropertyId = item.PropertyId
			if channelInfo.CurrState != tkrpc.ChannelState_Close {
				channelInfo.CurrState = item.CurrState
				channelInfo.PeerIda = item.PeerIda
				channelInfo.PeerIdb = item.PeerIdb
				channelInfo.AmountA = item.AmountA
				channelInfo.AmountB = item.AmountB
			}
		}
		if item.IsAlice {
			channelInfo.ObdNodeIda = item.NodeId
		} else {
			channelInfo.ObdNodeIdb = item.NodeId
		}
		err=Orm.Save(channelInfo).Error
		if err != nil {
			return &tkrpc.EmptyRes{}, err
		}
	}
	return &tkrpc.EmptyRes{},nil
}

func (s *ImpInfoServer) GetChannelInfo(ctx context.Context, filter *tkrpc.SimpleFilter) ( info *tkrpc.ChannelInfo, err error) {
	err=Orm.First(info,tkrpc.ChannelInfo{ChannelId: filter.ChannelId}).Error
	return
}

func (s *ImpInfoServer) GetChannels(ctx context.Context, req *tkrpc.ListReq) (res *tkrpc.NodeInfosRes, err error) {
	res=new(tkrpc.NodeInfosRes)
	Orm.Count(&res.Count)
	err=Orm.Offset(req.Offset()).Limit(req.Limit()).Find(&res.Results).Error
	return
}

func (s *ImpInfoServer) UpdateHtlcInfo(ctx context.Context, req *tkrpc.HtlcInfo) (*tkrpc.HtlcInfo, error) {
	err := req.ValidteClientData()
	if err != nil {
		return nil, err
	}
	info := new(tkrpc.HtlcInfo)
	err = Orm.FirstOrCreate(info, tkrpc.HtlcInfo{Path: req.Path, H: req.H}).Error
	if err == nil {
		if tool.CheckIsString(&req.R) == false && tool.CheckIsString(&req.R) {
			info.R = req.R
		}
		info.DirectionFlag = req.DirectionFlag
		info.CurrChannelId = req.CurrChannelId
		err = Orm.Save(info).Error
	}
	return info, err
}

func (s *ImpInfoServer) HtlcGetPath(ctx context.Context, req *tkrpc.HtlcGetPathReq) (*tkrpc.HtlcGetPathRes, error) {
	panic("implement me")
}
func (s *ImpInfoServer) GetHtlcInfo(ctx context.Context, req *tkrpc.GetHtlcInfoReq) (res *tkrpc.HtlcInfo, err error) {
	err = Orm.First(res, req).Error
	return
}

//func (s *ImpInfoServer) mustEmbedUnimplementedInfoTrackerServer() {
//	panic("implement me")
//}

/*
//every HeartBeat will set nodeInfo.isonline=true; every HeartBeat disconnect will set  nodeInfo.isonline=false;
	HeartBeat(ctx context.Context, opts ...grpc.CallOption) (InfoTracker_HeartBeatClient, error)
	//map to old func Logout Login
	UpdateNodeInfo(ctx context.Context, in *UpdateNodeInfoReq, opts ...grpc.CallOption) (*EmptyRes, error)
	//map to old func userLogout userLogin
	UpdateUserInfo(ctx context.Context, in *UpdateUserInfoReq, opts ...grpc.CallOption) (*EmptyRes, error)
	//map to old func updateUsers
	UpdateUserInfos(ctx context.Context, in *UpdateUserInfosReq, opts ...grpc.CallOption) (*EmptyRes, error)
	//map to old func: GetUserState , GetUserP2pNodeId
	//Request:SetUserInfoReq{user_id,node_Id}
	//old GetUserP2pNodeId use request:SetUserInfoReq{user_id} ;  may not work when user login on multi node
	GetUserInfo(ctx context.Context, in *UpdateUserInfoReq, opts ...grpc.CallOption) (*UserInfo, error)
	//Map to old function GetAllUsers
	GetUserInfos(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (*UserInfosRes, error)
	//map old function GetAllObdNodes
	GetNodes(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (*NodeInfosRes, error)
	UpdateChannelInfo(ctx context.Context, in *ChannelInfo, opts ...grpc.CallOption) (*ChannelInfo, error)
	//map to old ChannelService.GetChannelState
	GetChannelInfo(ctx context.Context, in *SimpleFilter, opts ...grpc.CallOption) (*ChannelInfo, error)
	GetChannels(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (*NodeInfosRes, error)
	// map old function updateHtlcInfo
	UpdateHtlcInfo(ctx context.Context, in *HtlcInfo, opts ...grpc.CallOption) (*HtlcInfo, error)
	//map old function getPath
	HtlcGetPath(ctx context.Context, in *HtlcGetPathReq, opts ...grpc.CallOption) (*HtlcGetPathRes, error)
	//map old function GetHtlcCurrState
	GetHtlcInfo(ctx context.Context, in *GetHtlcInfoReq, opts ...grpc.CallOption) (*HtlcInfo, error)
*/

func getUserState(userId string) bool {
	uinfo:=new(tkrpc.UserInfo)
	Orm.First(uinfo,tkrpc.UserInfo{UserId: userId})
	return uinfo.IsOnline==1
}