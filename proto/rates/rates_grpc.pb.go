// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/rates/rates.proto

package rates

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
	RatesService_GetRates_FullMethodName    = "/rates.RatesService/GetRates"
	RatesService_Healthcheck_FullMethodName = "/rates.RatesService/Healthcheck"
)

// RatesServiceClient is the client API for RatesService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// RatesService provides USDT exchange rates from Grinex
type RatesServiceClient interface {
	// GetRates retrieves current USDT exchange rates
	GetRates(ctx context.Context, in *GetRatesRequest, opts ...grpc.CallOption) (*GetRatesResponse, error)
	// Healthcheck checks service health status
	Healthcheck(ctx context.Context, in *HealthcheckRequest, opts ...grpc.CallOption) (*HealthcheckResponse, error)
}

type ratesServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRatesServiceClient(cc grpc.ClientConnInterface) RatesServiceClient {
	return &ratesServiceClient{cc}
}

func (c *ratesServiceClient) GetRates(ctx context.Context, in *GetRatesRequest, opts ...grpc.CallOption) (*GetRatesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetRatesResponse)
	err := c.cc.Invoke(ctx, RatesService_GetRates_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ratesServiceClient) Healthcheck(ctx context.Context, in *HealthcheckRequest, opts ...grpc.CallOption) (*HealthcheckResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HealthcheckResponse)
	err := c.cc.Invoke(ctx, RatesService_Healthcheck_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RatesServiceServer is the server API for RatesService service.
// All implementations must embed UnimplementedRatesServiceServer
// for forward compatibility.
//
// RatesService provides USDT exchange rates from Grinex
type RatesServiceServer interface {
	// GetRates retrieves current USDT exchange rates
	GetRates(context.Context, *GetRatesRequest) (*GetRatesResponse, error)
	// Healthcheck checks service health status
	Healthcheck(context.Context, *HealthcheckRequest) (*HealthcheckResponse, error)
	mustEmbedUnimplementedRatesServiceServer()
}

// UnimplementedRatesServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRatesServiceServer struct{}

func (UnimplementedRatesServiceServer) GetRates(context.Context, *GetRatesRequest) (*GetRatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRates not implemented")
}
func (UnimplementedRatesServiceServer) Healthcheck(context.Context, *HealthcheckRequest) (*HealthcheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Healthcheck not implemented")
}
func (UnimplementedRatesServiceServer) mustEmbedUnimplementedRatesServiceServer() {}
func (UnimplementedRatesServiceServer) testEmbeddedByValue()                      {}

// UnsafeRatesServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RatesServiceServer will
// result in compilation errors.
type UnsafeRatesServiceServer interface {
	mustEmbedUnimplementedRatesServiceServer()
}

func RegisterRatesServiceServer(s grpc.ServiceRegistrar, srv RatesServiceServer) {
	// If the following call pancis, it indicates UnimplementedRatesServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RatesService_ServiceDesc, srv)
}

func _RatesService_GetRates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRatesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RatesServiceServer).GetRates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RatesService_GetRates_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RatesServiceServer).GetRates(ctx, req.(*GetRatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RatesService_Healthcheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthcheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RatesServiceServer).Healthcheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RatesService_Healthcheck_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RatesServiceServer).Healthcheck(ctx, req.(*HealthcheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RatesService_ServiceDesc is the grpc.ServiceDesc for RatesService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RatesService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rates.RatesService",
	HandlerType: (*RatesServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRates",
			Handler:    _RatesService_GetRates_Handler,
		},
		{
			MethodName: "Healthcheck",
			Handler:    _RatesService_Healthcheck_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/rates/rates.proto",
}
