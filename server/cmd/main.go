package main

import (
	"conference/server/internal/app"
	"conference/server/internal/ports/grpcserver"
	"conference/server/internal/sound/soundwav"
	"conference/server/internal/soundadapters/reporedis"
	"conference/server/internal/userInfo/infoRepoRedis"
	"context"
	"flag"

	"log"

	"go.uber.org/zap"
)

type ServerLoggers struct {
	ServerLogger *zap.Logger
	RepoLogger   *zap.Logger
}

// TODO: move sound interface choosing there
// TODO: add app logger (now server logger and app logger are merged)
func RunServer(addr string, sRepoAddr string, uInfRepoAddr string, slg ServerLoggers, opts ...grpcserver.ServerOption) {
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

	lis, s := grpcserver.NewServer(app.NewApp(reporedis.NewRepo(context.Background(), sRepoAddr, slg.RepoLogger), soundwav.NewEmptySound(), slg.ServerLogger), infoRepoRedis.NewRepo(context.Background(), uInfRepoAddr, slg.RepoLogger), slg.ServerLogger, addr, opts...)

	slg.ServerLogger.Info("Server is listenning", zap.String("addr", addr))
	if err := s.Serve(lis); err != nil {
		slg.ServerLogger.Fatal("Failed to serve", zap.Error(err))
	}
}

func main() {
	var serverAddr = flag.String("serveraddr", ":9090", "Server listening address")
	var sRedisAddr = flag.String("sredisaddr", ":8088", "Redis for sound listening address")
	var uInfRedisAddr = flag.String("uinfredisaddr", ":8089", "Redis for user info listening address")
	flag.Parse()

	if serverAddr == nil || sRedisAddr == nil || uInfRedisAddr == nil {
		log.Fatal("Can't parse server address or repoaddress from flags")
	}

	RunServer(*serverAddr, *sRedisAddr, *uInfRedisAddr, ServerLoggers{})
}
