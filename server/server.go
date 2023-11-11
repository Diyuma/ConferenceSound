package server

import (
	"encoding/binary"
	"flag"
	"log"
	"os"
	"time"

	"github.com/valyala/fasthttp"
)

var addr = flag.String("addrs", "localhost:8080", "server side http service address")
var bc = flag.Int("bc", 0, "Business complexity in ns")     // not litterally business complexity, but some prework and run routine
var rsz = flag.Int("respsz", 10, "http response byte size") // need >= 8 to use Time marks (GenSomeByteDataWithTime)

func GenSomeByteData(sz int) []byte {
	res := make([]byte, sz)
	for i := 0; i < sz; i++ {
		res[i] = byte(i % 256)
	}

	return res
}

func GetUnixTimeInNs() uint64 {
	return uint64(time.Now().UnixNano())
}

func GenSomeByteDataWithTime(sz int) []byte {
	res := make([]byte, sz)
	binary.LittleEndian.PutUint64(res, GetUnixTimeInNs())
	res[8] = 0

	for i := 9; i < sz; i++ {
		res[i] = byte(i % 256)
	}

	return res
}

func handler(ctx *fasthttp.RequestCtx) {
	msg := ctx.Request.Body()

	time.Sleep(time.Duration(*bc) * time.Nanosecond)
	var res []byte
	if len(msg) >= 9 && msg[8] == 0 {
		res = append(msg[:8], GenSomeByteData(*rsz-8)...)
	} else {
		res = GenSomeByteData(*rsz)
	}

	ctx.SetBody(res)
	return
}

func RunServer() {
	flag.Parse()

	f, err := os.Create("logs_server")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.SetOutput(f)

	s := &fasthttp.Server{Handler: handler}

	if err := s.ListenAndServe(*addr); err != nil {
		log.Fatalln(err)
	}
}
