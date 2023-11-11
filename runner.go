package main

import (
	"conferenceTestHTTP/client"
	"conferenceTestHTTP/server"
	"flag"
)

var pt = flag.String("pt", "client", "Type of programm to run")

func main() {
	flag.Parse()

	if *pt == "client" {
		client.RunClient()
	} else {
		server.RunServer()
	}
}
