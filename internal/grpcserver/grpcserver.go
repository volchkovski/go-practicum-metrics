// Package grpcserver provides gRPC server initialization and management.
package grpcserver

import (
	"errors"
	"fmt"
	"net"

	"github.com/volchkovski/go-practicum-metrics/internal/grpcapi"
	"github.com/volchkovski/go-practicum-metrics/internal/grpcapi/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GRPCServer wraps the gRPC server with lifecycle management
type GRPCServer struct {
	server   *grpc.Server
	address  string
	logger   *zap.SugaredLogger
	notify   chan error
	listener net.Listener
}

// New creates a new gRPC server instance
func New(address string, service grpcapi.MetricService, logger *zap.SugaredLogger, trustedSubnet string) *GRPCServer {
	// Desugar logger for gRPC interceptors
	zapLogger := logger.Desugar()
	
	// Create gRPC server with interceptors
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcapi.RecoveryInterceptor(zapLogger),
			grpcapi.LoggingInterceptor(zapLogger),
			grpcapi.SubnetInterceptor(trustedSubnet),
		),
	)

	// Register the metrics service
	metricsServer := grpcapi.NewMetricsServer(service)
	pb.RegisterMetricsServiceServer(server, metricsServer)

	return &GRPCServer{
		server:  server,
		address: address,
		logger:  logger,
		notify:  make(chan error, 1),
	}
}

// Start begins listening and serving gRPC requests
func (s *GRPCServer) Start() {
	go func() {
		var err error
		s.listener, err = net.Listen("tcp", s.address)
		if err != nil {
			s.notify <- fmt.Errorf("failed to listen on %s: %w", s.address, err)
			return
		}

		s.logger.Info("gRPC server starting", zap.String("address", s.address))

		if err := s.server.Serve(s.listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			s.notify <- fmt.Errorf("gRPC server error: %w", err)
			return
		}

		s.notify <- nil
	}()
}

// Shutdown gracefully stops the gRPC server
func (s *GRPCServer) Shutdown() error {
	s.logger.Info("shutting down gRPC server")
	s.server.GracefulStop()
	return nil
}

// Notify returns a channel that receives server errors
func (s *GRPCServer) Notify() <-chan error {
	return s.notify
}

