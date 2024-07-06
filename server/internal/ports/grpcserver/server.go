package grpcserver

import (
	"conference/internal/app"
	"conference/internal/ports/grpcserver/protosound"
	"conference/internal/userInfo"
	"log"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	protosound.UnimplementedSoundServiceServer
	app    app.App
	uInf   userInfo.Repository
	logger *zap.Logger
}

type ServerOption func(*Server)

func NewServer(a app.App, uInfRepo userInfo.Repository, lr *zap.Logger, addr string, opts ...ServerOption) (net.Listener, *grpc.Server, *Server) {
	grpcS := grpc.NewServer()
	server := &Server{app: a, uInf: uInfRepo, logger: lr}

	for _, opt := range opts {
		opt(server)
	}

	protosound.RegisterSoundServiceServer(grpcS, server)

	server.logger.Info("server ready to listen")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return lis, grpcS, server
}
