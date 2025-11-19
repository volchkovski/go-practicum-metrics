package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/volchkovski/go-practicum-metrics/internal/grpcapi/pb"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// GRPCClient handles communication with the metrics server via gRPC
type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.MetricsServiceClient
	logger *zap.SugaredLogger
}

// NewGRPCClient creates a new gRPC client for metrics
func NewGRPCClient(serverAddr string, logger *zap.SugaredLogger) (*GRPCClient, error) {
	// Create connection with retry and timeout
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(retryInterceptor(3)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: pb.NewMetricsServiceClient(conn),
		logger: logger,
	}, nil
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// PushGauge sends a gauge metric to the server
func (c *GRPCClient) PushGauge(ctx context.Context, metric *m.GaugeMetric) error {
	req := &pb.PushGaugeRequest{
		Metric: &pb.GaugeMetric{
			Name:  metric.Name,
			Value: metric.Value,
		},
	}

	resp, err := c.client.PushGauge(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to push gauge metric: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("server reported failure for gauge metric")
	}

	return nil
}

// PushCounter sends a counter metric to the server
func (c *GRPCClient) PushCounter(ctx context.Context, metric *m.CounterMetric) error {
	req := &pb.PushCounterRequest{
		Metric: &pb.CounterMetric{
			Name:  metric.Name,
			Value: metric.Value,
		},
	}

	resp, err := c.client.PushCounter(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to push counter metric: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("server reported failure for counter metric")
	}

	return nil
}

// PushMetrics sends multiple metrics in a batch
func (c *GRPCClient) PushMetrics(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
	pbGauges := make([]*pb.GaugeMetric, 0, len(gauges))
	for _, g := range gauges {
		pbGauges = append(pbGauges, &pb.GaugeMetric{
			Name:  g.Name,
			Value: g.Value,
		})
	}

	pbCounters := make([]*pb.CounterMetric, 0, len(counters))
	for _, ct := range counters {
		pbCounters = append(pbCounters, &pb.CounterMetric{
			Name:  ct.Name,
			Value: ct.Value,
		})
	}

	req := &pb.PushMetricsRequest{
		Gauges:   pbGauges,
		Counters: pbCounters,
	}

	resp, err := c.client.PushMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to push metrics batch: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("server reported failure for metrics batch")
	}

	c.logger.Info("pushed metrics batch via gRPC",
		zap.Int("gauges", len(gauges)),
		zap.Int("counters", len(counters)),
	)

	return nil
}

// Ping checks the server health
func (c *GRPCClient) Ping(ctx context.Context) error {
	resp, err := c.client.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	if !resp.GetHealthy() {
		return fmt.Errorf("server reported unhealthy status")
	}

	return nil
}

// retryInterceptor adds retry logic to gRPC calls
func retryInterceptor(maxRetries int) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			// Check if error is retryable
			st, ok := status.FromError(err)
			if !ok {
				return err
			}

			// Retry on specific error codes
			switch st.Code() {
			case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
				if attempt < maxRetries {
					// Exponential backoff
					wait := time.Duration(attempt+1) * time.Second
					time.Sleep(wait)
					continue
				}
			default:
				return err
			}
		}
		return err
	}
}

