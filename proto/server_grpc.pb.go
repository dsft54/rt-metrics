// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.5
// source: proto/server.proto

package proto

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

// MetricClient is the client API for Metric service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricClient interface {
	AddMetric(ctx context.Context, in *AddMetricRequest, opts ...grpc.CallOption) (*AddMetricResponse, error)
	AddMetrics(ctx context.Context, in *AddMetricsRequest, opts ...grpc.CallOption) (*AddMetricsResponse, error)
}

type metricClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricClient(cc grpc.ClientConnInterface) MetricClient {
	return &metricClient{cc}
}

func (c *metricClient) AddMetric(ctx context.Context, in *AddMetricRequest, opts ...grpc.CallOption) (*AddMetricResponse, error) {
	out := new(AddMetricResponse)
	err := c.cc.Invoke(ctx, "/proto.Metric/AddMetric", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricClient) AddMetrics(ctx context.Context, in *AddMetricsRequest, opts ...grpc.CallOption) (*AddMetricsResponse, error) {
	out := new(AddMetricsResponse)
	err := c.cc.Invoke(ctx, "/proto.Metric/AddMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricServer is the server API for Metric service.
// All implementations must embed UnimplementedMetricServer
// for forward compatibility
type MetricServer interface {
	AddMetric(context.Context, *AddMetricRequest) (*AddMetricResponse, error)
	AddMetrics(context.Context, *AddMetricsRequest) (*AddMetricsResponse, error)
	mustEmbedUnimplementedMetricServer()
}

// UnimplementedMetricServer must be embedded to have forward compatible implementations.
type UnimplementedMetricServer struct {
}

func (UnimplementedMetricServer) AddMetric(context.Context, *AddMetricRequest) (*AddMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddMetric not implemented")
}
func (UnimplementedMetricServer) AddMetrics(context.Context, *AddMetricsRequest) (*AddMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddMetrics not implemented")
}
func (UnimplementedMetricServer) mustEmbedUnimplementedMetricServer() {}

// UnsafeMetricServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricServer will
// result in compilation errors.
type UnsafeMetricServer interface {
	mustEmbedUnimplementedMetricServer()
}

func RegisterMetricServer(s grpc.ServiceRegistrar, srv MetricServer) {
	s.RegisterService(&Metric_ServiceDesc, srv)
}

func _Metric_AddMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServer).AddMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Metric/AddMetric",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServer).AddMetric(ctx, req.(*AddMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metric_AddMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServer).AddMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Metric/AddMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServer).AddMetrics(ctx, req.(*AddMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Metric_ServiceDesc is the grpc.ServiceDesc for Metric service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Metric_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Metric",
	HandlerType: (*MetricServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddMetric",
			Handler:    _Metric_AddMetric_Handler,
		},
		{
			MethodName: "AddMetrics",
			Handler:    _Metric_AddMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/server.proto",
}