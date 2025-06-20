// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.2
// source: pkg/api/proto/get/get.proto

package get

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	GetService_Get_FullMethodName         = "/GetService/Get"
	GetService_GetRealTime_FullMethodName = "/GetService/GetRealTime"
)

// GetServiceClient is the client API for GetService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GetServiceClient interface {
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	GetRealTime(ctx context.Context, in *GetRealTimeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[GetResponse], error)
}

type getServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGetServiceClient(cc grpc.ClientConnInterface) GetServiceClient {
	return &getServiceClient{cc}
}

func (c *getServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, GetService_Get_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *getServiceClient) GetRealTime(ctx context.Context, in *GetRealTimeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[GetResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &GetService_ServiceDesc.Streams[0], GetService_GetRealTime_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[GetRealTimeRequest, GetResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GetService_GetRealTimeClient = grpc.ServerStreamingClient[GetResponse]

// GetServiceServer is the server API for GetService service.
// All implementations must embed UnimplementedGetServiceServer
// for forward compatibility.
type GetServiceServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	GetRealTime(*GetRealTimeRequest, grpc.ServerStreamingServer[GetResponse]) error
	mustEmbedUnimplementedGetServiceServer()
}

// UnimplementedGetServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedGetServiceServer struct{}

func (UnimplementedGetServiceServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedGetServiceServer) GetRealTime(*GetRealTimeRequest, grpc.ServerStreamingServer[GetResponse]) error {
	return status.Errorf(codes.Unimplemented, "method GetRealTime not implemented")
}
func (UnimplementedGetServiceServer) mustEmbedUnimplementedGetServiceServer() {}
func (UnimplementedGetServiceServer) testEmbeddedByValue()                    {}

// UnsafeGetServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GetServiceServer will
// result in compilation errors.
type UnsafeGetServiceServer interface {
	mustEmbedUnimplementedGetServiceServer()
}

func RegisterGetServiceServer(s grpc.ServiceRegistrar, srv GetServiceServer) {
	// If the following call pancis, it indicates UnimplementedGetServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GetService_ServiceDesc, srv)
}

func _GetService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GetServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GetService_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GetServiceServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GetService_GetRealTime_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetRealTimeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GetServiceServer).GetRealTime(m, &grpc.GenericServerStream[GetRealTimeRequest, GetResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type GetService_GetRealTimeServer = grpc.ServerStreamingServer[GetResponse]

// GetService_ServiceDesc is the grpc.ServiceDesc for GetService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GetService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "GetService",
	HandlerType: (*GetServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _GetService_Get_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetRealTime",
			Handler:       _GetService_GetRealTime_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/api/proto/get/get.proto",
}
