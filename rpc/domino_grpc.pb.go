// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package rpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GameServiceClient is the client API for GameService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GameServiceClient interface {
	// join game request
	Join(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (*JoinResponse, error)
	// send/receive update about game states
	Update(ctx context.Context, opts ...grpc.CallOption) (GameService_UpdateClient, error)
}

type gameServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGameServiceClient(cc grpc.ClientConnInterface) GameServiceClient {
	return &gameServiceClient{cc}
}

func (c *gameServiceClient) Join(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (*JoinResponse, error) {
	out := new(JoinResponse)
	err := c.cc.Invoke(ctx, "/domino.GameService/Join", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gameServiceClient) Update(ctx context.Context, opts ...grpc.CallOption) (GameService_UpdateClient, error) {
	stream, err := c.cc.NewStream(ctx, &GameService_ServiceDesc.Streams[0], "/domino.GameService/Update", opts...)
	if err != nil {
		return nil, err
	}
	x := &gameServiceUpdateClient{stream}
	return x, nil
}

type GameService_UpdateClient interface {
	Send(*GameUpdate) error
	Recv() (*GameUpdate, error)
	grpc.ClientStream
}

type gameServiceUpdateClient struct {
	grpc.ClientStream
}

func (x *gameServiceUpdateClient) Send(m *GameUpdate) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gameServiceUpdateClient) Recv() (*GameUpdate, error) {
	m := new(GameUpdate)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GameServiceServer is the server API for GameService service.
// All implementations must embed UnimplementedGameServiceServer
// for forward compatibility
type GameServiceServer interface {
	// join game request
	Join(context.Context, *JoinRequest) (*JoinResponse, error)
	// send/receive update about game states
	Update(GameService_UpdateServer) error
	mustEmbedUnimplementedGameServiceServer()
}

// UnimplementedGameServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGameServiceServer struct {
}

func (UnimplementedGameServiceServer) Join(context.Context, *JoinRequest) (*JoinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Join not implemented")
}
func (UnimplementedGameServiceServer) Update(GameService_UpdateServer) error {
	return status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedGameServiceServer) mustEmbedUnimplementedGameServiceServer() {}

// UnsafeGameServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GameServiceServer will
// result in compilation errors.
type UnsafeGameServiceServer interface {
	mustEmbedUnimplementedGameServiceServer()
}

func RegisterGameServiceServer(s grpc.ServiceRegistrar, srv GameServiceServer) {
	s.RegisterService(&GameService_ServiceDesc, srv)
}

func _GameService_Join_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GameServiceServer).Join(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/domino.GameService/Join",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GameServiceServer).Join(ctx, req.(*JoinRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GameService_Update_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GameServiceServer).Update(&gameServiceUpdateServer{stream})
}

type GameService_UpdateServer interface {
	Send(*GameUpdate) error
	Recv() (*GameUpdate, error)
	grpc.ServerStream
}

type gameServiceUpdateServer struct {
	grpc.ServerStream
}

func (x *gameServiceUpdateServer) Send(m *GameUpdate) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gameServiceUpdateServer) Recv() (*GameUpdate, error) {
	m := new(GameUpdate)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GameService_ServiceDesc is the grpc.ServiceDesc for GameService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GameService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "domino.GameService",
	HandlerType: (*GameServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Join",
			Handler:    _GameService_Join_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Update",
			Handler:       _GameService_Update_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "domino.proto",
}