package client

import (
	"bytes"
	"conferenceTestHTTP/server"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var addr = flag.String("addrc", "http://127.0.0.1:8080", "client side http service address")
var rsz = flag.Int("reqsz", 16, "http request byte size")
var tw = flag.Int64("tw", 1e9/5000, "Amount of time we can wait response in ns") // 5k times per second resp want to get
var ars = flag.Int("ars", 10, "Amount of responses per second")

func RunClient() {
	flag.Parse()
	var stWr sync.Mutex

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

	ticker := time.NewTicker(time.Duration(1e9 / *ars) * time.Nanosecond) // second = 1e9 * ns
	defer ticker.Stop()

	client := http.Client{}

	for {
		select {
		case <-ticker.C:
			request, err := http.NewRequest(http.MethodGet, *addr, bytes.NewReader(server.GenSomeByteDataWithTime(*rsz)))
			if err != nil {
				log.Println("Error during req creation occur:", err)
				return
			}

			resp, err := client.Do(request)
			if err != nil {
				log.Println("Error during req do occur:", err)
				return
			}

			msg, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error during req read occur:", err)
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
	}
}
