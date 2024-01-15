package grpcserver

import (
	"homework/server/internal/app"
	"homework/server/internal/ports/grpcserver/proto"
	"log"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedSoundServiceServer
	app    app.App
	logger *zap.Logger
}

type ServerOption func(*server)

func NewServer(a app.App, lr *zap.Logger, addr string, opts ...ServerOption) (net.Listener, *grpc.Server) {
	grpcS := grpc.NewServer()
	server := &server{app: a, logger: lr}

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
