// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: volc.proto

package grpc

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

// BackendClient is the client API for Backend service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BackendClient interface {
	WatchUpload(ctx context.Context, in *WatchUploadRequest, opts ...grpc.CallOption) (Backend_WatchUploadClient, error)
}

type backendClient struct {
	cc grpc.ClientConnInterface
}

func NewBackendClient(cc grpc.ClientConnInterface) BackendClient {
	return &backendClient{cc}
}

func (c *backendClient) WatchUpload(ctx context.Context, in *WatchUploadRequest, opts ...grpc.CallOption) (Backend_WatchUploadClient, error) {
	stream, err := c.cc.NewStream(ctx, &Backend_ServiceDesc.Streams[0], "/backend/WatchUpload", opts...)
	if err != nil {
		return nil, err
	}
	x := &backendWatchUploadClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Backend_WatchUploadClient interface {
	Recv() (*WatchUploadResponse, error)
	grpc.ClientStream
}

type backendWatchUploadClient struct {
	grpc.ClientStream
}

func (x *backendWatchUploadClient) Recv() (*WatchUploadResponse, error) {
	m := new(WatchUploadResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BackendServer is the server API for Backend service.
// All implementations must embed UnimplementedBackendServer
// for forward compatibility
type BackendServer interface {
	WatchUpload(*WatchUploadRequest, Backend_WatchUploadServer) error
	mustEmbedUnimplementedBackendServer()
}

// UnimplementedBackendServer must be embedded to have forward compatible implementations.
type UnimplementedBackendServer struct {
}

func (UnimplementedBackendServer) WatchUpload(*WatchUploadRequest, Backend_WatchUploadServer) error {
	return status.Errorf(codes.Unimplemented, "method WatchUpload not implemented")
}
func (UnimplementedBackendServer) mustEmbedUnimplementedBackendServer() {}

// UnsafeBackendServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BackendServer will
// result in compilation errors.
type UnsafeBackendServer interface {
	mustEmbedUnimplementedBackendServer()
}

func RegisterBackendServer(s grpc.ServiceRegistrar, srv BackendServer) {
	s.RegisterService(&Backend_ServiceDesc, srv)
}

func _Backend_WatchUpload_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(WatchUploadRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BackendServer).WatchUpload(m, &backendWatchUploadServer{stream})
}

type Backend_WatchUploadServer interface {
	Send(*WatchUploadResponse) error
	grpc.ServerStream
}

type backendWatchUploadServer struct {
	grpc.ServerStream
}

func (x *backendWatchUploadServer) Send(m *WatchUploadResponse) error {
	return x.ServerStream.SendMsg(m)
}

// Backend_ServiceDesc is the grpc.ServiceDesc for Backend service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Backend_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "backend",
	HandlerType: (*BackendServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "WatchUpload",
			Handler:       _Backend_WatchUpload_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "volc.proto",
}
