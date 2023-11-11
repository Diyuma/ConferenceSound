package client

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"conferenceTestHTTPWebSocket/server"

	"github.com/fasthttp/websocket"
)

var addr = flag.String("addrc", "localhost:8080", "client side http service address")
var rsz = flag.Int("reqsz", 16, "http request byte size")
var tw = flag.Int64("tw", 1e9/5000, "Amount of time we can wait response in ns") // 5k times per second resp want to get
var ars = flag.Int("ars", 10, "Amount of responses per second")

func closeConnectionGracefully(c *websocket.Conn) {
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error during close occur:", err)
	}
}

func RunClient() {
	flag.Parse()

	f, err := os.Create("logs_client")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.SetOutput(f)

	fs, err := os.Create("logs_statistic")
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	var stWr sync.Mutex

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Error during connection occur", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Error during read occur:", err)
				return
			}

			tnow := server.GetUnixTimeInNs()
			ts := binary.LittleEndian.Uint64(msg[:8])

			stWr.Lock()
			fmt.Fprintf(fs, "%d ", tnow-ts)
			stWr.Unlock()

			if tnow-ts >= uint64(*tw) {
				log.Printf("Got response in %d ns!", tnow-ts)
			}
		}
	}()

	ticker := time.NewTicker(time.Duration(1e9 / *ars) * time.Nanosecond) // second = 1e9 * ns
	defer ticker.Stop()

	for {
		select {
		case <-done:
			closeConnectionGracefully(c)
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, server.GenSomeByteDataWithTime(*rsz))
			if err != nil {
				log.Println("Error during write occur:", err)
				return
			}
		case <-interrupt:
			closeConnectionGracefully(c)
			return
		}
	}
}
