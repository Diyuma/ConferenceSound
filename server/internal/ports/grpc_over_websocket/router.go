package grpcoverwebsocket

import (
	"net/http"
)

func appRouter(s *server) {
	http.HandleFunc("/getsound", s.GetSoundHandler)
	http.HandleFunc("/sendsound", s.SendSoundHandler)
}
