package main

import (
	"io"
	"log"
	"net/http"
)

func loadFile(url string) string {
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal("Failed to load html to send:", err)
	}

	f := make([]byte, 0)
	for {
		body := make([]byte, 1000)
		n, err := resp.Body.Read(body)
		if err != nil && err != io.EOF {
			log.Fatal("Failed to parse loaded html to send:", err)
		}
		f = append(f, body[:n]...)
		if err == io.EOF {
			break
		}
	}

	return string(f)
}

func runServer(htmlInd []byte, jsMain []byte) {
	http.HandleFunc("/conference", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(htmlInd); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to write response data: ", err)
		}
	})

	http.HandleFunc("/front_files/main.js", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(jsMain); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to write response data: ", err)
		}
	})

	log.Fatal(http.ListenAndServe(":8086", nil))
}

// TODO - not load from github?
func main() {
	htmlInd := loadFile("https://raw.github.com/Diyuma/ConferenceClient/main/mainBranch/index.html")
	jsMain := loadFile("https://raw.github.com/Diyuma/ConferenceClient/main/mainBranch/dist/main.js")

	runServer([]byte(htmlInd), []byte(jsMain))
}
