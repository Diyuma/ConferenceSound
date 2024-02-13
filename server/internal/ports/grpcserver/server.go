package grpcserver

import (
	"conference/server/internal/app"
	"conference/server/internal/ports/grpcserver/proto"
	"conference/server/internal/userInfo"
	"log"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedSoundServiceServer
	app    app.App
	uInf   userInfo.Repository
	logger *zap.Logger
}

type ServerOption func(*server)

func NewServer(a app.App, uInfRepo userInfo.Repository, lr *zap.Logger, addr string, opts ...ServerOption) (net.Listener, *grpc.Server) {
	grpcS := grpc.NewServer()
	server := &server{app: a, uInf: uInfRepo, logger: lr}

	for _, opt := range opts {
		opt(server)
	}

	proto.RegisterSoundServiceServer(grpcS, server)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return lis, grpcS
}
