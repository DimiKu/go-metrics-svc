package grpcserver

import (
	"go-metric-svc/internal/config"
	"go-metric-svc/internal/orchestrator"
	"go-metric-svc/internal/proto"
	"go-metric-svc/internal/service/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
)

type GRPCServer struct {
	collectorService *server.MetricCollectorSvc
	cfg              config.ServerConfig
	server           *grpc.Server

	log *zap.SugaredLogger
}

func NewGRPCServer(
	service *server.MetricCollectorSvc,
	cfg config.ServerConfig,
	log *zap.SugaredLogger,
) *GRPCServer {
	return &GRPCServer{
		collectorService: service,
		cfg:              cfg,
		log:              log,
	}
}

func (g *GRPCServer) StartGRPCServer() error {
	lis, err := net.Listen("tcp", g.cfg.Addr)
	if err != nil {
		g.log.Fatalf("gRPC server failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	g.server = grpcServer

	grpcOrch := orchestrator.NewOrchestrator(g.collectorService, g.log)
	proto.RegisterMetricsServiceServer(grpcServer, grpcOrch)

	g.log.Infof("gRPC server listening on %s", g.cfg.Addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}

	return nil
}

func (g *GRPCServer) Shutdown() error {
	g.log.Info("start gRPC server shutting down")
	g.server.GracefulStop()
	return nil
}
