// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: proxy.proto

package toolrpc

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// LuckPkApiClient is the client API for LuckPkApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LuckPkApiClient interface {
	RegistTlsKey(ctx context.Context, in *RegistTlsKeyReq, opts ...grpc.CallOption) (*empty.Empty, error)
	HeartBeat(ctx context.Context, opts ...grpc.CallOption) (LuckPkApi_HeartBeatClient, error)
}

type luckPkApiClient struct {
	cc grpc.ClientConnInterface
}

func NewLuckPkApiClient(cc grpc.ClientConnInterface) LuckPkApiClient {
	return &luckPkApiClient{cc}
}

func (c *luckPkApiClient) RegistTlsKey(ctx context.Context, in *RegistTlsKeyReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/toolrpc.luckPkApi/RegistTlsKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *luckPkApiClient) HeartBeat(ctx context.Context, opts ...grpc.CallOption) (LuckPkApi_HeartBeatClient, error) {
	stream, err := c.cc.NewStream(ctx, &LuckPkApi_ServiceDesc.Streams[0], "/toolrpc.luckPkApi/HeartBeat", opts...)
	if err != nil {
		return nil, err
	}
	x := &luckPkApiHeartBeatClient{stream}
	return x, nil
}

type LuckPkApi_HeartBeatClient interface {
	Send(*empty.Empty) error
	CloseAndRecv() (*empty.Empty, error)
	grpc.ClientStream
}

type luckPkApiHeartBeatClient struct {
	grpc.ClientStream
}

func (x *luckPkApiHeartBeatClient) Send(m *empty.Empty) error {
	return x.ClientStream.SendMsg(m)
}

func (x *luckPkApiHeartBeatClient) CloseAndRecv() (*empty.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(empty.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// LuckPkApiServer is the server API for LuckPkApi service.
// All implementations must embed UnimplementedLuckPkApiServer
// for forward compatibility
type LuckPkApiServer interface {
	RegistTlsKey(context.Context, *RegistTlsKeyReq) (*empty.Empty, error)
	HeartBeat(LuckPkApi_HeartBeatServer) error
	mustEmbedUnimplementedLuckPkApiServer()
}

// UnimplementedLuckPkApiServer must be embedded to have forward compatible implementations.
type UnimplementedLuckPkApiServer struct {
}

func (UnimplementedLuckPkApiServer) RegistTlsKey(context.Context, *RegistTlsKeyReq) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegistTlsKey not implemented")
}
func (UnimplementedLuckPkApiServer) HeartBeat(LuckPkApi_HeartBeatServer) error {
	return status.Errorf(codes.Unimplemented, "method HeartBeat not implemented")
}
func (UnimplementedLuckPkApiServer) mustEmbedUnimplementedLuckPkApiServer() {}

// UnsafeLuckPkApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LuckPkApiServer will
// result in compilation errors.
type UnsafeLuckPkApiServer interface {
	mustEmbedUnimplementedLuckPkApiServer()
}

func RegisterLuckPkApiServer(s grpc.ServiceRegistrar, srv LuckPkApiServer) {
	s.RegisterService(&LuckPkApi_ServiceDesc, srv)
}

func _LuckPkApi_RegistTlsKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegistTlsKeyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LuckPkApiServer).RegistTlsKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/toolrpc.luckPkApi/RegistTlsKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LuckPkApiServer).RegistTlsKey(ctx, req.(*RegistTlsKeyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _LuckPkApi_HeartBeat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LuckPkApiServer).HeartBeat(&luckPkApiHeartBeatServer{stream})
}

type LuckPkApi_HeartBeatServer interface {
	SendAndClose(*empty.Empty) error
	Recv() (*empty.Empty, error)
	grpc.ServerStream
}

type luckPkApiHeartBeatServer struct {
	grpc.ServerStream
}

func (x *luckPkApiHeartBeatServer) SendAndClose(m *empty.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *luckPkApiHeartBeatServer) Recv() (*empty.Empty, error) {
	m := new(empty.Empty)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// LuckPkApi_ServiceDesc is the grpc.ServiceDesc for LuckPkApi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LuckPkApi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "toolrpc.luckPkApi",
	HandlerType: (*LuckPkApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegistTlsKey",
			Handler:    _LuckPkApi_RegistTlsKey_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "HeartBeat",
			Handler:       _LuckPkApi_HeartBeat_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "proxy.proto",
}