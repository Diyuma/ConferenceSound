package grpcoverwebsocket

import (
	"conference/internal/ports/grpcserver/protosound"
	"context"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type soundServiceGetSoundServerWebsocket struct {
	unImplServerStream
	conn *websocket.Conn
	lg   *zap.Logger
}

func NewGSSW(lg *zap.Logger, conn *websocket.Conn) soundServiceGetSoundServerWebsocket {
	return soundServiceGetSoundServerWebsocket{conn: conn, lg: lg.With(zap.String("app", "soundServiceGetSoundServerWebsocket"))}
}

func (gssw soundServiceGetSoundServerWebsocket) Send(msg *protosound.ChatServerMessage) error {
	buf, err := proto.Marshal(msg)
	if err != nil {
		gssw.lg.Error("failed to marshal response", zap.Error(err))
		return err
	}
	if err = gssw.conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
		gssw.lg.Error("failed to write response", zap.Error(err))
		return err
	}
	return nil
}

type unImplServerStream struct {
	lg *zap.Logger
}

func (uss unImplServerStream) Context() context.Context {
	uss.lg.Fatal("tried to get access to unimplemented function")
	return context.TODO()
}
func (uss unImplServerStream) RecvMsg(m any) error {
	uss.lg.Fatal("tried to get access to unimplemented function")
	return nil
}
func (uss unImplServerStream) SendHeader(metadata.MD) error {
	uss.lg.Fatal("tried to get access to unimplemented function")
	return nil
}
func (uss unImplServerStream) SendMsg(m any) error {
	uss.lg.Fatal("tried to get access to unimplemented function")
	return nil
}
func (uss unImplServerStream) SetHeader(metadata.MD) error {
	uss.lg.Fatal("tried to get access to unimplemented function")
	return nil
}
func (uss unImplServerStream) SetTrailer(metadata.MD) {
	uss.lg.Fatal("tried to get access to unimplemented function")
}
