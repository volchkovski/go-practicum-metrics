package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/volchkovski/go-practicum-metrics/internal/grpcapi/pb"
	"github.com/volchkovski/go-practicum-metrics/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockMetricsServiceClient implements pb.MetricsServiceClient for testing
type mockMetricsServiceClient struct {
	pushGaugeFunc   func(context.Context, *pb.PushGaugeRequest, ...grpc.CallOption) (*pb.PushGaugeResponse, error)
	pushCounterFunc func(context.Context, *pb.PushCounterRequest, ...grpc.CallOption) (*pb.PushCounterResponse, error)
	pushMetricsFunc func(context.Context, *pb.PushMetricsRequest, ...grpc.CallOption) (*pb.PushMetricsResponse, error)
	pingFunc        func(context.Context, *pb.PingRequest, ...grpc.CallOption) (*pb.PingResponse, error)
}

func (m *mockMetricsServiceClient) PushGauge(ctx context.Context, req *pb.PushGaugeRequest, opts ...grpc.CallOption) (*pb.PushGaugeResponse, error) {
	if m.pushGaugeFunc != nil {
		return m.pushGaugeFunc(ctx, req, opts...)
	}
	return &pb.PushGaugeResponse{Success: true}, nil
}

func (m *mockMetricsServiceClient) PushCounter(ctx context.Context, req *pb.PushCounterRequest, opts ...grpc.CallOption) (*pb.PushCounterResponse, error) {
	if m.pushCounterFunc != nil {
		return m.pushCounterFunc(ctx, req, opts...)
	}
	return &pb.PushCounterResponse{Success: true}, nil
}

func (m *mockMetricsServiceClient) PushMetrics(ctx context.Context, req *pb.PushMetricsRequest, opts ...grpc.CallOption) (*pb.PushMetricsResponse, error) {
	if m.pushMetricsFunc != nil {
		return m.pushMetricsFunc(ctx, req, opts...)
	}
	return &pb.PushMetricsResponse{Success: true}, nil
}

func (m *mockMetricsServiceClient) Ping(ctx context.Context, req *pb.PingRequest, opts ...grpc.CallOption) (*pb.PingResponse, error) {
	if m.pingFunc != nil {
		return m.pingFunc(ctx, req, opts...)
	}
	return &pb.PingResponse{Healthy: true}, nil
}

// Implement remaining methods from pb.MetricsServiceClient
func (m *mockMetricsServiceClient) GetGauge(ctx context.Context, req *pb.GetGaugeRequest, opts ...grpc.CallOption) (*pb.GetGaugeResponse, error) {
	return &pb.GetGaugeResponse{}, nil
}

func (m *mockMetricsServiceClient) GetCounter(ctx context.Context, req *pb.GetCounterRequest, opts ...grpc.CallOption) (*pb.GetCounterResponse, error) {
	return &pb.GetCounterResponse{}, nil
}

func (m *mockMetricsServiceClient) GetAllMetrics(ctx context.Context, req *pb.GetAllMetricsRequest, opts ...grpc.CallOption) (*pb.GetAllMetricsResponse, error) {
	return &pb.GetAllMetricsResponse{}, nil
}

func TestNewGRPCClient(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("creates client with invalid server address without blocking", func(t *testing.T) {
		// With grpc.NewClient (non-blocking), client creation succeeds even with invalid address
		// Connection errors only occur when actual RPC calls are made
		client, err := NewGRPCClient("invalid:99999", logger)
		assert.NoError(t, err, "NewClient should not fail with invalid address as it doesn't connect immediately")
		assert.NotNil(t, client, "Client should be created")
		
		// Cleanup
		if client != nil {
			_ = client.Close()
		}
	})
}

func TestGRPCClient_Close(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("closes successfully", func(t *testing.T) {
		// Create a mock client with a nil connection
		client := &GRPCClient{
			conn:   nil,
			client: &mockMetricsServiceClient{},
			logger: logger,
		}

		err := client.Close()
		assert.NoError(t, err)
	})
}

func TestGRPCClient_PushGauge(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("pushes gauge successfully", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pushGaugeFunc: func(ctx context.Context, req *pb.PushGaugeRequest, opts ...grpc.CallOption) (*pb.PushGaugeResponse, error) {
				assert.Equal(t, "test_gauge", req.Metric.Name)
				assert.Equal(t, 42.0, req.Metric.Value)
				return &pb.PushGaugeResponse{Success: true}, nil
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		metric := &models.GaugeMetric{
			Name:  "test_gauge",
			Value: 42.0,
		}

		err := client.PushGauge(context.Background(), metric)
		assert.NoError(t, err)
	})

	t.Run("handles push failure", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pushGaugeFunc: func(ctx context.Context, req *pb.PushGaugeRequest, opts ...grpc.CallOption) (*pb.PushGaugeResponse, error) {
				return &pb.PushGaugeResponse{Success: false}, nil
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		metric := &models.GaugeMetric{
			Name:  "test_gauge",
			Value: 42.0,
		}

		err := client.PushGauge(context.Background(), metric)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server reported failure")
	})

	t.Run("handles gRPC error", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pushGaugeFunc: func(ctx context.Context, req *pb.PushGaugeRequest, opts ...grpc.CallOption) (*pb.PushGaugeResponse, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		metric := &models.GaugeMetric{
			Name:  "test_gauge",
			Value: 42.0,
		}

		err := client.PushGauge(context.Background(), metric)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to push gauge metric")
	})
}

