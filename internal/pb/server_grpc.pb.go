// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: server.proto

package pb

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MetricCollectorClient is the client API for MetricCollector service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricCollectorClient interface {
	UpdatesAllMetricsJSON(ctx context.Context, in *RequestListMetrics, opts ...grpc.CallOption) (*emptypb.Empty, error)
	UpdateOneMetricsJSON(ctx context.Context, in *RequestMetrics, opts ...grpc.CallOption) (*emptypb.Empty, error)
	UpdateOneMetrics(ctx context.Context, in *RequestMetricsString, opts ...grpc.CallOption) (*emptypb.Empty, error)
	PingDataBase(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetValue(ctx context.Context, in *RequestMetricsName, opts ...grpc.CallOption) (*ResponseString, error)
	GetValueJSON(ctx context.Context, in *RequestGetMetrics, opts ...grpc.CallOption) (*ResponseMetrics, error)
	GetListMetrics(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ResponseListMetrics, error)
}

type metricCollectorClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricCollectorClient(cc grpc.ClientConnInterface) MetricCollectorClient {
	return &metricCollectorClient{cc}
}

func (c *metricCollectorClient) UpdatesAllMetricsJSON(ctx context.Context, in *RequestListMetrics, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/UpdatesAllMetricsJSON", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) UpdateOneMetricsJSON(ctx context.Context, in *RequestMetrics, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/UpdateOneMetricsJSON", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) UpdateOneMetrics(ctx context.Context, in *RequestMetricsString, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/UpdateOneMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) PingDataBase(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/PingDataBase", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) GetValue(ctx context.Context, in *RequestMetricsName, opts ...grpc.CallOption) (*ResponseString, error) {
	out := new(ResponseString)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/GetValue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) GetValueJSON(ctx context.Context, in *RequestGetMetrics, opts ...grpc.CallOption) (*ResponseMetrics, error) {
	out := new(ResponseMetrics)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/GetValueJSON", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) GetListMetrics(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ResponseListMetrics, error) {
	out := new(ResponseListMetrics)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/GetListMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricCollectorServer is the server API for MetricCollector service.
// All implementations must embed UnimplementedMetricCollectorServer
// for forward compatibility
type MetricCollectorServer interface {
	UpdatesAllMetricsJSON(context.Context, *RequestListMetrics) (*emptypb.Empty, error)
	UpdateOneMetricsJSON(context.Context, *RequestMetrics) (*emptypb.Empty, error)
	UpdateOneMetrics(context.Context, *RequestMetricsString) (*emptypb.Empty, error)
	PingDataBase(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	GetValue(context.Context, *RequestMetricsName) (*ResponseString, error)
	GetValueJSON(context.Context, *RequestGetMetrics) (*ResponseMetrics, error)
	GetListMetrics(context.Context, *emptypb.Empty) (*ResponseListMetrics, error)
	mustEmbedUnimplementedMetricCollectorServer()
}

// UnimplementedMetricCollectorServer must be embedded to have forward compatible implementations.
type UnimplementedMetricCollectorServer struct {
}

func (UnimplementedMetricCollectorServer) UpdatesAllMetricsJSON(context.Context, *RequestListMetrics) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatesAllMetricsJSON not implemented")
}
func (UnimplementedMetricCollectorServer) UpdateOneMetricsJSON(context.Context, *RequestMetrics) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOneMetricsJSON not implemented")
}
func (UnimplementedMetricCollectorServer) UpdateOneMetrics(context.Context, *RequestMetricsString) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOneMetrics not implemented")
}
func (UnimplementedMetricCollectorServer) PingDataBase(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PingDataBase not implemented")
}
func (UnimplementedMetricCollectorServer) GetValue(context.Context, *RequestMetricsName) (*ResponseString, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValue not implemented")
}
func (UnimplementedMetricCollectorServer) GetValueJSON(context.Context, *RequestGetMetrics) (*ResponseMetrics, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValueJSON not implemented")
}
func (UnimplementedMetricCollectorServer) GetListMetrics(context.Context, *emptypb.Empty) (*ResponseListMetrics, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetListMetrics not implemented")
}
func (UnimplementedMetricCollectorServer) mustEmbedUnimplementedMetricCollectorServer() {}

// UnsafeMetricCollectorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricCollectorServer will
// result in compilation errors.
type UnsafeMetricCollectorServer interface {
	mustEmbedUnimplementedMetricCollectorServer()
}

func RegisterMetricCollectorServer(s grpc.ServiceRegistrar, srv MetricCollectorServer) {
	s.RegisterService(&MetricCollector_ServiceDesc, srv)
}

func _MetricCollector_UpdatesAllMetricsJSON_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestListMetrics)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricCollectorServer).UpdatesAllMetricsJSON(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/handlers.MetricCollector/UpdatesAllMetricsJSON",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricCollectorServer).UpdatesAllMetricsJSON(ctx, req.(*RequestListMetrics))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_UpdateOneMetricsJSON_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestMetrics)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricCollectorServer).UpdateOneMetricsJSON(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/handlers.MetricCollector/UpdateOneMetricsJSON",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricCollectorServer).UpdateOneMetricsJSON(ctx, req.(*RequestMetrics))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_UpdateOneMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestMetricsString)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricCollectorServer).UpdateOneMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/handlers.MetricCollector/UpdateOneMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricCollectorServer).UpdateOneMetrics(ctx, req.(*RequestMetricsString))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_PingDataBase_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricCollectorServer).PingDataBase(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/handlers.MetricCollector/PingDataBase",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricCollectorServer).PingDataBase(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_GetValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestMetricsName)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricCollectorServer).GetValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/handlers.MetricCollector/GetValue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricCollectorServer).GetValue(ctx, req.(*RequestMetricsName))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_GetValueJSON_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestGetMetrics)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricCollectorServer).GetValueJSON(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/handlers.MetricCollector/GetValueJSON",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricCollectorServer).GetValueJSON(ctx, req.(*RequestGetMetrics))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_GetListMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricCollectorServer).GetListMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/handlers.MetricCollector/GetListMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricCollectorServer).GetListMetrics(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// MetricCollector_ServiceDesc is the grpc.ServiceDesc for MetricCollector service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetricCollector_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "handlers.MetricCollector",
	HandlerType: (*MetricCollectorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpdatesAllMetricsJSON",
			Handler:    _MetricCollector_UpdatesAllMetricsJSON_Handler,
		},
		{
			MethodName: "UpdateOneMetricsJSON",
			Handler:    _MetricCollector_UpdateOneMetricsJSON_Handler,
		},
		{
			MethodName: "UpdateOneMetrics",
			Handler:    _MetricCollector_UpdateOneMetrics_Handler,
		},
		{
			MethodName: "PingDataBase",
			Handler:    _MetricCollector_PingDataBase_Handler,
		},
		{
			MethodName: "GetValue",
			Handler:    _MetricCollector_GetValue_Handler,
		},
		{
			MethodName: "GetValueJSON",
			Handler:    _MetricCollector_GetValueJSON_Handler,
		},
		{
			MethodName: "GetListMetrics",
			Handler:    _MetricCollector_GetListMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}
