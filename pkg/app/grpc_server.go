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

// Comprueba en compilación que *GRPCServer cumple Component
var _ Component = (*GRPCServer)(nil)

// GRPCServer representa el servidor gRPC de un servicio como Component
type GRPCServer struct {
	addr            string
	logger          *slog.Logger
	shutdownTimeout time.Duration
	server          *grpc.Server
	health          *health.Server
}

// NewGRPCServer crea el servidor gRPC con health check y reflection
func NewGRPCServer(addr string, logger *slog.Logger, shutdownTimeout time.Duration, opts ...grpc.ServerOption) *GRPCServer {
	server := grpc.NewServer(opts...)

	// Registra el health check global, el que consultan Kubernetes y los balanceadores
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)

	// Permite llamar al servidor con grpcurl sin pasarle los .proto
	reflection.Register(server)

	return &GRPCServer{
		addr:            addr,
		logger:          logger,
		shutdownTimeout: shutdownTimeout,
		server:          server,
		health:          healthServer,
	}
}

// Server devuelve el *grpc.Server para registrar servicios antes de llamar a Run
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// Name devuelve el nombre del componente en los logs
func (s *GRPCServer) Name() string {
	return "grpc-server"
}

// Run sirve hasta que ctx se cancela y apaga de forma gradual con un plazo máximo
func (s *GRPCServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", s.addr, err)
	}

	// Anuncia el puerto ya reservado: desde aquí las conexiones entrantes esperan en cola a Serve
	s.logger.Info("grpc server listening", "addr", s.addr)

	// Sirve en otra goroutine porque Serve bloquea; el buffer evita que se quede colgada
	serveErr := make(chan error, 1)
	go func() { serveErr <- s.server.Serve(lis) }()

	// Espera a que Serve falle o a que llegue la orden de parar
	select {
	// Devuelve el fallo de Serve sin esperar a ctx
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	// Marca NOT_SERVING antes de cerrar para que el health check no mienta durante el apagado
	s.health.Shutdown()

	// Lanza GracefulStop aparte para poder abandonarlo si se pasa del plazo
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	// Espera a que terminen las RPC en curso o a que venza el plazo
	select {
	case <-stopped:
		return nil
	// Corta en seco las RPC que sigan vivas para que el proceso pueda morir
	case <-time.After(s.shutdownTimeout):
		s.logger.Warn("graceful stop timed out, forcing stop", "timeout", s.shutdownTimeout)
		s.server.Stop()
		return nil
	}
}
