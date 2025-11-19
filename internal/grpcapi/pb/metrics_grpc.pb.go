// gRPC service stubs for MetricsService

package pb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MetricsService_PushGauge_FullMethodName     = "/metrics.v1.MetricsService/PushGauge"
	MetricsService_PushCounter_FullMethodName   = "/metrics.v1.MetricsService/PushCounter"
	MetricsService_PushMetrics_FullMethodName   = "/metrics.v1.MetricsService/PushMetrics"
	MetricsService_GetGauge_FullMethodName      = "/metrics.v1.MetricsService/GetGauge"
	MetricsService_GetCounter_FullMethodName    = "/metrics.v1.MetricsService/GetCounter"
	MetricsService_GetAllMetrics_FullMethodName = "/metrics.v1.MetricsService/GetAllMetrics"
	MetricsService_Ping_FullMethodName          = "/metrics.v1.MetricsService/Ping"
)

// MetricsServiceClient is the client API for MetricsService service.
type MetricsServiceClient interface {
	PushGauge(ctx context.Context, in *PushGaugeRequest, opts ...grpc.CallOption) (*PushGaugeResponse, error)
	PushCounter(ctx context.Context, in *PushCounterRequest, opts ...grpc.CallOption) (*PushCounterResponse, error)
	PushMetrics(ctx context.Context, in *PushMetricsRequest, opts ...grpc.CallOption) (*PushMetricsResponse, error)
	GetGauge(ctx context.Context, in *GetGaugeRequest, opts ...grpc.CallOption) (*GetGaugeResponse, error)
	GetCounter(ctx context.Context, in *GetCounterRequest, opts ...grpc.CallOption) (*GetCounterResponse, error)
	GetAllMetrics(ctx context.Context, in *GetAllMetricsRequest, opts ...grpc.CallOption) (*GetAllMetricsResponse, error)
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
}

type metricsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricsServiceClient(cc grpc.ClientConnInterface) MetricsServiceClient {
	return &metricsServiceClient{cc}
}

func (c *metricsServiceClient) PushGauge(ctx context.Context, in *PushGaugeRequest, opts ...grpc.CallOption) (*PushGaugeResponse, error) {
	out := new(PushGaugeResponse)
	err := c.cc.Invoke(ctx, MetricsService_PushGauge_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsServiceClient) PushCounter(ctx context.Context, in *PushCounterRequest, opts ...grpc.CallOption) (*PushCounterResponse, error) {
	out := new(PushCounterResponse)
	err := c.cc.Invoke(ctx, MetricsService_PushCounter_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsServiceClient) PushMetrics(ctx context.Context, in *PushMetricsRequest, opts ...grpc.CallOption) (*PushMetricsResponse, error) {
	out := new(PushMetricsResponse)
	err := c.cc.Invoke(ctx, MetricsService_PushMetrics_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsServiceClient) GetGauge(ctx context.Context, in *GetGaugeRequest, opts ...grpc.CallOption) (*GetGaugeResponse, error) {
	out := new(GetGaugeResponse)
	err := c.cc.Invoke(ctx, MetricsService_GetGauge_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsServiceClient) GetCounter(ctx context.Context, in *GetCounterRequest, opts ...grpc.CallOption) (*GetCounterResponse, error) {
	out := new(GetCounterResponse)
	err := c.cc.Invoke(ctx, MetricsService_GetCounter_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsServiceClient) GetAllMetrics(ctx context.Context, in *GetAllMetricsRequest, opts ...grpc.CallOption) (*GetAllMetricsResponse, error) {
	out := new(GetAllMetricsResponse)
	err := c.cc.Invoke(ctx, MetricsService_GetAllMetrics_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, MetricsService_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricsServiceServer is the server API for MetricsService service.
type MetricsServiceServer interface {
	PushGauge(context.Context, *PushGaugeRequest) (*PushGaugeResponse, error)
	PushCounter(context.Context, *PushCounterRequest) (*PushCounterResponse, error)
	PushMetrics(context.Context, *PushMetricsRequest) (*PushMetricsResponse, error)
	GetGauge(context.Context, *GetGaugeRequest) (*GetGaugeResponse, error)
	GetCounter(context.Context, *GetCounterRequest) (*GetCounterResponse, error)
	GetAllMetrics(context.Context, *GetAllMetricsRequest) (*GetAllMetricsResponse, error)
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	mustEmbedUnimplementedMetricsServiceServer()
}

// UnimplementedMetricsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMetricsServiceServer struct{}

func (UnimplementedMetricsServiceServer) PushGauge(context.Context, *PushGaugeRequest) (*PushGaugeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushGauge not implemented")
}
func (UnimplementedMetricsServiceServer) PushCounter(context.Context, *PushCounterRequest) (*PushCounterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushCounter not implemented")
}
func (UnimplementedMetricsServiceServer) PushMetrics(context.Context, *PushMetricsRequest) (*PushMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushMetrics not implemented")
}
func (UnimplementedMetricsServiceServer) GetGauge(context.Context, *GetGaugeRequest) (*GetGaugeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGauge not implemented")
}
func (UnimplementedMetricsServiceServer) GetCounter(context.Context, *GetCounterRequest) (*GetCounterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCounter not implemented")
}
func (UnimplementedMetricsServiceServer) GetAllMetrics(context.Context, *GetAllMetricsRequest) (*GetAllMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllMetrics not implemented")
}
func (UnimplementedMetricsServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedMetricsServiceServer) mustEmbedUnimplementedMetricsServiceServer() {}

// RegisterMetricsServiceServer registers the service with gRPC server
func RegisterMetricsServiceServer(s grpc.ServiceRegistrar, srv MetricsServiceServer) {
	s.RegisterService(&MetricsService_ServiceDesc, srv)
}

func _MetricsService_PushGauge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushGaugeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).PushGauge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_PushGauge_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).PushGauge(ctx, req.(*PushGaugeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricsService_PushCounter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushCounterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).PushCounter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_PushCounter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).PushCounter(ctx, req.(*PushCounterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricsService_PushMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).PushMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_PushMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).PushMetrics(ctx, req.(*PushMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricsService_GetGauge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGaugeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).GetGauge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_GetGauge_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).GetGauge(ctx, req.(*GetGaugeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricsService_GetCounter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCounterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).GetCounter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_GetCounter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).GetCounter(ctx, req.(*GetCounterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricsService_GetAllMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).GetAllMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_GetAllMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).GetAllMetrics(ctx, req.(*GetAllMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricsService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricsService_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MetricsService_ServiceDesc is the grpc.ServiceDesc for MetricsService service.
var MetricsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "metrics.v1.MetricsService",
	HandlerType: (*MetricsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PushGauge",
			Handler:    _MetricsService_PushGauge_Handler,
		},
		{
			MethodName: "PushCounter",
			Handler:    _MetricsService_PushCounter_Handler,
		},
		{
			MethodName: "PushMetrics",
			Handler:    _MetricsService_PushMetrics_Handler,
		},
		{
			MethodName: "GetGauge",
			Handler:    _MetricsService_GetGauge_Handler,
		},
		{
			MethodName: "GetCounter",
			Handler:    _MetricsService_GetCounter_Handler,
		},
		{
			MethodName: "GetAllMetrics",
			Handler:    _MetricsService_GetAllMetrics_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _MetricsService_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/metrics/v1/metrics.proto",
}

