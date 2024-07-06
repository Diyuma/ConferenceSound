package grpcoverwebsocket

import (
	"conference/internal/ports/grpcserver"
	"context"
	"log"
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
	s := server{mu: &sync.Mutex{}, lg: lg, ctx: context.Background(), grpcS: grpcServer}
	var httpSrv http.Server
	httpSrv.Addr = addr
	http.HandleFunc("/getsound", s.GetSoundHandler)
	http.HandleFunc("/sendsound", s.SendSoundHandler)
	log.Println("websocket server is listenning")
	err := httpSrv.ListenAndServe()
	if err != nil {
		lg.Error("failed to listen and serve websocket conn", zap.Error(err))
	}
	return err
}
