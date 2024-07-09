package grpcoverwebsocket

import (
	"conference/internal/ports/grpcserver"
	"context"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

type server struct {
	mu    *sync.Mutex
	lg    *zap.Logger
	ctx   context.Context
	grpcS *grpcserver.Server
}

func RunServerWebSocket(addr string, grpcServer *grpcserver.Server, lg *zap.Logger) error {
	s := server{mu: &sync.Mutex{}, lg: lg.With(zap.String("app", "websocketserver")), ctx: context.Background(), grpcS: grpcServer}
	var httpSrv http.Server
	httpSrv.Addr = addr
	appRouter(&s)
	lg.Info("websocket server is listenning", zap.String("addr", addr))
	err := httpSrv.ListenAndServe()
	if err != nil {
		lg.Error("failed to listen and serve websocket conn", zap.Error(err))
	}
	return err
}
