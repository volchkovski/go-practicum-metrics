package grpcapi

import (
	"context"

	"github.com/volchkovski/go-practicum-metrics/internal/grpcapi/pb"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricService defines the interface for metric operations
type MetricService interface {
	PushGaugeMetric(ctx context.Context, metric *m.GaugeMetric) error
	PushCounterMetric(ctx context.Context, metric *m.CounterMetric) error
	PushMetrics(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error
	GetGaugeMetric(ctx context.Context, name string) (*m.GaugeMetric, error)
	GetCounterMetric(ctx context.Context, name string) (*m.CounterMetric, error)
	GetAllGaugeMetrics(ctx context.Context) ([]*m.GaugeMetric, error)
	GetAllCounterMetrics(ctx context.Context) ([]*m.CounterMetric, error)
	PingDB(ctx context.Context) error
}

// MetricsServer implements the gRPC MetricsService interface
type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	service MetricService
}

// NewMetricsServer creates a new gRPC metrics server
func NewMetricsServer(service MetricService) *MetricsServer {
	return &MetricsServer{
		service: service,
	}
}

// PushGauge handles storing a single gauge metric
func (s *MetricsServer) PushGauge(ctx context.Context, req *pb.PushGaugeRequest) (*pb.PushGaugeResponse, error) {
	if req.GetMetric() == nil {
		return nil, status.Error(codes.InvalidArgument, "metric is required")
	}

	metric := &m.GaugeMetric{
		Name:  req.GetMetric().GetName(),
		Value: req.GetMetric().GetValue(),
	}

	if err := s.service.PushGaugeMetric(ctx, metric); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to push gauge metric: %v", err)
	}

	return &pb.PushGaugeResponse{Success: true}, nil
}

// PushCounter handles storing a single counter metric
func (s *MetricsServer) PushCounter(ctx context.Context, req *pb.PushCounterRequest) (*pb.PushCounterResponse, error) {
	if req.GetMetric() == nil {
		return nil, status.Error(codes.InvalidArgument, "metric is required")
	}

	metric := &m.CounterMetric{
		Name:  req.GetMetric().GetName(),
		Value: req.GetMetric().GetValue(),
	}

	if err := s.service.PushCounterMetric(ctx, metric); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to push counter metric: %v", err)
	}

	return &pb.PushCounterResponse{Success: true}, nil
}

// PushMetrics handles storing multiple metrics in batch
func (s *MetricsServer) PushMetrics(ctx context.Context, req *pb.PushMetricsRequest) (*pb.PushMetricsResponse, error) {
	gauges := make([]*m.GaugeMetric, 0, len(req.GetGauges()))
	for _, g := range req.GetGauges() {
		gauges = append(gauges, &m.GaugeMetric{
			Name:  g.GetName(),
			Value: g.GetValue(),
		})
	}

	counters := make([]*m.CounterMetric, 0, len(req.GetCounters()))
	for _, c := range req.GetCounters() {
		counters = append(counters, &m.CounterMetric{
			Name:  c.GetName(),
			Value: c.GetValue(),
		})
	}

	if err := s.service.PushMetrics(ctx, gauges, counters); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to push metrics: %v", err)
	}

	return &pb.PushMetricsResponse{Success: true}, nil
}

// GetGauge retrieves a specific gauge metric by name
func (s *MetricsServer) GetGauge(ctx context.Context, req *pb.GetGaugeRequest) (*pb.GetGaugeResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "metric name is required")
	}

	metric, err := s.service.GetGaugeMetric(ctx, req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "gauge metric not found: %v", err)
	}

	return &pb.GetGaugeResponse{
		Metric: &pb.GaugeMetric{
			Name:  metric.Name,
			Value: metric.Value,
		},
	}, nil
}

// GetCounter retrieves a specific counter metric by name
func (s *MetricsServer) GetCounter(ctx context.Context, req *pb.GetCounterRequest) (*pb.GetCounterResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "metric name is required")
	}

	metric, err := s.service.GetCounterMetric(ctx, req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "counter metric not found: %v", err)
	}

	return &pb.GetCounterResponse{
		Metric: &pb.CounterMetric{
			Name:  metric.Name,
			Value: metric.Value,
		},
	}, nil
}

// GetAllMetrics retrieves all stored metrics
func (s *MetricsServer) GetAllMetrics(ctx context.Context, req *pb.GetAllMetricsRequest) (*pb.GetAllMetricsResponse, error) {
	gauges, err := s.service.GetAllGaugeMetrics(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get gauge metrics: %v", err)
	}

	counters, err := s.service.GetAllCounterMetrics(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get counter metrics: %v", err)
	}

	pbGauges := make([]*pb.GaugeMetric, 0, len(gauges))
	for _, g := range gauges {
		pbGauges = append(pbGauges, &pb.GaugeMetric{
			Name:  g.Name,
			Value: g.Value,
		})
	}

	pbCounters := make([]*pb.CounterMetric, 0, len(counters))
	for _, c := range counters {
		pbCounters = append(pbCounters, &pb.CounterMetric{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	return &pb.GetAllMetricsResponse{
		Gauges:   pbGauges,
		Counters: pbCounters,
	}, nil
}

// Ping checks the health of the service
func (s *MetricsServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	if err := s.service.PingDB(ctx); err != nil {
		return &pb.PingResponse{Healthy: false}, nil
	}
	return &pb.PingResponse{Healthy: true}, nil
}

