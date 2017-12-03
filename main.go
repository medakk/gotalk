package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/medakk/audiostream/client"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/ws/", client.ServeClient)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
