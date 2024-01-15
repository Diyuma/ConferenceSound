package main

import (
	"context"
	"flag"
	"homework/server/internal/app"
	"homework/server/internal/ports/grpcserver"
	"homework/server/internal/sound/soundwav"
	"homework/server/internal/soundadapters/reporedis"

	"log"

	"go.uber.org/zap"
)

type ServerLoggers struct {
	ServerLogger *zap.Logger
	RepoLogger   *zap.Logger
}

func RunServer(addr string, repoaddr string, slg ServerLoggers, opts ...grpcserver.ServerOption) {
	var err error

	if slg.ServerLogger == nil {
		slg.ServerLogger, err = zap.NewProduction(zap.WithCaller(true))
		if err != nil {
			log.Fatalf("Failed to init logger: %v", err)
		}
		defer slg.ServerLogger.Sync()
	}

	if slg.RepoLogger == nil {
		slg.RepoLogger, err = zap.NewProduction(zap.WithCaller(true))
		if err != nil {
			log.Fatalf("Failed to init logger: %v", err)
		}
		defer slg.RepoLogger.Sync()
	}

	lis, s := grpcserver.NewServer(app.NewApp(reporedis.NewRepo(context.Background(), repoaddr, slg.RepoLogger), soundwav.NewEmptySound()), slg.ServerLogger, addr, opts...)

	slg.ServerLogger.Info("Server is listenning", zap.String("addr", addr))
	if err := s.Serve(lis); err != nil {
		slg.ServerLogger.Fatal("Failed to serve", zap.Error(err))
	}
}

func main() {
	var serverAddr = flag.String("serveraddr", ":8081", "Server listening address")
	var redisAddr = flag.String("serveraddr", ":8088", "Redis listening address")
	flag.Parse()

	if serverAddr == nil || redisAddr == nil {
		log.Fatal("Can't parse server address or repoaddress from flags")
	}

	RunServer(*serverAddr, *redisAddr, ServerLoggers{})
}
