package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// GRPCServer es un Component que expone un servidor gRPC con health check y reflection.
type GRPCServer struct {
	addr            string
	logger          *slog.Logger
	shutdownTimeout time.Duration
	server          *grpc.Server
	health          *health.Server
}

// NewGRPCServer crea el servidor con health check y reflection ya registrados.
func NewGRPCServer(addr string, logger *slog.Logger, shutdownTimeout time.Duration, opts ...grpc.ServerOption) *GRPCServer {
	server := grpc.NewServer(opts...)

	healthServer := health.NewServer()
	// Servicio "" = estado global del servidor, no de un servicio concreto.
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)

	reflection.Register(server)

	return &GRPCServer{
		addr:            addr,
		logger:          logger,
		shutdownTimeout: shutdownTimeout,
		server:          server,
		health:          healthServer,
	}
}

// Server expone el *grpc.Server para que el servicio registre sus propios servicios.
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// Name identifica el componente en los logs de app.Run.
func (s *GRPCServer) Name() string {
	return "grpc-server"
}

// Run sirve hasta que ctx se cancela y hace un apagado gradual con límite de tiempo.
func (s *GRPCServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", s.addr, err)
	}

	// Serve bloquea; va en goroutine con canal con buffer para no dejarla colgada.
	serveErr := make(chan error, 1)
	go func() { serveErr <- s.server.Serve(lis) }()

	// El primero en llegar gana: un fallo de Serve aborta sin esperar a ctx.
	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	// NOT_SERVING antes de dejar de aceptar tráfico, para que el health check no mienta.
	s.health.Shutdown()

	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	// Si GracefulStop tarda más que shutdownTimeout, Stop corta en seco.
	select {
	case <-stopped:
		return nil
	case <-time.After(s.shutdownTimeout):
		s.logger.Warn("graceful stop timed out, forcing stop", "timeout", s.shutdownTimeout)
		s.server.Stop()
		return nil
	}
}
