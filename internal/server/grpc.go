package server

import (
	"context"
	"errors"
	"net"

	"github.com/andrearcaina/hyperion/internal/logger"
	grpctransport "github.com/andrearcaina/hyperion/internal/transport/grpc"
	hyperionv1 "github.com/andrearcaina/hyperion/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	address string
	server  *grpc.Server
	logger  *logger.Logger
}

func NewGRPCServer(address string, logger *logger.Logger, handler *grpctransport.Handler) *GRPCServer {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			logger.NewGRPCLoggingInterceptor(nil),
		),
	)
	hyperionv1.RegisterHyperionServer(server, handler)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	reflection.Register(server)

	return &GRPCServer{address: address, server: server, logger: logger}
}

func (s *GRPCServer) Run() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.logger.Info(context.Background(), "starting gRPC server", "address", s.address)
	if err := s.server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}

	return nil
}

func (s *GRPCServer) Close(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}
