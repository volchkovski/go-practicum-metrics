package grpcapi

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestLoggingInterceptor(t *testing.T) {
	logger := zap.NewNop()

	t.Run("logs successful request", func(t *testing.T) {
		interceptor := LoggingInterceptor(logger)
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "response", nil
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.NoError(t, err)
		assert.Equal(t, "response", resp)
	})

	t.Run("logs failed request", func(t *testing.T) {
		interceptor := LoggingInterceptor(logger)
		
		expectedErr := errors.New("test error")
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, expectedErr
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("logs request with gRPC status error", func(t *testing.T) {
		interceptor := LoggingInterceptor(logger)
		
		expectedErr := status.Error(codes.InvalidArgument, "invalid argument")
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, expectedErr
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestSubnetInterceptor(t *testing.T) {
	t.Run("allows all requests when trusted subnet is empty", func(t *testing.T) {
		interceptor := SubnetInterceptor("")
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "response", nil
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.NoError(t, err)
		assert.Equal(t, "response", resp)
	})

	t.Run("returns error when metadata is missing", func(t *testing.T) {
		interceptor := SubnetInterceptor("192.168.1.0/24")
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "response", nil
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Contains(t, st.Message(), "missing metadata")
	})

	t.Run("returns error when X-Real-IP header is missing", func(t *testing.T) {
		interceptor := SubnetInterceptor("192.168.1.0/24")
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "response", nil
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		// Create context with empty metadata
		md := metadata.New(map[string]string{})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		resp, err := interceptor(ctx, "request", info, handler)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
		assert.Contains(t, st.Message(), "missing X-Real-IP header")
	})

	t.Run("allows request with X-Real-IP header", func(t *testing.T) {
		interceptor := SubnetInterceptor("192.168.1.0/24")
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "response", nil
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		// Create context with X-Real-IP header
		md := metadata.New(map[string]string{
			"x-real-ip": "192.168.1.100",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		resp, err := interceptor(ctx, "request", info, handler)
		
		// Currently the implementation allows all requests (TODO in code)
		assert.NoError(t, err)
		assert.Equal(t, "response", resp)
	})
}

func TestRecoveryInterceptor(t *testing.T) {
	logger := zap.NewNop()

	t.Run("recovers from panic", func(t *testing.T) {
		interceptor := RecoveryInterceptor(logger)
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("test panic")
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Contains(t, st.Message(), "internal server error")
	})

	t.Run("does not affect normal execution", func(t *testing.T) {
		interceptor := RecoveryInterceptor(logger)
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "response", nil
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.NoError(t, err)
		assert.Equal(t, "response", resp)
	})

	t.Run("recovers from panic with nil value", func(t *testing.T) {
		interceptor := RecoveryInterceptor(logger)
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			panic(nil)
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("recovers from panic with error", func(t *testing.T) {
		interceptor := RecoveryInterceptor(logger)
		
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			panic(errors.New("panic error"))
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("allows handler errors to pass through", func(t *testing.T) {
		interceptor := RecoveryInterceptor(logger)
		
		expectedErr := status.Error(codes.NotFound, "not found")
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, expectedErr
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
	})
}

