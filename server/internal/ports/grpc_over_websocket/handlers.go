package grpcoverwebsocket

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *server) GetSoundHandler(w http.ResponseWriter, r *http.Request) {
	s.lg.Debug("Got websocket GetSound request")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.lg.Error("failed to upgrade connection", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	if err := s.getSoundWorker(conn); errors.Is(err, errors.Join(ErrorIncorrectMessageData, ErrorIncorrectMessageType)) {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) SendSoundHandler(w http.ResponseWriter, r *http.Request) {
	s.lg.Debug("Got websocket SendSound request")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.lg.Error("failed to upgrade connection", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	if err := s.sendSoundWorker(conn); errors.Is(err, errors.Join(ErrorIncorrectMessageData, ErrorIncorrectMessageType)) {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