func TestGRPCClient_PushCounter(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("pushes counter successfully", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pushCounterFunc: func(ctx context.Context, req *pb.PushCounterRequest, opts ...grpc.CallOption) (*pb.PushCounterResponse, error) {
				assert.Equal(t, "test_counter", req.Metric.Name)
				assert.Equal(t, int64(100), req.Metric.Value)
				return &pb.PushCounterResponse{Success: true}, nil
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		metric := &models.CounterMetric{
			Name:  "test_counter",
			Value: 100,
		}

		err := client.PushCounter(context.Background(), metric)
		assert.NoError(t, err)
	})

	t.Run("handles push failure", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pushCounterFunc: func(ctx context.Context, req *pb.PushCounterRequest, opts ...grpc.CallOption) (*pb.PushCounterResponse, error) {
				return &pb.PushCounterResponse{Success: false}, nil
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		metric := &models.CounterMetric{
			Name:  "test_counter",
			Value: 100,
		}

		err := client.PushCounter(context.Background(), metric)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server reported failure")
	})
}

func TestGRPCClient_PushMetrics(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("pushes metrics batch successfully", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pushMetricsFunc: func(ctx context.Context, req *pb.PushMetricsRequest, opts ...grpc.CallOption) (*pb.PushMetricsResponse, error) {
				assert.Len(t, req.Gauges, 2)
				assert.Len(t, req.Counters, 1)
				return &pb.PushMetricsResponse{Success: true}, nil
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		gauges := []*models.GaugeMetric{
			{Name: "gauge1", Value: 1.0},
			{Name: "gauge2", Value: 2.0},
		}
		counters := []*models.CounterMetric{
			{Name: "counter1", Value: 10},
		}

		err := client.PushMetrics(context.Background(), gauges, counters)
		assert.NoError(t, err)
	})

	t.Run("handles push failure", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pushMetricsFunc: func(ctx context.Context, req *pb.PushMetricsRequest, opts ...grpc.CallOption) (*pb.PushMetricsResponse, error) {
				return &pb.PushMetricsResponse{Success: false}, nil
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		err := client.PushMetrics(context.Background(), nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server reported failure")
	})
}

func TestGRPCClient_Ping(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("pings successfully", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pingFunc: func(ctx context.Context, req *pb.PingRequest, opts ...grpc.CallOption) (*pb.PingResponse, error) {
				return &pb.PingResponse{Healthy: true}, nil
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		err := client.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("handles unhealthy server", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pingFunc: func(ctx context.Context, req *pb.PingRequest, opts ...grpc.CallOption) (*pb.PingResponse, error) {
				return &pb.PingResponse{Healthy: false}, nil
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		err := client.Ping(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unhealthy status")
	})

	t.Run("handles ping error", func(t *testing.T) {
		mockClient := &mockMetricsServiceClient{
			pingFunc: func(ctx context.Context, req *pb.PingRequest, opts ...grpc.CallOption) (*pb.PingResponse, error) {
				return nil, status.Error(codes.Unavailable, "server unavailable")
			},
		}

		client := &GRPCClient{
			client: mockClient,
			logger: logger,
		}

		err := client.Ping(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping failed")
	})
}

func TestRetryInterceptor(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		interceptor := retryInterceptor(3)
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		}

		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		assert.NoError(t, err)
	})

	t.Run("retries on unavailable error", func(t *testing.T) {
		attempts := 0
		interceptor := retryInterceptor(2)
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			attempts++
			if attempts < 2 {
				return status.Error(codes.Unavailable, "unavailable")
			}
			return nil
		}

		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("fails after max retries", func(t *testing.T) {
		attempts := 0
		interceptor := retryInterceptor(2)
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			attempts++
			return status.Error(codes.Unavailable, "unavailable")
		}

		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		require.Error(t, err)
		assert.Equal(t, 3, attempts) // Initial + 2 retries
		assert.Equal(t, codes.Unavailable, status.Code(err))
	})

	t.Run("does not retry on non-retryable error", func(t *testing.T) {
		attempts := 0
		interceptor := retryInterceptor(2)
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			attempts++
			return status.Error(codes.InvalidArgument, "invalid argument")
		}

		err := interceptor(context.Background(), "/test", nil, nil, nil, invoker)
		require.Error(t, err)
		assert.Equal(t, 1, attempts) // No retries for InvalidArgument
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

