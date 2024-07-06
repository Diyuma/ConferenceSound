package grpcoverwebsocket

import (
	"conference/internal/ports/grpcserver/protosound"
	"context"
	"errors"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

var ErrorIncorrectMessageType error = errors.New("incorrect message type - binary message expected")
var ErrorIncorrectMessageData error = errors.New("incorrect message data")

type soundService_GetSoundServerWebsocket struct {
	unImplServerStream
	conn *websocket.Conn
	lg   *zap.Logger
}

func New(lg *zap.Logger, conn *websocket.Conn) soundService_GetSoundServerWebsocket {
	return soundService_GetSoundServerWebsocket{conn: conn, lg: lg.With(zap.String("app", "soundService_GetSoundServerWebsocket"))}
}

func (gssw soundService_GetSoundServerWebsocket) Send(msg *protosound.ChatServerMessage) error {
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

func (s *server) getSoundWorker(conn *websocket.Conn) error {
	mt, buf, err := conn.ReadMessage()
	if err != nil {
		s.lg.Error("failed to read message", zap.Error(err))
		return ErrorIncorrectMessageData
	}
	if mt != websocket.BinaryMessage {
		s.lg.Warn("incorrect message type", zap.Int("msg type", mt))
		return ErrorIncorrectMessageType
	}

	var clientInfMsg protosound.ClientInfoMessage
	if err := proto.Unmarshal(buf, &clientInfMsg); err != nil {
		s.lg.Error("failed to unmarshall msg", zap.Error(err))
		return err
	}
	if err := s.grpcS.GetSound(&clientInfMsg, soundService_GetSoundServerWebsocket{}); err != nil {
		s.lg.Error("failed to get sound", zap.Error(err))
		return err
	}
	return nil
}

func (s *server) sendSoundWorker(conn *websocket.Conn) error {
	for {
		mt, buf, err := conn.ReadMessage()
		if err != nil {
			s.lg.Error("failed to read message", zap.Error(err))
			return ErrorIncorrectMessageData
		}
		if mt != websocket.BinaryMessage {
			s.lg.Warn("incorrect message type", zap.Int("msg type", mt))
			return ErrorIncorrectMessageType
		}

		var chatClientMsg protosound.ChatClientMessage
		if err := proto.Unmarshal(buf, &chatClientMsg); err != nil {
			s.lg.Error("failed to unmarshall msg", zap.Error(err))
			return err
		}
		msg, err := s.grpcS.SendSound(s.ctx, &chatClientMsg)
		if err != nil {
			s.lg.Error("failed to send sound", zap.Error(err))
			return err
		}
		buf, err = proto.Marshal(msg)
		if err != nil {
			s.lg.Error("failed to marshal response", zap.Error(err))
		}
		if err = conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
			s.lg.Error("failed to write response", zap.Error(err))
			return err
		}
	}
}
