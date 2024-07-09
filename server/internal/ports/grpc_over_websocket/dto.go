package grpcoverwebsocket

import (
	"conference/internal/ports/grpcserver/protosound"
	"errors"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var ErrorIncorrectMessageType error = errors.New("incorrect message type - binary message expected")
var ErrorIncorrectMessageData error = errors.New("incorrect message data")

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
	if err := s.grpcS.GetSound(&clientInfMsg, NewGSSW(s.lg, conn)); err != nil {
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
