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

// Comprobación en compilación de que *GRPCServer sigue cumpliendo la interfaz Component.
var _ Component = (*GRPCServer)(nil)

// GRPCServer es el Component que expone el servidor gRPC de un servicio.
//
// Trae de serie dos servicios estándar que no dependen del dominio: el health
// check, que usan Kubernetes y los balanceadores para saber si mandar tráfico,
// y reflection, que permite llamar al servidor con grpcurl sin pasarle los
// .proto. Los servicios propios se registran desde fuera con Server().
type GRPCServer struct {
	addr            string
	logger          *slog.Logger
	shutdownTimeout time.Duration
	server          *grpc.Server
	health          *health.Server
}

// NewGRPCServer prepara el servidor, pero no escucha hasta que se llama a Run.
//
// Ese hueco entre construir y escuchar es deliberado: es cuando el servicio
// registra sus propios servicios gRPC, que deben estar todos antes de aceptar
// la primera petición.
func NewGRPCServer(addr string, logger *slog.Logger, shutdownTimeout time.Duration, opts ...grpc.ServerOption) *GRPCServer {
	server := grpc.NewServer(opts...)

	healthServer := health.NewServer()

	// Servicio "" = estado global del servidor, no el de un servicio concreto.
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)

	// Expone el catálogo de servicios y sus contratos para clientes como grpcurl.
	reflection.Register(server)

	return &GRPCServer{
		addr:            addr,
		logger:          logger,
		shutdownTimeout: shutdownTimeout,
		server:          server,
		health:          healthServer,
	}
}

// Server da acceso al *grpc.Server para registrar los servicios del dominio.
//
// Solo debe usarse antes de Run: registrar con el servidor ya sirviendo hace
// que gRPC entre en pánico.
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// Name identifica el componente en los logs de Run.
func (s *GRPCServer) Name() string {
	return "grpc-server"
}

// Run escucha y sirve hasta que ctx se cancela, y entonces apaga en dos tiempos.
//
// Primero GracefulStop, que deja de aceptar conexiones nuevas pero espera a que
// terminen las RPC en curso, para no cortar una petición a medias. Como esa
// espera no tiene límite y una RPC colgada dejaría el proceso sin apagarse
// nunca, compite contra shutdownTimeout: si vence, Stop corta en seco.
// Un apagado, gradual o forzado, devuelve nil; solo un fallo del propio
// servidor devuelve error.
func (s *GRPCServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", s.addr, err)
	}

	// Serve bloquea, así que va aparte; el buffer evita que la goroutine quede colgada.
	serveErr := make(chan error, 1)
	go func() { serveErr <- s.server.Serve(lis) }()

	// Gana el primero: si Serve muere solo, se devuelve el error sin esperar a ctx.
	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	// NOT_SERVING antes de empezar a cerrar: así el health check no miente durante el apagado.
	s.health.Shutdown()

	// GracefulStop también bloquea, y aquí sí hay que poder rendirse.
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil

	// Vencido el plazo, se cortan las RPC que sigan vivas: el proceso debe morir.
	case <-time.After(s.shutdownTimeout):
		s.logger.Warn("graceful stop timed out, forcing stop", "timeout", s.shutdownTimeout)
		s.server.Stop()
		return nil
	}
}
