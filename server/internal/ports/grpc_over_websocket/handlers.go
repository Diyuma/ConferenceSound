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

/*buf := make([]byte, 1000)
for i := 0; i < 1000; i++ {
	buf[i] = byte(i % 128)
}
for i := 0; i < 1000; i++ {
	_, buf, err = conn.ReadMessage()
	if err != nil {
		log.Println("Failed to read message, error:", err)
	}
	if i == 0 {
		log.Println("got buffer", buf)
	}
	//conn.WriteMessage(websocket.BinaryMessage, buf)
	msg := &protosound.ClientInfoMessage{ConfId: 10, UserId: 25}

	str := ""
	c, d := proto.Marshal(msg)
	for _, ch := range c {
		str += fmt.Sprint(int(ch))
		str += " "
	}
	s.lg.Info("buf", zap.Any("buf", str))
	a, b := msg.Descriptor()
	s.lg.Info("buf", zap.Any("buf2", a), zap.Any("buf2", b))
	c, d = proto.Marshal(msg) //msg.Descriptor()
	s.lg.Info("buf", zap.Any("buf2", c), zap.Any("buf2", d))
	//e, f := grpc.Codec.Marshal(*msg) //msg.Descriptor()
	//s.lg.Info("buf", zap.Any("buf2", e), zap.Any("buf2", f))
	conn.WriteMessage(websocket.BinaryMessage, []byte(msg.String()))
}*/
