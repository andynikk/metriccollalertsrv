// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: server.proto

package handlers

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

// MetricCollectorClient is the client API for MetricCollector service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricCollectorClient interface {
	UpdatesAllMetricsJSON(ctx context.Context, in *UpdatesRequest, opts ...grpc.CallOption) (*TextErrResponse, error)
	UpdateOneMetricsJSON(ctx context.Context, in *UpdateStrRequest, opts ...grpc.CallOption) (*TextErrResponse, error)
	UpdateOneMetrics(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*TextErrResponse, error)
	PingDataBases(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*TextErrResponse, error)
	GetValueJSON(ctx context.Context, in *UpdatesRequest, opts ...grpc.CallOption) (*FullResponse, error)
	GetValue(ctx context.Context, in *UpdatesRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	GetListMetrics(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*StatusResponse, error)
}

type metricCollectorClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricCollectorClient(cc grpc.ClientConnInterface) MetricCollectorClient {
	return &metricCollectorClient{cc}
}

func (c *metricCollectorClient) UpdatesAllMetricsJSON(ctx context.Context, in *UpdatesRequest, opts ...grpc.CallOption) (*TextErrResponse, error) {
	out := new(TextErrResponse)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/UpdatesAllMetricsJSON", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) UpdateOneMetricsJSON(ctx context.Context, in *UpdateStrRequest, opts ...grpc.CallOption) (*TextErrResponse, error) {
	out := new(TextErrResponse)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/UpdateOneMetricsJSON", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) UpdateOneMetrics(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*TextErrResponse, error) {
	out := new(TextErrResponse)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/UpdateOneMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) PingDataBases(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*TextErrResponse, error) {
	out := new(TextErrResponse)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/PingDataBases", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) GetValueJSON(ctx context.Context, in *UpdatesRequest, opts ...grpc.CallOption) (*FullResponse, error) {
	out := new(FullResponse)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/GetValueJSON", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) GetValue(ctx context.Context, in *UpdatesRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, "/handlers.MetricCollector/GetValue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricCollectorClient) GetListMetrics(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
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
	UpdatesAllMetricsJSON(context.Context, *UpdatesRequest) (*TextErrResponse, error)
	UpdateOneMetricsJSON(context.Context, *UpdateStrRequest) (*TextErrResponse, error)
	UpdateOneMetrics(context.Context, *UpdateRequest) (*TextErrResponse, error)
	PingDataBases(context.Context, *EmptyRequest) (*TextErrResponse, error)
	GetValueJSON(context.Context, *UpdatesRequest) (*FullResponse, error)
	GetValue(context.Context, *UpdatesRequest) (*StatusResponse, error)
	GetListMetrics(context.Context, *EmptyRequest) (*StatusResponse, error)
	mustEmbedUnimplementedMetricCollectorServer()
}

// UnimplementedMetricCollectorServer must be embedded to have forward compatible implementations.
type UnimplementedMetricCollectorServer struct {
}

func (UnimplementedMetricCollectorServer) UpdatesAllMetricsJSON(context.Context, *UpdatesRequest) (*TextErrResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatesAllMetricsJSON not implemented")
}
func (UnimplementedMetricCollectorServer) UpdateOneMetricsJSON(context.Context, *UpdateStrRequest) (*TextErrResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOneMetricsJSON not implemented")
}
func (UnimplementedMetricCollectorServer) UpdateOneMetrics(context.Context, *UpdateRequest) (*TextErrResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOneMetrics not implemented")
}
func (UnimplementedMetricCollectorServer) PingDataBases(context.Context, *EmptyRequest) (*TextErrResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PingDataBases not implemented")
}
func (UnimplementedMetricCollectorServer) GetValueJSON(context.Context, *UpdatesRequest) (*FullResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValueJSON not implemented")
}
func (UnimplementedMetricCollectorServer) GetValue(context.Context, *UpdatesRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValue not implemented")
}
func (UnimplementedMetricCollectorServer) GetListMetrics(context.Context, *EmptyRequest) (*StatusResponse, error) {
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
	in := new(UpdatesRequest)
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
		return srv.(MetricCollectorServer).UpdatesAllMetricsJSON(ctx, req.(*UpdatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_UpdateOneMetricsJSON_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateStrRequest)
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
		return srv.(MetricCollectorServer).UpdateOneMetricsJSON(ctx, req.(*UpdateStrRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_UpdateOneMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
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
		return srv.(MetricCollectorServer).UpdateOneMetrics(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_PingDataBases_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricCollectorServer).PingDataBases(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/handlers.MetricCollector/PingDataBases",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricCollectorServer).PingDataBases(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_GetValueJSON_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatesRequest)
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
		return srv.(MetricCollectorServer).GetValueJSON(ctx, req.(*UpdatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_GetValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatesRequest)
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
		return srv.(MetricCollectorServer).GetValue(ctx, req.(*UpdatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricCollector_GetListMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
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
		return srv.(MetricCollectorServer).GetListMetrics(ctx, req.(*EmptyRequest))
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
			MethodName: "PingDataBases",
			Handler:    _MetricCollector_PingDataBases_Handler,
		},
		{
			MethodName: "GetValueJSON",
			Handler:    _MetricCollector_GetValueJSON_Handler,
		},
		{
			MethodName: "GetValue",
			Handler:    _MetricCollector_GetValue_Handler,
		},
		{
			MethodName: "GetListMetrics",
			Handler:    _MetricCollector_GetListMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}